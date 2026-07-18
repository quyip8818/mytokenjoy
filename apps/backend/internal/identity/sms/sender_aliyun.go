package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	"github.com/alibabacloud-go/tea/tea"
)

// AliyunSender sends verification codes via Alibaba Cloud SMS.
type AliyunSender struct {
	client       *dysmsapi.Client
	signName     string
	templateCode string
	logger       *slog.Logger
}

// AliyunConfig holds verification SMS settings.
type AliyunConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	SignName        string
	TemplateCode    string
	Endpoint        string
}

// NewAliyunSender creates a sender for SMS verification codes.
// Returns a no-op sender if credentials are incomplete.
func NewAliyunSender(cfg AliyunConfig, logger *slog.Logger) Sender {
	keyID := strings.TrimSpace(cfg.AccessKeyID)
	keySecret := strings.TrimSpace(cfg.AccessKeySecret)
	signName := strings.TrimSpace(cfg.SignName)
	templateCode := strings.TrimSpace(cfg.TemplateCode)

	if keyID == "" || keySecret == "" || signName == "" || templateCode == "" {
		logger.Info("sms: aliyun sender disabled (incomplete config)")
		return &logSender{logger: logger}
	}

	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		endpoint = "dysmsapi.aliyuncs.com"
	}

	client, err := dysmsapi.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(keyID),
		AccessKeySecret: tea.String(keySecret),
		Endpoint:        tea.String(endpoint),
	})
	if err != nil {
		logger.Error("sms: failed to create aliyun client", "error", err)
		return &logSender{logger: logger}
	}

	return &AliyunSender{
		client:       client,
		signName:     signName,
		templateCode: templateCode,
		logger:       logger,
	}
}

func (s *AliyunSender) SendCode(ctx context.Context, phone string, code string) error {
	// Strip +86 prefix for Aliyun API (expects bare number or with country code).
	phoneNum := phone
	if strings.HasPrefix(phone, "+86") {
		phoneNum = strings.TrimPrefix(phone, "+86")
	}

	paramJSON, _ := json.Marshal(map[string]string{"code": code})

	resp, err := s.client.SendSms(&dysmsapi.SendSmsRequest{
		PhoneNumbers:  tea.String(phoneNum),
		SignName:      tea.String(s.signName),
		TemplateCode:  tea.String(s.templateCode),
		TemplateParam: tea.String(string(paramJSON)),
	})
	if err != nil {
		return fmt.Errorf("aliyun sms send: %w", err)
	}
	if resp.Body == nil {
		return fmt.Errorf("aliyun sms: nil response body")
	}
	if respCode := tea.StringValue(resp.Body.Code); respCode != "OK" {
		return fmt.Errorf("aliyun sms rejected: %s - %s", respCode, tea.StringValue(resp.Body.Message))
	}

	s.logger.Debug("sms code sent", "phone", phoneNum, "bizId", tea.StringValue(resp.Body.BizId))
	return nil
}

// logSender is a fallback that logs instead of sending (dev mode).
type logSender struct {
	logger *slog.Logger
}

func (s *logSender) SendCode(_ context.Context, phone string, code string) error {
	s.logger.Info("sms: [DEV] verification code", "phone", phone, "code", code)
	return nil
}
