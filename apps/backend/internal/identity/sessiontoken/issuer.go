package sessiontoken

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	CompanyID int64  `json:"company_id"`
	Sid       string `json:"sid"`
	jwt.RegisteredClaims
}

type Issuer interface {
	Issue(companyID int64, memberID string) (string, error)
	Parse(token string) (Claims, error)
	Secret() []byte
}

type issuer struct {
	secret []byte
	ttl    time.Duration
}

func NewIssuer(secret string, ttlSec int) (Issuer, error) {
	if secret == "" {
		return nil, fmt.Errorf("session secret is required")
	}
	if ttlSec <= 0 {
		ttlSec = 86400
	}
	return &issuer{
		secret: []byte(secret),
		ttl:    time.Duration(ttlSec) * time.Second,
	}, nil
}

func (i *issuer) Secret() []byte {
	return append([]byte(nil), i.secret...)
}

func (i *issuer) Issue(companyID int64, memberID string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		CompanyID: companyID,
		Sid:       newSessionID(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   memberID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(i.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(i.secret)
}

func (i *issuer) Parse(token string) (Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return i.secret, nil
	})
	if err != nil {
		return Claims{}, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return Claims{}, fmt.Errorf("invalid token")
	}
	return *claims, nil
}

func newSessionID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}
