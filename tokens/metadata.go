// | KB @CerbeRus - Nexus Invest Team
// tokens/metadata.go

package tokens

type TokenMetadata struct {
	Name        string
	Symbol      string
	Decimals    uint8
	Description string
	Standard    string
	Compiler    string
	Version     string
}

func NewMetadata(name, symbol string, decimals uint8, standard string) TokenMetadata {
	return TokenMetadata{
		Name:     name,
		Symbol:   symbol,
		Decimals: decimals,
		Standard: standard,
	}
}
func (tm TokenMetadata) ToJSON() string {
	return `{"name":"` + tm.Name + `","symbol":"` + tm.Symbol + `"}`
}
