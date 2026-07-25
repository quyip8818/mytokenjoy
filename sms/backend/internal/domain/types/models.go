package types

import (
	"time"

	"github.com/google/uuid"
)

// ========== 数据库 Entity（对应表结构）==========

type User struct {
	ID           uuid.UUID `json:"id"          db:"id"`
	Username     string    `json:"username"    db:"username"`
	PasswordHash string    `json:"-"           db:"password_hash"`
	RealName     string    `json:"realName"    db:"real_name"`
	Email        *string   `json:"email"       db:"email"`
	Role         string    `json:"role"        db:"role"`
	Status       int       `json:"status"      db:"status"`
	CreatedAt    time.Time `json:"createdAt"   db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt"   db:"updated_at"`
}

type Session struct {
	Token     string    `db:"token"`
	UserID    uuid.UUID `db:"user_id"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

type Supplier struct {
	ID          uuid.UUID  `json:"id"          db:"id"`
	Name        string     `json:"name"        db:"name"`
	Code        string     `json:"code"        db:"code"`
	Category    *string    `json:"category"    db:"category"`
	Website     *string    `json:"website"     db:"website"`
	Status      string     `json:"status"      db:"status"`
	Description *string    `json:"description" db:"description"`
	CreatedBy   *uuid.UUID `json:"createdBy"   db:"created_by"`
	CreatedAt   time.Time  `json:"createdAt"   db:"created_at"`
	UpdatedAt   time.Time  `json:"updatedAt"   db:"updated_at"`
}

type SupplierContact struct {
	ID         uuid.UUID `json:"id"         db:"id"`
	SupplierID uuid.UUID `json:"supplierId" db:"supplier_id"`
	Name       string    `json:"name"       db:"name"`
	Position   *string   `json:"position"   db:"position"`
	Phone      *string   `json:"phone"      db:"phone"`
	Email      *string   `json:"email"      db:"email"`
	IsPrimary  bool      `json:"isPrimary"  db:"is_primary"`
	CreatedAt  time.Time `json:"createdAt"  db:"created_at"`
}

type AiModel struct {
	ID            uuid.UUID `json:"id"            db:"id"`
	SupplierID    uuid.UUID `json:"supplierId"    db:"supplier_id"`
	ModelName     string    `json:"modelName"     db:"model_name"`
	ModelID       *string   `json:"modelId"       db:"model_id"`
	ModelType     *string   `json:"modelType"     db:"model_type"`
	ContextLength *int      `json:"contextLength" db:"context_length"`
	InputPrice    *float64  `json:"inputPrice"    db:"input_price"`
	OutputPrice   *float64  `json:"outputPrice"   db:"output_price"`
	Discount      *float64  `json:"discount"      db:"discount"`
	Status        string    `json:"status"        db:"status"`
	Description   *string   `json:"description"   db:"description"`
	CreatedAt     time.Time `json:"createdAt"     db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt"     db:"updated_at"`
	// JOIN 字段
	SupplierName *string `json:"supplierName,omitempty" db:"supplier_name"`
}

type Contract struct {
	ID         uuid.UUID  `json:"id"           db:"id"`
	SupplierID uuid.UUID  `json:"supplierId"   db:"supplier_id"`
	ContractNo string     `json:"contractNo"   db:"contract_no"`
	Title      string     `json:"title"        db:"title"`
	Amount     *float64   `json:"amount"       db:"amount"`
	SignDate   *DateOnly  `json:"signDate"     db:"sign_date"`
	StartDate  *DateOnly  `json:"startDate"    db:"start_date"`
	EndDate    *DateOnly  `json:"endDate"      db:"end_date"`
	Status     string     `json:"status"       db:"status"`
	Remarks    *string    `json:"remarks"      db:"remarks"`
	CreatedBy  *uuid.UUID `json:"createdBy"    db:"created_by"`
	CreatedAt  time.Time  `json:"createdAt"    db:"created_at"`
	UpdatedAt  time.Time  `json:"updatedAt"    db:"updated_at"`
	// JOIN 字段
	SupplierName *string `json:"supplierName,omitempty" db:"supplier_name"`
}

type ContractAttachment struct {
	ID         uuid.UUID  `json:"id"         db:"id"`
	ContractID uuid.UUID  `json:"contractId" db:"contract_id"`
	FileName   string     `json:"fileName"   db:"file_name"`
	FilePath   string     `json:"-"          db:"file_path"`
	FileSize   int64      `json:"fileSize"   db:"file_size"`
	UploadedBy *uuid.UUID `json:"uploadedBy" db:"uploaded_by"`
	CreatedAt  time.Time  `json:"createdAt"  db:"created_at"`
}

type PurchaseOrder struct {
	ID          uuid.UUID  `json:"id"           db:"id"`
	OrderNo     string     `json:"orderNo"      db:"order_no"`
	SupplierID  uuid.UUID  `json:"supplierId"   db:"supplier_id"`
	ContractID  *uuid.UUID `json:"contractId"   db:"contract_id"`
	TotalAmount *float64   `json:"totalAmount"  db:"total_amount"`
	OrderDate   *DateOnly  `json:"orderDate"    db:"order_date"`
	Status      string     `json:"status"       db:"status"`
	Description *string    `json:"description"  db:"description"`
	CreatedBy   *uuid.UUID `json:"createdBy"    db:"created_by"`
	CreatedAt   time.Time  `json:"createdAt"    db:"created_at"`
	UpdatedAt   time.Time  `json:"updatedAt"    db:"updated_at"`
	// JOIN 字段
	SupplierName *string `json:"supplierName,omitempty" db:"supplier_name"`
	ContractNo   *string `json:"contractNo,omitempty"   db:"contract_no"`
	CreatorName  *string `json:"creatorName,omitempty"  db:"creator_name"`
}

type Evaluation struct {
	ID          uuid.UUID `json:"id"            db:"id"`
	SupplierID  uuid.UUID `json:"supplierId"    db:"supplier_id"`
	EvaluatorID uuid.UUID `json:"evaluatorId"   db:"evaluator_id"`
	Period      string    `json:"period"        db:"period"`
	Quality     int       `json:"quality"       db:"quality"`
	Performance int       `json:"performance"   db:"performance"`
	Price       int       `json:"price"         db:"price"`
	Service     int       `json:"service"       db:"service"`
	Compliance  int       `json:"compliance"    db:"compliance"`
	TotalScore  float64   `json:"totalScore"    db:"total_score"`
	Grade       string    `json:"grade"         db:"grade"`
	Comment     *string   `json:"comment"       db:"comment"`
	CreatedAt   time.Time `json:"createdAt"     db:"created_at"`
	// JOIN 字段
	SupplierName  *string `json:"supplierName,omitempty"  db:"supplier_name"`
	EvaluatorName *string `json:"evaluatorName,omitempty" db:"evaluator_name"`
}

type EvaluationWeight struct {
	ID        uuid.UUID `json:"id"        db:"id"`
	Dimension string    `json:"dimension" db:"dimension"`
	Weight    int       `json:"weight"    db:"weight"`
}
