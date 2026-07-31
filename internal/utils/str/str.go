package str

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
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

func ExtractURLFromIframeTag(s string) (string, error) {
	u, err := url.ParseRequestURI(s)
	if err == nil && u.Scheme != "" && u.Host != "" {

		return s, nil
	}

	inQuote := false
	extractedURL := ""

	for _, r := range s {
		if r == '"' {
			if !inQuote {
				inQuote = true

				continue
			}

			break
		}

		if inQuote {
			extractedURL += string(r)
		}
	}

	u, err = url.ParseRequestURI(extractedURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid string that should contain a url")
	}

	return extractedURL, nil
}
