// | KB @CerbeRus - Nexus Invest Team
// api/admin.go — админские маршруты: выдача/отзыв API-ключей, имена и роли кошельков.

package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

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
	now := time.Now().UTC()
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

// AdminListWallets возвращает список кошельков с именами и ролями.
// GET /api/v1/admin/wallets
func (s *Server) AdminListWallets(c *gin.Context) {
	if !s.RequireAdmin(c) {
		return
	}
	ctx := c.Request.Context()
	roleFilter := c.Query("role")
	q := `
		SELECT w.id, w.account_id, w.address, w.public_key, w.signer_wallet_id, w.name, w.role, w.created_at
		FROM public.wallets w
		WHERE 1=1`
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
		if err := rows.Scan(&id, &accountID, &address, &publicKey, &signerWalletID, &name, &role, &createdAt); err != nil {
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
		list = append(list, gin.H{
			"id":               id,
			"account_id":       accountID,
			"address":          address,
			"name":             nm,
			"role":             rl,
			"signer_wallet_id": sw,
			"created_at":       createdAt.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"wallets": list}})
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
