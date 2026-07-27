package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/company"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
)

// ApplyNotifications seeds notification_log entries for the demo company.
// Provides a variety of categories, read/unread states, and grouped notifications.
func ApplyNotifications(ctx context.Context, st store.Store) error {
	if _, ok := company.FromContext(ctx); !ok {
		ctx = company.DefaultContext(contract.DefaultCompanyID)
	}

	// Idempotency: skip if notifications already exist.
	count, err := st.Notification().GetUnreadCount(ctx, contract.IDMemberAdmin)
	if err != nil {
		return fmt.Errorf("check notification_log: %w", err)
	}
	if count > 0 {
		return nil
	}

	entries := buildSeedNotifications()
	for _, e := range entries {
		if err := st.Notification().Append(ctx, e); err != nil {
			return fmt.Errorf("seed notification %s: %w", e.ID, err)
		}
	}

	// Mark some as read to simulate realistic inbox state.
	readIDs := []uuid.UUID{idNotif5, idNotif6, idNotif7, idNotif8}
	for _, id := range readIDs {
		if err := st.Notification().MarkRead(ctx, id); err != nil {
			return fmt.Errorf("seed mark-read %s: %w", id, err)
		}
	}

	// Archive one to populate the archived tab.
	if err := st.Notification().Archive(ctx, idNotif8); err != nil {
		return fmt.Errorf("seed archive %s: %w", idNotif8, err)
	}

	return nil
}

// Deterministic seed IDs for notification entries.
var (
	idNotif1 = uuid.MustParse("00000000-0000-7000-8000-0000000ac001")
	idNotif2 = uuid.MustParse("00000000-0000-7000-8000-0000000ac002")
	idNotif3 = uuid.MustParse("00000000-0000-7000-8000-0000000ac003")
	idNotif4 = uuid.MustParse("00000000-0000-7000-8000-0000000ac004")
	idNotif5 = uuid.MustParse("00000000-0000-7000-8000-0000000ac005")
	idNotif6 = uuid.MustParse("00000000-0000-7000-8000-0000000ac006")
	idNotif7 = uuid.MustParse("00000000-0000-7000-8000-0000000ac007")
	idNotif8 = uuid.MustParse("00000000-0000-7000-8000-0000000ac008")
)

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func buildSeedNotifications() []types.NotificationLogEntry {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	if loc == nil {
		loc = time.UTC
	}

	t := func(s string) time.Time {
		ts, _ := time.ParseInLocation("2006-01-02 15:04:05", s, loc)
		return ts.UTC()
	}

	return []types.NotificationLogEntry{
		// 1. Budget alert — unread, grouped with #3
		{
			ID:        idNotif1,
			Channel:   domainnotification.ChannelInApp,
			EventType: domainnotification.EventBudgetAlertReached,
			UserID:    contract.IDMemberAdmin,
			Title:     "预算使用率达到 90%",
			Body:      "项目「智能客服」本月预算已使用 ¥2,340 / ¥2,600",
			Payload:   mustJSON(map[string]any{"ruleID": contract.IDAlertRule1.String(), "projectID": contract.IDProject1.String(), "threshold": 90}),
			SendOK:    true,
			Category:  domainnotification.CategoryBudgetAlert,
			GroupKey:  fmt.Sprintf("budget:%s:%s", contract.IDAlertRule1, contract.DemoBudgetPeriod),
			CreatedAt: t("2026-07-27 09:30:00"),
		},
		// 2. Key expiring soon — unread
		{
			ID:        idNotif2,
			Channel:   domainnotification.ChannelInApp,
			EventType: domainnotification.EventKeyExpiringSoon,
			UserID:    contract.IDMemberAdmin,
			Title:     "API Key 即将到期",
			Body:      "Key「prod-gw」将于 7 天后到期，请及时更换",
			Payload:   mustJSON(map[string]any{"keyID": contract.IDPlatformKey1.String(), "keyName": "prod-gw", "daysLeft": 7}),
			SendOK:    true,
			Category:  domainnotification.CategoryKeyExpiration,
			GroupKey:  fmt.Sprintf("key_expiry:%s", contract.IDPlatformKey1),
			CreatedAt: t("2026-07-27 08:00:00"),
		},
		// 3. Budget alert — unread, same group as #1 (tests grouping)
		{
			ID:        idNotif3,
			Channel:   domainnotification.ChannelInApp,
			EventType: domainnotification.EventBudgetAlertReached,
			UserID:    contract.IDMemberAdmin,
			Title:     "预算使用率达到 80%",
			Body:      "项目「智能客服」本月预算已使用 ¥2,080 / ¥2,600",
			Payload:   mustJSON(map[string]any{"ruleID": contract.IDAlertRule1.String(), "projectID": contract.IDProject1.String(), "threshold": 80}),
			SendOK:    true,
			Category:  domainnotification.CategoryBudgetAlert,
			GroupKey:  fmt.Sprintf("budget:%s:%s", contract.IDAlertRule1, contract.DemoBudgetPeriod),
			CreatedAt: t("2026-07-26 14:00:00"),
		},
		// 4. Overrun blocked — unread
		{
			ID:        idNotif4,
			Channel:   domainnotification.ChannelInApp,
			EventType: domainnotification.EventOverrunBlocked,
			UserID:    contract.IDMemberAdmin,
			Title:     "超支已拦截",
			Body:      "项目「数据分析平台」预算耗尽，已自动停用相关 Key",
			Payload:   mustJSON(map[string]any{"projectID": contract.IDProject4.String(), "keyID": contract.IDPlatformKey3.String()}),
			SendOK:    true,
			Category:  domainnotification.CategoryOverrun,
			GroupKey:  fmt.Sprintf("overrun:%s:2026-07-25", contract.IDProject4),
			CreatedAt: t("2026-07-25 16:45:00"),
		},
		// 5. Security event — read
		{
			ID:        idNotif5,
			Channel:   domainnotification.ChannelInApp,
			EventType: domainnotification.EventSecurityLoginNewDevice,
			UserID:    contract.IDMemberAdmin,
			Title:     "新设备登录",
			Body:      "检测到从新设备（macOS / Chrome）登录您的账号",
			Payload:   mustJSON(map[string]any{"device": "macOS", "browser": "Chrome", "ip": "203.0.113.42"}),
			SendOK:    true,
			Category:  domainnotification.CategorySecurityEvent,
			GroupKey:  "",
			CreatedAt: t("2026-07-24 10:15:00"),
		},
		// 6. System maintenance — read
		{
			ID:        idNotif6,
			Channel:   domainnotification.ChannelInApp,
			EventType: domainnotification.EventSystemMaintenanceScheduled,
			UserID:    contract.IDMemberAdmin,
			Title:     "系统维护通知",
			Body:      "计划于 7月28日 02:00-04:00 进行系统升级，届时服务可能短暂中断",
			Payload:   mustJSON(map[string]any{"startTime": "2026-07-28T02:00:00+08:00", "endTime": "2026-07-28T04:00:00+08:00"}),
			SendOK:    true,
			Category:  domainnotification.CategorySystemMaintenance,
			GroupKey:  "maintenance:2026-07-28",
			CreatedAt: t("2026-07-23 18:00:00"),
		},
		// 7. Usage weekly report — read
		{
			ID:        idNotif7,
			Channel:   domainnotification.ChannelInApp,
			EventType: domainnotification.EventUsageWeeklyReport,
			UserID:    contract.IDMemberAdmin,
			Title:     "本周用量报告",
			Body:      "本周总消耗 ¥1,256，较上周增长 12%",
			Payload:   mustJSON(map[string]any{"weekStart": "2026-07-14", "weekEnd": "2026-07-20", "totalCost": 1256, "growthPct": 12}),
			SendOK:    true,
			Category:  domainnotification.CategoryUsageReport,
			GroupKey:  "",
			CreatedAt: t("2026-07-21 09:00:00"),
		},
		// 8. Key expired — read + archived (tests archived tab)
		{
			ID:        idNotif8,
			Channel:   domainnotification.ChannelInApp,
			EventType: domainnotification.EventKeyExpired,
			UserID:    contract.IDMemberAdmin,
			Title:     "API Key 已过期",
			Body:      "Key「test-dev」已过期并自动停用",
			Payload:   mustJSON(map[string]any{"keyID": contract.IDPlatformKey2.String(), "keyName": "test-dev"}),
			SendOK:    true,
			Category:  domainnotification.CategoryKeyExpiration,
			GroupKey:  fmt.Sprintf("key_expiry:%s", contract.IDPlatformKey2),
			CreatedAt: t("2026-07-18 11:30:00"),
		},
	}
}
