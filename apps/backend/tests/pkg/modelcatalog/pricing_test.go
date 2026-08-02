package modelcatalog_test

import (
	"math"
	"testing"

	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
)

func TestPriceFromRatio(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		modelRatio      float64
		completionRatio float64
		wantInput       float64
		wantOutput      float64
	}{
		{name: "gpt-4o typical", modelRatio: 1.25, completionRatio: 4.0, wantInput: 2.5, wantOutput: 10.0},
		{name: "equal ratio", modelRatio: 1.0, completionRatio: 1.0, wantInput: 2.0, wantOutput: 2.0},
		{name: "zero ratio", modelRatio: 0, completionRatio: 0, wantInput: 0, wantOutput: 0},
		{name: "completion 2x", modelRatio: 5.0, completionRatio: 2.0, wantInput: 10.0, wantOutput: 20.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, output := modelcatalog.PriceFromRatio(tt.modelRatio, tt.completionRatio)
			if math.Abs(input-tt.wantInput) > 1e-9 {
				t.Errorf("inputPrice = %f, want %f", input, tt.wantInput)
			}
			if math.Abs(output-tt.wantOutput) > 1e-9 {
				t.Errorf("outputPrice = %f, want %f", output, tt.wantOutput)
			}
		})
	}
}

func TestRatioFromPrice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		inputPrice     float64
		outputPrice    float64
		wantModelRatio float64
		wantCompletion float64
	}{
		{name: "gpt-4o typical", inputPrice: 2.5, outputPrice: 10.0, wantModelRatio: 1.25, wantCompletion: 4.0},
		{name: "equal price", inputPrice: 2.0, outputPrice: 2.0, wantModelRatio: 1.0, wantCompletion: 1.0},
		{name: "zero price", inputPrice: 0, outputPrice: 0, wantModelRatio: 0, wantCompletion: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelRatio, completionRatio := modelcatalog.RatioFromPrice(tt.inputPrice, tt.outputPrice)
			if math.Abs(modelRatio-tt.wantModelRatio) > 1e-9 {
				t.Errorf("modelRatio = %f, want %f", modelRatio, tt.wantModelRatio)
			}
			if math.Abs(completionRatio-tt.wantCompletion) > 1e-9 {
				t.Errorf("completionRatio = %f, want %f", completionRatio, tt.wantCompletion)
			}
		})
	}
}

func TestPriceRatioRoundtrip(t *testing.T) {
	t.Parallel()
	// Price → Ratio → Price should be identity.
	cases := [][2]float64{{2.5, 10.0}, {1.0, 3.0}, {0.5, 0.5}, {100.0, 200.0}}
	for _, c := range cases {
		inputPrice, outputPrice := c[0], c[1]
		mr, cr := modelcatalog.RatioFromPrice(inputPrice, outputPrice)
		gotInput, gotOutput := modelcatalog.PriceFromRatio(mr, cr)
		if math.Abs(gotInput-inputPrice) > 1e-9 || math.Abs(gotOutput-outputPrice) > 1e-9 {
			t.Errorf("roundtrip(%f,%f) → ratio(%f,%f) → price(%f,%f)", inputPrice, outputPrice, mr, cr, gotInput, gotOutput)
		}
	}
}
