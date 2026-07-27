package newapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"sms/backend/internal/domain/newapisync"
	newapiclient "sms/backend/internal/integration/newapi"
)

func TestClient_ListCurrentRatios(t *testing.T) {
	t.Parallel()

	modelRatio := map[string]float64{"gpt-4o": 30, "claude-3": 22.5}
	completionRatio := map[string]float64{"gpt-4o": 2, "claude-3": 2}

	mr, _ := json.Marshal(modelRatio)
	cr, _ := json.Marshal(completionRatio)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/option/" {
			http.Error(w, "not found", 404)
			return
		}
		resp := map[string]any{
			"data": []map[string]string{{
				"ModelRatio":      string(mr),
				"CompletionRatio": string(cr),
			}},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := newapiclient.NewClientForTest(srv.URL, "test-token")
	ratios, err := client.ListCurrentRatios(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(ratios) != 2 {
		t.Fatalf("expected 2 ratios, got %d", len(ratios))
	}

	// gpt-4o: ratio=30, CR=2 → input=60, output=120
	gpt := ratios["gpt-4o"]
	if gpt.InputPrice != 60 {
		t.Fatalf("expected input 60, got %f", gpt.InputPrice)
	}
	if gpt.OutputPrice != 120 {
		t.Fatalf("expected output 120, got %f", gpt.OutputPrice)
	}

	// claude-3: ratio=22.5, CR=2 → input=45, output=90
	claude := ratios["claude-3"]
	if claude.InputPrice != 45 {
		t.Fatalf("expected input 45, got %f", claude.InputPrice)
	}
	if claude.OutputPrice != 90 {
		t.Fatalf("expected output 90, got %f", claude.OutputPrice)
	}
}

func TestClient_SyncPricing_MergesCorrectly(t *testing.T) {
	t.Parallel()

	existingMR := map[string]float64{"existing-model": 5}
	existingCR := map[string]float64{"existing-model": 1.5}

	var putCalls int32
	var writtenMR, writtenCR map[string]float64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/option/" {
			mr, _ := json.Marshal(existingMR)
			cr, _ := json.Marshal(existingCR)
			resp := map[string]any{
				"data": []map[string]string{
					{"key": "ModelRatio", "value": string(mr)},
					{"key": "CompletionRatio", "value": string(cr)},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		if r.Method == http.MethodPut && r.URL.Path == "/api/option/" {
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			call := atomic.AddInt32(&putCalls, 1)
			if call == 1 {
				json.Unmarshal([]byte(body["value"]), &writtenMR)
			} else {
				json.Unmarshal([]byte(body["value"]), &writtenCR)
			}
			w.WriteHeader(200)
			w.Write([]byte(`{}`))
			return
		}
		http.Error(w, "not found", 404)
	}))
	defer srv.Close()

	client := newapiclient.NewClientForTest(srv.URL, "test-token")
	err := client.SyncPricing(context.Background(), []newapisync.PricingEntry{
		{ModelID: "gpt-4o", InputPrice: 60, OutputPrice: 120},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify merge: existing-model preserved + gpt-4o added
	if writtenMR["existing-model"] != 5 {
		t.Fatalf("expected existing-model ratio preserved, got %f", writtenMR["existing-model"])
	}
	// gpt-4o: inputPrice=60 → ratio=30
	if writtenMR["gpt-4o"] != 30 {
		t.Fatalf("expected gpt-4o ratio 30, got %f", writtenMR["gpt-4o"])
	}
	if writtenCR["existing-model"] != 1.5 {
		t.Fatalf("expected existing-model CR preserved, got %f", writtenCR["existing-model"])
	}
	// gpt-4o: outputPrice/inputPrice = 120/60 = 2
	if writtenCR["gpt-4o"] != 2 {
		t.Fatalf("expected gpt-4o CR 2, got %f", writtenCR["gpt-4o"])
	}
}

func TestClient_SyncPricing_SkipsZeroInputPrice(t *testing.T) {
	t.Parallel()

	var putCalled bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			resp := map[string]any{"data": []map[string]string{{}}}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		if r.Method == http.MethodPut {
			putCalled = true
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			// The entry with 0 input price should be skipped, map should be empty
			var m map[string]float64
			json.Unmarshal([]byte(body["value"]), &m)
			if len(m) != 0 {
				t.Errorf("expected empty map, got %v", m)
			}
			w.WriteHeader(200)
			w.Write([]byte(`{}`))
			return
		}
	}))
	defer srv.Close()

	client := newapiclient.NewClientForTest(srv.URL, "test-token")
	err := client.SyncPricing(context.Background(), []newapisync.PricingEntry{
		{ModelID: "bad-model", InputPrice: 0, OutputPrice: 50},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !putCalled {
		t.Fatal("expected PUT to be called (writing empty maps)")
	}
}

func TestClient_UpsertModelRatio(t *testing.T) {
	t.Parallel()

	var writtenMR map[string]float64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			resp := map[string]any{"data": []map[string]string{{
				"ModelRatio":      `{"other": 10}`,
				"CompletionRatio": `{"other": 1}`,
			}}}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		if r.Method == http.MethodPut {
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			if body["key"] == "ModelRatio" {
				json.Unmarshal([]byte(body["value"]), &writtenMR)
			}
			w.WriteHeader(200)
			w.Write([]byte(`{}`))
			return
		}
	}))
	defer srv.Close()

	client := newapiclient.NewClientForTest(srv.URL, "test-token")
	err := client.UpsertModelRatio(context.Background(), "deepseek-chat", 2, 4)
	if err != nil {
		t.Fatal(err)
	}

	// deepseek-chat: ratio = 2/2 = 1, existing "other" = 10 preserved
	if writtenMR["other"] != 10 {
		t.Fatalf("expected other=10, got %f", writtenMR["other"])
	}
	if writtenMR["deepseek-chat"] != 1 {
		t.Fatalf("expected deepseek-chat=1, got %f", writtenMR["deepseek-chat"])
	}
}
