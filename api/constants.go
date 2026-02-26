package api

import "time"

const (
	// API URLs (основной узел ГАНИМЕД)
	RestURL = "http://31.128.41.155:8182/api/v1"
	RpcURL  = "http://31.128.41.155:8181"
	WsURL   = "ws://31.128.41.155:8183/ws"

	// Домены: документация vs нода для подключения
	ApiDocHost = "api.gnd-net.com"       // Только описание/документация API
	NodeHost   = "main-node.gnd-net.com" // Нода, к которой идёт подключение (REST/RPC/WS)
	// RestURLSecure = "https://main-node.gnd-net.com/api/v1"

	// Стандарт токенов/контрактов ГАНИМЕД
	TokenStandardGNDst1 = "GND-st1"

	// API Key (для тестов)
	ApiKey = "test_api_key"

	// Timeouts
	HttpTimeout = 5 * time.Second
	WsTimeout   = 5 * time.Second
)
