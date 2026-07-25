package types

var SupplierStatuses = []string{"potential", "active", "frozen", "blacklisted"}
var ContractStatuses = []string{"draft", "active", "expired", "terminated"}
var OrderStatuses = []string{"pending", "approved", "delivered", "completed", "cancelled"}
var ModelStatuses = []string{"available", "deprecated"}
var EvalGrades = []string{"A", "B", "C", "D"}

func IsValidStatus(val string, allowed []string) bool {
	for _, s := range allowed {
		if s == val {
			return true
		}
	}
	return false
}

// OrderTransitions 订单状态流转
var OrderTransitions = map[string][]string{
	"pending":   {"approved", "cancelled"},
	"approved":  {"delivered", "cancelled"},
	"delivered": {"completed", "cancelled"},
}

func IsValidTransition(from, to string) bool {
	allowed, ok := OrderTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}
