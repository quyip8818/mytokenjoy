package contract

import (
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

// Demo consumption numbers shared by budget tree, platform keys, projects, and usage bucket scaling.

var DemoLeafDeptConsumed = map[uuid.UUID]float64{
	IDDept3: 21000,
	IDDept4: 11200,
	IDDept5: 6000,
	IDDept6: 14300,
	IDDept7: 8500,
	IDDept8: 6500,
}

var DemoPlatformKeyConsumed = map[uuid.UUID]float64{
	IDPlatformKey1: 3200,
	IDPlatformKey2: 450,
	IDPlatformKey3: 7800,
	IDPlatformKey4: 2000,
	IDPlatformKey5: 4500,
	IDPlatformKey6: 4200,
}

var DemoProjectConsumed = map[uuid.UUID]float64{
	IDProject1: 18500,
	IDProject4: 4200,
}

func init() {
	ppu := float64(common.DefaultPointsPerUnit)
	scaleMap := func(m map[uuid.UUID]float64) {
		for k, v := range m {
			m[k] = v * ppu
		}
	}
	scaleMap(DemoLeafDeptConsumed)
	scaleMap(DemoPlatformKeyConsumed)
	scaleMap(DemoProjectConsumed)
}

func DemoRootConsumed() float64 {
	var total float64
	for _, consumed := range DemoLeafDeptConsumed {
		total += consumed
	}
	return total
}
