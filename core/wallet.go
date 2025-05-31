package core

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"github.com/btcsuite/btcutil/base58"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"golang.org/x/crypto/ripemd160"
)

// Допустимые префиксы
var validPrefixes = []string{"GND", "GND_", "GN", "GN_"}

type Wallet struct {
	PrivateKey *secp256k1.PrivateKey
	Address    string
}

func NewWallet() (*Wallet, error) {
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, err
	}
	pubKey := privKey.PubKey().SerializeUncompressed()
	shaHash := sha256.Sum256(pubKey[1:])
	ripemdHasher := ripemd160.New()
	_, err = ripemdHasher.Write(shaHash[:])
	if err != nil {
		return nil, err
	}
	pubKeyHash := ripemdHasher.Sum(nil)
	prefix := randomPrefix()
	payload := append(prefix, pubKeyHash...)
	checksum := checksum(payload)
	fullPayload := append(payload, checksum...)
	address := base58.Encode(fullPayload)
	return &Wallet{
		PrivateKey: privKey,
		Address:    address,
	}, nil
}

func randomPrefix() []byte {
	b := make([]byte, 1)
	_, err := rand.Read(b)
	if err != nil {
		return []byte(validPrefixes[0])
	}
	return []byte(validPrefixes[int(b[0])%len(validPrefixes)])
}

func checksum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	return second[:4]
}

func ValidateAddress(address string) bool {
	decoded := base58.Decode(address)
	if len(decoded) < 8 {
		return false
	}
	var matchedPrefix []byte
	for _, prefix := range validPrefixes {
		if len(decoded) > len(prefix) && string(decoded[:len(prefix)]) == string(prefix) {
			matchedPrefix = prefix
			break
		}
	}
	if matchedPrefix == nil {
		return false
	}
	payload := decoded[:len(decoded)-4]
	checksumBytes := decoded[len(decoded)-4:]
	calcChecksum := checksum(payload)
	for i := 0; i < 4; i++ {
		if checksumBytes[i] != calcChecksum[i] {
			return false
		}
	}
	return true
}

func (w *Wallet) PrivateKeyHex() string {
	return hex.EncodeToString(w.PrivateKey.Serialize())
}

func (w *Wallet) PublicKeyHex() string {
	return hex.EncodeToString(w.PrivateKey.PubKey().SerializeUncompressed())
}
