package servicer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

var verificationKey = []byte{142, 80, 107, 188, 92, 20, 197, 218, 205, 136, 179, 124, 29, 252, 213, 190}

func VerificationToken(id string) string {
	return base64.StdEncoding.EncodeToString(verificationToken(id))
}

func VerifyToken(id, token string) bool {
	tok, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return false
	}

	return hmac.Equal(verificationToken(id), tok)
}

func verificationToken(id string) []byte {
	mac := hmac.New(sha256.New, verificationKey)
	mac.Write([]byte(id))
	return mac.Sum(nil)
}
