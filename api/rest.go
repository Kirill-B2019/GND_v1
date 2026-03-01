// | KB @CerbeRus - Nexus Invest Team
// api/rest.go

package api

import (
	"GND/audit"
	"GND/core"
	"GND/tokens/deployer"
	"GND/tokens/interfaces"
	"GND/tokens/registry"
	tokentypes "GND/tokens/types"
	"GND/types"
	"GND/vm"
	"GND/vm/compiler"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"math/big"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

// Server представляет HTTP сервер
type Server struct {
	router   *gin.Engine
	db       *pgxpool.Pool
	core     *core.Blockchain
	mempool  *core.Mempool
	deployer *deployer.Deployer
	cfg      *core.Config
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
		"timestamp": time.Now().Format(time.RFC3339),
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

// CreateWallet создаёт новый кошелёк. Требуется заголовок X-API-Key.
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
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    wallet,
	})
}

// GetBalance возвращает все балансы токенов кошелька из token_balances с полями из tokens (standard, symbol, name, decimals, is_verified)
func (s *Server) GetBalance(c *gin.Context) {
	address := c.Param("address")
	balances := []core.WalletTokenBalance{}
	if s.core.Pool != nil {
		var err error
		balances, err = core.GetWalletTokenBalances(c.Request.Context(), s.core.Pool, address)
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

// SendTransaction отправляет транзакцию
func (s *Server) SendTransaction(c *gin.Context) {
	var txData struct {
		From      string   `json:"from"`
		To        string   `json:"to"`
		Value     *big.Int `json:"value"`
		Fee       *big.Int `json:"fee"`
		Nonce     uint64   `json:"nonce"`
		Type      string   `json:"type"`
		Data      string   `json:"data"`
		Signature string   `json:"signature"`
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

	tx := &core.Transaction{
		Sender:     fromAddr,
		Recipient:  toAddr,
		Value:      value,
		Nonce:      int64(txData.Nonce),
		Data:       []byte(txData.Data),
		Signature:  []byte(txData.Signature),
		GasLimit:   21000, // Стандартный лимит газа для простой транзакции
		GasPrice:   fee,
		Symbol:     "GND",
		IsVerified: true, // транзакции с нативной монетой GND верифицированы
		Timestamp:  time.Now().UTC(),
		Status:     "pending",
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

// GetTransaction возвращает информацию о транзакции
func (s *Server) GetTransaction(c *gin.Context) {
	hash := c.Param("hash")
	tx, err := s.core.GetTransaction(hash)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error:   "Транзакция не найдена: " + err.Error(),
			Code:    http.StatusNotFound,
		})
		return
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    tx,
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
		Name        string                 `json:"name"`
		Standard    string                 `json:"standard"`
		Owner       string                 `json:"owner"`
		Compiler    string                 `json:"compiler"`
		Version     string                 `json:"version"`
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
		Name:        paramsData.Name,
		Standard:    paramsData.Standard,
		Owner:       paramsData.Owner,
		Compiler:    paramsData.Compiler,
		Version:     paramsData.Version,
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
	address, err := s.core.DeployContract(&params)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Ошибка деплоя контракта: " + err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
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
	solc := compiler.DefaultSolidityCompiler{SolcPath: "solc"}
	metadata := compiler.ContractMetadata{
		Name:     req.Name,
		Standard: req.Standard,
		Compiler: "solc",
		Version:  "0.8.0",
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
		Name        string   `json:"name"`
		Symbol      string   `json:"symbol"`
		Decimals    uint8    `json:"decimals"`
		TotalSupply *big.Int `json:"total_supply"`
		Owner       string   `json:"owner"`
		Standard    string   `json:"standard"`
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

	ownerWalletCreated := false
	if strings.TrimSpace(req.Owner) == "" {
		if s.core == nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "Поле owner обязательно: сервис создания кошельков недоступен",
				Code:    http.StatusBadRequest,
			})
			return
		}
		wallet, err := s.core.CreateWallet(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Error:   "Ошибка создания кошелька для владельца токена: " + err.Error(),
				Code:    http.StatusInternalServerError,
			})
			return
		}
		req.Owner = string(wallet.Address)
		ownerWalletCreated = true
	}

	params := tokentypes.TokenParams{
		Name:        req.Name,
		Symbol:      req.Symbol,
		Decimals:    req.Decimals,
		TotalSupply: req.TotalSupply,
		Owner:       req.Owner,
		Standard:    req.Standard,
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
	data := gin.H{
		"address":              token.GetAddress(),
		"name":                 token.GetName(),
		"symbol":               token.GetSymbol(),
		"decimals":             token.GetDecimals(),
		"total_supply":         token.GetTotalSupply().String(),
		"standard":             token.GetStandard(),
		"owner":                req.Owner,
		"owner_wallet_created": ownerWalletCreated,
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

func (s *Server) setupRoutes() {
	api := s.router.Group("/api/v1")

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
	api.GET("/transactions", s.GetTransactionsList) // список ожидающих (как /mempool)
	api.GET("/mempool", s.GetMempool)

	// Блоки
	api.GET("/block/latest", s.GetLatestBlock)
	api.GET("/block/:number", s.GetBlockByNumber)

	// Контракты
	api.POST("/contract", s.DeployContract)
	api.POST("/contract/compile", s.CompileContract)
	api.POST("/contract/analyze", s.AnalyzeContract)
	api.GET("/contract/:address", s.GetContract)

	// Токены (создание — по API-ключу; операции — без ключа в текущей реализации)
	api.POST("/token/deploy", s.DeployToken)
	// Токены
	api.POST("/token/transfer", func(c *gin.Context) {
		var req struct {
			TokenAddress string   `json:"token_address"`
			From         string   `json:"from"`
			To           string   `json:"to"`
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
			err := gnd.Transfer(c.Request.Context(), req.From, req.To, req.Amount)
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
	}
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
func StartRESTServer(bc *core.Blockchain, mp *core.Mempool, cfg *core.Config, pool *pgxpool.Pool, evmInstance *vm.EVM, signerCreator core.SignerWalletCreator) {
	if signerCreator != nil {
		bc.SignerCreator = signerCreator
	}
	// Инициализируем метрики блоков из текущей цепи (LastBlockTime, TotalBlocks и т.д.)
	if latest, err := bc.GetLatestBlock(); err == nil {
		core.InitBlockMetricsFromBlock(latest)
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
