package contract

import (
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

// Demo consumption numbers shared by budget tree, platform keys, projects, and usage bucket scaling.
// Values are in display currency (CNY) and scaled to quota in init().

var DemoLeafDeptConsumed = map[uuid.UUID]int64{
	IDDept3: 0,
	IDDept4: 0,
	IDDept5: 0,
	IDDept6: 0,
	IDDept7: 0,
	IDDept8: 0,
}

var DemoPlatformKeyConsumed = map[uuid.UUID]int64{
	IDPlatformKey1: 0,
	IDPlatformKey2: 0,
	IDPlatformKey3: 0,
	IDPlatformKey4: 0,
	IDPlatformKey5: 0,
	IDPlatformKey6: 0,
}

var DemoProjectConsumed = map[uuid.UUID]int64{
	IDProject1: 0,
	IDProject4: 0,
}

func init() {
	qpu := common.DefaultQuotaPerUnit
	scaleMap := func(m map[uuid.UUID]int64, display map[uuid.UUID]float64) {
		for k, v := range display {
			m[k] = common.QuotaFromAmount(v, qpu)
		}
	}
	scaleMap(DemoLeafDeptConsumed, map[uuid.UUID]float64{
		IDDept3: 21000,
		IDDept4: 11200,
		IDDept5: 6000,
		IDDept6: 14300,
		IDDept7: 8500,
		IDDept8: 6500,
	})
	scaleMap(DemoPlatformKeyConsumed, map[uuid.UUID]float64{
		IDPlatformKey1: 3200,
		IDPlatformKey2: 450,
		IDPlatformKey3: 7800,
		IDPlatformKey4: 2000,
		IDPlatformKey5: 4500,
		IDPlatformKey6: 4200,
	})
	scaleMap(DemoProjectConsumed, map[uuid.UUID]float64{
		IDProject1: 18500,
		IDProject4: 4200,
	})
}

func DemoRootConsumed() int64 {
	var total int64
	for _, consumed := range DemoLeafDeptConsumed {
		total += consumed
	}
	return total
}
