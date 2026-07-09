package billing_test

import (
	"testing"

	"github.com/tokenjoy/backend/seed/runtime"
)

func TestListRechargeRecordsSeeded(t *testing.T) {
	t.Parallel()
	svc, st, ctx := newBillingService(t, nil)
	if err := runtime.ApplyRechargeOrders(ctx, st); err != nil {
		t.Fatal(err)
	}
	records, err := svc.ListRechargeRecords(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 5 {
		t.Fatalf("expected 5 seeded records, got %d", len(records))
	}
	if records[0].OrderID != "ORD202606190001" {
		t.Fatalf("expected newest order first, got %+v", records[0])
	}
	if records[0].Method != "alipay" || records[0].InvoiceStatus != "none" {
		t.Fatalf("expected persisted display fields, got %+v", records[0])
	}
}

func TestListRechargeRecordsMapsPendingStatus(t *testing.T) {
	t.Parallel()
	svc, st, ctx := newBillingService(t, nil)
	if err := runtime.ApplyRechargeOrders(ctx, st); err != nil {
		t.Fatal(err)
	}
	records, err := svc.ListRechargeRecords(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range records {
		if record.ID == "tu-4" {
			if record.Status != "pending" || record.PaidAmount != 0 {
				t.Fatalf("expected pending unpaid record, got %+v", record)
			}
			return
		}
		if record.ID == "tu-1" && record.Status != "confirmed" {
			t.Fatalf("expected confirmed status, got %+v", record)
		}
	}
	t.Fatal("expected tu-4 in records")
}
