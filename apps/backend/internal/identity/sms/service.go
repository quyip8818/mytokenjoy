package sms

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Service manages SMS verification code lifecycle.
type Service struct {
	rdb    *goredis.Client
	sender Sender
	logger *slog.Logger
}

// Sender is the interface for actually delivering SMS messages.
type Sender interface {
	SendCode(ctx context.Context, phone string, code string) error
}

// NewService creates an SMS verification service.
// Returns nil if redisURL is empty (SMS disabled).
func NewService(redisURL string, sender Sender, logger *slog.Logger) (*Service, error) {
	if redisURL == "" {
		return nil, nil
	}
	opt, err := goredis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("sms: parse REDIS_URL: %w", err)
	}
	rdb := goredis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("sms: redis ping: %w", err)
	}
	return &Service{rdb: rdb, sender: sender, logger: logger}, nil
}

// Redis key patterns:
//   sms:code:{phone}     → code, TTL 5min
//   sms:lock:{phone}     → "1", TTL 60s (send interval)
//   sms:daily:{phone}    → counter, TTL to end of day
//   sms:attempts:{phone} → counter, TTL 15min (verify failure count)

func keyCode(phone string) string     { return "sms:code:" + phone }
func keyLock(phone string) string     { return "sms:lock:" + phone }
func keyDaily(phone string) string    { return "sms:daily:" + phone }
func keyAttempts(phone string) string { return "sms:attempts:" + phone }

const (
	codeTTL     = 5 * time.Minute
	lockTTL     = 60 * time.Second
	attemptsTTL = 15 * time.Minute
	dailyLimit  = 10
	maxAttempts = 5
	codeLength  = 6
)

// SendResult indicates why a send was rejected, if at all.
type SendResult struct {
	OK         bool
	RetryAfter int // seconds until next send allowed (when locked)
	Reason     string
}

// Send generates and sends a verification code to the given phone number.
func (s *Service) Send(ctx context.Context, phone string) SendResult {
	// Check send interval lock.
	lockKey := keyLock(phone)
	ttl, err := s.rdb.TTL(ctx, lockKey).Result()
	if err == nil && ttl > 0 {
		return SendResult{OK: false, RetryAfter: int(ttl.Seconds()), Reason: "send too frequent"}
	}

	// Check daily limit.
	dailyKey := keyDaily(phone)
	count, _ := s.rdb.Get(ctx, dailyKey).Int64()
	if count >= dailyLimit {
		return SendResult{OK: false, Reason: "daily limit exceeded"}
	}

	// Generate code.
	code := generateCode(codeLength)

	// Store code with TTL.
	pipe := s.rdb.Pipeline()
	pipe.Set(ctx, keyCode(phone), code, codeTTL)
	pipe.Set(ctx, lockKey, "1", lockTTL)
	pipe.Incr(ctx, dailyKey)
	// Set daily key expiry to end of today if not already set.
	pipe.ExpireNX(ctx, dailyKey, timeUntilEndOfDay())
	// Reset attempts on new code send.
	pipe.Del(ctx, keyAttempts(phone))
	if _, err := pipe.Exec(ctx); err != nil {
		s.logger.Error("sms: redis pipeline failed", "error", err, "phone", phone)
		return SendResult{OK: false, Reason: "internal error"}
	}

	// Send via external provider.
	if err := s.sender.SendCode(ctx, phone, code); err != nil {
		s.logger.Error("sms: send failed", "error", err, "phone", phone)
		// Code is stored; user can retry verify if SMS arrives late.
		// Don't roll back Redis state to avoid abuse.
		return SendResult{OK: false, Reason: "send failed"}
	}

	return SendResult{OK: true}
}

// VerifyResult indicates why a verification failed, if at all.
type VerifyResult struct {
	OK     bool
	Locked bool // true when max attempts exceeded
	Reason string
}

// Verify checks the code for the given phone number.
func (s *Service) Verify(ctx context.Context, phone, code string) VerifyResult {
	attemptsKey := keyAttempts(phone)

	// Check if locked due to too many attempts.
	attempts, _ := s.rdb.Get(ctx, attemptsKey).Int64()
	if attempts >= maxAttempts {
		return VerifyResult{OK: false, Locked: true, Reason: "too many attempts, try again later"}
	}

	// Get stored code.
	stored, err := s.rdb.Get(ctx, keyCode(phone)).Result()
	if err == goredis.Nil {
		return VerifyResult{OK: false, Reason: "code expired or not sent"}
	}
	if err != nil {
		s.logger.Error("sms: redis get code failed", "error", err)
		return VerifyResult{OK: false, Reason: "internal error"}
	}

	if stored != code {
		// Increment attempts.
		pipe := s.rdb.Pipeline()
		pipe.Incr(ctx, attemptsKey)
		pipe.Expire(ctx, attemptsKey, attemptsTTL)
		_, _ = pipe.Exec(ctx)
		return VerifyResult{OK: false, Reason: "invalid code"}
	}

	// Success — clean up.
	pipe := s.rdb.Pipeline()
	pipe.Del(ctx, keyCode(phone))
	pipe.Del(ctx, keyAttempts(phone))
	_, _ = pipe.Exec(ctx)

	return VerifyResult{OK: true}
}

// Close releases Redis resources.
func (s *Service) Close() error {
	if s == nil || s.rdb == nil {
		return nil
	}
	return s.rdb.Close()
}

// generateCode generates a random numeric code of the given length.
func generateCode(length int) string {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		// Fallback: should never happen with crypto/rand.
		return "123456"
	}
	code := n.String()
	// Pad with leading zeros.
	for len(code) < length {
		code = "0" + code
	}
	return code
}

// timeUntilEndOfDay returns duration until 23:59:59 today (local time, suitable for China).
func timeUntilEndOfDay() time.Duration {
	now := time.Now()
	loc := now.Location()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, loc)
	d := endOfDay.Sub(now)
	if d <= 0 {
		d = 24 * time.Hour
	}
	return d
}

// FormatPhone normalizes a phone number for storage.
// Ensures the phone starts with country code prefix.
func FormatPhone(phone string) string {
	// Simple normalization: if already has +86 or starts with 1 (Chinese mobile), keep as-is.
	// In production, use a proper phone library.
	if len(phone) == 11 && phone[0] == '1' {
		return "+86" + phone
	}
	if len(phone) > 0 && phone[0] == '+' {
		return phone
	}
	return phone
}
