package contract

import "github.com/tokenjoy/backend/internal/pkg/common"

// Demo consumption numbers shared by budget tree, platform keys, projects, and usage bucket scaling.

var DemoLeafDeptConsumed = map[string]float64{
	IDDept3:  21000,
	IDDept4:  11200,
	"dept-5": 6000,
	"dept-6": 14300,
	"dept-7": 8500,
	"dept-8": 6500,
}

var DemoPlatformKeyConsumed = map[string]float64{
	"plk-1":    3200,
	"plk-1b":   450,
	"plk-2":    7800,
	"plk-4":    2000,
	"plk-5":    4500,
	"plk-bg-1": 4200,
}

var DemoProjectConsumed = map[string]float64{
	IDProject1: 18500,
	IDProject4: 4200,
}

func init() {
	ppu := float64(common.DefaultPointsPerUnit)
	scaleMap := func(m map[string]float64) {
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
