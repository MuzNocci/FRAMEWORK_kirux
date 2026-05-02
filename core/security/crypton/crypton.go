package crypton

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidSignature = errors.New("invalid signature")

func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func Sign(payload, secret string) (string, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	encoded := base64.RawURLEncoding.EncodeToString([]byte(payload))
	return encoded + "." + sig, nil
}

func Verify(token, secret string) (string, error) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return "", ErrInvalidSignature
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", ErrInvalidSignature
	}

	expected, err := Sign(string(payload), secret)
	if err != nil || expected != token {
		return "", ErrInvalidSignature
	}

	return string(payload), nil
}

func RandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}
