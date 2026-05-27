package service

import "crypto/rand"

func generateID() (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const idLen = 8

	b := make([]byte, idLen)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	out := make([]byte, idLen)

	for i := range b {
		out[i] = alphabet[int(b[i])%len(alphabet)]
	}

	return string(out), nil
}
