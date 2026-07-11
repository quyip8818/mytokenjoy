package gateway

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
)

type Prechecker interface {
	Run(ctx context.Context, keyHash string, model string, skipModelCheck bool) error
}

type PrecheckService struct {
	loader store.GatewayPrecheckRepository
	clock  clock.Clock
}

func NewPrecheckService(loader store.GatewayPrecheckRepository, clk clock.Clock) *PrecheckService {
	return &PrecheckService{
		loader: loader,
		clock:  clock.OrDefault(clk),
	}
}

func (p *PrecheckService) Run(ctx context.Context, keyHash string, model string, skipModelCheck bool) error {
	row, err := p.loader.LoadPrecheckContext(ctx, keyHash, clock.NowUTC(p.clock))
	if err != nil {
		return err
	}
	if row == nil {
		return fmt.Errorf("platform key not found")
	}
	return Evaluate(PrecheckContextFromStore(row), model, skipModelCheck)
}

var _ Prechecker = (*PrecheckService)(nil)
