// TODO(real): replace in-memory budget approvals with persistent workflow and DB storage.
package budget

import (
	"context"
	"sync"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
)

var (
	approvalsMu     sync.RWMutex
	approvalsByCo   = make(map[int64][]types.BudgetApproval)
	approvalsSeeded = make(map[int64]bool)
)

func seedBudgetApprovals(companyID int64) []types.BudgetApproval {
	_ = companyID
	resolved1 := "2026-06-20 11:30"
	resolved2 := "2026-06-15 16:45"
	resolved3 := "2026-06-25 17:30"
	return []types.BudgetApproval{
		{
			ID: "appr-1", ApplicantName: "张三", DepartmentName: "后端组",
			Amount: 500, Reason: "本月额度用尽，需完成搜索优化任务",
			Status: "pending", CreatedAt: "2026-06-28 14:30",
		},
		{
			ID: "appr-1b", ApplicantName: "张三", DepartmentName: "后端组",
			Amount: 300, Reason: "RAG 管道调试需额外调用",
			Status: "approved", CreatedAt: "2026-06-20 09:00", ResolvedAt: &resolved1,
		},
		{
			ID: "appr-1c", ApplicantName: "张三", DepartmentName: "后端组",
			Amount: 200, Reason: "紧急修复线上搜索问题",
			Status: "approved", CreatedAt: "2026-06-15 16:00", ResolvedAt: &resolved2,
		},
		{
			ID: "appr-2", ApplicantName: "赵六", DepartmentName: "后端组",
			Amount: 300, Reason: "调试 RAG 管道需额外调用",
			Status: "pending", CreatedAt: "2026-06-29 09:15",
		},
		{
			ID: "appr-3", ApplicantName: "吴十", DepartmentName: "产品部",
			Amount: 200, Reason: "产品文档生成",
			Status: "approved", CreatedAt: "2026-06-25 16:00", ResolvedAt: &resolved3,
		},
	}
}

func loadCompanyApprovals(companyID int64) []types.BudgetApproval {
	approvalsMu.Lock()
	defer approvalsMu.Unlock()
	if !approvalsSeeded[companyID] {
		seeded := seedBudgetApprovals(companyID)
		copied := append([]types.BudgetApproval{}, seeded...)
		approvalsByCo[companyID] = copied
		approvalsSeeded[companyID] = true
	}
	return append([]types.BudgetApproval{}, approvalsByCo[companyID]...)
}

func saveCompanyApprovals(companyID int64, items []types.BudgetApproval) {
	approvalsMu.Lock()
	defer approvalsMu.Unlock()
	approvalsByCo[companyID] = append([]types.BudgetApproval{}, items...)
	approvalsSeeded[companyID] = true
}

func (s *service) ListApprovals(ctx context.Context) ([]types.BudgetApproval, error) {
	if err := s.delayer.Wait(ctx, 100*time.Millisecond); err != nil {
		return nil, err
	}
	return loadCompanyApprovals(company.CompanyID(ctx)), nil
}

func (s *service) ResolveApproval(ctx context.Context, id string, input types.ResolveBudgetApprovalInput) (types.BudgetApproval, error) {
	if err := s.delayer.Wait(ctx, 100*time.Millisecond); err != nil {
		return types.BudgetApproval{}, err
	}
	if input.Status != "approved" && input.Status != "rejected" {
		return types.BudgetApproval{}, domain.Validation("invalid status")
	}
	if input.Status == "rejected" && (input.RejectReason == nil || *input.RejectReason == "") {
		return types.BudgetApproval{}, domain.Validation("reject reason required")
	}
	companyID := company.CompanyID(ctx)
	items := loadCompanyApprovals(companyID)
	idx := -1
	for i := range items {
		if items[i].ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return types.BudgetApproval{}, domain.NotFound("Not found")
	}
	if items[idx].Status != "pending" {
		return types.BudgetApproval{}, domain.Validation("approval already resolved")
	}
	now := time.Now().UTC().Format("2006-01-02 15:04")
	items[idx].Status = input.Status
	items[idx].ResolvedAt = &now
	if input.Status == "rejected" {
		items[idx].RejectReason = input.RejectReason
	}
	saveCompanyApprovals(companyID, items)
	return items[idx], nil
}
