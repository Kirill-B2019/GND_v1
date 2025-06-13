package api

import "time"

const (
	// API URLs
	RestURL = "http://45.12.72.15:8182/api"
	RpcURL  = "http://45.12.72.15:8181"
	WsURL   = "ws://45.12.72.15:8183/ws"

	// API Key
	ApiKey = "test_api_key"

	// Timeouts
	HttpTimeout = 5 * time.Second
	WsTimeout   = 5 * time.Second
)
