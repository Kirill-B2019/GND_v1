// api/rest.go

package api

import (
	"GND/core"
	"GND/tokens/interfaces"
	"GND/tokens/registry"
	"GND/types"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"math/big"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	blockchain *core.Blockchain
	mempool    *core.Mempool
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
	router *gin.Engine
	db     *pgxpool.Pool
	core   *core.Blockchain
}

// NewServer создает новый экземпляр сервера
func NewServer(db *pgxpool.Pool, blockchain *core.Blockchain) *Server {
	server := &Server{
		router: gin.Default(),
		db:     db,
		core:   blockchain,
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

// HealthCheck возвращает статус сервера
func (s *Server) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: gin.H{
			"status":    "ok",
			"version":   "1.0.0",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	})
}

// CreateWallet создает новый кошелек
func (s *Server) CreateWallet(c *gin.Context) {
	wallet, err := s.core.CreateWallet()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Ошибка создания кошелька: " + err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    wallet,
	})
}

// GetBalance возвращает баланс кошелька
func (s *Server) GetBalance(c *gin.Context) {
	address := c.Param("address")
	balance := s.core.GetBalance(address, "GND") // Using GND as default symbol
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: gin.H{
			"address": address,
			"balance": balance.String(),
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

	tx := &core.Transaction{
		Sender:    fromAddr,
		Recipient: toAddr,
		Value:     txData.Value,
		Nonce:     int64(txData.Nonce),
		Data:      []byte(txData.Data),
		Signature: []byte(txData.Signature),
		GasLimit:  21000, // Стандартный лимит газа для простой транзакции
		GasPrice:  txData.Fee,
	}
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

// GetLatestBlock возвращает последний блок
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

func (s *Server) setupRoutes() {
	api := s.router.Group("/api/v1")

	// Метрики
	api.GET("/metrics", s.GetMetrics)
	api.GET("/metrics/transactions", s.GetTransactionMetrics)
	api.GET("/metrics/fees", s.GetFeeMetrics)
	api.GET("/alerts", s.GetAlerts)
	api.POST("/alerts/thresholds", s.SetAlertThresholds)

	// Здоровье
	api.GET("/health", s.HealthCheck)

	// Кошельки
	api.POST("/wallet", s.CreateWallet)
	api.GET("/wallet/:address/balance", s.GetBalance)

	// Транзакции
	api.POST("/transaction", s.SendTransaction)
	api.GET("/transaction/:hash", s.GetTransaction)

	// Блоки
	api.GET("/block/latest", s.GetLatestBlock)
	api.GET("/block/:number", s.GetBlockByNumber)

	// Контракты
	api.POST("/contract", s.DeployContract)
	api.GET("/contract/:address", s.GetContract)

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
}

// StartRESTServer запускает REST API сервер
func StartRESTServer(bc *core.Blockchain, mp *core.Mempool, cfg *core.Config, pool *pgxpool.Pool) {
	server := NewServer(pool, bc)
	if err := server.Start(fmt.Sprintf(":%d", cfg.Server.REST.Port)); err != nil {
		log.Fatalf("Ошибка запуска REST сервера: %v", err)
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
