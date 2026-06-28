package types

import (
	"encoding/json"
	"fmt"
)

type FeishuCredential struct {
	Platform  Platform `json:"platform"`
	AppID     string   `json:"appId"`
	AppSecret string   `json:"appSecret"`
}

type DingtalkCredential struct {
	Platform  Platform `json:"platform"`
	CorpID    string   `json:"corpId"`
	AppKey    string   `json:"appKey"`
	AppSecret string   `json:"appSecret"`
}

type WecomCredential struct {
	Platform Platform `json:"platform"`
	CorpID   string   `json:"corpId"`
	Secret   string   `json:"secret"`
	AgentID  string   `json:"agentId"`
}

type Credential struct {
	Platform Platform
	Feishu   *FeishuCredential
	Dingtalk *DingtalkCredential
	Wecom    *WecomCredential
}

type StoredCredential struct {
	Platform  Platform
	Encrypted []byte
}

func DecodeCredential(dec *json.Decoder) (Credential, error) {
	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		return Credential{}, err
	}
	return decodeCredentialJSON(raw)
}

func decodeCredentialJSON(raw json.RawMessage) (Credential, error) {
	var probe struct {
		Platform Platform `json:"platform"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return Credential{}, fmt.Errorf("invalid credential body")
	}
	if probe.Platform == "" {
		return Credential{}, fmt.Errorf("platform is required")
	}

	cred := Credential{Platform: probe.Platform}
	switch probe.Platform {
	case PlatformFeishu:
		var fc FeishuCredential
		if err := json.Unmarshal(raw, &fc); err != nil {
			return Credential{}, fmt.Errorf("invalid feishu credential")
		}
		cred.Feishu = &fc
	case PlatformDingtalk:
		var dc DingtalkCredential
		if err := json.Unmarshal(raw, &dc); err != nil {
			return Credential{}, fmt.Errorf("invalid dingtalk credential")
		}
		cred.Dingtalk = &dc
	case PlatformWecom:
		var wc WecomCredential
		if err := json.Unmarshal(raw, &wc); err != nil {
			return Credential{}, fmt.Errorf("invalid wecom credential")
		}
		cred.Wecom = &wc
	default:
		return Credential{}, fmt.Errorf("unsupported platform")
	}
	return cred, nil
}

func (c Credential) Validate() error {
	switch c.Platform {
	case PlatformFeishu:
		if c.Feishu == nil || c.Feishu.AppID == "" || c.Feishu.AppSecret == "" {
			return fmt.Errorf("appId and appSecret are required")
		}
	case PlatformDingtalk:
		if c.Dingtalk == nil || c.Dingtalk.CorpID == "" || c.Dingtalk.AppKey == "" || c.Dingtalk.AppSecret == "" {
			return fmt.Errorf("corpId, appKey and appSecret are required")
		}
	case PlatformWecom:
		if c.Wecom == nil || c.Wecom.CorpID == "" || c.Wecom.Secret == "" || c.Wecom.AgentID == "" {
			return fmt.Errorf("corpId, secret and agentId are required")
		}
	default:
		return fmt.Errorf("unsupported platform")
	}
	return nil
}

func MarshalCredentialPayload(c Credential) ([]byte, error) {
	switch c.Platform {
	case PlatformFeishu:
		return json.Marshal(c.Feishu)
	case PlatformDingtalk:
		return json.Marshal(c.Dingtalk)
	case PlatformWecom:
		return json.Marshal(c.Wecom)
	default:
		return nil, fmt.Errorf("unsupported platform")
	}
}

func UnmarshalCredentialPayload(platform Platform, raw []byte) (Credential, error) {
	cred := Credential{Platform: platform}
	switch platform {
	case PlatformFeishu:
		var fc FeishuCredential
		if err := json.Unmarshal(raw, &fc); err != nil {
			return Credential{}, err
		}
		cred.Feishu = &fc
	case PlatformDingtalk:
		var dc DingtalkCredential
		if err := json.Unmarshal(raw, &dc); err != nil {
			return Credential{}, err
		}
		cred.Dingtalk = &dc
	case PlatformWecom:
		var wc WecomCredential
		if err := json.Unmarshal(raw, &wc); err != nil {
			return Credential{}, err
		}
		cred.Wecom = &wc
	default:
		return Credential{}, fmt.Errorf("unsupported platform")
	}
	return cred, nil
}
