// | KB @CerberRus00 - Nexus Invest Team
// api/admin.go — админские маршруты: выдача/отзыв API-ключей, имена и роли кошельков.

package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"GND/core"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

const adminKeyPrefix = "gnd_"
const adminKeyBytes = 32

// RequireAdmin проверяет заголовок X-Admin-Token. При неверном/отсутствующем возвращает 401 и false.
func (s *Server) RequireAdmin(c *gin.Context) bool {
	token := c.GetHeader("X-Admin-Token")
	if !ValidateAdminToken(token) {
		c.JSON(http.StatusUnauthorized, APIResponse{
			Success: false,
			Error:   "Неверный или отсутствующий X-Admin-Token",
			Code:    http.StatusUnauthorized,
		})
		return false
	}
	return true
}

// AdminCreateKey создаёт новый API-ключ. Ключ возвращается один раз в ответе.
// POST /api/v1/admin/keys
func (s *Server) AdminCreateKey(c *gin.Context) {
	if !s.RequireAdmin(c) {
		return
	}
	var req struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
		ExpiresAt   string   `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Неверный JSON", Code: http.StatusBadRequest})
		return
	}
	if req.Name == "" {
		req.Name = "API Key"
	}
	raw := make([]byte, adminKeyBytes)
	if _, err := rand.Read(raw); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: "Ошибка генерации ключа", Code: http.StatusInternalServerError})
		return
	}
	keySecret := hex.EncodeToString(raw)
	plainKey := adminKeyPrefix + keySecret
	keyHash := HashKey(plainKey)
	keyPrefix := KeyPrefix(plainKey)

	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Неверный формат expires_at (ожидается RFC3339)", Code: http.StatusBadRequest})
			return
		}
		expiresAt = &t
	}

	perms := req.Permissions
	if perms == nil {
		perms = []string{}
	}
	permsJSON, _ := json.Marshal(perms)

	ctx := c.Request.Context()
	now := core.BlockchainNow()
	var id int
	err := s.db.QueryRow(ctx, `
		INSERT INTO public.api_keys (name, key_prefix, key_hash, permissions, expires_at, disabled, created_at)
		VALUES ($1, $2, $3, $4::jsonb, $5, false, $6)
		RETURNING id`,
		req.Name, keyPrefix, keyHash, permsJSON, expiresAt, now).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: "Ошибка сохранения ключа: " + err.Error(), Code: http.StatusInternalServerError})
		return
	}
	if s.core != nil && s.core.Pool != nil && s.core.Genesis != nil {
		_ = core.RecordAdminTransaction(ctx, s.core.Pool, s.core.Genesis.ID, "api_key_create", "GND_SYSTEM", "GND_SYSTEM", req.Name)
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: gin.H{
			"id":          id,
			"key":         plainKey,
			"name":        req.Name,
			"key_prefix":  keyPrefix,
			"permissions": perms,
			"expires_at":  req.ExpiresAt,
			"created_at":  now.Format(time.RFC3339),
		},
	})
}

// AdminListKeys возвращает список ключей без самого ключа.
// GET /api/v1/admin/keys
func (s *Server) AdminListKeys(c *gin.Context) {
	if !s.RequireAdmin(c) {
		return
	}
	ctx := c.Request.Context()
	rows, err := s.db.Query(ctx, `
		SELECT id, name, key_prefix, permissions, created_at, expires_at, disabled
		FROM public.api_keys ORDER BY id DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error(), Code: http.StatusInternalServerError})
		return
	}
	defer rows.Close()
	var list []gin.H
	for rows.Next() {
		var id int
		var name, keyPrefix *string
		var permsRaw []byte
		var createdAt time.Time
		var expiresAt *time.Time
		var disabled bool
		if err := rows.Scan(&id, &name, &keyPrefix, &permsRaw, &createdAt, &expiresAt, &disabled); err != nil {
			continue
		}
		perms := []string{}
		if len(permsRaw) > 0 {
			_ = json.Unmarshal(permsRaw, &perms)
		}
		nm := ""
		if name != nil {
			nm = *name
		}
		kp := ""
		if keyPrefix != nil {
			kp = *keyPrefix
		}
		exp := ""
		if expiresAt != nil {
			exp = expiresAt.Format(time.RFC3339)
		}
		list = append(list, gin.H{
			"id":          id,
			"name":        nm,
			"key_prefix":  kp,
			"permissions": perms,
			"created_at":  createdAt.Format(time.RFC3339),
			"expires_at":  exp,
			"disabled":    disabled,
		})
	}
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"keys": list}})
}

// AdminRecordTransaction записывает в БД транзакцию административного действия (например contract_verify).
// POST /api/v1/admin/record-transaction. Тело: { "type": "contract_verify", "sender": "GND_SYSTEM", "recipient": "<address>", "payload": "" }
func (s *Server) AdminRecordTransaction(c *gin.Context) {
	if !s.RequireAdmin(c) {
		return
	}
	var req struct {
		Type      string `json:"type"`
		Sender    string `json:"sender"`
		Recipient string `json:"recipient"`
		Payload   string `json:"payload"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Type == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите type (contract_verify и др.)", Code: http.StatusBadRequest})
		return
	}
	if req.Sender == "" {
		req.Sender = "GND_SYSTEM"
	}
	if req.Recipient == "" {
		req.Recipient = "GND_SYSTEM"
	}
	ctx := c.Request.Context()
	if s.core == nil || s.core.Pool == nil {
		c.JSON(http.StatusServiceUnavailable, APIResponse{Success: false, Error: "Блокчейн недоступен", Code: http.StatusServiceUnavailable})
		return
	}
	var genesisID int64
	if s.core.Genesis != nil {
		genesisID = s.core.Genesis.ID
	}
	if err := core.RecordAdminTransaction(ctx, s.core.Pool, genesisID, req.Type, req.Sender, req.Recipient, req.Payload); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: "Ошибка записи транзакции: " + err.Error(), Code: http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"recorded": true, "type": req.Type}})
}

// AdminRevokeKey отключает ключ (disabled = true).
// POST /api/v1/admin/keys/:id/revoke или DELETE /api/v1/admin/keys/:id
func (s *Server) AdminRevokeKey(c *gin.Context) {
	if !s.RequireAdmin(c) {
		return
	}
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите id ключа", Code: http.StatusBadRequest})
		return
	}
	ctx := c.Request.Context()
	cmd, err := s.db.Exec(ctx, "UPDATE public.api_keys SET disabled = true WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error(), Code: http.StatusInternalServerError})
		return
	}
	if cmd.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, APIResponse{Success: false, Error: "Ключ не найден", Code: http.StatusNotFound})
		return
	}
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"revoked": true}})
}

// AdminListWallets возвращает список кошельков с именами и ролями (без мягко удалённых).
// GET /api/v1/admin/wallets
func (s *Server) AdminListWallets(c *gin.Context) {
	if !s.RequireAdmin(c) {
		return
	}
	ctx := c.Request.Context()
	roleFilter := c.Query("role")
	q := `
		SELECT w.id, w.account_id, w.address, w.public_key, w.signer_wallet_id, w.name, w.role, w.created_at, sw.disabled AS signer_disabled
		FROM public.wallets w
		LEFT JOIN public.signer_wallets sw ON sw.id = w.signer_wallet_id
		WHERE COALESCE(w.disabled, false) = false`
	args := []interface{}{}
	if roleFilter != "" {
		q += " AND w.role = $1"
		args = append(args, roleFilter)
	}
	q += " ORDER BY w.created_at DESC"
	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error(), Code: http.StatusInternalServerError})
		return
	}
	defer rows.Close()
	var list []gin.H
	for rows.Next() {
		var id, accountID int
		var address, publicKey string
		var signerWalletID *string
		var name, role *string
		var createdAt time.Time
		var signerDisabled *bool
		if err := rows.Scan(&id, &accountID, &address, &publicKey, &signerWalletID, &name, &role, &createdAt, &signerDisabled); err != nil {
			continue
		}
		nm := ""
		if name != nil {
			nm = *name
		}
		rl := ""
		if role != nil {
			rl = *role
		}
		sw := ""
		if signerWalletID != nil {
			sw = *signerWalletID
		}
		blocked := signerDisabled != nil && *signerDisabled
		list = append(list, gin.H{
			"id":               id,
			"account_id":       accountID,
			"address":          address,
			"name":             nm,
			"role":             rl,
			"signer_wallet_id": sw,
			"created_at":       createdAt.Format(time.RFC3339),
			"blocked":          blocked,
		})
	}
	// private_key не возвращаем; добавляем массив metadata (лого, адрес контракта)
	metadata := s.getTokenMetadata(c.Request.Context())
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"wallets": list, "metadata": metadata}})
}

// AdminUpdateWallet обновляет имя и/или роль кошелька по адресу.
// PATCH /api/v1/admin/wallets/:address
func (s *Server) AdminUpdateWallet(c *gin.Context) {
	if !s.RequireAdmin(c) {
		return
	}
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите address", Code: http.StatusBadRequest})
		return
	}
	var req struct {
		Name *string `json:"name"`
		Role *string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Неверный JSON", Code: http.StatusBadRequest})
		return
	}
	if req.Name == nil && req.Role == nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите name и/или role", Code: http.StatusBadRequest})
		return
	}
	ctx := c.Request.Context()
	var tag pgconn.CommandTag
	var err error
	if req.Name != nil && req.Role != nil {
		tag, err = s.db.Exec(ctx, `
			UPDATE public.wallets SET name = $1, role = $2 WHERE address = $3`,
			*req.Name, *req.Role, address)
	} else if req.Name != nil {
		tag, err = s.db.Exec(ctx, `
			UPDATE public.wallets SET name = $1 WHERE address = $2`,
			*req.Name, address)
	} else {
		tag, err = s.db.Exec(ctx, `
			UPDATE public.wallets SET role = $1 WHERE address = $2`,
			*req.Role, address)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error(), Code: http.StatusInternalServerError})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, APIResponse{Success: false, Error: "Кошелёк не найден", Code: http.StatusNotFound})
		return
	}
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"updated": true}})
}

// AdminDisableWallet блокирует подписание: устанавливает signer_wallets.disabled = true.
// POST /api/v1/admin/wallets/:address/disable
func (s *Server) AdminDisableWallet(c *gin.Context) {
	if !s.RequireAdmin(c) {
		return
	}
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите address", Code: http.StatusBadRequest})
		return
	}
	ctx := c.Request.Context()
	tag, err := s.db.Exec(ctx, `
		UPDATE public.signer_wallets SET disabled = true, updated_at = NOW()
		WHERE id = (SELECT signer_wallet_id FROM public.wallets WHERE address = $1 AND COALESCE(disabled, false) = false)`,
		address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error(), Code: http.StatusInternalServerError})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, APIResponse{Success: false, Error: "Кошелёк не найден или без signer_wallet", Code: http.StatusNotFound})
		return
	}
	recordWalletTransaction(s, c.Request.Context(), "wallet_disable", address)
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"disabled": true}})
}

// AdminEnableWallet снимает блокировку подписания: signer_wallets.disabled = false.
// POST /api/v1/admin/wallets/:address/enable
func (s *Server) AdminEnableWallet(c *gin.Context) {
	if !s.RequireAdmin(c) {
		return
	}
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите address", Code: http.StatusBadRequest})
		return
	}
	ctx := c.Request.Context()
	tag, err := s.db.Exec(ctx, `
		UPDATE public.signer_wallets SET disabled = false, updated_at = NOW()
		WHERE id = (SELECT signer_wallet_id FROM public.wallets WHERE address = $1 AND COALESCE(disabled, false) = false)`,
		address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error(), Code: http.StatusInternalServerError})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, APIResponse{Success: false, Error: "Кошелёк не найден или без signer_wallet", Code: http.StatusNotFound})
		return
	}
	recordWalletTransaction(s, c.Request.Context(), "wallet_enable", address)
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"enabled": true}})
}

// AdminDeleteWallet мягкое удаление: wallets.disabled = true (скрывается из списка).
// DELETE /api/v1/admin/wallets/:address или POST /api/v1/admin/wallets/:address/delete
func (s *Server) AdminDeleteWallet(c *gin.Context) {
	if !s.RequireAdmin(c) {
		return
	}
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите address", Code: http.StatusBadRequest})
		return
	}
	ctx := c.Request.Context()
	tag, err := s.db.Exec(ctx, `UPDATE public.wallets SET disabled = true WHERE address = $1`, address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error(), Code: http.StatusInternalServerError})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, APIResponse{Success: false, Error: "Кошелёк не найден", Code: http.StatusNotFound})
		return
	}
	recordWalletTransaction(s, c.Request.Context(), "wallet_delete", address)
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"deleted": true}})
}

// recordWalletTransaction записывает в БД транзакцию операции с кошельком (disable, enable, delete).
func recordWalletTransaction(s *Server, ctx context.Context, txType, address string) {
	pool := s.db
	if s.core != nil && s.core.Pool != nil {
		pool = s.core.Pool
	}
	genesisID := int64(0)
	if s.core != nil && s.core.Genesis != nil {
		genesisID = s.core.Genesis.ID
	}
	if pool != nil {
		if err := core.RecordAdminTransaction(ctx, pool, genesisID, txType, "GND_ADMIN", address, ""); err != nil {
			// логируем в stdout, т.к. в admin нет доступа к log
			fmt.Printf("[REST] запись транзакции %s в gnd_db.transactions: %v\n", txType, err)
		}
	}
}

// AdminContractDisable записывает транзакцию блокировки контракта. POST /api/v1/admin/contracts/:address/disable
func (s *Server) AdminContractDisable(c *gin.Context) {
	if !s.RequireAdmin(c) {
		return
	}
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите address контракта", Code: http.StatusBadRequest})
		return
	}
	recordWalletTransaction(s, c.Request.Context(), "contract_disable", address)
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"recorded": true, "type": "contract_disable"}})
}

// AdminContractDelete записывает транзакцию удаления контракта. POST /api/v1/admin/contracts/:address/delete
func (s *Server) AdminContractDelete(c *gin.Context) {
	if !s.RequireAdmin(c) {
		return
	}
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите address контракта", Code: http.StatusBadRequest})
		return
	}
	recordWalletTransaction(s, c.Request.Context(), "contract_delete", address)
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"recorded": true, "type": "contract_delete"}})
}

// AdminUpdateContractABI обновляет только ABI контракта по адресу. PATCH /api/v1/admin/contracts/:address/abi. Body: {"abi": [...]}.
func (s *Server) AdminUpdateContractABI(c *gin.Context) {
	if !s.RequireAdmin(c) {
		return
	}
	address := strings.TrimSpace(c.Param("address"))
	if address == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите address контракта", Code: http.StatusBadRequest})
		return
	}
	var req struct {
		ABI json.RawMessage `json:"abi"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Неверный формат тела. Ожидается {\"abi\": [...]}", Code: http.StatusBadRequest})
		return
	}
	if len(req.ABI) == 0 {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите abi (JSON-массив ABI)", Code: http.StatusBadRequest})
		return
	}
	pool := s.db
	if s.core != nil && s.core.Pool != nil {
		pool = s.core.Pool
	}
	if pool == nil {
		c.JSON(http.StatusServiceUnavailable, APIResponse{Success: false, Error: "БД недоступна", Code: http.StatusServiceUnavailable})
		return
	}
	if err := core.UpdateContractABI(c.Request.Context(), pool, address, []byte(req.ABI)); err != nil {
		if strings.Contains(err.Error(), "не найден") {
			c.JSON(http.StatusNotFound, APIResponse{Success: false, Error: err.Error(), Code: http.StatusNotFound})
			return
		}
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: err.Error(), Code: http.StatusBadRequest})
		return
	}
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"address": address, "message": "ABI обновлён"}})
}

// AdminTokenDisable записывает транзакцию блокировки токена. POST /api/v1/admin/tokens/:id/disable
func (s *Server) AdminTokenDisable(c *gin.Context) {
	if !s.RequireAdmin(c) {
		return
	}
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите id токена", Code: http.StatusBadRequest})
		return
	}
	recordWalletTransaction(s, c.Request.Context(), "token_disable", id)
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"recorded": true, "type": "token_disable"}})
}

// AdminTokenDelete записывает транзакцию удаления токена. POST /api/v1/admin/tokens/:id/delete
func (s *Server) AdminTokenDelete(c *gin.Context) {
	if !s.RequireAdmin(c) {
		return
	}
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите id токена", Code: http.StatusBadRequest})
		return
	}
	recordWalletTransaction(s, c.Request.Context(), "token_delete", id)
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"recorded": true, "type": "token_delete"}})
}
