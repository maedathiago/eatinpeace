package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/maedathiago/eatinpeace/internal/domain"
)

var (
	ErrNotFound  = errors.New("not found")
	ErrInvalid   = errors.New("invalid command")
	ErrConflict  = errors.New("conflict")
	ErrForbidden = errors.New("forbidden")
)

type Store interface {
	SaveRestaurant(context.Context, domain.Restaurant) error
	SaveShift(context.Context, domain.ServiceShift) error
	GetShift(context.Context, string) (domain.ServiceShift, error)
	SaveTable(context.Context, domain.Table) error
	GetTable(context.Context, string) (domain.Table, error)
	SaveStaffMember(context.Context, domain.StaffMember) error

	AppendEvent(context.Context, domain.OperationalEvent) error
	ListEvents(context.Context, EventFilter) ([]domain.OperationalEvent, error)

	SaveTableSession(context.Context, domain.TableSession) error
	GetTableSession(context.Context, string) (domain.TableSession, error)
	ListTableSessions(context.Context, string) ([]domain.TableSession, error)

	SaveOrder(context.Context, domain.Order) error
	GetOrder(context.Context, string) (domain.Order, error)
	ListOrders(context.Context, OrderFilter) ([]domain.Order, error)

	SaveFloorTask(context.Context, domain.FloorTask) error
	GetFloorTask(context.Context, string) (domain.FloorTask, error)
	ListFloorTasks(context.Context, TaskFilter) ([]domain.FloorTask, error)

	SaveComplaint(context.Context, domain.Complaint) error
	GetComplaint(context.Context, string) (domain.Complaint, error)
	ListComplaints(context.Context, ComplaintFilter) ([]domain.Complaint, error)

	SaveBillHandoff(context.Context, domain.BillHandoff) error
	GetBillHandoff(context.Context, string) (domain.BillHandoff, error)
	ListBillHandoffs(context.Context, BillFilter) ([]domain.BillHandoff, error)
}

type EventFilter struct {
	RestaurantID   string
	ShiftID        string
	TableSessionID string
}

type OrderFilter struct {
	ShiftID        string
	TableSessionID string
}

type TaskFilter struct {
	ShiftID              string
	TableID              string
	Status               domain.TaskStatus
	ResponsibleID        string
	RelatedOrderID       string
	RelatedComplaintID   string
	RelatedBillHandoffID string
}

type ComplaintFilter struct {
	ShiftID        string
	TableSessionID string
}

type BillFilter struct {
	ShiftID        string
	TableSessionID string
}

type Service struct {
	store    Store
	now      func() time.Time
	policies domain.SLAPolicies
}

func NewService(store Store) *Service {
	return &Service{
		store:    store,
		now:      func() time.Time { return time.Now().UTC() },
		policies: domain.DefaultSLAPolicies(),
	}
}

func (s *Service) SetClock(now func() time.Time) {
	s.now = now
}

func (s *Service) SetSLAPolicies(policies domain.SLAPolicies) {
	s.policies = policies
}

func (s *Service) SeedPilotFixtures(ctx context.Context) error {
	now := s.now()
	restaurant := domain.Restaurant{ID: "rest_pilot", Name: "Restaurante Piloto"}
	shift := domain.ServiceShift{
		ID:           "shift_pilot_open",
		RestaurantID: restaurant.ID,
		Name:         "Turno piloto",
		OpenedAt:     now,
		Status:       domain.ShiftOpen,
	}
	tables := []domain.Table{
		{ID: "table_01", RestaurantID: restaurant.ID, Label: "Mesa 1"},
		{ID: "table_02", RestaurantID: restaurant.ID, Label: "Mesa 2"},
		{ID: "table_03", RestaurantID: restaurant.ID, Label: "Mesa 3"},
	}
	staff := []domain.StaffMember{
		{ID: "staff_waiter", RestaurantID: restaurant.ID, Name: "Garcom Piloto", Role: "waiter"},
		{ID: "staff_lead", RestaurantID: restaurant.ID, Name: "Lider de Sala", Role: "floor_lead"},
		{ID: "staff_kitchen", RestaurantID: restaurant.ID, Name: "Cozinha", Role: "kitchen"},
		{ID: "staff_cashier", RestaurantID: restaurant.ID, Name: "Caixa", Role: "cashier"},
	}
	if err := s.store.SaveRestaurant(ctx, restaurant); err != nil {
		return err
	}
	if err := s.store.SaveShift(ctx, shift); err != nil {
		return err
	}
	for _, table := range tables {
		if err := s.store.SaveTable(ctx, table); err != nil {
			return err
		}
	}
	for _, member := range staff {
		if err := s.store.SaveStaffMember(ctx, member); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) CreateEvent(ctx context.Context, event domain.OperationalEvent) (domain.OperationalEvent, error) {
	event = s.prepareEvent(event)
	if err := event.Validate(); err != nil {
		return domain.OperationalEvent{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if err := s.store.AppendEvent(ctx, event); err != nil {
		return domain.OperationalEvent{}, err
	}
	return event, nil
}

type OpenTableSessionCommand struct {
	RestaurantID string
	ShiftID      string
	TableID      string
	Source       domain.Source
	ActorID      string
	QRToken      string
}

type WriteResult[T any] struct {
	Resource        T      `json:"resource"`
	EventID         string `json:"event_id"`
	CorrelationID   string `json:"correlation_id"`
	CustomerMessage string `json:"customer_message,omitempty"`
	FloorTaskID     string `json:"floor_task_id,omitempty"`
}

func (s *Service) OpenTableSession(ctx context.Context, cmd OpenTableSessionCommand) (WriteResult[domain.TableSession], error) {
	if strings.TrimSpace(cmd.RestaurantID) == "" || strings.TrimSpace(cmd.ShiftID) == "" || strings.TrimSpace(cmd.TableID) == "" {
		return WriteResult[domain.TableSession]{}, fmt.Errorf("%w: restaurant_id, shift_id and table_id are required", ErrInvalid)
	}
	if _, err := s.store.GetTable(ctx, cmd.TableID); err != nil {
		return WriteResult[domain.TableSession]{}, err
	}
	now := s.now()
	session := domain.TableSession{
		ID:           newID("ts"),
		RestaurantID: cmd.RestaurantID,
		ShiftID:      cmd.ShiftID,
		TableID:      cmd.TableID,
		Status:       domain.TableSessionActive,
		OpenedAt:     now,
	}
	event := s.prepareEvent(domain.OperationalEvent{
		RestaurantID:   cmd.RestaurantID,
		ShiftID:        cmd.ShiftID,
		TableID:        cmd.TableID,
		TableSessionID: session.ID,
		EventType:      domain.EventTableSessionOpened,
		OccurredAt:     now,
		Source:         defaultSource(cmd.Source, domain.SourceCustomer),
		ActorID:        cmd.ActorID,
		Payload:        map[string]any{"qr_token_present": cmd.QRToken != ""},
	})
	if err := event.Validate(); err != nil {
		return WriteResult[domain.TableSession]{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if err := s.store.SaveTableSession(ctx, session); err != nil {
		return WriteResult[domain.TableSession]{}, err
	}
	if err := s.store.AppendEvent(ctx, event); err != nil {
		return WriteResult[domain.TableSession]{}, err
	}
	return WriteResult[domain.TableSession]{
		Resource:        session,
		EventID:         event.ID,
		CorrelationID:   event.CorrelationID,
		CustomerMessage: "Sua mesa foi registrada. A equipe acompanha o andamento por aqui.",
	}, nil
}

type CreateOrderCommand struct {
	RestaurantID   string
	ShiftID        string
	TableID        string
	TableSessionID string
	Source         domain.Source
	ActorID        string
	Items          []CreateOrderItem
}

type CreateOrderItem struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Notes    string `json:"notes,omitempty"`
}

func (s *Service) CreateOrder(ctx context.Context, cmd CreateOrderCommand) (WriteResult[domain.Order], error) {
	if strings.TrimSpace(cmd.RestaurantID) == "" || strings.TrimSpace(cmd.ShiftID) == "" || strings.TrimSpace(cmd.TableID) == "" || strings.TrimSpace(cmd.TableSessionID) == "" {
		return WriteResult[domain.Order]{}, fmt.Errorf("%w: restaurant_id, shift_id, table_id and table_session_id are required", ErrInvalid)
	}
	if len(cmd.Items) == 0 {
		return WriteResult[domain.Order]{}, fmt.Errorf("%w: at least one order item is required", ErrInvalid)
	}
	now := s.now()
	orderID := newID("ord")
	items := make([]domain.OrderItem, 0, len(cmd.Items))
	for _, item := range cmd.Items {
		if strings.TrimSpace(item.Name) == "" {
			return WriteResult[domain.Order]{}, fmt.Errorf("%w: order item name is required", ErrInvalid)
		}
		quantity := item.Quantity
		if quantity <= 0 {
			quantity = 1
		}
		items = append(items, domain.OrderItem{ID: newID("item"), OrderID: orderID, Name: item.Name, Quantity: quantity, Notes: item.Notes})
	}
	order := domain.Order{
		ID:             orderID,
		RestaurantID:   cmd.RestaurantID,
		ShiftID:        cmd.ShiftID,
		TableID:        cmd.TableID,
		TableSessionID: cmd.TableSessionID,
		Status:         domain.OrderReceived,
		Items:          items,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	event := s.prepareEvent(domain.OperationalEvent{
		RestaurantID:   cmd.RestaurantID,
		ShiftID:        cmd.ShiftID,
		TableID:        cmd.TableID,
		TableSessionID: cmd.TableSessionID,
		EventType:      domain.EventOrderCreated,
		OccurredAt:     now,
		Source:         defaultSource(cmd.Source, domain.SourceStaff),
		ActorID:        cmd.ActorID,
		Payload:        map[string]any{"order_id": order.ID, "items_count": len(items), "status": string(order.Status)},
	})
	if err := event.Validate(); err != nil {
		return WriteResult[domain.Order]{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if err := s.store.SaveOrder(ctx, order); err != nil {
		return WriteResult[domain.Order]{}, err
	}
	if err := s.store.AppendEvent(ctx, event); err != nil {
		return WriteResult[domain.Order]{}, err
	}
	return WriteResult[domain.Order]{
		Resource:        order,
		EventID:         event.ID,
		CorrelationID:   event.CorrelationID,
		CustomerMessage: "Pedido registrado. A equipe vai atualizar o status por aqui.",
	}, nil
}

type UpdateOrderStatusCommand struct {
	OrderID       string
	Status        domain.OrderStatus
	Source        domain.Source
	ActorID       string
	Reason        string
	Override      bool
	CorrelationID string
}

func (s *Service) UpdateOrderStatus(ctx context.Context, cmd UpdateOrderStatusCommand) (WriteResult[domain.Order], error) {
	order, err := s.store.GetOrder(ctx, cmd.OrderID)
	if err != nil {
		return WriteResult[domain.Order]{}, err
	}
	overrideReason := ""
	if cmd.Override {
		overrideReason = cmd.Reason
		if strings.TrimSpace(overrideReason) == "" {
			return WriteResult[domain.Order]{}, fmt.Errorf("%w: override reason is required", ErrInvalid)
		}
	}
	if err := domain.ValidateOrderTransition(order.Status, cmd.Status, overrideReason); err != nil {
		return WriteResult[domain.Order]{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	now := s.now()
	previous := order.Status
	order.Status = cmd.Status
	order.UpdatedAt = now
	if cmd.Status == domain.OrderReady {
		order.ReadyAt = &now
	}
	if cmd.Status == domain.OrderDelivered {
		order.DeliveredAt = &now
	}
	eventType := eventForOrderStatus(cmd.Status)
	event := s.prepareEvent(domain.OperationalEvent{
		RestaurantID:   order.RestaurantID,
		ShiftID:        order.ShiftID,
		TableID:        order.TableID,
		TableSessionID: order.TableSessionID,
		EventType:      eventType,
		OccurredAt:     now,
		Source:         defaultSource(cmd.Source, domain.SourceStaff),
		ActorID:        cmd.ActorID,
		CorrelationID:  cmd.CorrelationID,
		Payload: map[string]any{
			"order_id":        order.ID,
			"previous_status": string(previous),
			"status":          string(order.Status),
			"reason":          cmd.Reason,
			"override":        cmd.Override,
		},
	})
	if err := event.Validate(); err != nil {
		return WriteResult[domain.Order]{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if err := s.store.SaveOrder(ctx, order); err != nil {
		return WriteResult[domain.Order]{}, err
	}
	if err := s.store.AppendEvent(ctx, event); err != nil {
		return WriteResult[domain.Order]{}, err
	}
	result := WriteResult[domain.Order]{
		Resource:      order,
		EventID:       event.ID,
		CorrelationID: event.CorrelationID,
	}
	switch cmd.Status {
	case domain.OrderReady:
		task, err := s.createOrUpdateTask(ctx, domain.FloorTask{
			RestaurantID:   order.RestaurantID,
			ShiftID:        order.ShiftID,
			TableID:        order.TableID,
			TableSessionID: order.TableSessionID,
			Type:           domain.TaskOrderReady,
			Status:         domain.TaskOpen,
			Severity:       domain.SeverityMedium,
			PriorityReason: "Pedido pronto aguardando retirada",
			RelatedOrderID: order.ID,
			DueAt:          ptrTime(now.Add(s.policies.ReadyPickupAfter)),
			SourceEventID:  event.ID,
		})
		if err != nil {
			return WriteResult[domain.Order]{}, err
		}
		result.FloorTaskID = task.ID
		result.CustomerMessage = "Seu pedido ficou pronto e a equipe foi avisada para retirar."
	case domain.OrderPickedUp, domain.OrderDelivered:
		_ = s.resolveOpenTasksForOrder(ctx, order.ID, domain.TaskOrderReady, "Pedido retirado para entrega")
		result.CustomerMessage = "O status do pedido foi atualizado."
	case domain.OrderDelayRisk:
		result.CustomerMessage = "A equipe identificou risco de atraso e esta acompanhando."
	case domain.OrderDelayed:
		result.CustomerMessage = "A equipe foi avisada sobre o atraso do pedido."
	default:
		result.CustomerMessage = "O status do pedido foi atualizado."
	}
	return result, nil
}

func eventForOrderStatus(status domain.OrderStatus) domain.EventType {
	switch status {
	case domain.OrderDelayRisk:
		return domain.EventOrderDelayRiskDetected
	case domain.OrderDelayed:
		return domain.EventOrderDelayed
	case domain.OrderReady:
		return domain.EventOrderReady
	case domain.OrderPickedUp:
		return domain.EventOrderPickedUp
	case domain.OrderDelivered:
		return domain.EventOrderDelivered
	default:
		return domain.EventOrderStatusChanged
	}
}

type CreateFloorTaskCommand struct {
	RestaurantID         string
	ShiftID              string
	TableID              string
	TableSessionID       string
	Type                 domain.TaskType
	Severity             domain.Severity
	PriorityReason       string
	DueAt                *time.Time
	Source               domain.Source
	ActorID              string
	RelatedOrderID       string
	RelatedComplaintID   string
	RelatedBillHandoffID string
}

func (s *Service) CreateFloorTask(ctx context.Context, cmd CreateFloorTaskCommand) (WriteResult[domain.FloorTask], error) {
	if strings.TrimSpace(cmd.RestaurantID) == "" || strings.TrimSpace(cmd.ShiftID) == "" || strings.TrimSpace(string(cmd.Type)) == "" {
		return WriteResult[domain.FloorTask]{}, fmt.Errorf("%w: restaurant_id, shift_id and type are required", ErrInvalid)
	}
	if strings.TrimSpace(cmd.PriorityReason) == "" {
		return WriteResult[domain.FloorTask]{}, fmt.Errorf("%w: priority_reason is required", ErrInvalid)
	}
	now := s.now()
	task := domain.FloorTask{
		ID:                   newID("task"),
		RestaurantID:         cmd.RestaurantID,
		ShiftID:              cmd.ShiftID,
		TableID:              cmd.TableID,
		TableSessionID:       cmd.TableSessionID,
		Type:                 cmd.Type,
		Status:               domain.TaskOpen,
		Severity:             defaultSeverity(cmd.Severity, domain.SeverityMedium),
		PriorityReason:       cmd.PriorityReason,
		RelatedOrderID:       cmd.RelatedOrderID,
		RelatedComplaintID:   cmd.RelatedComplaintID,
		RelatedBillHandoffID: cmd.RelatedBillHandoffID,
		DueAt:                cmd.DueAt,
		CreatedAt:            now,
	}
	event := s.prepareEvent(domain.OperationalEvent{
		RestaurantID:   cmd.RestaurantID,
		ShiftID:        cmd.ShiftID,
		TableID:        cmd.TableID,
		TableSessionID: cmd.TableSessionID,
		EventType:      domain.EventFloorTaskCreated,
		OccurredAt:     now,
		Source:         defaultSource(cmd.Source, domain.SourceStaff),
		ActorID:        cmd.ActorID,
		Payload: map[string]any{
			"task_id":         task.ID,
			"type":            string(task.Type),
			"severity":        string(task.Severity),
			"priority_reason": task.PriorityReason,
		},
	})
	if err := event.Validate(); err != nil {
		return WriteResult[domain.FloorTask]{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	task.SourceEventID = event.ID
	if err := s.store.SaveFloorTask(ctx, task); err != nil {
		return WriteResult[domain.FloorTask]{}, err
	}
	if err := s.store.AppendEvent(ctx, event); err != nil {
		return WriteResult[domain.FloorTask]{}, err
	}
	return WriteResult[domain.FloorTask]{Resource: task, EventID: event.ID, CorrelationID: event.CorrelationID}, nil
}

type UpdateFloorTaskCommand struct {
	TaskID         string
	Action         string
	ActorID        string
	AssigneeID     string
	Reason         string
	ResolutionNote string
}

func (s *Service) UpdateFloorTask(ctx context.Context, cmd UpdateFloorTaskCommand) (WriteResult[domain.FloorTask], error) {
	if strings.TrimSpace(cmd.ActorID) == "" {
		return WriteResult[domain.FloorTask]{}, fmt.Errorf("%w: actor_id is required", ErrInvalid)
	}
	task, err := s.store.GetFloorTask(ctx, cmd.TaskID)
	if err != nil {
		return WriteResult[domain.FloorTask]{}, err
	}
	now := s.now()
	eventType := domain.EventFloorTaskClaimed
	switch cmd.Action {
	case "claim":
		task.Status = domain.TaskClaimed
		task.ResponsibleID = firstNonEmpty(cmd.AssigneeID, cmd.ActorID)
		task.ClaimedAt = &now
	case "start":
		task.Status = domain.TaskInProgress
		task.ResponsibleID = firstNonEmpty(task.ResponsibleID, cmd.AssigneeID, cmd.ActorID)
		task.StartedAt = &now
	case "reassign":
		if strings.TrimSpace(cmd.Reason) == "" {
			return WriteResult[domain.FloorTask]{}, fmt.Errorf("%w: reassignment reason is required", ErrInvalid)
		}
		if strings.TrimSpace(cmd.AssigneeID) == "" {
			return WriteResult[domain.FloorTask]{}, fmt.Errorf("%w: assignee_id is required", ErrInvalid)
		}
		task.ResponsibleID = cmd.AssigneeID
		task.Status = domain.TaskClaimed
		task.ClaimedAt = &now
	case "resolve":
		task.Status = domain.TaskResolved
		task.ResolvedAt = &now
		eventType = domain.EventFloorTaskResolved
	case "cancel":
		if strings.TrimSpace(cmd.Reason) == "" {
			return WriteResult[domain.FloorTask]{}, fmt.Errorf("%w: cancellation reason is required", ErrInvalid)
		}
		task.Status = domain.TaskCancelled
		task.ResolvedAt = &now
		eventType = domain.EventFloorTaskResolved
	default:
		return WriteResult[domain.FloorTask]{}, fmt.Errorf("%w: unknown task action %q", ErrInvalid, cmd.Action)
	}
	event := s.prepareEvent(domain.OperationalEvent{
		RestaurantID:   task.RestaurantID,
		ShiftID:        task.ShiftID,
		TableID:        task.TableID,
		TableSessionID: task.TableSessionID,
		EventType:      eventType,
		OccurredAt:     now,
		Source:         domain.SourceStaff,
		ActorID:        cmd.ActorID,
		Payload: map[string]any{
			"task_id":         task.ID,
			"action":          cmd.Action,
			"assignee_id":     task.ResponsibleID,
			"reason":          cmd.Reason,
			"resolution_note": cmd.ResolutionNote,
		},
	})
	if err := event.Validate(); err != nil {
		return WriteResult[domain.FloorTask]{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if err := s.store.SaveFloorTask(ctx, task); err != nil {
		return WriteResult[domain.FloorTask]{}, err
	}
	if err := s.store.AppendEvent(ctx, event); err != nil {
		return WriteResult[domain.FloorTask]{}, err
	}
	return WriteResult[domain.FloorTask]{Resource: task, EventID: event.ID, CorrelationID: event.CorrelationID}, nil
}

func (s *Service) ListFloorTasks(ctx context.Context, filter TaskFilter) ([]domain.FloorTask, error) {
	tasks, err := s.store.ListFloorTasks(ctx, filter)
	if err != nil {
		return nil, err
	}
	domain.SortFloorTasks(tasks, s.now())
	return tasks, nil
}

type OpenComplaintCommand struct {
	RestaurantID   string
	ShiftID        string
	TableID        string
	TableSessionID string
	OrderID        string
	Source         domain.Source
	ActorID        string
	Reason         string
	Description    string
	Severity       domain.Severity
}

func (s *Service) OpenComplaint(ctx context.Context, cmd OpenComplaintCommand) (WriteResult[domain.Complaint], error) {
	if strings.TrimSpace(cmd.RestaurantID) == "" || strings.TrimSpace(cmd.ShiftID) == "" || strings.TrimSpace(cmd.TableSessionID) == "" {
		return WriteResult[domain.Complaint]{}, fmt.Errorf("%w: restaurant_id, shift_id and table_session_id are required", ErrInvalid)
	}
	if strings.TrimSpace(cmd.Reason) == "" {
		return WriteResult[domain.Complaint]{}, fmt.Errorf("%w: complaint reason is required", ErrInvalid)
	}
	now := s.now()
	complaint := domain.Complaint{
		ID:             newID("cmp"),
		RestaurantID:   cmd.RestaurantID,
		ShiftID:        cmd.ShiftID,
		TableID:        cmd.TableID,
		TableSessionID: cmd.TableSessionID,
		OrderID:        cmd.OrderID,
		Reason:         cmd.Reason,
		Severity:       defaultSeverity(cmd.Severity, domain.SeverityMedium),
		Status:         domain.ComplaintClassified,
		OpenedAt:       now,
	}
	event := s.prepareEvent(domain.OperationalEvent{
		RestaurantID:   complaint.RestaurantID,
		ShiftID:        complaint.ShiftID,
		TableID:        complaint.TableID,
		TableSessionID: complaint.TableSessionID,
		EventType:      domain.EventComplaintOpened,
		OccurredAt:     now,
		Source:         defaultSource(cmd.Source, domain.SourceCustomer),
		ActorID:        cmd.ActorID,
		Payload: map[string]any{
			"complaint_id": complaint.ID,
			"reason":       complaint.Reason,
			"severity":     string(complaint.Severity),
			"description":  cmd.Description,
			"order_id":     complaint.OrderID,
		},
	})
	if err := event.Validate(); err != nil {
		return WriteResult[domain.Complaint]{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if err := s.store.SaveComplaint(ctx, complaint); err != nil {
		return WriteResult[domain.Complaint]{}, err
	}
	if err := s.store.AppendEvent(ctx, event); err != nil {
		return WriteResult[domain.Complaint]{}, err
	}
	task, err := s.createOrUpdateTask(ctx, domain.FloorTask{
		RestaurantID:       complaint.RestaurantID,
		ShiftID:            complaint.ShiftID,
		TableID:            complaint.TableID,
		TableSessionID:     complaint.TableSessionID,
		Type:               domain.TaskComplaint,
		Status:             domain.TaskOpen,
		Severity:           complaint.Severity,
		PriorityReason:     "Reclamacao registrada: " + complaint.Reason,
		RelatedComplaintID: complaint.ID,
		DueAt:              ptrTime(now.Add(s.policies.ComplaintAssignAfter)),
		SourceEventID:      event.ID,
	})
	if err != nil {
		return WriteResult[domain.Complaint]{}, err
	}
	return WriteResult[domain.Complaint]{
		Resource:        complaint,
		EventID:         event.ID,
		CorrelationID:   event.CorrelationID,
		FloorTaskID:     task.ID,
		CustomerMessage: "Reclamacao registrada. A equipe responsavel foi avisada.",
	}, nil
}

type UpdateComplaintCommand struct {
	ComplaintID       string
	Action            string
	ActorID           string
	AssigneeID        string
	Severity          domain.Severity
	Reason            string
	ResolutionSummary string
	Note              string
}

func (s *Service) UpdateComplaint(ctx context.Context, cmd UpdateComplaintCommand) (WriteResult[domain.Complaint], error) {
	if strings.TrimSpace(cmd.ActorID) == "" {
		return WriteResult[domain.Complaint]{}, fmt.Errorf("%w: actor_id is required", ErrInvalid)
	}
	complaint, err := s.store.GetComplaint(ctx, cmd.ComplaintID)
	if err != nil {
		return WriteResult[domain.Complaint]{}, err
	}
	now := s.now()
	eventType := domain.EventComplaintClassified
	switch cmd.Action {
	case "classify":
		if cmd.Severity != "" {
			complaint.Severity = cmd.Severity
		}
		if strings.TrimSpace(cmd.Reason) != "" {
			complaint.Reason = cmd.Reason
		}
		complaint.Status = domain.ComplaintClassified
	case "assign":
		if strings.TrimSpace(cmd.AssigneeID) == "" {
			return WriteResult[domain.Complaint]{}, fmt.Errorf("%w: assignee_id is required", ErrInvalid)
		}
		complaint.ResponsibleID = cmd.AssigneeID
		complaint.Status = domain.ComplaintAssigned
		eventType = domain.EventComplaintAssigned
	case "record_first_response":
		complaint.Status = domain.ComplaintInProgress
		complaint.FirstResponseAt = &now
		eventType = domain.EventComplaintFirstResponse
	case "resolve":
		complaint.Status = domain.ComplaintResolved
		complaint.ResolvedAt = &now
		complaint.ResolutionSummary = cmd.ResolutionSummary
		eventType = domain.EventComplaintResolved
	case "reopen":
		complaint.Status = domain.ComplaintReopened
		complaint.ResolvedAt = nil
		complaint.ResolutionSummary = ""
		eventType = domain.EventComplaintOpened
	default:
		return WriteResult[domain.Complaint]{}, fmt.Errorf("%w: unknown complaint action %q", ErrInvalid, cmd.Action)
	}
	event := s.prepareEvent(domain.OperationalEvent{
		RestaurantID:   complaint.RestaurantID,
		ShiftID:        complaint.ShiftID,
		TableID:        complaint.TableID,
		TableSessionID: complaint.TableSessionID,
		EventType:      eventType,
		OccurredAt:     now,
		Source:         domain.SourceStaff,
		ActorID:        cmd.ActorID,
		Payload: map[string]any{
			"complaint_id":       complaint.ID,
			"action":             cmd.Action,
			"assignee_id":        complaint.ResponsibleID,
			"reason":             complaint.Reason,
			"severity":           string(complaint.Severity),
			"resolution_summary": complaint.ResolutionSummary,
			"note":               cmd.Note,
		},
	})
	if err := event.Validate(); err != nil {
		return WriteResult[domain.Complaint]{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if err := s.store.SaveComplaint(ctx, complaint); err != nil {
		return WriteResult[domain.Complaint]{}, err
	}
	if err := s.store.AppendEvent(ctx, event); err != nil {
		return WriteResult[domain.Complaint]{}, err
	}
	if cmd.Action == "assign" {
		_ = s.claimOpenTasksForComplaint(ctx, complaint.ID, complaint.ResponsibleID)
	}
	if cmd.Action == "resolve" {
		_ = s.resolveOpenTasksForComplaint(ctx, complaint.ID, "Reclamacao resolvida")
	}
	return WriteResult[domain.Complaint]{Resource: complaint, EventID: event.ID, CorrelationID: event.CorrelationID}, nil
}

type RequestBillHandoffCommand struct {
	RestaurantID      string
	ShiftID           string
	TableID           string
	TableSessionID    string
	Source            domain.Source
	ActorID           string
	HandoffTarget     string
	ExternalReference string
	Note              string
}

func (s *Service) RequestBillHandoff(ctx context.Context, cmd RequestBillHandoffCommand) (WriteResult[domain.BillHandoff], error) {
	if strings.TrimSpace(cmd.RestaurantID) == "" || strings.TrimSpace(cmd.ShiftID) == "" || strings.TrimSpace(cmd.TableID) == "" {
		return WriteResult[domain.BillHandoff]{}, fmt.Errorf("%w: restaurant_id, shift_id and table_id are required", ErrInvalid)
	}
	now := s.now()
	blockReasons, err := s.billBlockReasons(ctx, cmd.TableSessionID)
	if err != nil {
		return WriteResult[domain.BillHandoff]{}, err
	}
	status := domain.BillRequested
	eventType := domain.EventBillRequested
	if len(blockReasons) > 0 {
		status = domain.BillBlocked
		eventType = domain.EventBillHandoffBlocked
	}
	bill := domain.BillHandoff{
		ID:             newID("bill"),
		RestaurantID:   cmd.RestaurantID,
		ShiftID:        cmd.ShiftID,
		TableID:        cmd.TableID,
		TableSessionID: cmd.TableSessionID,
		Status:         status,
		BlockReason:    strings.Join(blockReasons, "; "),
		RequestedAt:    now,
	}
	event := s.prepareEvent(domain.OperationalEvent{
		RestaurantID:   bill.RestaurantID,
		ShiftID:        bill.ShiftID,
		TableID:        bill.TableID,
		TableSessionID: bill.TableSessionID,
		EventType:      eventType,
		OccurredAt:     now,
		Source:         defaultSource(cmd.Source, domain.SourceCustomer),
		ActorID:        cmd.ActorID,
		Payload: map[string]any{
			"bill_handoff_id":    bill.ID,
			"handoff_target":     cmd.HandoffTarget,
			"external_reference": cmd.ExternalReference,
			"note":               cmd.Note,
			"status":             string(bill.Status),
			"block_reasons":      blockReasons,
		},
	})
	if err := event.Validate(); err != nil {
		return WriteResult[domain.BillHandoff]{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if err := s.store.SaveBillHandoff(ctx, bill); err != nil {
		return WriteResult[domain.BillHandoff]{}, err
	}
	if err := s.store.AppendEvent(ctx, event); err != nil {
		return WriteResult[domain.BillHandoff]{}, err
	}
	severity := domain.SeverityMedium
	reason := "Conta solicitada aguardando handoff"
	if bill.Status == domain.BillBlocked {
		severity = domain.SeverityHigh
		reason = "Conta solicitada com pendencia operacional: " + bill.BlockReason
	}
	task, err := s.createOrUpdateTask(ctx, domain.FloorTask{
		RestaurantID:         bill.RestaurantID,
		ShiftID:              bill.ShiftID,
		TableID:              bill.TableID,
		TableSessionID:       bill.TableSessionID,
		Type:                 domain.TaskBillHandoff,
		Status:               domain.TaskOpen,
		Severity:             severity,
		PriorityReason:       reason,
		RelatedBillHandoffID: bill.ID,
		DueAt:                ptrTime(now.Add(s.policies.BillHandoffAcceptedAfter)),
		SourceEventID:        event.ID,
	})
	if err != nil {
		return WriteResult[domain.BillHandoff]{}, err
	}
	message := "Conta solicitada. A equipe foi avisada."
	if bill.Status == domain.BillBlocked {
		message = "Conta solicitada, mas existe pendencia operacional que a equipe precisa revisar."
	}
	return WriteResult[domain.BillHandoff]{
		Resource:        bill,
		EventID:         event.ID,
		CorrelationID:   event.CorrelationID,
		FloorTaskID:     task.ID,
		CustomerMessage: message,
	}, nil
}

type UpdateBillHandoffCommand struct {
	BillHandoffID     string
	Action            string
	ActorID           string
	AssigneeID        string
	Reason            string
	ExternalReference string
}

func (s *Service) UpdateBillHandoff(ctx context.Context, cmd UpdateBillHandoffCommand) (WriteResult[domain.BillHandoff], error) {
	if strings.TrimSpace(cmd.ActorID) == "" {
		return WriteResult[domain.BillHandoff]{}, fmt.Errorf("%w: actor_id is required", ErrInvalid)
	}
	bill, err := s.store.GetBillHandoff(ctx, cmd.BillHandoffID)
	if err != nil {
		return WriteResult[domain.BillHandoff]{}, err
	}
	now := s.now()
	eventType := domain.EventBillHandoffAccepted
	switch cmd.Action {
	case "accept":
		bill.Status = domain.BillInReview
		bill.ResponsibleID = firstNonEmpty(cmd.AssigneeID, cmd.ActorID)
		bill.AcceptedAt = &now
	case "send_to_existing_system":
		bill.Status = domain.BillSentToExistingSystem
		bill.ResponsibleID = firstNonEmpty(bill.ResponsibleID, cmd.AssigneeID, cmd.ActorID)
	case "await_cashier":
		bill.Status = domain.BillAwaitingCashierAction
		bill.ResponsibleID = firstNonEmpty(bill.ResponsibleID, cmd.AssigneeID, cmd.ActorID)
	case "block":
		if strings.TrimSpace(cmd.Reason) == "" {
			return WriteResult[domain.BillHandoff]{}, fmt.Errorf("%w: block reason is required", ErrInvalid)
		}
		bill.Status = domain.BillBlocked
		bill.BlockReason = cmd.Reason
		eventType = domain.EventBillHandoffBlocked
	case "confirm_close":
		if bill.Status == domain.BillBlocked {
			return WriteResult[domain.BillHandoff]{}, fmt.Errorf("%w: blocked bill handoff cannot be closed", ErrConflict)
		}
		bill.Status = domain.BillClosed
		bill.ClosedAt = &now
		eventType = domain.EventBillCloseConfirmed
	default:
		return WriteResult[domain.BillHandoff]{}, fmt.Errorf("%w: unknown bill action %q", ErrInvalid, cmd.Action)
	}
	event := s.prepareEvent(domain.OperationalEvent{
		RestaurantID:   bill.RestaurantID,
		ShiftID:        bill.ShiftID,
		TableID:        bill.TableID,
		TableSessionID: bill.TableSessionID,
		EventType:      eventType,
		OccurredAt:     now,
		Source:         domain.SourceCashier,
		ActorID:        cmd.ActorID,
		Payload: map[string]any{
			"bill_handoff_id":    bill.ID,
			"action":             cmd.Action,
			"assignee_id":        bill.ResponsibleID,
			"reason":             cmd.Reason,
			"external_reference": cmd.ExternalReference,
			"status":             string(bill.Status),
		},
	})
	if err := event.Validate(); err != nil {
		return WriteResult[domain.BillHandoff]{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if err := s.store.SaveBillHandoff(ctx, bill); err != nil {
		return WriteResult[domain.BillHandoff]{}, err
	}
	if err := s.store.AppendEvent(ctx, event); err != nil {
		return WriteResult[domain.BillHandoff]{}, err
	}
	if cmd.Action == "accept" || cmd.Action == "confirm_close" {
		_ = s.claimOrResolveTasksForBill(ctx, bill.ID, bill.ResponsibleID, cmd.Action == "confirm_close")
	}
	if cmd.Action == "confirm_close" {
		_ = s.closeTableSession(ctx, bill.TableSessionID, cmd.ActorID)
	}
	return WriteResult[domain.BillHandoff]{Resource: bill, EventID: event.ID, CorrelationID: event.CorrelationID}, nil
}

func (s *Service) EvaluateSLA(ctx context.Context, shiftID string) ([]domain.FloorTask, error) {
	now := s.now()
	var created []domain.FloorTask
	orders, err := s.store.ListOrders(ctx, OrderFilter{ShiftID: shiftID})
	if err != nil {
		return nil, err
	}
	for _, order := range orders {
		if order.Status == domain.OrderPreparing && !order.UpdatedAt.Add(s.policies.OrderStaleAfter).After(now) {
			event, err := s.CreateEvent(ctx, domain.OperationalEvent{
				RestaurantID:   order.RestaurantID,
				ShiftID:        order.ShiftID,
				TableID:        order.TableID,
				TableSessionID: order.TableSessionID,
				EventType:      domain.EventOrderStaleDetected,
				OccurredAt:     now,
				Source:         domain.SourceSystem,
				Payload:        map[string]any{"order_id": order.ID, "status": string(order.Status)},
			})
			if err != nil {
				return nil, err
			}
			task, err := s.createOrUpdateTask(ctx, domain.FloorTask{
				RestaurantID:   order.RestaurantID,
				ShiftID:        order.ShiftID,
				TableID:        order.TableID,
				TableSessionID: order.TableSessionID,
				Type:           domain.TaskOrderStale,
				Status:         domain.TaskOpen,
				Severity:       domain.SeverityMedium,
				PriorityReason: "Pedido sem atualizacao recente",
				RelatedOrderID: order.ID,
				DueAt:          ptrTime(now),
				SourceEventID:  event.ID,
			})
			if err != nil {
				return nil, err
			}
			created = append(created, task)
		}
		if order.Status == domain.OrderPreparing && !order.UpdatedAt.Add(s.policies.OrderDelayRiskAfter).After(now) {
			res, err := s.UpdateOrderStatus(ctx, UpdateOrderStatusCommand{OrderID: order.ID, Status: domain.OrderDelayRisk, Source: domain.SourceSystem, Reason: "sla_delay_risk"})
			if err != nil {
				return nil, err
			}
			task, err := s.createOrUpdateTask(ctx, domain.FloorTask{
				RestaurantID:   order.RestaurantID,
				ShiftID:        order.ShiftID,
				TableID:        order.TableID,
				TableSessionID: order.TableSessionID,
				Type:           domain.TaskOrderDelayed,
				Status:         domain.TaskOpen,
				Severity:       domain.SeverityMedium,
				PriorityReason: "Pedido em risco de atraso",
				RelatedOrderID: order.ID,
				DueAt:          ptrTime(now.Add(2 * time.Minute)),
				SourceEventID:  res.EventID,
			})
			if err != nil {
				return nil, err
			}
			created = append(created, task)
		}
		if (order.Status == domain.OrderPreparing || order.Status == domain.OrderDelayRisk) && !order.UpdatedAt.Add(s.policies.OrderDelayedAfter).After(now) {
			res, err := s.UpdateOrderStatus(ctx, UpdateOrderStatusCommand{OrderID: order.ID, Status: domain.OrderDelayed, Source: domain.SourceSystem, Reason: "sla_delayed"})
			if err != nil {
				return nil, err
			}
			task, err := s.createOrUpdateTask(ctx, domain.FloorTask{
				RestaurantID:   order.RestaurantID,
				ShiftID:        order.ShiftID,
				TableID:        order.TableID,
				TableSessionID: order.TableSessionID,
				Type:           domain.TaskOrderDelayed,
				Status:         domain.TaskOpen,
				Severity:       domain.SeverityHigh,
				PriorityReason: "Pedido atrasado acima do SLA",
				RelatedOrderID: order.ID,
				DueAt:          ptrTime(now),
				SourceEventID:  res.EventID,
			})
			if err != nil {
				return nil, err
			}
			created = append(created, task)
		}
		if order.Status == domain.OrderReady && order.ReadyAt != nil && !order.ReadyAt.Add(s.policies.ReadyPickupAfter).After(now) {
			task, err := s.createOrUpdateTask(ctx, domain.FloorTask{
				RestaurantID:   order.RestaurantID,
				ShiftID:        order.ShiftID,
				TableID:        order.TableID,
				TableSessionID: order.TableSessionID,
				Type:           domain.TaskOrderReady,
				Status:         domain.TaskOpen,
				Severity:       domain.SeverityHigh,
				PriorityReason: "Pedido pronto parado acima do SLA",
				RelatedOrderID: order.ID,
				DueAt:          ptrTime(now),
			})
			if err != nil {
				return nil, err
			}
			created = append(created, task)
		}
	}
	complaints, err := s.store.ListComplaints(ctx, ComplaintFilter{ShiftID: shiftID})
	if err != nil {
		return nil, err
	}
	for _, complaint := range complaints {
		if domain.ComplaintNeedsEscalation(complaint, now, s.policies.ComplaintAssignAfter) {
			complaint.EscalatedToLeaderAt = &now
			if err := s.store.SaveComplaint(ctx, complaint); err != nil {
				return nil, err
			}
			task, err := s.createOrUpdateTask(ctx, domain.FloorTask{
				RestaurantID:       complaint.RestaurantID,
				ShiftID:            complaint.ShiftID,
				TableID:            complaint.TableID,
				TableSessionID:     complaint.TableSessionID,
				Type:               domain.TaskComplaint,
				Status:             domain.TaskOpen,
				Severity:           domain.SeverityHigh,
				PriorityReason:     "Reclamacao sem responsavel acima do SLA",
				RelatedComplaintID: complaint.ID,
				DueAt:              ptrTime(now),
			})
			if err != nil {
				return nil, err
			}
			created = append(created, task)
		}
	}
	return created, nil
}

type Timeline struct {
	TableSession domain.TableSession       `json:"table_session"`
	Table        domain.Table              `json:"table"`
	Orders       []domain.Order            `json:"orders"`
	FloorTasks   []domain.FloorTask        `json:"floor_tasks"`
	Complaints   []domain.Complaint        `json:"complaints"`
	BillHandoffs []domain.BillHandoff      `json:"bill_handoffs"`
	Events       []domain.OperationalEvent `json:"events"`
}

func (s *Service) Timeline(ctx context.Context, tableSessionID string) (Timeline, error) {
	session, err := s.store.GetTableSession(ctx, tableSessionID)
	if err != nil {
		return Timeline{}, err
	}
	table, err := s.store.GetTable(ctx, session.TableID)
	if err != nil {
		return Timeline{}, err
	}
	orders, err := s.store.ListOrders(ctx, OrderFilter{TableSessionID: tableSessionID})
	if err != nil {
		return Timeline{}, err
	}
	tasks, err := s.store.ListFloorTasks(ctx, TaskFilter{ShiftID: session.ShiftID})
	if err != nil {
		return Timeline{}, err
	}
	filteredTasks := tasks[:0]
	for _, task := range tasks {
		if task.TableSessionID == tableSessionID {
			filteredTasks = append(filteredTasks, task)
		}
	}
	complaints, err := s.store.ListComplaints(ctx, ComplaintFilter{TableSessionID: tableSessionID})
	if err != nil {
		return Timeline{}, err
	}
	bills, err := s.store.ListBillHandoffs(ctx, BillFilter{TableSessionID: tableSessionID})
	if err != nil {
		return Timeline{}, err
	}
	events, err := s.store.ListEvents(ctx, EventFilter{TableSessionID: tableSessionID})
	if err != nil {
		return Timeline{}, err
	}
	return Timeline{
		TableSession: session,
		Table:        table,
		Orders:       orders,
		FloorTasks:   filteredTasks,
		Complaints:   complaints,
		BillHandoffs: bills,
		Events:       events,
	}, nil
}

func (s *Service) CloseShift(ctx context.Context, shiftID, actorID string) (WriteResult[domain.ServiceShift], error) {
	if strings.TrimSpace(actorID) == "" {
		return WriteResult[domain.ServiceShift]{}, fmt.Errorf("%w: actor_id is required", ErrInvalid)
	}
	shift, err := s.store.GetShift(ctx, shiftID)
	if err != nil {
		return WriteResult[domain.ServiceShift]{}, err
	}
	now := s.now()
	shift.Status = domain.ShiftClosed
	shift.ClosedAt = &now
	event := s.prepareEvent(domain.OperationalEvent{
		RestaurantID: shift.RestaurantID,
		ShiftID:      shift.ID,
		EventType:    domain.EventShiftClosed,
		OccurredAt:   now,
		Source:       domain.SourceStaff,
		ActorID:      actorID,
		Payload:      map[string]any{"shift_id": shift.ID},
	})
	if err := event.Validate(); err != nil {
		return WriteResult[domain.ServiceShift]{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if err := s.store.SaveShift(ctx, shift); err != nil {
		return WriteResult[domain.ServiceShift]{}, err
	}
	if err := s.store.AppendEvent(ctx, event); err != nil {
		return WriteResult[domain.ServiceShift]{}, err
	}
	return WriteResult[domain.ServiceShift]{Resource: shift, EventID: event.ID, CorrelationID: event.CorrelationID}, nil
}

func (s *Service) Metrics(ctx context.Context, shiftID string) (domain.ShiftMetrics, error) {
	now := s.now()
	shift, err := s.store.GetShift(ctx, shiftID)
	if err != nil {
		return domain.ShiftMetrics{}, err
	}
	events, err := s.store.ListEvents(ctx, EventFilter{ShiftID: shiftID})
	if err != nil {
		return domain.ShiftMetrics{}, err
	}
	orders, err := s.store.ListOrders(ctx, OrderFilter{ShiftID: shiftID})
	if err != nil {
		return domain.ShiftMetrics{}, err
	}
	tasks, err := s.store.ListFloorTasks(ctx, TaskFilter{ShiftID: shiftID})
	if err != nil {
		return domain.ShiftMetrics{}, err
	}
	complaints, err := s.store.ListComplaints(ctx, ComplaintFilter{ShiftID: shiftID})
	if err != nil {
		return domain.ShiftMetrics{}, err
	}
	bills, err := s.store.ListBillHandoffs(ctx, BillFilter{ShiftID: shiftID})
	if err != nil {
		return domain.ShiftMetrics{}, err
	}
	metrics := domain.ShiftMetrics{ShiftID: shiftID, TotalEvents: len(events), Closed: shift.Status == domain.ShiftClosed}
	for _, order := range orders {
		if (order.Status == domain.OrderPreparing || order.Status == domain.OrderDelayRisk) && !order.UpdatedAt.Add(s.policies.OrderStaleAfter).After(now) {
			metrics.OrdersWithoutRecentUpdate++
		}
		if order.Status == domain.OrderDelayed {
			metrics.DelayedOrders++
		}
		if order.Status == domain.OrderReady {
			metrics.ReadyOrdersAwaitingPickup++
		}
	}
	for _, task := range tasks {
		if task.Status == domain.TaskOpen || task.Status == domain.TaskClaimed || task.Status == domain.TaskInProgress {
			metrics.OpenTasks++
			if task.DueAt != nil && !task.DueAt.After(now) {
				metrics.OverdueTasks++
			}
		}
	}
	for _, complaint := range complaints {
		if domain.ComplaintNeedsEscalation(complaint, now, s.policies.ComplaintAssignAfter) {
			metrics.ComplaintsWithoutOwnerOverSLA++
		}
	}
	for _, bill := range bills {
		if bill.ResponsibleID == "" && bill.Status != domain.BillClosed && !bill.RequestedAt.Add(s.policies.BillHandoffAcceptedAfter).After(now) {
			metrics.BillHandoffsWithoutOwnerOverSLA++
		}
	}
	return metrics, nil
}

func (s *Service) prepareEvent(event domain.OperationalEvent) domain.OperationalEvent {
	if event.ID == "" {
		event.ID = newID("evt")
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = s.now()
	}
	if event.CorrelationID == "" {
		event.CorrelationID = newID("corr")
	}
	if event.Payload == nil {
		event.Payload = map[string]any{}
	}
	return event
}

func (s *Service) createOrUpdateTask(ctx context.Context, task domain.FloorTask) (domain.FloorTask, error) {
	filter := TaskFilter{
		ShiftID:              task.ShiftID,
		RelatedOrderID:       task.RelatedOrderID,
		RelatedComplaintID:   task.RelatedComplaintID,
		RelatedBillHandoffID: task.RelatedBillHandoffID,
	}
	existing, err := s.store.ListFloorTasks(ctx, filter)
	if err != nil {
		return domain.FloorTask{}, err
	}
	for _, current := range existing {
		if current.Type == task.Type && current.Status != domain.TaskResolved && current.Status != domain.TaskCancelled {
			if task.Severity.Rank() > current.Severity.Rank() {
				current.Severity = task.Severity
			}
			if task.PriorityReason != "" {
				current.PriorityReason = task.PriorityReason
			}
			if task.DueAt != nil {
				current.DueAt = task.DueAt
			}
			if err := s.store.SaveFloorTask(ctx, current); err != nil {
				return domain.FloorTask{}, err
			}
			return current, nil
		}
	}
	now := s.now()
	task.ID = newID("task")
	task.CreatedAt = now
	if task.Status == "" {
		task.Status = domain.TaskOpen
	}
	if task.Severity == "" {
		task.Severity = domain.SeverityMedium
	}
	if err := s.store.SaveFloorTask(ctx, task); err != nil {
		return domain.FloorTask{}, err
	}
	event := s.prepareEvent(domain.OperationalEvent{
		RestaurantID:   task.RestaurantID,
		ShiftID:        task.ShiftID,
		TableID:        task.TableID,
		TableSessionID: task.TableSessionID,
		EventType:      domain.EventFloorTaskCreated,
		OccurredAt:     now,
		Source:         domain.SourceSystem,
		Payload:        map[string]any{"task_id": task.ID, "type": string(task.Type), "priority_reason": task.PriorityReason},
	})
	if task.SourceEventID != "" {
		event.CorrelationID = task.SourceEventID
	}
	if err := event.Validate(); err != nil {
		return domain.FloorTask{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if err := s.store.AppendEvent(ctx, event); err != nil {
		return domain.FloorTask{}, err
	}
	return task, nil
}

func (s *Service) resolveOpenTasksForOrder(ctx context.Context, orderID string, taskType domain.TaskType, note string) error {
	tasks, err := s.store.ListFloorTasks(ctx, TaskFilter{RelatedOrderID: orderID})
	if err != nil {
		return err
	}
	now := s.now()
	for _, task := range tasks {
		if task.Type == taskType && task.Status != domain.TaskResolved && task.Status != domain.TaskCancelled {
			task.Status = domain.TaskResolved
			task.ResolvedAt = &now
			if err := s.store.SaveFloorTask(ctx, task); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) claimOpenTasksForComplaint(ctx context.Context, complaintID, assigneeID string) error {
	tasks, err := s.store.ListFloorTasks(ctx, TaskFilter{RelatedComplaintID: complaintID})
	if err != nil {
		return err
	}
	now := s.now()
	for _, task := range tasks {
		if task.Status == domain.TaskOpen {
			task.Status = domain.TaskClaimed
			task.ResponsibleID = assigneeID
			task.ClaimedAt = &now
			if err := s.store.SaveFloorTask(ctx, task); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) resolveOpenTasksForComplaint(ctx context.Context, complaintID, note string) error {
	tasks, err := s.store.ListFloorTasks(ctx, TaskFilter{RelatedComplaintID: complaintID})
	if err != nil {
		return err
	}
	now := s.now()
	for _, task := range tasks {
		if task.Status != domain.TaskResolved && task.Status != domain.TaskCancelled {
			task.Status = domain.TaskResolved
			task.ResolvedAt = &now
			if err := s.store.SaveFloorTask(ctx, task); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) claimOrResolveTasksForBill(ctx context.Context, billID, assigneeID string, resolve bool) error {
	tasks, err := s.store.ListFloorTasks(ctx, TaskFilter{RelatedBillHandoffID: billID})
	if err != nil {
		return err
	}
	now := s.now()
	for _, task := range tasks {
		if task.Status == domain.TaskResolved || task.Status == domain.TaskCancelled {
			continue
		}
		if resolve {
			task.Status = domain.TaskResolved
			task.ResolvedAt = &now
		} else {
			task.Status = domain.TaskClaimed
			task.ResponsibleID = assigneeID
			task.ClaimedAt = &now
		}
		if err := s.store.SaveFloorTask(ctx, task); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) billBlockReasons(ctx context.Context, tableSessionID string) ([]string, error) {
	var reasons []string
	if strings.TrimSpace(tableSessionID) == "" {
		return reasons, nil
	}
	complaints, err := s.store.ListComplaints(ctx, ComplaintFilter{TableSessionID: tableSessionID})
	if err != nil {
		return nil, err
	}
	for _, complaint := range complaints {
		if complaint.Status != domain.ComplaintResolved {
			reasons = append(reasons, "reclamacao_aberta")
			break
		}
	}
	orders, err := s.store.ListOrders(ctx, OrderFilter{TableSessionID: tableSessionID})
	if err != nil {
		return nil, err
	}
	for _, order := range orders {
		if order.Status != domain.OrderDelivered && order.Status != domain.OrderCancelled {
			reasons = append(reasons, "pedido_nao_entregue")
			break
		}
	}
	return reasons, nil
}

func (s *Service) closeTableSession(ctx context.Context, tableSessionID, actorID string) error {
	if strings.TrimSpace(tableSessionID) == "" {
		return nil
	}
	session, err := s.store.GetTableSession(ctx, tableSessionID)
	if err != nil {
		return err
	}
	if session.Status == domain.TableSessionClosed {
		return nil
	}
	now := s.now()
	session.Status = domain.TableSessionClosed
	session.ClosedAt = &now
	if err := s.store.SaveTableSession(ctx, session); err != nil {
		return err
	}
	event := s.prepareEvent(domain.OperationalEvent{
		RestaurantID:   session.RestaurantID,
		ShiftID:        session.ShiftID,
		TableID:        session.TableID,
		TableSessionID: session.ID,
		EventType:      domain.EventTableSessionClosed,
		OccurredAt:     now,
		Source:         domain.SourceStaff,
		ActorID:        actorID,
		Payload:        map[string]any{"table_session_id": session.ID},
	})
	if err := event.Validate(); err != nil {
		return err
	}
	return s.store.AppendEvent(ctx, event)
}

func defaultSource(value, fallback domain.Source) domain.Source {
	if value == "" {
		return fallback
	}
	return value
}

func defaultSeverity(value, fallback domain.Severity) domain.Severity {
	if value == "" {
		return fallback
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func newID(prefix string) string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return prefix + "_" + hex.EncodeToString(b[:])
}
