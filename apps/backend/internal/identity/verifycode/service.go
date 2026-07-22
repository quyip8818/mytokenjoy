// Package verifycode manages verification code lifecycle (generate, store, send, verify)
// for both SMS and Email channels. Delivery is delegated to the notification infrastructure.
package verifycode

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	goredis "github.com/redis/go-redis/v9"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
)

// Notifier is the subset of notification.Service that verifycode needs.
type Notifier interface {
	SendDirect(ctx context.Context, channel string, address string, msg domainnotification.RenderedMessage) error
}

// Service manages verification code lifecycle for any channel (SMS, Email).
type Service struct {
	rdb        *goredis.Client
	notifier   Notifier
	logger     *slog.Logger
	skipVerify bool // dev-only: bypass send + accept any code
}

// Config holds the settings for creating a verifycode.Service.
type Config struct {
	RedisURL   string
	SkipVerify bool
}

// NewService creates a verification code service.
// Returns nil if RedisURL is empty (service disabled).
// When SkipVerify is true, Send always succeeds without sending and Verify always passes.
func NewService(cfg Config, notifier Notifier, logger *slog.Logger) (*Service, error) {
	if cfg.RedisURL == "" {
		return nil, nil
	}
	if cfg.SkipVerify {
		return &Service{skipVerify: true, logger: logger, notifier: notifier}, nil
	}
	opt, err := goredis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("verifycode: parse REDIS_URL: %w", err)
	}
	rdb := goredis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("verifycode: redis ping: %w", err)
	}
	return &Service{rdb: rdb, notifier: notifier, logger: logger}, nil
}

// Redis key patterns (channel-aware):
//   vc:code:{channel}:{address}     → code, TTL 5min
//   vc:lock:{channel}:{address}     → "1", TTL 60s (send interval)
//   vc:daily:{channel}:{address}    → counter, TTL to end of day
//   vc:attempts:{channel}:{address} → counter, TTL 15min (verify failure count)

func keyCode(channel, addr string) string     { return "vc:code:" + channel + ":" + addr }
func keyLock(channel, addr string) string     { return "vc:lock:" + channel + ":" + addr }
func keyDaily(channel, addr string) string    { return "vc:daily:" + channel + ":" + addr }
func keyAttempts(channel, addr string) string { return "vc:attempts:" + channel + ":" + addr }

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

// Send generates a verification code and delivers it via the specified channel.
// channel: "sms" or "email". address: phone number or email address.
func (s *Service) Send(ctx context.Context, channel, address string) SendResult {
	if s.skipVerify {
		return SendResult{OK: true}
	}

	// Check send interval lock.
	lockKey := keyLock(channel, address)
	ttl, err := s.rdb.TTL(ctx, lockKey).Result()
	if err == nil && ttl > 0 {
		return SendResult{OK: false, RetryAfter: int(ttl.Seconds()), Reason: "send too frequent"}
	}

	// Check daily limit.
	dailyKey := keyDaily(channel, address)
	count, _ := s.rdb.Get(ctx, dailyKey).Int64()
	if count >= dailyLimit {
		return SendResult{OK: false, Reason: "daily limit exceeded"}
	}

	// Generate code.
	code := generateCode(codeLength)

	// Store code with TTL.
	pipe := s.rdb.Pipeline()
	pipe.Set(ctx, keyCode(channel, address), code, codeTTL)
	pipe.Set(ctx, lockKey, "1", lockTTL)
	pipe.Incr(ctx, dailyKey)
	pipe.ExpireNX(ctx, dailyKey, timeUntilEndOfDay())
	pipe.Del(ctx, keyAttempts(channel, address))
	if _, err := pipe.Exec(ctx); err != nil {
		s.logger.Error("verifycode: redis pipeline failed", "error", err, "channel", channel, "address", address)
		return SendResult{OK: false, Reason: "internal error"}
	}

	// Deliver via notification channel.
	msg := domainnotification.RenderedMessage{
		Title: "验证码",
		Body:  fmt.Sprintf("您的验证码是：%s，5分钟内有效。", code),
		Payload: map[string]any{
			"code":      code,
			"eventType": "verification_code",
		},
	}
	if err := s.notifier.SendDirect(ctx, channel, address, msg); err != nil {
		s.logger.Error("verifycode: send failed", "error", err, "channel", channel, "address", address)
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

// Verify checks the code for the given channel and address.
func (s *Service) Verify(ctx context.Context, channel, address, code string) VerifyResult {
	if s.skipVerify {
		return VerifyResult{OK: true}
	}

	attemptsKey := keyAttempts(channel, address)

	// Check if locked due to too many attempts.
	attempts, _ := s.rdb.Get(ctx, attemptsKey).Int64()
	if attempts >= maxAttempts {
		return VerifyResult{OK: false, Locked: true, Reason: "too many attempts, try again later"}
	}

	// Get stored code.
	stored, err := s.rdb.Get(ctx, keyCode(channel, address)).Result()
	if err == goredis.Nil {
		return VerifyResult{OK: false, Reason: "code expired or not sent"}
	}
	if err != nil {
		s.logger.Error("verifycode: redis get code failed", "error", err)
		return VerifyResult{OK: false, Reason: "internal error"}
	}

	if stored != code {
		pipe := s.rdb.Pipeline()
		pipe.Incr(ctx, attemptsKey)
		pipe.Expire(ctx, attemptsKey, attemptsTTL)
		_, _ = pipe.Exec(ctx)
		return VerifyResult{OK: false, Reason: "invalid code"}
	}

	// Success — clean up.
	pipe := s.rdb.Pipeline()
	pipe.Del(ctx, keyCode(channel, address))
	pipe.Del(ctx, keyAttempts(channel, address))
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

// --- Helpers ---

func generateCode(length int) string {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "123456"
	}
	code := n.String()
	for len(code) < length {
		code = "0" + code
	}
	return code
}

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
func FormatPhone(phone string) string {
	if len(phone) == 11 && phone[0] == '1' {
		return "+86" + phone
	}
	if len(phone) > 0 && phone[0] == '+' {
		return phone
	}
	return phone
}
