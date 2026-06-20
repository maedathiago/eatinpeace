package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/maedathiago/eatinpeace/internal/application"
	"github.com/maedathiago/eatinpeace/internal/domain"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) http.Handler {
	return &Handler{service: service}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	if path == "" {
		path = "/"
	}
	switch {
	case r.Method == http.MethodGet && path == "/healthz":
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	case path == "/v1/events" && r.Method == http.MethodPost:
		h.createEvent(w, r)
	case path == "/v1/table-sessions" && r.Method == http.MethodPost:
		h.openTableSession(w, r)
	case strings.HasPrefix(path, "/v1/table-sessions/") && strings.HasSuffix(path, "/timeline") && r.Method == http.MethodGet:
		h.getTimeline(w, r, strings.TrimSuffix(strings.TrimPrefix(path, "/v1/table-sessions/"), "/timeline"))
	case path == "/v1/orders" && r.Method == http.MethodPost:
		h.createOrder(w, r)
	case strings.HasPrefix(path, "/v1/orders/") && strings.HasSuffix(path, "/status") && r.Method == http.MethodPatch:
		h.updateOrderStatus(w, r, strings.TrimSuffix(strings.TrimPrefix(path, "/v1/orders/"), "/status"))
	case path == "/v1/floor-tasks" && r.Method == http.MethodPost:
		h.createFloorTask(w, r)
	case path == "/v1/floor-tasks" && r.Method == http.MethodGet:
		h.listFloorTasks(w, r)
	case strings.HasPrefix(path, "/v1/floor-tasks/") && r.Method == http.MethodPatch:
		h.updateFloorTask(w, r, strings.TrimPrefix(path, "/v1/floor-tasks/"))
	case path == "/v1/complaints" && r.Method == http.MethodPost:
		h.openComplaint(w, r)
	case strings.HasPrefix(path, "/v1/complaints/") && r.Method == http.MethodPatch:
		h.updateComplaint(w, r, strings.TrimPrefix(path, "/v1/complaints/"))
	case path == "/v1/bill-handoffs" && r.Method == http.MethodPost:
		h.requestBillHandoff(w, r)
	case strings.HasPrefix(path, "/v1/bill-handoffs/") && r.Method == http.MethodPatch:
		h.updateBillHandoff(w, r, strings.TrimPrefix(path, "/v1/bill-handoffs/"))
	case path == "/v1/sla/evaluate" && r.Method == http.MethodPost:
		h.evaluateSLA(w, r)
	case strings.HasPrefix(path, "/v1/service-shifts/") && strings.HasSuffix(path, "/close") && r.Method == http.MethodPost:
		h.closeShift(w, r, strings.TrimSuffix(strings.TrimPrefix(path, "/v1/service-shifts/"), "/close"))
	case strings.HasPrefix(path, "/v1/service-shifts/") && strings.HasSuffix(path, "/metrics") && r.Method == http.MethodGet:
		h.getMetrics(w, r, strings.TrimSuffix(strings.TrimPrefix(path, "/v1/service-shifts/"), "/metrics"))
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (h *Handler) createEvent(w http.ResponseWriter, r *http.Request) {
	var event domain.OperationalEvent
	if !decodeJSON(w, r, &event) {
		return
	}
	created, err := h.service.CreateEvent(r.Context(), event)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) openTableSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RestaurantID string        `json:"restaurant_id"`
		ShiftID      string        `json:"shift_id"`
		TableID      string        `json:"table_id"`
		Source       domain.Source `json:"source"`
		ActorID      string        `json:"actor_id"`
		QRToken      string        `json:"qr_token"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.OpenTableSession(r.Context(), application.OpenTableSessionCommand{
		RestaurantID: req.RestaurantID,
		ShiftID:      req.ShiftID,
		TableID:      req.TableID,
		Source:       req.Source,
		ActorID:      req.ActorID,
		QRToken:      req.QRToken,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, writeResponse(result.Resource, result))
}

func (h *Handler) getTimeline(w http.ResponseWriter, r *http.Request, id string) {
	timeline, err := h.service.Timeline(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, timeline)
}

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RestaurantID   string                        `json:"restaurant_id"`
		ShiftID        string                        `json:"shift_id"`
		TableID        string                        `json:"table_id"`
		TableSessionID string                        `json:"table_session_id"`
		Source         domain.Source                 `json:"source"`
		ActorID        string                        `json:"actor_id"`
		Items          []application.CreateOrderItem `json:"items"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.CreateOrder(r.Context(), application.CreateOrderCommand{
		RestaurantID:   req.RestaurantID,
		ShiftID:        req.ShiftID,
		TableID:        req.TableID,
		TableSessionID: req.TableSessionID,
		Source:         req.Source,
		ActorID:        req.ActorID,
		Items:          req.Items,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, writeResponse(result.Resource, result))
}

func (h *Handler) updateOrderStatus(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Status   domain.OrderStatus `json:"status"`
		Source   domain.Source      `json:"source"`
		ActorID  string             `json:"actor_id"`
		Reason   string             `json:"reason"`
		Override bool               `json:"override"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.UpdateOrderStatus(r.Context(), application.UpdateOrderStatusCommand{
		OrderID:  id,
		Status:   req.Status,
		Source:   req.Source,
		ActorID:  req.ActorID,
		Reason:   req.Reason,
		Override: req.Override,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, writeResponse(result.Resource, result))
}

func (h *Handler) createFloorTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RestaurantID         string          `json:"restaurant_id"`
		ShiftID              string          `json:"shift_id"`
		TableID              string          `json:"table_id"`
		TableSessionID       string          `json:"table_session_id"`
		Type                 domain.TaskType `json:"type"`
		Severity             domain.Severity `json:"severity"`
		PriorityReason       string          `json:"priority_reason"`
		Source               domain.Source   `json:"source"`
		ActorID              string          `json:"actor_id"`
		RelatedOrderID       string          `json:"related_order_id"`
		RelatedComplaintID   string          `json:"related_complaint_id"`
		RelatedBillHandoffID string          `json:"related_bill_handoff_id"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.CreateFloorTask(r.Context(), application.CreateFloorTaskCommand{
		RestaurantID:         req.RestaurantID,
		ShiftID:              req.ShiftID,
		TableID:              req.TableID,
		TableSessionID:       req.TableSessionID,
		Type:                 req.Type,
		Severity:             req.Severity,
		PriorityReason:       req.PriorityReason,
		Source:               req.Source,
		ActorID:              req.ActorID,
		RelatedOrderID:       req.RelatedOrderID,
		RelatedComplaintID:   req.RelatedComplaintID,
		RelatedBillHandoffID: req.RelatedBillHandoffID,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, writeResponse(result.Resource, result))
}

func (h *Handler) listFloorTasks(w http.ResponseWriter, r *http.Request) {
	status := domain.TaskStatus(r.URL.Query().Get("status"))
	tasks, err := h.service.ListFloorTasks(r.Context(), application.TaskFilter{
		ShiftID:       r.URL.Query().Get("shift_id"),
		TableID:       r.URL.Query().Get("table_id"),
		Status:        status,
		ResponsibleID: r.URL.Query().Get("assignee_id"),
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
}

func (h *Handler) updateFloorTask(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Action         string `json:"action"`
		ActorID        string `json:"actor_id"`
		AssigneeID     string `json:"assignee_id"`
		Reason         string `json:"reason"`
		ResolutionNote string `json:"resolution_note"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.UpdateFloorTask(r.Context(), application.UpdateFloorTaskCommand{
		TaskID:         id,
		Action:         req.Action,
		ActorID:        req.ActorID,
		AssigneeID:     req.AssigneeID,
		Reason:         req.Reason,
		ResolutionNote: req.ResolutionNote,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, writeResponse(result.Resource, result))
}

func (h *Handler) openComplaint(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RestaurantID   string          `json:"restaurant_id"`
		ShiftID        string          `json:"shift_id"`
		TableID        string          `json:"table_id"`
		TableSessionID string          `json:"table_session_id"`
		OrderID        string          `json:"related_order_id"`
		Source         domain.Source   `json:"source"`
		ActorID        string          `json:"actor_id"`
		Reason         string          `json:"reason_code"`
		Description    string          `json:"description"`
		Severity       domain.Severity `json:"severity"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.OpenComplaint(r.Context(), application.OpenComplaintCommand{
		RestaurantID:   req.RestaurantID,
		ShiftID:        req.ShiftID,
		TableID:        req.TableID,
		TableSessionID: req.TableSessionID,
		OrderID:        req.OrderID,
		Source:         req.Source,
		ActorID:        req.ActorID,
		Reason:         req.Reason,
		Description:    req.Description,
		Severity:       req.Severity,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, writeResponse(result.Resource, result))
}

func (h *Handler) updateComplaint(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Action            string          `json:"action"`
		ActorID           string          `json:"actor_id"`
		AssigneeID        string          `json:"assignee_id"`
		Severity          domain.Severity `json:"severity"`
		Reason            string          `json:"reason_code"`
		ResolutionSummary string          `json:"resolution_code"`
		Note              string          `json:"note"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.UpdateComplaint(r.Context(), application.UpdateComplaintCommand{
		ComplaintID:       id,
		Action:            req.Action,
		ActorID:           req.ActorID,
		AssigneeID:        req.AssigneeID,
		Severity:          req.Severity,
		Reason:            req.Reason,
		ResolutionSummary: req.ResolutionSummary,
		Note:              req.Note,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, writeResponse(result.Resource, result))
}

func (h *Handler) requestBillHandoff(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RestaurantID      string        `json:"restaurant_id"`
		ShiftID           string        `json:"shift_id"`
		TableID           string        `json:"table_id"`
		TableSessionID    string        `json:"table_session_id"`
		Source            domain.Source `json:"source"`
		ActorID           string        `json:"actor_id"`
		HandoffTarget     string        `json:"handoff_target"`
		ExternalReference string        `json:"external_reference"`
		Note              string        `json:"note"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.RequestBillHandoff(r.Context(), application.RequestBillHandoffCommand{
		RestaurantID:      req.RestaurantID,
		ShiftID:           req.ShiftID,
		TableID:           req.TableID,
		TableSessionID:    req.TableSessionID,
		Source:            req.Source,
		ActorID:           req.ActorID,
		HandoffTarget:     req.HandoffTarget,
		ExternalReference: req.ExternalReference,
		Note:              req.Note,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, writeResponse(result.Resource, result))
}

func (h *Handler) updateBillHandoff(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Action            string `json:"action"`
		ActorID           string `json:"actor_id"`
		AssigneeID        string `json:"assignee_id"`
		Reason            string `json:"reason"`
		ExternalReference string `json:"external_reference"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.UpdateBillHandoff(r.Context(), application.UpdateBillHandoffCommand{
		BillHandoffID:     id,
		Action:            req.Action,
		ActorID:           req.ActorID,
		AssigneeID:        req.AssigneeID,
		Reason:            req.Reason,
		ExternalReference: req.ExternalReference,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, writeResponse(result.Resource, result))
}

func (h *Handler) evaluateSLA(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ShiftID string `json:"shift_id"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	tasks, err := h.service.EvaluateSLA(r.Context(), req.ShiftID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
}

func (h *Handler) closeShift(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		ActorID string `json:"actor_id"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	result, err := h.service.CloseShift(r.Context(), id, req.ActorID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, writeResponse(result.Resource, result))
}

func (h *Handler) getMetrics(w http.ResponseWriter, r *http.Request, id string) {
	metrics, err := h.service.Metrics(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, metrics)
}

func writeResponse[T any](resource T, result application.WriteResult[T]) map[string]any {
	return map[string]any{
		"resource":         resource,
		"event_id":         result.EventID,
		"correlation_id":   result.CorrelationID,
		"customer_message": result.CustomerMessage,
		"floor_task_id":    result.FloorTaskID,
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, application.ErrInvalid):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, application.ErrConflict):
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
