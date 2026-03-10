// | KB @CerbeRus - Nexus Invest Team
// api/rest.go

package api

import (
	"GND/audit"
	"GND/core"
	"GND/tokens/deployer"
	"GND/tokens/interfaces"
	"GND/tokens/registry"
	"GND/tokens/standards/gndst1"
	tokentypes "GND/tokens/types"
	"GND/types"
	"GND/vm"
	"GND/vm/compiler"
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"math/big"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UploadDirTokenLogos — каталог для загруженных логотипов токенов (относительно рабочей директории).
const UploadDirTokenLogos = "uploads/token_logos"

// MaxLogoFileSize — максимальный размер файла логотипа в байтах (2 МБ).
const MaxLogoFileSize = 2 * 1024 * 1024

// APIResponse — унифицированная структура ответа API
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    int         `json:"code,omitempty"`
}

func sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Ошибка кодирования JSON: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
	}
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	sendJSON(w, APIResponse{
		Success: false,
		Error:   message,
		Code:    statusCode,
	}, statusCode)
}

// RecoverMiddleware защищает от паник в хендлерах
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("Восстановление после паники: %v", rec)
				sendError(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware — заглушка для проверки API-ключа (в будущем реализовать полноценную проверку)

// AdminSigner — интерфейс подписи digest по кошельку (signer_wallet_id). Используется для подписи транзакций из админки.
type AdminSigner interface {
	SignDigest(ctx context.Context, walletID uuid.UUID, digest []byte) ([]byte, error)
}

// Server представляет HTTP сервер
type Server struct {
	router      *gin.Engine
	db          *pgxpool.Pool
	core        *core.Blockchain
	mempool     *core.Mempool
	deployer    *deployer.Deployer
	cfg         *core.Config
	evm         *vm.EVM     // для вызова контрактов (view) и отправки транзакций (write)
	adminSigner AdminSigner // опционально: для подписи транзакций от имени кошелька при запросе из админки
}

// NewServer создает новый экземпляр сервера. deployer может быть nil — тогда POST /token/deploy недоступен. cfg опционально — для health (chain_id, subnet_id).
func NewServer(db *pgxpool.Pool, blockchain *core.Blockchain, mempool *core.Mempool, tokenDeployer *deployer.Deployer, cfg *core.Config) *Server {
	server := &Server{
		router:   gin.Default(),
		db:       db,
		core:     blockchain,
		mempool:  mempool,
		deployer: tokenDeployer,
		cfg:      cfg,
	}
	// События Transfer/Approval токенов GND-st1 рассылаются подписчикам WebSocket (порт 8183).
	gndst1.TokenEventNotifier = func(contract, eventType, from, to, amount string) {
		NotifyContractEvent(map[string]interface{}{
			"contract": contract,
			"type":     eventType,
			"from":     from,
			"to":       to,
			"amount":   amount,
		})
	}
	server.setupRoutes()
	return server
}

// Start запускает сервер
func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}

// GetMetrics возвращает текущие метрики блокчейна
func (s *Server) GetMetrics(c *gin.Context) {
	metrics := core.GetMetrics()
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    metrics,
	})
}

// GetTransactionMetrics возвращает метрики по транзакциям
func (s *Server) GetTransactionMetrics(c *gin.Context) {
	metrics := core.GetMetrics()
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    metrics.TransactionMetrics,
	})
}

// GetFeeMetrics возвращает метрики по комиссиям
func (s *Server) GetFeeMetrics(c *gin.Context) {
	metrics := core.GetMetrics()
	feeMetrics := struct {
		AverageFee      float64
		MinFee          *big.Int
		MaxFee          *big.Int
		TotalFee        *big.Int
		FeeDistribution map[string]uint64
		TypeMetrics     map[string]*core.TransactionTypeMetrics
	}{
		AverageFee:      metrics.TransactionMetrics.AverageFee,
		MinFee:          metrics.TransactionMetrics.MinFee,
		MaxFee:          metrics.TransactionMetrics.MaxFee,
		TotalFee:        metrics.TransactionMetrics.TotalFee,
		FeeDistribution: metrics.TransactionMetrics.FeeDistribution,
		TypeMetrics:     metrics.TransactionMetrics.TypeMetrics,
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    feeMetrics,
	})
}

// GetAlerts возвращает текущие алерты
func (s *Server) GetAlerts(c *gin.Context) {
	metrics := core.GetMetrics()
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    metrics.Alerts.AlertHistory,
	})
}

// SetAlertThresholds устанавливает пороговые значения для алертов
func (s *Server) SetAlertThresholds(c *gin.Context) {
	var thresholds struct {
		HighFee     *big.Int `json:"high_fee"`
		LowFee      *big.Int `json:"low_fee"`
		HighLatency int64    `json:"high_latency"`
		HighCPU     float64  `json:"high_cpu"`
		HighMemory  uint64   `json:"high_memory"`
	}

	if err := c.ShouldBindJSON(&thresholds); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Неверный формат данных",
			Code:    http.StatusBadRequest,
		})
		return
	}

	core.SetAlertThresholds(
		thresholds.HighFee,
		thresholds.LowFee,
		time.Duration(thresholds.HighLatency)*time.Millisecond,
		thresholds.HighCPU,
		thresholds.HighMemory,
	)

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    "Пороговые значения обновлены",
	})
}

// HealthCheck возвращает статус сервера и идентификаторы сети (для мостов и подсетей)
func (s *Server) HealthCheck(c *gin.Context) {
	data := gin.H{
		"status":    "ok",
		"version":   "1.0.0",
		"timestamp": core.BlockchainNow().Format(time.RFC3339),
	}
	if s.cfg != nil {
		data["network_id"] = s.cfg.NetworkID
		data["chain_id"] = s.cfg.ChainID
		if s.cfg.SubnetID != "" {
			data["subnet_id"] = s.cfg.SubnetID
		}
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

// getTokenMetadata возвращает массив метаданных токенов (logo_url, contract_address) из БД. Не возвращает private_key.
func (s *Server) getTokenMetadata(ctx context.Context) []gin.H {
	if s.db == nil {
		return nil
	}
	rows, err := s.db.Query(ctx, `
		SELECT COALESCE(c.address, ''), COALESCE(t.logo_url, '')
		FROM tokens t
		LEFT JOIN contracts c ON c.id = t.contract_id
		ORDER BY t.symbol`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var list []gin.H
	for rows.Next() {
		var addr, logo string
		if err := rows.Scan(&addr, &logo); err != nil {
			continue
		}
		list = append(list, gin.H{"logo_url": logo, "contract_address": addr})
	}
	return list
}

// CreateWallet создаёт новый кошелёк. Требуется заголовок X-API-Key.
// В ответе нет private_key; добавлен массив metadata (лого, адрес контракта).
func (s *Server) CreateWallet(c *gin.Context) {
	apiKey := c.GetHeader("X-API-Key")
	if !ValidateAPIKey(c.Request.Context(), s.db, apiKey) {
		c.JSON(http.StatusUnauthorized, APIResponse{
			Success: false,
			Error:   "Неверный или отсутствующий X-API-Key",
			Code:    http.StatusUnauthorized,
		})
		return
	}
	wallet, err := s.core.CreateWallet(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Ошибка создания кошелька: " + err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}
	if wallet.SignerWalletID != nil {
		log.Printf("[REST] Кошелёк создан через signing_service, запись в signer_wallets: %s (адрес: %s)", wallet.SignerWalletID.String(), wallet.Address)
	}
	// Ответ без private_key: только address, public_key, signer_wallet_id и metadata
	data := gin.H{
		"address":  string(wallet.Address),
		"metadata": s.getTokenMetadata(c.Request.Context()),
	}
	if wallet.PrivateKey != nil {
		data["public_key"] = wallet.PublicKeyHex()
	} else {
		data["public_key"] = ""
	}
	if wallet.SignerWalletID != nil {
		data["signer_wallet_id"] = wallet.SignerWalletID.String()
	}
	pool := s.db
	if s.core != nil && s.core.Pool != nil {
		pool = s.core.Pool
	}
	genesisID := int64(0)
	if s.core != nil && s.core.Genesis != nil {
		genesisID = s.core.Genesis.ID
	}
	if pool != nil {
		if errTx := core.RecordAdminTransaction(c.Request.Context(), pool, genesisID, "wallet_create", "GND_SYSTEM", string(wallet.Address), ""); errTx != nil {
			log.Printf("[REST] запись транзакции wallet_create в gnd_db.transactions: %v", errTx)
		}
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

// GetNativeCoinBalance возвращает баланс нативной монеты (GND, GANI) для адреса. GET /api/v1/coin/:symbol/balance/:owner
func (s *Server) GetNativeCoinBalance(c *gin.Context) {
	symbol := strings.TrimSpace(strings.ToUpper(c.Param("symbol")))
	owner := strings.TrimSpace(c.Param("owner"))
	if !core.IsNativeSymbol(symbol) {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Допустимые символы нативных монет: GND, GANI",
			Code:    http.StatusBadRequest,
		})
		return
	}
	if owner == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Укажите owner (адрес)",
			Code:    http.StatusBadRequest,
		})
		return
	}
	var balance *big.Int
	if s.core != nil && s.core.State != nil {
		balance = s.core.State.GetBalance(types.Address(owner), symbol)
	} else {
		balance = big.NewInt(0)
	}
	if balance == nil {
		balance = big.NewInt(0)
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: gin.H{
			"symbol":  symbol,
			"owner":   owner,
			"balance": balance.String(),
		},
	})
}

// GetNativeCoinSupply возвращает total_supply и circulating_supply нативной монеты из БД. GET /api/v1/coin/:symbol/supply
func (s *Server) GetNativeCoinSupply(c *gin.Context) {
	symbol := strings.TrimSpace(strings.ToUpper(c.Param("symbol")))
	if !core.IsNativeSymbol(symbol) {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Допустимые символы нативных монет: GND, GANI",
			Code:    http.StatusBadRequest,
		})
		return
	}
	if s.core == nil || s.core.Pool == nil {
		c.JSON(http.StatusServiceUnavailable, APIResponse{
			Success: false,
			Error:   "Сервис недоступен",
			Code:    http.StatusServiceUnavailable,
		})
		return
	}
	tok, err := core.GetTokenBySymbol(c.Request.Context(), s.core.Pool, symbol)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error:   "Монета не найдена: " + symbol,
			Code:    http.StatusNotFound,
		})
		return
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: gin.H{
			"symbol":             symbol,
			"name":               tok.Name,
			"decimals":           tok.Decimals,
			"total_supply":       tok.TotalSupply,
			"circulating_supply": tok.CirculatingSupply,
		},
	})
}

// GetBalance возвращает все балансы токенов кошелька из token_balances с полями из tokens (standard, symbol, name, decimals, is_verified)
// Включает нативные монеты (GND, GANI) из native_balances.
func (s *Server) GetBalance(c *gin.Context) {
	address := c.Param("address")
	balances := []core.WalletTokenBalance{}
	if s.core.Pool != nil {
		var err error
		var nc *core.NativeContractsConfig
		if s.cfg != nil && s.cfg.NativeContracts != nil {
			nc = s.cfg.NativeContracts
		}
		balances, err = core.GetWalletTokenBalances(c.Request.Context(), s.core.Pool, address, nc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Error:   err.Error(),
				Code:    http.StatusInternalServerError,
			})
			return
		}
		if balances == nil {
			balances = []core.WalletTokenBalance{}
		}
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: gin.H{
			"address":  address,
			"balances": balances,
		},
	})
}

// decodeSignatureHex декодирует подпись из hex-строки (0x или без префикса). Иначе возвращает []byte(s) как есть.
func decodeSignatureHex(s string) []byte {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		s = s[2:]
	}
	if len(s) == 128 {
		b, err := hex.DecodeString(s)
		if err == nil && len(b) == 64 {
			return b
		}
	}
	return []byte(s)
}

// SendTransaction отправляет транзакцию
func (s *Server) SendTransaction(c *gin.Context) {
	var txData struct {
		From            string   `json:"from"`
		To              string   `json:"to"`
		Value           *big.Int `json:"value"`
		Fee             *big.Int `json:"fee"`
		Nonce           uint64   `json:"nonce"`
		Type            string   `json:"type"`
		Data            string   `json:"data"`
		Signature       string   `json:"signature"`
		SenderPublicKey string   `json:"sender_public_key"` // hex публичного ключа P-256 для проверки подписи
	}
	if err := c.ShouldBindJSON(&txData); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Неверный формат данных транзакции",
			Code:    http.StatusBadRequest,
		})
		return
	}

	fromAddr, err := types.ParseAddress(txData.From)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Неверный формат адреса отправителя: " + err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	toAddr, err := types.ParseAddress(txData.To)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Неверный формат адреса получателя: " + err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	value := txData.Value
	if value == nil {
		value = big.NewInt(0)
	}
	fee := txData.Fee
	if fee == nil {
		fee = big.NewInt(0)
	}

	sigBytes := decodeSignatureHex(txData.Signature)

	tx := &core.Transaction{
		Sender:             fromAddr,
		Recipient:          toAddr,
		Value:              value,
		Nonce:              int64(txData.Nonce),
		Data:               []byte(txData.Data),
		Signature:          sigBytes,
		GasLimit:           21000, // Стандартный лимит газа для простой транзакции
		GasPrice:           fee,
		Symbol:             "GND",
		IsVerified:         false, // пользовательские транзакции требуют проверки подписи
		SenderPublicKeyHex: strings.TrimSpace(txData.SenderPublicKey),
		Timestamp:          core.BlockchainNow(),
		Status:             "pending",
	}
	tx.Hash = tx.CalculateHash()

	result, err := s.core.SendTransaction(tx)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Ошибка отправки транзакции: " + err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    result,
	})
}

// transactionResponse формирует объект для JSON-ответа: contract_id как null или число (не sql.NullInt64).
func transactionResponse(tx *core.Transaction) gin.H {
	var contractID interface{}
	if tx.ContractID.Valid {
		contractID = tx.ContractID.Int64
	} else {
		contractID = nil
	}
	dataHex := ""
	if len(tx.Data) > 0 {
		dataHex = "x" + hex.EncodeToString(tx.Data)
	}
	payloadHex := ""
	if len(tx.Payload) > 0 {
		payloadHex = "x" + hex.EncodeToString(tx.Payload)
	} else if len(tx.Data) > 0 {
		payloadHex = "x" + hex.EncodeToString(tx.Data)
	}
	sigHex := ""
	if len(tx.Signature) > 0 {
		sigHex = hex.EncodeToString(tx.Signature)
	}
	value := "0"
	if tx.Value != nil {
		value = tx.Value.String()
	}
	fee := interface{}(nil)
	if tx.Fee != nil {
		fee = tx.Fee.String()
	}
	gasPrice := int64(1)
	if tx.GasPrice != nil {
		gasPrice = tx.GasPrice.Int64()
	}
	data := gin.H{
		"id": tx.ID, "sender": tx.Sender.String(), "recipient": tx.Recipient.String(),
		"value": value, "data": dataHex, "nonce": tx.Nonce,
		"gas_limit": tx.GasLimit, "gas_price": gasPrice,
		"signature": sigHex, "hash": tx.Hash, "fee": fee,
		"type": tx.Type, "status": tx.Status, "timestamp": tx.Timestamp,
		"block_id": tx.BlockID, "contract_id": contractID, "payload": payloadHex,
		"symbol": tx.Symbol, "is_verified": tx.IsVerified,
	}
	return data
}

// transactionResponseWithBlockNumber возвращает ответ по транзакции и добавляет block_number (номер в цепи) по block_id для сканера.
func (s *Server) transactionResponseWithBlockNumber(c *gin.Context, tx *core.Transaction) gin.H {
	data := transactionResponse(tx)
	pool := s.db
	if s.core != nil && s.core.Pool != nil {
		pool = s.core.Pool
	}
	if pool != nil && tx != nil && tx.BlockID > 0 {
		if blockNum, err := core.GetBlockIndexByID(c.Request.Context(), pool, int64(tx.BlockID)); err == nil {
			data["block_number"] = blockNum
		}
	}
	return data
}

// GetTransaction возвращает информацию о транзакции (сначала в памяти/мемпуле, затем в gnd_db.transactions).
func (s *Server) GetTransaction(c *gin.Context) {
	hash := strings.TrimSpace(c.Param("hash"))
	if hash == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите хеш транзакции", Code: http.StatusBadRequest})
		return
	}
	tx, err := s.core.GetTransaction(hash)
	if err != nil || tx == nil {
		// Транзакция может быть в БД (подтверждённая), но не в памяти — ищем в gnd_db.transactions
		pool := s.db
		if s.core != nil && s.core.Pool != nil {
			pool = s.core.Pool
		}
		if pool != nil {
			tx, err = core.LoadTransactionByHash(c.Request.Context(), pool, hash)
			if err == nil && tx != nil {
				c.JSON(http.StatusOK, APIResponse{Success: true, Data: s.transactionResponseWithBlockNumber(c, tx)})
				return
			}
		}
		msg := "Транзакция не найдена"
		if err != nil {
			msg = "Транзакция не найдена: " + err.Error()
		}
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error:   msg,
			Code:    http.StatusNotFound,
		})
		return
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    s.transactionResponseWithBlockNumber(c, tx),
	})
}

// GetLatestBlock возвращает последний блок (с полем Transactions — список транзакций)
func (s *Server) GetLatestBlock(c *gin.Context) {
	block, err := s.core.GetLatestBlock()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Ошибка получения последнего блока: " + err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}
	if block != nil {
		if s.db != nil {
			txs, _ := core.LoadTransactionsForBlock(c.Request.Context(), s.db, block.ID)
			if txs != nil {
				block.Transactions = txs
			} else {
				block.Transactions = []*core.Transaction{}
			}
		} else {
			block.Transactions = []*core.Transaction{}
		}
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    block,
	})
}

// GetBlockByNumber возвращает блок по номеру
func (s *Server) GetBlockByNumber(c *gin.Context) {
	numberStr := c.Param("number")
	number, err := strconv.ParseUint(numberStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Неверный номер блока",
			Code:    http.StatusBadRequest,
		})
		return
	}
	block, err := s.core.GetBlockByNumber(number)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Ошибка получения блока: " + err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}
	if block == nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error:   "Блок не найден",
			Code:    http.StatusNotFound,
		})
		return
	}
	if s.db != nil {
		txs, _ := core.LoadTransactionsForBlock(c.Request.Context(), s.db, block.ID)
		if txs != nil {
			block.Transactions = txs
		} else {
			block.Transactions = []*core.Transaction{}
		}
	}
	if block.Transactions == nil {
		block.Transactions = []*core.Transaction{}
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    block,
	})
}

// DeployContract деплоит новый контракт
func (s *Server) DeployContract(c *gin.Context) {
	var paramsData struct {
		From        string                 `json:"from"`
		Bytecode    string                 `json:"bytecode"`
		ABI         json.RawMessage        `json:"abi"`
		Name        string                 `json:"name"`
		Symbol      string                 `json:"symbol"`
		Standard    string                 `json:"standard"`
		Owner       string                 `json:"owner"`
		Compiler    string                 `json:"compiler"`
		Version     string                 `json:"version"`
		License     string                 `json:"license"`
		Params      map[string]interface{} `json:"params"`
		Description string                 `json:"description"`
		MetadataCID string                 `json:"metadata_cid"`
		Metadata    json.RawMessage        `json:"metadata"`
		SourceCode  string                 `json:"source_code"`
		GasLimit    uint64                 `json:"gas_limit"`
		GasPrice    *big.Int               `json:"gas_price"`
		Nonce       uint64                 `json:"nonce"`
		Signature   string                 `json:"signature"`
		TotalSupply *big.Int               `json:"total_supply"`
	}
	if err := c.ShouldBindJSON(&paramsData); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Неверный формат данных",
			Code:    http.StatusBadRequest,
		})
		return
	}
	params := core.ContractParams{
		From:        paramsData.From,
		Bytecode:    paramsData.Bytecode,
		ABI:         paramsData.ABI,
		Name:        paramsData.Name,
		Symbol:      paramsData.Symbol,
		Standard:    paramsData.Standard,
		Owner:       paramsData.Owner,
		Compiler:    paramsData.Compiler,
		Version:     paramsData.Version,
		License:     paramsData.License,
		Params:      paramsData.Params,
		Description: paramsData.Description,
		MetadataCID: paramsData.MetadataCID,
		Metadata:    paramsData.Metadata,
		SourceCode:  paramsData.SourceCode,
		GasLimit:    paramsData.GasLimit,
		GasPrice:    paramsData.GasPrice,
		Nonce:       paramsData.Nonce,
		Signature:   paramsData.Signature,
		TotalSupply: paramsData.TotalSupply,
	}
	// При деплое с адреса gndself_address комиссия за деплой не взимается (при наличии такой логики в core).
	address, err := s.core.DeployContract(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Ошибка деплоя контракта: " + err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}
	// Записываем транзакцию деплоя в БД (для отображения в админке и истории)
	pool := s.core.Pool
	if pool == nil {
		pool = s.db
	}
	genesisID := int64(0)
	if s.core != nil && s.core.Genesis != nil {
		genesisID = s.core.Genesis.ID
	}
	if pool != nil {
		if errTx := core.RecordAdminTransaction(c.Request.Context(), pool, genesisID, "contract_deploy", params.From, address, address); errTx != nil {
			log.Printf("[REST] запись транзакции contract_deploy в gnd_db.transactions: %v", errTx)
		}
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    gin.H{"address": address},
	})
}

// GetContract возвращает информацию о контракте
func (s *Server) GetContract(c *gin.Context) {
	address := c.Param("address")
	contract, err := s.core.GetContract(address)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error:   "Контракт не найден: " + err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    contract,
	})
}

// GetContractState возвращает состояние контракта (name, symbol, total_supply, balances по адресам). GET /api/v1/contract/:address/state?addresses=addr1,addr2
func (s *Server) GetContractState(c *gin.Context) {
	address := strings.TrimSpace(c.Param("address"))
	if address == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите address контракта", Code: http.StatusBadRequest})
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
	addressesParam := c.Query("addresses")
	var accountAddresses []string
	if addressesParam != "" {
		for _, a := range strings.Split(addressesParam, ",") {
			a = strings.TrimSpace(a)
			if a != "" {
				accountAddresses = append(accountAddresses, a)
			}
		}
	}
	state, err := core.GetContractState(c.Request.Context(), pool, address, accountAddresses)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{Success: false, Error: err.Error(), Code: http.StatusNotFound})
		return
	}
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: state})
}

// GetContractView возвращает данные для вкладки «Просмотр контракта»: исходный код контракта (Solidity), ABI,
// список методов/функций самого контракта (view и write из ABI). Просмотр/чтение/запись относятся к методам контракта.
// GET /api/v1/contract/:address/view
func (s *Server) GetContractView(c *gin.Context) {
	address := strings.TrimSpace(c.Param("address"))
	if address == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите address контракта", Code: http.StatusBadRequest})
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
	ctx := c.Request.Context()
	contract, err := core.LoadContract(ctx, pool, address)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{Success: false, Error: err.Error(), Code: http.StatusNotFound})
		return
	}
	var abiJSON json.RawMessage
	if len(contract.ABI) > 0 {
		abiJSON = contract.ABI
	} else {
		abiJSON = json.RawMessage("[]")
	}
	viewFuncs, writeFuncs, _ := ParseABIFunctions(contract.ABI)
	out := gin.H{
		"address":         address,
		"source_code":     contract.SourceCode,
		"abi":             abiJSON,
		"view_functions":  viewFuncs,
		"write_functions": writeFuncs,
		"compiler":        contract.Compiler,
	}
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: out})
}

// ContractCall выполняет вызов view/constant метода контракта (без создания транзакции). POST /api/v1/contract/:address/call
// Body: { "data": "0x..." [, "from": "GN_..."] }. data — ABI-encoded calldata (selector + args).
func (s *Server) ContractCall(c *gin.Context) {
	address := strings.TrimSpace(c.Param("address"))
	if address == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите address контракта", Code: http.StatusBadRequest})
		return
	}
	if s.evm == nil {
		c.JSON(http.StatusServiceUnavailable, APIResponse{Success: false, Error: "EVM недоступен для вызова контракта", Code: http.StatusServiceUnavailable})
		return
	}
	var req struct {
		Data string `json:"data"` // hex с префиксом 0x
		From string `json:"from"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Data == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите data (hex calldata)", Code: http.StatusBadRequest})
		return
	}
	data, err := decodeHex(req.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Неверный hex data: " + err.Error(), Code: http.StatusBadRequest})
		return
	}
	from := strings.TrimSpace(req.From)
	if from == "" {
		from = "0x0000000000000000000000000000000000000000"
	}
	gasLimit := uint64(300000)
	result, err := s.evm.CallContractStatic(from, address, data, gasLimit, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Вызов контракта: " + err.Error(), Code: http.StatusBadRequest})
		return
	}
	returnHex := ""
	success := result != nil && result.Error == nil
	dataMap := gin.H{"return_data": returnHex, "success": success}
	if result != nil && len(result.ReturnData) > 0 {
		returnHex = "0x" + hex.EncodeToString(result.ReturnData)
		dataMap["return_data"] = returnHex
		// Для uint256 (32 байта) добавляем десятичное значение для отображения (totalSupply, balanceOf и т.п.)
		if len(result.ReturnData) == 32 {
			val := new(big.Int).SetBytes(result.ReturnData)
			dataMap["return_data_decoded"] = val.String()
		}
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    dataMap,
	})
}

// ContractSend создаёт и отправляет транзакцию вызова метода контракта (transfer, approve и т.д.). POST /api/v1/contract/:address/send
// Body: { "from": "GN_...", "data": "0x...", "value": "0", "gas_limit": 100000 }.
func (s *Server) ContractSend(c *gin.Context) {
	address := strings.TrimSpace(c.Param("address"))
	if address == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите address контракта", Code: http.StatusBadRequest})
		return
	}
	if s.core == nil || s.evm == nil {
		c.JSON(http.StatusServiceUnavailable, APIResponse{Success: false, Error: "Нода недоступна для отправки транзакции", Code: http.StatusServiceUnavailable})
		return
	}
	var req struct {
		From            string `json:"from"`
		Data            string `json:"data"`
		Value           string `json:"value"`
		GasLimit        uint64 `json:"gas_limit"`
		Signature       string `json:"signature"`
		SenderPublicKey string `json:"sender_public_key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Неверный формат данных", Code: http.StatusBadRequest})
		return
	}
	if strings.TrimSpace(req.From) == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите from (адрес отправителя)", Code: http.StatusBadRequest})
		return
	}
	if strings.TrimSpace(req.Data) == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите data (hex calldata)", Code: http.StatusBadRequest})
		return
	}
	data, err := decodeHex(req.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Неверный hex data: " + err.Error(), Code: http.StatusBadRequest})
		return
	}
	val := big.NewInt(0)
	if req.Value != "" {
		val.SetString(req.Value, 10)
	}
	gasLimit := req.GasLimit
	if gasLimit == 0 {
		gasLimit = 200000
	}
	fromAddr := strings.TrimSpace(req.From)
	tx := &core.Transaction{
		Sender:             types.Address(fromAddr),
		Recipient:          types.Address(address),
		Data:               data,
		GasLimit:           gasLimit,
		GasPrice:           big.NewInt(1),
		Value:              val,
		Timestamp:          core.BlockchainNow(),
		Type:               "contract_call",
		Status:             "pending",
		Symbol:             "GND",
		Signature:          decodeSignatureHex(req.Signature),
		SenderPublicKeyHex: strings.TrimSpace(req.SenderPublicKey),
		IsVerified:         false,
	}
	if s.core != nil && s.core.State != nil {
		tx.Nonce = s.core.State.GetNonce(types.Address(fromAddr))
	}
	tx.Hash = tx.CalculateHash()
	// Подпись из админки: если подпись не передана, но передан X-Admin-Token и кошелёк from управляется нодой (signer_wallet_id), подписываем транзакцию нодой.
	if len(tx.Signature) == 0 && tx.SenderPublicKeyHex == "" && s.adminSigner != nil && s.db != nil && ValidateAdminToken(c.GetHeader("X-Admin-Token")) {
		walletID, errSigner := core.GetSignerWalletIDByAddress(c.Request.Context(), s.db, fromAddr)
		if errSigner == nil {
			sig, errSig := s.adminSigner.SignDigest(c.Request.Context(), walletID, []byte(tx.Hash))
			if errSig == nil {
				tx.Signature = sig
				tx.IsVerified = true
			}
		}
	}
	hash, err := s.core.SendTransaction(tx)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Отправка транзакции: " + err.Error(), Code: http.StatusBadRequest})
		return
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    gin.H{"hash": hash, "message": "Транзакция отправлена в мемпул"},
	})
}

// AdminContractCall вызывает view/constant метод контракта по id (для страницы /admin/contracts/:id). POST /api/v1/admin/contracts/:id/call
func (s *Server) AdminContractCall(c *gin.Context) {
	address, err := s.resolveContractAddressByID(c)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{Success: false, Error: err.Error(), Code: http.StatusNotFound})
		return
	}
	c.Params = append(c.Params, gin.Param{Key: "address", Value: address})
	s.ContractCall(c)
}

// AdminContractSend отправляет транзакцию вызова метода контракта по id (для страницы /admin/contracts/:id). POST /api/v1/admin/contracts/:id/send
func (s *Server) AdminContractSend(c *gin.Context) {
	if !s.RequireAdmin(c) {
		return
	}
	address, err := s.resolveContractAddressByID(c)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{Success: false, Error: err.Error(), Code: http.StatusNotFound})
		return
	}
	c.Params = append(c.Params, gin.Param{Key: "address", Value: address})
	s.ContractSend(c)
}

// resolveContractAddressByID возвращает адрес контракта по id из c.Param("id"); pool берётся из s.db или s.core.Pool.
func (s *Server) resolveContractAddressByID(c *gin.Context) (string, error) {
	idStr := strings.TrimSpace(c.Param("id"))
	if idStr == "" {
		return "", fmt.Errorf("укажите id контракта")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return "", fmt.Errorf("неверный id контракта: %s", idStr)
	}
	pool := s.db
	if s.core != nil && s.core.Pool != nil {
		pool = s.core.Pool
	}
	if pool == nil {
		return "", fmt.Errorf("БД недоступна")
	}
	return core.GetContractAddressByID(c.Request.Context(), pool, id)
}

func decodeHex(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
	}
	return hex.DecodeString(s)
}

// CompileContract компилирует исходный код Solidity. POST /contract/compile
func (s *Server) CompileContract(c *gin.Context) {
	var req struct {
		Source   string `json:"source"`
		Name     string `json:"name"`
		Standard string `json:"standard"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Неверный формат данных", Code: http.StatusBadRequest})
		return
	}
	if req.Source == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Поле source обязательно", Code: http.StatusBadRequest})
		return
	}
	if req.Standard == "" {
		req.Standard = "GND-st1"
	}
	if req.Name == "" {
		req.Name = "Contract"
	}
	solcPath := "solc"
	if s.cfg != nil && s.cfg.EVM.SolcPath != "" {
		solcPath = s.cfg.EVM.SolcPath
	}
	solc := compiler.DefaultSolidityCompiler{SolcPath: solcPath}
	metadata := compiler.ContractMetadata{
		Name:     req.Name,
		Standard: req.Standard,
		Compiler: "solc",
		Version:  "0.8.20",
	}
	result, err := solc.Compile([]byte(req.Source), metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Ошибка компиляции: " + err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: gin.H{
			"bytecode": result.Bytecode,
			"abi":      result.ABI,
			"warnings": result.Warnings,
			"errors":   result.Errors,
		},
	})
}

// AnalyzeContract выполняет проверки безопасности по исходному коду. POST /contract/analyze
func (s *Server) AnalyzeContract(c *gin.Context) {
	var req struct {
		Source string `json:"source"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Неверный формат данных", Code: http.StatusBadRequest})
		return
	}
	if req.Source == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Поле source обязательно", Code: http.StatusBadRequest})
		return
	}
	issues := audit.RunSecurityChecksSource(req.Source, "source")
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    gin.H{"issues": issues},
	})
}

// DeployToken создаёт и регистрирует токен. Требуется заголовок X-API-Key (внешняя система / api-keys).
// Поле owner необязательно: если пусто — создаётся новый кошелёк и назначается владельцем токена (регистрация в один клик).
func (s *Server) DeployToken(c *gin.Context) {
	if s.deployer == nil {
		c.JSON(http.StatusServiceUnavailable, APIResponse{
			Success: false,
			Error:   "Сервис деплоя токенов недоступен",
			Code:    http.StatusServiceUnavailable,
		})
		return
	}
	apiKey := c.GetHeader("X-API-Key")
	if !ValidateAPIKey(c.Request.Context(), s.db, apiKey) {
		c.JSON(http.StatusUnauthorized, APIResponse{
			Success: false,
			Error:   "Неверный или отсутствующий X-API-Key",
			Code:    http.StatusUnauthorized,
		})
		return
	}
	var req struct {
		Name         string   `json:"name"`
		Symbol       string   `json:"symbol"`
		Decimals     uint8    `json:"decimals"`
		TotalSupply  *big.Int `json:"total_supply"`
		Owner        string   `json:"owner"`
		Standard     string   `json:"standard"`
		LogoURL      string   `json:"logo_url"`
		DeployWallet string   `json:"deploy_wallet"` // опциональный кошелёк деплоя (оплачивает газ)
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Неверный формат тела запроса: " + err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}
	if req.Standard == "" {
		req.Standard = "GND-st1"
	}

	// Определяем кошелёк деплоя (оплачивает газ) и владельца.
	// Сценарии:
	// - deploy_wallet не задан, owner задан: деплой и владение от одного кошелька (как раньше).
	// - deploy_wallet не задан, owner пустой: создаём кошелёк, он же owner и деплойер (как раньше).
	// - deploy_wallet задан, owner пустой: деплой выполняется от deploy_wallet, токен считается системным (владелец не задан).
	// - deploy_wallet задан, owner задан: деплой от deploy_wallet, владелец — owner.
	deployFrom := strings.TrimSpace(req.DeployWallet)
	ownerWalletCreated := false

	owner := strings.TrimSpace(req.Owner)
	gndself := ""
	if s.cfg != nil && s.cfg.NativeContracts != nil && s.cfg.NativeContracts.GndselfAddress != "" {
		gndself = strings.TrimSpace(s.cfg.NativeContracts.GndselfAddress)
	}
	if deployFrom == "" {
		if owner == "" {
			if gndself != "" {
				// Адрес не указан — владелец и деплойер = gndself_address (системный токен).
				owner = gndself
				deployFrom = gndself
			} else if s.core == nil {
				c.JSON(http.StatusBadRequest, APIResponse{
					Success: false,
					Error:   "Поле owner или deploy_wallet обязательно: сервис создания кошельков недоступен",
					Code:    http.StatusBadRequest,
				})
				return
			} else {
				wallet, err := s.core.CreateWallet(c.Request.Context())
				if err != nil {
					c.JSON(http.StatusInternalServerError, APIResponse{
						Success: false,
						Error:   "Ошибка создания кошелька для владельца токена: " + err.Error(),
						Code:    http.StatusInternalServerError,
					})
					return
				}
				owner = string(wallet.Address)
				deployFrom = owner
				ownerWalletCreated = true
			}
		} else {
			deployFrom = owner
		}
	} else {
		// deploy_wallet задан; owner может быть пустым (системный токен) или отдельным.
		// Если owner не указан — присваиваем gndself_address (системный токен).
		if owner == "" && gndself != "" {
			owner = gndself
		}
	}

	// При owner или deployFrom = gndself_address комиссия за деплой не взимается.
	skipDeployFee := gndself != "" && (owner == gndself || deployFrom == gndself)

	params := tokentypes.TokenParams{
		Name:          req.Name,
		Symbol:        req.Symbol,
		Decimals:      req.Decimals,
		TotalSupply:   req.TotalSupply,
		Owner:         owner,
		Standard:      req.Standard,
		LogoURL:       strings.TrimSpace(req.LogoURL),
		Deployer:      deployFrom,
		SkipDeployFee: skipDeployFee,
	}
	token, err := s.deployer.DeployToken(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Ошибка деплоя токена: " + err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}
	pool := s.db
	if s.core != nil && s.core.Pool != nil {
		pool = s.core.Pool
	}
	genesisID := int64(0)
	if s.core != nil && s.core.Genesis != nil {
		genesisID = s.core.Genesis.ID
	}
	if pool != nil {
		payload := req.Symbol
		if req.Name != "" {
			payload = req.Name + "|" + req.Symbol
		}
		if errTx := core.RecordAdminTransaction(c.Request.Context(), pool, genesisID, "token_deploy", deployFrom, token.GetAddress(), payload); errTx != nil {
			log.Printf("[REST] запись транзакции token_deploy в gnd_db.transactions: %v", errTx)
		}
	}
	data := gin.H{
		"address":              token.GetAddress(),
		"name":                 token.GetName(),
		"symbol":               token.GetSymbol(),
		"decimals":             token.GetDecimals(),
		"total_supply":         token.GetTotalSupply().String(),
		"standard":             token.GetStandard(),
		"owner":                owner,
		"owner_wallet_created": ownerWalletCreated,
		"logo_url":             params.LogoURL,
		"deploy_fee_waived":    skipDeployFee, // true при owner = gndself_address
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

// TokenLogoUpload загружает файл логотипа, проверяет 250x250 и тип картинки, сохраняет в uploads/token_logos и обновляет tokens.logo_url.
// Требуется X-API-Key. Form: file (обязательно), token_id или symbol или token_address (один из них).
func (s *Server) TokenLogoUpload(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, APIResponse{Success: false, Error: "БД недоступна", Code: http.StatusServiceUnavailable})
		return
	}
	if !ValidateAPIKey(c.Request.Context(), s.db, c.GetHeader("X-API-Key")) {
		c.JSON(http.StatusUnauthorized, APIResponse{Success: false, Error: "Неверный или отсутствующий X-API-Key", Code: http.StatusUnauthorized})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Отсутствует поле file с изображением", Code: http.StatusBadRequest})
		return
	}
	tokenIDStr := strings.TrimSpace(c.PostForm("token_id"))
	symbol := strings.TrimSpace(c.PostForm("symbol"))
	tokenAddress := strings.TrimSpace(c.PostForm("token_address"))
	if tokenIDStr == "" && symbol == "" && tokenAddress == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите token_id, symbol или token_address", Code: http.StatusBadRequest})
		return
	}

	// Открываем и читаем файл (лимит размера)
	fh, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Не удалось прочитать файл: " + err.Error(), Code: http.StatusBadRequest})
		return
	}
	defer fh.Close()
	data, err := io.ReadAll(io.LimitReader(fh, MaxLogoFileSize))
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: "Ошибка чтения файла", Code: http.StatusInternalServerError})
		return
	}
	if int64(len(data)) == MaxLogoFileSize {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Файл слишком большой (макс. 2 МБ)", Code: http.StatusBadRequest})
		return
	}
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if _, _, err := ValidateTokenLogo(data, contentType); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: err.Error(), Code: http.StatusBadRequest})
		return
	}

	// Определяем token_id в БД
	var tokenID int
	if tokenIDStr != "" {
		if id, err := strconv.Atoi(tokenIDStr); err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Некорректный token_id", Code: http.StatusBadRequest})
			return
		} else {
			tokenID = id
		}
	} else {
		if symbol != "" {
			err = s.db.QueryRow(c.Request.Context(), `SELECT id FROM tokens WHERE symbol = $1`, symbol).Scan(&tokenID)
		} else {
			err = s.db.QueryRow(c.Request.Context(), `SELECT t.id FROM tokens t JOIN contracts c ON c.id = t.contract_id WHERE c.address = $1`, tokenAddress).Scan(&tokenID)
		}
		if err != nil {
			c.JSON(http.StatusNotFound, APIResponse{Success: false, Error: "Токен не найден по symbol или token_address", Code: http.StatusNotFound})
			return
		}
	}

	// Сохраняем файл в uploads/token_logos
	if err := os.MkdirAll(UploadDirTokenLogos, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: "Не удалось создать каталог для загрузок", Code: http.StatusInternalServerError})
		return
	}
	ext := LogoExtByContentType(contentType)
	namePart := tokenIDStr
	if namePart == "" {
		namePart = symbol
	}
	if namePart == "" {
		namePart = tokenAddress
	}
	filename := SafeLogoFilename(namePart, uuid.New().String()[:8], ext)
	path := filepath.Join(UploadDirTokenLogos, filename)
	if err := os.WriteFile(path, data, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: "Не удалось сохранить файл: " + err.Error(), Code: http.StatusInternalServerError})
		return
	}
	logoURL := "/uploads/token_logos/" + filename

	// Обновляем tokens.logo_url
	_, err = s.db.Exec(c.Request.Context(), `UPDATE public.tokens SET logo_url = $1 WHERE id = $2`, logoURL, tokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: "Ошибка обновления logo_url: " + err.Error(), Code: http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"logo_url": logoURL}})
}

// TokenLogoSet устанавливает logo_url для токена по id/symbol/address (тело: logo_url). Требуется X-API-Key.
func (s *Server) TokenLogoSet(c *gin.Context) {
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, APIResponse{Success: false, Error: "БД недоступна", Code: http.StatusServiceUnavailable})
		return
	}
	if !ValidateAPIKey(c.Request.Context(), s.db, c.GetHeader("X-API-Key")) {
		c.JSON(http.StatusUnauthorized, APIResponse{Success: false, Error: "Неверный или отсутствующий X-API-Key", Code: http.StatusUnauthorized})
		return
	}
	var req struct {
		TokenID      *int   `json:"token_id"`
		Symbol       string `json:"symbol"`
		TokenAddress string `json:"token_address"`
		LogoURL      string `json:"logo_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.LogoURL) == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Тело запроса должно содержать logo_url (и один из: token_id, symbol, token_address)", Code: http.StatusBadRequest})
		return
	}
	logoURL := strings.TrimSpace(req.LogoURL)
	var tokenID int
	var err error
	if req.TokenID != nil && *req.TokenID > 0 {
		tokenID = *req.TokenID
	} else if req.Symbol != "" || req.TokenAddress != "" {
		sym := strings.TrimSpace(req.Symbol)
		addr := strings.TrimSpace(req.TokenAddress)
		if sym != "" {
			err = s.db.QueryRow(c.Request.Context(), `SELECT id FROM tokens WHERE symbol = $1`, sym).Scan(&tokenID)
		} else {
			err = s.db.QueryRow(c.Request.Context(), `SELECT t.id FROM tokens t JOIN contracts c ON c.id = t.contract_id WHERE c.address = $1`, addr).Scan(&tokenID)
		}
		if err != nil {
			c.JSON(http.StatusNotFound, APIResponse{Success: false, Error: "Токен не найден", Code: http.StatusNotFound})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите token_id, symbol или token_address", Code: http.StatusBadRequest})
		return
	}
	cmd, err := s.db.Exec(c.Request.Context(), `UPDATE public.tokens SET logo_url = $1 WHERE id = $2`, logoURL, tokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error(), Code: http.StatusInternalServerError})
		return
	}
	if cmd.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, APIResponse{Success: false, Error: "Токен не найден", Code: http.StatusNotFound})
		return
	}
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"logo_url": logoURL}})
}

func (s *Server) setupRoutes() {
	// Раздача загруженных файлов (логотипы токенов и коинов)
	s.router.Static("/uploads", "uploads")

	api := s.router.Group("/api/v1")
	// Дублируем раздачу uploads под /api/v1/uploads — чтобы работало за прокси, который проксирует только /api/v1
	api.Static("/uploads", "uploads")

	// Метрики
	api.GET("/metrics", s.GetMetrics)
	api.GET("/metrics/transactions", s.GetTransactionMetrics)
	api.GET("/metrics/fees", s.GetFeeMetrics)
	api.GET("/fees", s.GetFeeMetrics) // алиас для удобства
	api.GET("/alerts", s.GetAlerts)
	api.POST("/alerts/thresholds", s.SetAlertThresholds)

	// Здоровье
	api.GET("/health", s.HealthCheck)

	// Кошельки
	api.POST("/wallet", s.CreateWallet)
	api.GET("/wallet/:address/balance", s.GetBalance)

	// Транзакции и мемпул
	api.GET("/transaction", s.GetTransactionHelp)  // GET без хеша — подсказка (иначе 404)
	api.GET("/transaction/", s.GetTransactionHelp) // то же при запросе с завершающим слэшем
	api.POST("/transaction", s.SendTransaction)
	api.GET("/transaction/:hash", s.GetTransaction)
	api.GET("/transactions", s.GetTransactionsList)        // список ожидающих (как /mempool)
	api.GET("/transactions/list", s.GetTransactionsFromDB) // список из gnd_db.transactions (для админки)
	api.GET("/mempool", s.GetMempool)

	// Блоки
	api.GET("/block/latest", s.GetLatestBlock)
	api.GET("/block/:number", s.GetBlockByNumber)

	// Контракты
	api.POST("/contract", s.DeployContract)
	api.POST("/contract/compile", s.CompileContract)
	api.POST("/contract/analyze", s.AnalyzeContract)
	api.GET("/contract/:address", s.GetContract)
	// Состояние контракта (функции/геттеры: name, symbol, total_supply, balances). Query: addresses=addr1,addr2
	api.GET("/contract/:address/state", s.GetContractState)
	// Просмотр контракта: ABI, список view/write функций, базовая инфо. Чтение: POST /contract/:address/call. Запись: POST /contract/:address/send
	api.GET("/contract/:address/view", s.GetContractView)
	api.POST("/contract/:address/call", s.ContractCall)
	api.POST("/contract/:address/send", s.ContractSend)

	// Токены (создание — по API-ключу; операции — без ключа в текущей реализации)
	api.POST("/token/deploy", s.DeployToken)
	api.POST("/token/logo/upload", s.TokenLogoUpload)
	api.PATCH("/token/logo", s.TokenLogoSet)
	// Нативные монеты (GND, GANI): баланс по символу и предложение (total_supply, circulating_supply)
	api.GET("/coin/:symbol/balance/:owner", s.GetNativeCoinBalance)
	api.GET("/coin/:symbol/supply", s.GetNativeCoinSupply)
	// Токены (amount — строка или число). Для нативных монет: symbol=GND|GANI, token_address пустой.
	api.POST("/token/transfer", func(c *gin.Context) {
		var req struct {
			TokenAddress string `json:"token_address"`
			Symbol       string `json:"symbol"`
			From         string `json:"from"`
			To           string `json:"to"`
			Amount       string `json:"amount"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "Неверный формат данных",
				Code:    http.StatusBadRequest,
			})
			return
		}
		amount := new(big.Int)
		if _, ok := amount.SetString(strings.TrimSpace(req.Amount), 10); !ok {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "Некорректная сумма (amount)",
				Code:    http.StatusBadRequest,
			})
			return
		}
		symbol := strings.TrimSpace(req.Symbol)
		// Перевод нативной монеты (GND, GANI) через state
		if core.IsNativeSymbol(symbol) && (req.TokenAddress == "" || strings.TrimSpace(req.TokenAddress) == "") {
			st := core.GetState()
			if st == nil {
				c.JSON(http.StatusInternalServerError, APIResponse{
					Success: false,
					Error:   "Состояние ноды недоступно",
					Code:    http.StatusInternalServerError,
				})
				return
			}
			err := st.TransferToken(types.Address(req.From), types.Address(req.To), symbol, amount)
			if err != nil {
				c.JSON(http.StatusBadRequest, APIResponse{
					Success: false,
					Error:   "Ошибка перевода нативной монеты: " + err.Error(),
					Code:    http.StatusBadRequest,
				})
				return
			}
			if err := st.SaveToDB(0); err != nil {
				c.JSON(http.StatusInternalServerError, APIResponse{
					Success: false,
					Error:   "Ошибка сохранения состояния",
					Code:    http.StatusInternalServerError,
				})
				return
			}
			c.JSON(http.StatusOK, APIResponse{
				Success: true,
				Data:    "Перевод нативной монеты выполнен успешно",
			})
			return
		}
		// Контрактный токен по адресу
		token, err := registry.GetToken(req.TokenAddress)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Error:   "Ошибка получения токена: " + err.Error(),
				Code:    http.StatusInternalServerError,
			})
			return
		}
		if gnd, ok := token.(interfaces.TokenInterface); ok {
			err := gnd.Transfer(c.Request.Context(), req.From, req.To, amount)
			if err != nil {
				c.JSON(http.StatusInternalServerError, APIResponse{
					Success: false,
					Error:   "Ошибка перевода токенов: " + err.Error(),
					Code:    http.StatusInternalServerError,
				})
				return
			}
			c.JSON(http.StatusOK, APIResponse{
				Success: true,
				Data:    "Перевод выполнен успешно",
			})
			return
		}
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Неподдерживаемый стандарт токена",
			Code:    http.StatusBadRequest,
		})
	})

	api.POST("/token/approve", func(c *gin.Context) {
		var req struct {
			TokenAddress string   `json:"token_address"`
			Owner        string   `json:"owner"`
			Spender      string   `json:"spender"`
			Amount       *big.Int `json:"amount"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "Неверный формат данных",
				Code:    http.StatusBadRequest,
			})
			return
		}
		token, err := registry.GetToken(req.TokenAddress)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Error:   "Ошибка получения токена: " + err.Error(),
				Code:    http.StatusInternalServerError,
			})
			return
		}
		if gnd, ok := token.(interfaces.TokenInterface); ok {
			err := gnd.Approve(c.Request.Context(), req.Owner, req.Spender, req.Amount)
			if err != nil {
				c.JSON(http.StatusInternalServerError, APIResponse{
					Success: false,
					Error:   "Ошибка разрешения перевода: " + err.Error(),
					Code:    http.StatusInternalServerError,
				})
				return
			}
			c.JSON(http.StatusOK, APIResponse{
				Success: true,
				Data:    "Разрешение выдано успешно",
			})
			return
		}
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Неподдерживаемый стандарт токена",
			Code:    http.StatusBadRequest,
		})
	})

	api.GET("/token/:address/balance/:owner", func(c *gin.Context) {
		tokenAddress := c.Param("address")
		owner := c.Param("owner")

		token, err := registry.GetToken(tokenAddress)
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Error:   "Ошибка получения токена: " + err.Error(),
				Code:    http.StatusInternalServerError,
			})
			return
		}
		if gnd, ok := token.(interfaces.TokenInterface); ok {
			balance, err := gnd.GetBalance(c.Request.Context(), owner)
			if err != nil {
				c.JSON(http.StatusInternalServerError, APIResponse{
					Success: false,
					Error:   "Ошибка получения баланса: " + err.Error(),
					Code:    http.StatusInternalServerError,
				})
				return
			}
			c.JSON(http.StatusOK, APIResponse{
				Success: true,
				Data: gin.H{
					"address": tokenAddress,
					"owner":   owner,
					"balance": balance.String(),
				},
			})
			return
		}
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Неподдерживаемый стандарт токена",
			Code:    http.StatusBadRequest,
		})
	})

	// Админские маршруты (защита: X-Admin-Token = GND_ADMIN_SECRET)
	admin := api.Group("/admin")
	{
		admin.POST("/keys", s.AdminCreateKey)
		admin.GET("/keys", s.AdminListKeys)
		admin.POST("/keys/:id/revoke", s.AdminRevokeKey)
		admin.DELETE("/keys/:id", s.AdminRevokeKey)
		admin.GET("/wallets", s.AdminListWallets)
		admin.PATCH("/wallets/:address", s.AdminUpdateWallet)
		admin.POST("/wallets/:address/disable", s.AdminDisableWallet)
		admin.POST("/wallets/:address/enable", s.AdminEnableWallet)
		admin.DELETE("/wallets/:address", s.AdminDeleteWallet)
		admin.POST("/wallets/:address/delete", s.AdminDeleteWallet)
		admin.POST("/record-transaction", s.AdminRecordTransaction)
		// Контракты: запись транзакций блокировки/удаления (для GND_admin)
		admin.POST("/contracts/:address/disable", s.AdminContractDisable)
		admin.POST("/contracts/:address/delete", s.AdminContractDelete)
		admin.PATCH("/contracts/:address/abi", s.AdminUpdateContractABI)
		// Чтение/запись методов контракта по id (страница /admin/contracts/:id). Путь by-id избегает конфликта с :address.
		admin.POST("/contracts/by-id/:id/call", s.AdminContractCall)
		admin.POST("/contracts/by-id/:id/send", s.AdminContractSend)
		// Токены: запись транзакций блокировки/удаления (для GND_admin)
		admin.POST("/tokens/:id/disable", s.AdminTokenDisable)
		admin.POST("/tokens/:id/delete", s.AdminTokenDelete)
		// Состояния контрактов: запись слота storage (для GND_admin)
		admin.POST("/state/contract/:address/storage", s.AdminWriteContractStorageSlot)
	}

	// Состояния аккаунтов и контрактов (чтение) — для GND_admin и клиентов
	api.GET("/state/account/:address", s.GetAccountStateCurrent)
	api.GET("/state/account/:address/block/:blockId", s.GetAccountStateAtBlock)
	api.GET("/state/contract/:address/storage", s.GetContractStorage)
}

// GetAccountStateCurrent возвращает текущее состояние аккаунта из accounts. GET /api/v1/state/account/:address
func (s *Server) GetAccountStateCurrent(c *gin.Context) {
	address := strings.TrimSpace(c.Param("address"))
	if address == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите address", Code: http.StatusBadRequest})
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
	st, err := core.GetCurrentAccountState(c.Request.Context(), pool, address)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{Success: false, Error: err.Error(), Code: http.StatusNotFound})
		return
	}
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: st})
}

// GetAccountStateAtBlock возвращает снимок состояния аккаунта на блок. GET /api/v1/state/account/:address/block/:blockId
func (s *Server) GetAccountStateAtBlock(c *gin.Context) {
	address := strings.TrimSpace(c.Param("address"))
	blockIDStr := c.Param("blockId")
	if address == "" || blockIDStr == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите address и blockId", Code: http.StatusBadRequest})
		return
	}
	blockID, err := strconv.ParseInt(blockIDStr, 10, 64)
	if err != nil || blockID < 0 {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Некорректный blockId", Code: http.StatusBadRequest})
		return
	}
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, APIResponse{Success: false, Error: "БД недоступна", Code: http.StatusServiceUnavailable})
		return
	}
	st, err := core.GetAccountStateAtBlock(c.Request.Context(), s.db, address, blockID)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{Success: false, Error: err.Error(), Code: http.StatusNotFound})
		return
	}
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: st})
}

// GetContractStorage возвращает слоты storage контракта на блок. GET /api/v1/state/contract/:address/storage?block_id=123
func (s *Server) GetContractStorage(c *gin.Context) {
	address := strings.TrimSpace(c.Param("address"))
	if address == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите address контракта", Code: http.StatusBadRequest})
		return
	}
	blockIDStr := c.Query("block_id")
	if blockIDStr == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите block_id в query", Code: http.StatusBadRequest})
		return
	}
	blockID, err := strconv.ParseInt(blockIDStr, 10, 64)
	if err != nil || blockID < 0 {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Некорректный block_id", Code: http.StatusBadRequest})
		return
	}
	if s.db == nil {
		c.JSON(http.StatusServiceUnavailable, APIResponse{Success: false, Error: "БД недоступна", Code: http.StatusServiceUnavailable})
		return
	}
	slots, err := core.GetContractStorageAtBlock(c.Request.Context(), s.db, address, blockID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error(), Code: http.StatusInternalServerError})
		return
	}
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"address": address, "block_id": blockID, "slots": slots}})
}

// AdminWriteContractStorageSlot записывает слот storage контракта. POST /api/v1/admin/state/contract/:address/storage
// Body: {"block_id": 1, "slot_key": "0x...", "slot_value": "0x..."} ИЛИ {"block_id": 1, "slot_index": 0, "slot_value": "0x..."}
// slot_index (0, 1, 2...) — для NativeTokensController: 0=gndToken, 1=ganiToken.
func (s *Server) AdminWriteContractStorageSlot(c *gin.Context) {
	address := strings.TrimSpace(c.Param("address"))
	if address == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите address контракта", Code: http.StatusBadRequest})
		return
	}
	var req struct {
		BlockID   int64  `json:"block_id"`
		SlotKey   string `json:"slot_key"`
		SlotIndex *int64 `json:"slot_index"` // опционально: 0=gndToken, 1=ganiToken
		SlotValue string `json:"slot_value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Ожидается JSON: block_id, slot_key или slot_index, slot_value", Code: http.StatusBadRequest})
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
	slotKey := req.SlotKey
	if req.SlotIndex != nil {
		slotKey = "0x" + hex.EncodeToString(core.SlotKeyFromIndex(uint64(*req.SlotIndex)))
	}
	if slotKey == "" {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Укажите slot_key или slot_index", Code: http.StatusBadRequest})
		return
	}
	if err := core.WriteContractStorageSlot(c.Request.Context(), pool, req.BlockID, address, slotKey, req.SlotValue); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: err.Error(), Code: http.StatusBadRequest})
		return
	}
	// Все действия с контрактами формируют транзакции в блокчейне
	genesisID := int64(0)
	if s.core != nil && s.core.Genesis != nil {
		genesisID = s.core.Genesis.ID
	}
	payload := slotKey
	if payload == "" {
		payload = "storage_slot"
	}
	if pool != nil {
		if errTx := core.RecordAdminTransaction(c.Request.Context(), pool, genesisID, "contract_storage_write", "GND_ADMIN", address, payload); errTx != nil {
			log.Printf("[REST] запись транзакции contract_storage_write в gnd_db.transactions: %v", errTx)
		}
	}
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: "Слот записан"})
}

// GetTransactionHelp возвращает подсказку при GET /transaction без хеша (избегаем 404)
func (s *Server) GetTransactionHelp(c *gin.Context) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Error:   "Укажите хеш транзакции: GET /api/v1/transaction/:hash. Отправка: POST /api/v1/transaction. Список ожидающих: GET /api/v1/transactions или GET /api/v1/mempool",
		Code:    http.StatusBadRequest,
	})
}

// GetTransactionsList возвращает список ожидающих транзакций (то же, что /mempool)
func (s *Server) GetTransactionsList(c *gin.Context) {
	s.GetMempool(c)
}

// GetTransactionsFromDB возвращает список транзакций из gnd_db.transactions (для админки). GET /api/v1/transactions/list?limit=50&offset=0
func (s *Server) GetTransactionsFromDB(c *gin.Context) {
	pool := s.db
	if s.core != nil && s.core.Pool != nil {
		pool = s.core.Pool
	}
	if pool == nil {
		c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"list": []interface{}{}, "total": 0}})
		return
	}
	limit := 50
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	offset := 0
	if o := c.Query("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}
	ctx := c.Request.Context()
	var total int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM transactions`).Scan(&total)
	rows, err := pool.Query(ctx, `
		SELECT id, block_id, hash, sender, recipient, value, fee, nonce, type, status, timestamp
		FROM transactions
		ORDER BY timestamp DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error(), Code: http.StatusInternalServerError})
		return
	}
	defer rows.Close()
	var list []gin.H
	for rows.Next() {
		var id int
		var blockIDNull sql.NullInt64
		var hash, sender, recipient, valueStr, feeStr, txType, status string
		var nonce int64
		var ts time.Time
		if err := rows.Scan(&id, &blockIDNull, &hash, &sender, &recipient, &valueStr, &feeStr, &nonce, &txType, &status, &ts); err != nil {
			continue
		}
		blockID := 0
		if blockIDNull.Valid {
			blockID = int(blockIDNull.Int64)
		}
		item := gin.H{
			"id":        id,
			"block_id":  blockID,
			"hash":      hash,
			"sender":    sender,
			"recipient": recipient,
			"value":     valueStr,
			"fee":       feeStr,
			"nonce":     nonce,
			"type":      txType,
			"status":    status,
			"timestamp": ts.Format(time.RFC3339),
		}
		// block_number — номер в цепи (для сканера: GET /block/:number ищет по index, не по id)
		if blockID > 0 {
			if blockNum, err := core.GetBlockIndexByID(ctx, pool, int64(blockID)); err == nil {
				item["block_number"] = blockNum
			}
		}
		list = append(list, item)
	}
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"list": list, "total": total}})
}

// GetMempool возвращает размер мемпула и список хешей ожидающих транзакций (для проверки работы mempool)
func (s *Server) GetMempool(c *gin.Context) {
	if s.mempool == nil {
		c.JSON(http.StatusOK, APIResponse{Success: true, Data: gin.H{"size": 0, "pending_hashes": []string{}}})
		return
	}
	size := s.mempool.Size()
	pending := s.mempool.GetPendingTransactions()
	hashes := make([]string, 0, len(pending))
	for _, tx := range pending {
		hashes = append(hashes, tx.Hash)
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    gin.H{"size": size, "pending_hashes": hashes},
	})
}

// StartRESTServer запускает REST API сервер. evmInstance используется для деплоя токенов (POST /token/deploy); если nil — эндпоинт возвращает 503.
// signerCreator — опционально: при наличии новые кошельки создаются через signing_service (ключ в signer_wallets).
// Если signerCreator реализует AdminSigner, запросы из админки с пустой подписью подписываются нодой для кошельков с signer_wallet_id.
func StartRESTServer(bc *core.Blockchain, mp *core.Mempool, cfg *core.Config, pool *pgxpool.Pool, evmInstance *vm.EVM, signerCreator core.SignerWalletCreator) {
	if signerCreator != nil {
		bc.SignerCreator = signerCreator
	}
	var adminSigner AdminSigner
	if signerCreator != nil {
		if s, ok := signerCreator.(AdminSigner); ok {
			adminSigner = s
		}
	}
	// Инициализируем метрики блоков из текущей цепи (LastBlockTime, TotalBlocks, AverageBlockTime и т.д.)
	if latest, err := bc.GetLatestBlock(); err == nil {
		var prev *core.Block
		chainHeight := latest.Height
		if chainHeight == 0 && latest.Index > 0 {
			chainHeight = latest.Index
		}
		if pool != nil && chainHeight >= 1 {
			prev, _ = core.GetBlockByNumber(pool, chainHeight-1)
		}
		core.InitBlockMetricsFromBlock(latest, prev)
	}
	// Инициализируем метрики транзакций из БД (TotalTransactions, TypeMetrics, StatusMetrics, комиссии)
	pendingCount := 0
	if mp != nil {
		pendingCount = mp.Size()
	}
	core.InitTransactionMetricsFromDB(context.Background(), pool, pendingCount)
	var tokenDeployer *deployer.Deployer
	if evmInstance != nil && pool != nil {
		tokenDeployer = deployer.NewDeployer(pool, &noopEventManager{}, newEVMAdapter(evmInstance))
	}
	server := NewServer(pool, bc, mp, tokenDeployer, cfg)
	server.evm = evmInstance
	if adminSigner != nil {
		server.adminSigner = adminSigner
	}
	addr := fmt.Sprintf(":%d", cfg.Server.REST.Port)
	log.Printf("=== REST API Server запущен на %s ===\nДоступен по /api/v1/* (health, wallet, transaction, block, contract, token/deploy, token/transfer и др.)", addr)
	if err := server.Start(addr); err != nil {
		log.Printf("[REST] Ошибка запуска: %v. RPC и WebSocket продолжают работать. Освободите порт (см. docs/deployment-server.md) и перезапустите ноду для включения REST.", err)
	}
}

// Универсальный обработчик токенов
func universalHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TokenAddr string        `json:"tokenAddr"`
		Method    string        `json:"method"`
		Args      []interface{} `json:"args"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	token, err := registry.GetToken(req.TokenAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if gnd, ok := token.(interfaces.TokenInterface); ok {
		switch req.Method {
		case "transfer":
			if len(req.Args) != 3 {
				http.Error(w, "transfer требует from, to, amount", http.StatusBadRequest)
				return
			}
			from, to := req.Args[0].(string), req.Args[1].(string)
			amount := req.Args[2].(*big.Int)
			err := gnd.Transfer(r.Context(), from, to, amount)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "approve":
			if len(req.Args) != 3 {
				http.Error(w, "approve требует owner, spender, amount", http.StatusBadRequest)
				return
			}
			owner, spender := req.Args[0].(string), req.Args[1].(string)
			amount := req.Args[2].(*big.Int)
			err := gnd.Approve(r.Context(), owner, spender, amount)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		case "balanceOf":
			if len(req.Args) != 1 {
				http.Error(w, "balanceOf требует address", http.StatusBadRequest)
				return
			}
			addr := req.Args[0].(string)
			balance, err := gnd.GetBalance(r.Context(), addr)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"balance": balance.String()})
		default:
			http.Error(w, "неподдерживаемый метод", http.StatusBadRequest)
		}
	} else {
		http.Error(w, "токен не реализует интерфейс TokenInterface", http.StatusInternalServerError)
	}
}

// handleTransfer обрабатывает перевод токенов
func handleTransfer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TokenAddress string   `json:"token_address"`
		From         string   `json:"from"`
		To           string   `json:"to"`
		Amount       *big.Int `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	token, err := registry.GetToken(req.TokenAddress)
	if err != nil {
		sendError(w, "Ошибка получения токена: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if gnd, ok := token.(interfaces.TokenInterface); ok {
		err := gnd.Transfer(r.Context(), req.From, req.To, req.Amount)
		if err != nil {
			sendError(w, "Ошибка перевода токенов: "+err.Error(), http.StatusInternalServerError)
			return
		}
		sendJSON(w, APIResponse{
			Success: true,
			Data:    "Перевод выполнен успешно",
		}, http.StatusOK)
		return
	}
	sendError(w, "Неподдерживаемый стандарт токена", http.StatusBadRequest)
}

// handleTokenApprove обрабатывает разрешение на перевод токенов
func handleTokenApprove(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TokenAddress string   `json:"token_address"`
		Owner        string   `json:"owner"`
		Spender      string   `json:"spender"`
		Amount       *big.Int `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	token, err := registry.GetToken(req.TokenAddress)
	if err != nil {
		sendError(w, "Ошибка получения токена: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if gnd, ok := token.(interfaces.TokenInterface); ok {
		err := gnd.Approve(r.Context(), req.Owner, req.Spender, req.Amount)
		if err != nil {
			sendError(w, "Ошибка разрешения перевода: "+err.Error(), http.StatusInternalServerError)
			return
		}
		sendJSON(w, APIResponse{
			Success: true,
			Data:    "Разрешение выдано успешно",
		}, http.StatusOK)
		return
	}
	sendError(w, "Неподдерживаемый стандарт токена", http.StatusBadRequest)
}

// handleGetTokenBalance обрабатывает запрос баланса токена
func handleGetTokenBalance(w http.ResponseWriter, r *http.Request) {
	tokenAddress := mux.Vars(r)["address"]
	owner := mux.Vars(r)["owner"]

	token, err := registry.GetToken(tokenAddress)
	if err != nil {
		sendError(w, "Ошибка получения токена: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if gnd, ok := token.(interfaces.TokenInterface); ok {
		balance, err := gnd.GetBalance(r.Context(), owner)
		if err != nil {
			sendError(w, "Ошибка получения баланса: "+err.Error(), http.StatusInternalServerError)
			return
		}
		sendJSON(w, APIResponse{
			Success: true,
			Data: gin.H{
				"address": tokenAddress,
				"owner":   owner,
				"balance": balance.String(),
			},
		}, http.StatusOK)
		return
	}
	sendError(w, "Неподдерживаемый стандарт токена", http.StatusBadRequest)
}
