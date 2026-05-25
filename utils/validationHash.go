package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

//Spočítá kryptografický hash obsahu souboru
func CalculateHash(content []byte) string {
	hasher := sha256.New()
	hasher.Write(content)
	return hex.EncodeToString(hasher.Sum(nil))
}