package auth

import (
	"errors"
	"kyrux/core/security/crypton"
	"time"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrTokenExpired = errors.New("token expired")

type Claims struct {
	UserID    string
	ExpiresAt time.Time
}

type Authenticator struct {
	secret string
}

func New(secret string) *Authenticator {
	return &Authenticator{secret: secret}
}

func (a *Authenticator) GenerateToken(userID string, ttl time.Duration) (string, error) {
	payload := userID + "|" + time.Now().Add(ttl).Format(time.RFC3339)
	return crypton.Sign(payload, a.secret)
}

func (a *Authenticator) ValidateToken(token string) (*Claims, error) {
	payload, err := crypton.Verify(token, a.secret)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	var userID, expStr string
	for i, c := range payload {
		if c == '|' {
			userID = payload[:i]
			expStr = payload[i+1:]
			break
		}
	}

	exp, err := time.Parse(time.RFC3339, expStr)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if time.Now().After(exp) {
		return nil, ErrTokenExpired
	}

	return &Claims{UserID: userID, ExpiresAt: exp}, nil
}
