package api

import "time"

const (
	// API URLs (основной узел ГАНИМЕД)
	RestURL = "http://31.128.41.155:8182/api/v1"
	RpcURL  = "http://31.128.41.155:8181"
	WsURL   = "ws://31.128.41.155:8183/ws"

	// Домены (для документации и конфигов)
	ApiHost  = "api.gnd-net.com"       // Публичный API (документация, клиенты)
	RestHost = "main-node.gnd-net.com" // Узел (при необходимости разделения)
	// RestURLSecure = "https://api.gnd-net.com/api/v1"

	// Стандарт токенов/контрактов ГАНИМЕД
	TokenStandardGNDst1 = "GND-st1"

	// API Key (для тестов)
	ApiKey = "test_api_key"

	// Timeouts
	HttpTimeout = 5 * time.Second
	WsTimeout   = 5 * time.Second
)
