package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/maedathiago/eatinpeace/internal/application"
	"github.com/maedathiago/eatinpeace/internal/domain"
	"github.com/maedathiago/eatinpeace/internal/httpapi"
	"github.com/maedathiago/eatinpeace/internal/testsupport"
)

type writeResp[T any] struct {
	Resource        T      `json:"resource"`
	EventID         string `json:"event_id"`
	CorrelationID   string `json:"correlation_id"`
	CustomerMessage string `json:"customer_message"`
	FloorTaskID     string `json:"floor_task_id"`
}

func TestP0E2E(t *testing.T) {
	fixture := testsupport.NewFixture(t)
	handler := httpapi.NewHandler(fixture.Service)

	var sessionResp writeResp[domain.TableSession]
	body := doJSON(t, handler, http.MethodPost, "/v1/table-sessions", map[string]any{
		"restaurant_id": "rest_pilot",
		"shift_id":      "shift_pilot_open",
		"table_id":      "table_01",
		"source":        "customer",
		"qr_token":      "qr-demo",
	}, http.StatusCreated, &sessionResp)
	assertNoFinancialScope(t, body)
	if sessionResp.EventID == "" || sessionResp.CorrelationID == "" {
		t.Fatalf("session response missed audit fields: %#v", sessionResp)
	}

	var orderResp writeResp[domain.Order]
	doJSON(t, handler, http.MethodPost, "/v1/orders", map[string]any{
		"restaurant_id":    "rest_pilot",
		"shift_id":         "shift_pilot_open",
		"table_id":         "table_01",
		"table_session_id": sessionResp.Resource.ID,
		"source":           "staff",
		"actor_id":         "staff_waiter",
		"items": []map[string]any{
			{"name": "Prato piloto", "quantity": 1},
		},
	}, http.StatusCreated, &orderResp)

	var blockedBill writeResp[domain.BillHandoff]
	doJSON(t, handler, http.MethodPost, "/v1/bill-handoffs", map[string]any{
		"restaurant_id":    "rest_pilot",
		"shift_id":         "shift_pilot_open",
		"table_id":         "table_01",
		"table_session_id": sessionResp.Resource.ID,
		"source":           "customer",
		"handoff_target":   "cashier",
	}, http.StatusCreated, &blockedBill)
	if blockedBill.Resource.Status != domain.BillBlocked {
		t.Fatalf("bill with pending order status = %s, want blocked", blockedBill.Resource.Status)
	}

	var orderStatus writeResp[domain.Order]
	doJSON(t, handler, http.MethodPatch, "/v1/orders/"+orderResp.Resource.ID+"/status", map[string]any{
		"status":   "preparing",
		"source":   "kitchen",
		"actor_id": "staff_kitchen",
	}, http.StatusOK, &orderStatus)

	fixture.Advance(11 * time.Minute)
	var slaResp struct {
		Tasks []domain.FloorTask `json:"tasks"`
	}
	doJSON(t, handler, http.MethodPost, "/v1/sla/evaluate", map[string]any{"shift_id": "shift_pilot_open"}, http.StatusOK, &slaResp)
	if len(slaResp.Tasks) == 0 {
		t.Fatal("SLA evaluator did not create tasks")
	}

	doJSON(t, handler, http.MethodPatch, "/v1/orders/"+orderResp.Resource.ID+"/status", map[string]any{
		"status":   "ready",
		"source":   "kitchen",
		"actor_id": "staff_kitchen",
	}, http.StatusOK, &orderStatus)
	if orderStatus.FloorTaskID == "" {
		t.Fatal("ready order did not return pickup task id")
	}
	doJSON(t, handler, http.MethodPatch, "/v1/orders/"+orderResp.Resource.ID+"/status", map[string]any{
		"status":   "picked_up",
		"source":   "staff",
		"actor_id": "staff_waiter",
	}, http.StatusOK, &orderStatus)
	doJSON(t, handler, http.MethodPatch, "/v1/orders/"+orderResp.Resource.ID+"/status", map[string]any{
		"status":   "delivered",
		"source":   "staff",
		"actor_id": "staff_waiter",
	}, http.StatusOK, &orderStatus)

	var complaintResp writeResp[domain.Complaint]
	doJSON(t, handler, http.MethodPost, "/v1/complaints", map[string]any{
		"restaurant_id":    "rest_pilot",
		"shift_id":         "shift_pilot_open",
		"table_id":         "table_01",
		"table_session_id": sessionResp.Resource.ID,
		"source":           "customer",
		"reason_code":      "service_delay",
		"description":      "Demorou e ninguem explicou.",
		"severity":         "high",
	}, http.StatusCreated, &complaintResp)
	if complaintResp.FloorTaskID == "" {
		t.Fatal("complaint did not create linked task")
	}

	doJSON(t, handler, http.MethodPost, "/v1/bill-handoffs", map[string]any{
		"restaurant_id":    "rest_pilot",
		"shift_id":         "shift_pilot_open",
		"table_id":         "table_01",
		"table_session_id": sessionResp.Resource.ID,
		"source":           "customer",
		"handoff_target":   "cashier",
	}, http.StatusCreated, &blockedBill)
	if blockedBill.Resource.Status != domain.BillBlocked {
		t.Fatalf("bill with open complaint status = %s, want blocked", blockedBill.Resource.Status)
	}

	var complaintPatch writeResp[domain.Complaint]
	doJSON(t, handler, http.MethodPatch, "/v1/complaints/"+complaintResp.Resource.ID, map[string]any{
		"action":      "assign",
		"actor_id":    "staff_lead",
		"assignee_id": "staff_waiter",
	}, http.StatusOK, &complaintPatch)
	doJSON(t, handler, http.MethodPatch, "/v1/complaints/"+complaintResp.Resource.ID, map[string]any{
		"action":   "record_first_response",
		"actor_id": "staff_waiter",
		"note":     "Garcom conversou com a mesa.",
	}, http.StatusOK, &complaintPatch)
	doJSON(t, handler, http.MethodPatch, "/v1/complaints/"+complaintResp.Resource.ID, map[string]any{
		"action":          "resolve",
		"actor_id":        "staff_waiter",
		"resolution_code": "explicado_e_resolvido",
	}, http.StatusOK, &complaintPatch)
	doJSON(t, handler, http.MethodPatch, "/v1/complaints/"+complaintResp.Resource.ID, map[string]any{
		"action":   "reopen",
		"actor_id": "staff_lead",
		"note":     "Cliente pediu nova verificacao.",
	}, http.StatusOK, &complaintPatch)
	doJSON(t, handler, http.MethodPatch, "/v1/complaints/"+complaintResp.Resource.ID, map[string]any{
		"action":          "resolve",
		"actor_id":        "staff_lead",
		"resolution_code": "lider_confirmou_resolucao",
	}, http.StatusOK, &complaintPatch)

	var billResp writeResp[domain.BillHandoff]
	body = doJSON(t, handler, http.MethodPost, "/v1/bill-handoffs", map[string]any{
		"restaurant_id":    "rest_pilot",
		"shift_id":         "shift_pilot_open",
		"table_id":         "table_01",
		"table_session_id": sessionResp.Resource.ID,
		"source":           "customer",
		"handoff_target":   "cashier",
	}, http.StatusCreated, &billResp)
	assertNoFinancialScope(t, body)
	if billResp.Resource.Status != domain.BillRequested {
		t.Fatalf("bill status = %s, want requested", billResp.Resource.Status)
	}
	doJSON(t, handler, http.MethodPatch, "/v1/bill-handoffs/"+billResp.Resource.ID, map[string]any{
		"action":   "accept",
		"actor_id": "staff_cashier",
	}, http.StatusOK, &billResp)
	doJSON(t, handler, http.MethodPatch, "/v1/bill-handoffs/"+billResp.Resource.ID, map[string]any{
		"action":   "confirm_close",
		"actor_id": "staff_cashier",
	}, http.StatusOK, &billResp)
	if billResp.Resource.Status != domain.BillClosed {
		t.Fatalf("bill status = %s, want closed", billResp.Resource.Status)
	}

	var timeline application.Timeline
	doJSON(t, handler, http.MethodGet, "/v1/table-sessions/"+sessionResp.Resource.ID+"/timeline", nil, http.StatusOK, &timeline)
	if timeline.TableSession.Status != domain.TableSessionClosed || len(timeline.Events) == 0 {
		t.Fatalf("timeline did not reflect closed session/events: %#v", timeline)
	}

	var closeResp writeResp[domain.ServiceShift]
	doJSON(t, handler, http.MethodPost, "/v1/service-shifts/shift_pilot_open/close", map[string]any{
		"actor_id": "staff_lead",
	}, http.StatusOK, &closeResp)

	var metrics domain.ShiftMetrics
	doJSON(t, handler, http.MethodGet, "/v1/service-shifts/shift_pilot_open/metrics", nil, http.StatusOK, &metrics)
	if !metrics.Closed || metrics.TotalEvents == 0 {
		t.Fatalf("metrics missing closed shift evidence: %#v", metrics)
	}
}

func doJSON(t *testing.T, handler http.Handler, method, path string, payload any, wantStatus int, out any) []byte {
	t.Helper()
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatalf("encode payload: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	gotBody := rec.Body.Bytes()
	if rec.Code != wantStatus {
		t.Fatalf("%s %s status = %d want %d body=%s", method, path, rec.Code, wantStatus, gotBody)
	}
	if out != nil {
		if err := json.Unmarshal(gotBody, out); err != nil {
			t.Fatalf("decode response for %s %s: %v body=%s", method, path, err, gotBody)
		}
	}
	return gotBody
}

func assertNoFinancialScope(t *testing.T, body []byte) {
	t.Helper()
	for _, forbidden := range []string{"payment", "card", "tax", "invoice", "fiscal", "reconciliation", "inventory"} {
		if strings.Contains(strings.ToLower(string(body)), forbidden) {
			t.Fatalf("response leaked forbidden scope %q: %s", forbidden, body)
		}
	}
}
