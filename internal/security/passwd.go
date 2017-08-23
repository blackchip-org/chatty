package security

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"

	"golang.org/x/crypto/pbkdf2"
)

func RandomPassword() (string, error) {
	b := make([]byte, 33)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func Salt() ([]byte, error) {
	salt := make([]byte, 64)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func EncodePassword(pass []byte, salt []byte) []byte {
	const iters = 10000
	const keylen = 64
	encrypted := pbkdf2.Key(pass, salt, iters, keylen, sha512.New)
	return encrypted
}
