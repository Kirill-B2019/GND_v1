// tokens/registry.go

package tokens

var Tokens = map[string]*TokenInfo{} // Глобальный реестр токенов

func RegisterToken(addr string, token *GNDst1) {
	Tokens[addr] = &TokenInfo{
		Name:        token.name,
		Symbol:      token.symbol,
		Decimals:    token.decimals,
		TotalSupply: token.totalSupply,
		Address:     addr,
	}
}

func GetAllTokens() []*TokenInfo {
	var list []*TokenInfo
	for _, token := range Tokens {
		list = append(list, token)
	}
	return list
}
