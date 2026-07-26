package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashKey calcula el hash SHA-256 de una clave de autenticación antes de
// almacenarla en la base de datos. Esto garantiza que, en caso de compromiso
// de la DB, los atacantes obtengan hashes unidireccionales en lugar de las
// claves de firma JWT raw.
//
// El flujo es: raw_key → SHA-256 → hash (almacenado en DB).
// El JWT se firma con el hash, no con el raw key.
func HashKey(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
