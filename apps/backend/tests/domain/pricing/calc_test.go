package pricing_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/pricing"
)

func TestCalcQuota(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		inputTokens  int64
		outputTokens int64
		inputPrice   float64
		outputPrice  float64
		quotaPerUnit int64
		want         int64
	}{
		{
			name:         "deepseek-chat typical call",
			inputTokens:  1000,
			outputTokens: 500,
			inputPrice:   1.0, // 元/1M tokens
			outputPrice:  4.0,
			quotaPerUnit: 500000,
			// cost = (1000*1.0 + 500*4.0) / 1_000_000 = 0.003 元
			// quota = ceil(0.003 * 500000) = ceil(1500) = 1500
			want: 1500,
		},
		{
			name:         "zero tokens",
			inputTokens:  0,
			outputTokens: 0,
			inputPrice:   10.0,
			outputPrice:  40.0,
			quotaPerUnit: 500000,
			want:         0,
		},
		{
			name:         "zero price",
			inputTokens:  10000,
			outputTokens: 5000,
			inputPrice:   0,
			outputPrice:  0,
			quotaPerUnit: 500000,
			want:         0,
		},
		{
			name:         "ceil rounding up",
			inputTokens:  1,
			outputTokens: 1,
			inputPrice:   1.0,
			outputPrice:  1.0,
			quotaPerUnit: 500000,
			// cost = (1*1.0 + 1*1.0) / 1_000_000 = 0.000002 元
			// quota = ceil(0.000002 * 500000) = ceil(1.0) = 1
			want: 1,
		},
		{
			name:         "large call gpt-4o",
			inputTokens:  100000,
			outputTokens: 50000,
			inputPrice:   15.0,
			outputPrice:  60.0,
			quotaPerUnit: 500000,
			// cost = (100000*15.0 + 50000*60.0) / 1_000_000 = (1500000+3000000)/1000000 = 4.5 元
			// quota = ceil(4.5 * 500000) = 2250000
			want: 2250000,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := pricing.CalcQuota(tc.inputTokens, tc.outputTokens, tc.inputPrice, tc.outputPrice, tc.quotaPerUnit)
			if got != tc.want {
				t.Errorf("CalcQuota(%d, %d, %f, %f, %d) = %d, want %d",
					tc.inputTokens, tc.outputTokens, tc.inputPrice, tc.outputPrice, tc.quotaPerUnit, got, tc.want)
			}
		})
	}
}
