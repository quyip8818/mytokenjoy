package adminport

import "strings"

const (
	TokenStatusEnabled  = 1
	TokenStatusDisabled = 2

	ChannelStatusEnabled  = 1
	ChannelStatusDisabled = 2
)

func ProviderChannelType(provider string) int {
	switch strings.ToLower(provider) {
	case "openai":
		return 1
	case "anthropic":
		return 14
	case "deepseek":
		return 25
	case "qwen":
		return 17
	case "azure":
		return 3
	default:
		return 1
	}
}
