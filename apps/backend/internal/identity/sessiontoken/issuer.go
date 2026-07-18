package sessiontoken

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	CompanyID uuid.UUID `json:"company_id"`
	UserID    uuid.UUID `json:"user_id,omitempty"`
	Sid       string    `json:"sid"`
	jwt.RegisteredClaims
}

type Issuer interface {
	Issue(companyID uuid.UUID, memberID uuid.UUID) (string, error)
	IssueWithUser(companyID uuid.UUID, memberID uuid.UUID, userID uuid.UUID) (string, error)
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
		ttlSec = 900
	}
	return &issuer{
		secret: []byte(secret),
		ttl:    time.Duration(ttlSec) * time.Second,
	}, nil
}

func (i *issuer) Secret() []byte {
	return append([]byte(nil), i.secret...)
}

func (i *issuer) Issue(companyID uuid.UUID, memberID uuid.UUID) (string, error) {
	return i.IssueWithUser(companyID, memberID, uuid.Nil)
}

func (i *issuer) IssueWithUser(companyID uuid.UUID, memberID uuid.UUID, userID uuid.UUID) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		CompanyID: companyID,
		UserID:    userID,
		Sid:       newSessionID(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   memberID.String(),
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

// NewSessionID generates a random session ID for external callers (e.g. issueTokenPair).
func NewSessionID() string {
	return newSessionID()
}

// RandomHex generates n random bytes and returns their hex encoding (2n chars).
func RandomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// SHA256Hex returns the hex-encoded SHA-256 hash of s.
func SHA256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// IssueAccessToken is a standalone function (not on Issuer) that signs a JWT
// with a caller-supplied sid. Used by refresh and issueTokenPair flows where
// the session ID is managed externally.
func IssueAccessToken(secret []byte, ttl time.Duration, companyID, memberID, userID uuid.UUID, sid string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		CompanyID: companyID,
		UserID:    userID,
		Sid:       sid,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   memberID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}
