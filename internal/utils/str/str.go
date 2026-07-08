package str

import (
	"crypto/rand"
	"math/big"
	"net/http"
)

func ExtractInviteCode(r *http.Request) string {
	return r.PathValue("invite_code")
}

func GenerateInviteCode() (string, error) {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		result[i] = chars[num.Int64()]
	}

	return string(result), nil
}
