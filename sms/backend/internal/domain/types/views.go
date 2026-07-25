package types

import "github.com/google/uuid"

// ========== 聚合视图 / API Response DTO ==========

type SupplierDetail struct {
	Supplier
	Contacts    []SupplierContact `json:"contacts"`
	Models      []AiModel         `json:"models"`
	Contracts   []Contract        `json:"contracts"`
	Orders      []PurchaseOrder   `json:"orders"`
	Evaluations []Evaluation      `json:"evaluations"`
}

type ContractDetail struct {
	Contract
	Attachments []ContractAttachment `json:"attachments"`
}

// Dashboard
type DashboardCards struct {
	SupplierTotal   int `json:"supplierTotal"`
	ActiveSuppliers int `json:"activeSuppliers"`
	ModelTotal      int `json:"modelTotal"`
	ActiveContracts int `json:"activeContracts"`
}

type DashboardCharts struct {
	GradeDistribution    []LabelCount `json:"gradeDistribution"`
	ModelCountBySupplier []LabelCount `json:"modelCountBySupplier"`
}

type LabelCount struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type ExpiringContract struct {
	ID           uuid.UUID `json:"id"           db:"id"`
	Title        string    `json:"title"        db:"title"`
	ContractNo   string    `json:"contractNo"   db:"contract_no"`
	EndDate      string    `json:"endDate"      db:"end_date"`
	SupplierName string    `json:"supplierName" db:"supplier_name"`
}
