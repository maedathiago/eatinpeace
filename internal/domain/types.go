package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Source string

const (
	SourceCustomer    Source = "customer"
	SourceStaff       Source = "staff"
	SourceKitchen     Source = "kitchen"
	SourceCashier     Source = "cashier"
	SourceSystem      Source = "system"
	SourceIntegration Source = "integration"
)

type EventType string

const (
	EventTableSessionOpened     EventType = "table_session_opened"
	EventOrderCreated           EventType = "order_created"
	EventOrderStatusChanged     EventType = "order_status_changed"
	EventOrderStaleDetected     EventType = "order_stale_detected"
	EventOrderDelayRiskDetected EventType = "order_delay_risk_detected"
	EventOrderDelayed           EventType = "order_delayed"
	EventOrderReady             EventType = "order_ready"
	EventOrderPickedUp          EventType = "order_picked_up"
	EventOrderDelivered         EventType = "order_delivered"
	EventWaiterCalled           EventType = "waiter_called"
	EventFloorTaskCreated       EventType = "floor_task_created"
	EventFloorTaskClaimed       EventType = "floor_task_claimed"
	EventFloorTaskResolved      EventType = "floor_task_resolved"
	EventComplaintOpened        EventType = "complaint_opened"
	EventComplaintClassified    EventType = "complaint_classified"
	EventComplaintAssigned      EventType = "complaint_assigned"
	EventComplaintFirstResponse EventType = "complaint_first_response_recorded"
	EventComplaintResolved      EventType = "complaint_resolved"
	EventBillRequested          EventType = "bill_requested"
	EventBillHandoffAccepted    EventType = "bill_handoff_accepted"
	EventBillHandoffBlocked     EventType = "bill_handoff_blocked"
	EventBillCloseConfirmed     EventType = "bill_close_confirmed"
	EventTableSessionClosed     EventType = "table_session_closed"
	EventShiftClosed            EventType = "shift_closed"
)

type OperationalEvent struct {
	ID             string         `json:"id"`
	RestaurantID   string         `json:"restaurant_id"`
	ShiftID        string         `json:"shift_id,omitempty"`
	TableID        string         `json:"table_id,omitempty"`
	TableSessionID string         `json:"table_session_id,omitempty"`
	EventType      EventType      `json:"event_type"`
	OccurredAt     time.Time      `json:"occurred_at"`
	Source         Source         `json:"source"`
	ActorID        string         `json:"actor_id,omitempty"`
	Payload        map[string]any `json:"payload,omitempty"`
	CorrelationID  string         `json:"correlation_id"`
}

func (e OperationalEvent) Validate() error {
	if strings.TrimSpace(e.ID) == "" {
		return errors.New("event id is required")
	}
	if strings.TrimSpace(e.RestaurantID) == "" {
		return errors.New("restaurant_id is required")
	}
	if strings.TrimSpace(string(e.EventType)) == "" {
		return errors.New("event_type is required")
	}
	if e.OccurredAt.IsZero() {
		return errors.New("occurred_at is required")
	}
	if !validSource(e.Source) {
		return fmt.Errorf("invalid source %q", e.Source)
	}
	if strings.TrimSpace(e.CorrelationID) == "" {
		return errors.New("correlation_id is required")
	}
	return nil
}

func validSource(source Source) bool {
	switch source {
	case SourceCustomer, SourceStaff, SourceKitchen, SourceCashier, SourceSystem, SourceIntegration:
		return true
	default:
		return false
	}
}

type Restaurant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ServiceShiftStatus string

const (
	ShiftOpen   ServiceShiftStatus = "open"
	ShiftClosed ServiceShiftStatus = "closed"
)

type ServiceShift struct {
	ID           string             `json:"id"`
	RestaurantID string             `json:"restaurant_id"`
	Name         string             `json:"name"`
	OpenedAt     time.Time          `json:"opened_at"`
	ClosedAt     *time.Time         `json:"closed_at,omitempty"`
	Status       ServiceShiftStatus `json:"status"`
}

type Table struct {
	ID           string `json:"id"`
	RestaurantID string `json:"restaurant_id"`
	Label        string `json:"label"`
}

type TableSessionStatus string

const (
	TableSessionActive TableSessionStatus = "active"
	TableSessionClosed TableSessionStatus = "closed"
)

type TableSession struct {
	ID           string             `json:"id"`
	RestaurantID string             `json:"restaurant_id"`
	ShiftID      string             `json:"shift_id"`
	TableID      string             `json:"table_id"`
	Status       TableSessionStatus `json:"status"`
	OpenedAt     time.Time          `json:"opened_at"`
	ClosedAt     *time.Time         `json:"closed_at,omitempty"`
}

type OrderStatus string

const (
	OrderReceived  OrderStatus = "received"
	OrderPreparing OrderStatus = "preparing"
	OrderDelayRisk OrderStatus = "delay_risk"
	OrderDelayed   OrderStatus = "delayed"
	OrderReady     OrderStatus = "ready"
	OrderPickedUp  OrderStatus = "picked_up"
	OrderDelivered OrderStatus = "delivered"
	OrderCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID             string      `json:"id"`
	RestaurantID   string      `json:"restaurant_id"`
	ShiftID        string      `json:"shift_id"`
	TableID        string      `json:"table_id"`
	TableSessionID string      `json:"table_session_id"`
	Status         OrderStatus `json:"status"`
	Items          []OrderItem `json:"items,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	ReadyAt        *time.Time  `json:"ready_at,omitempty"`
	DeliveredAt    *time.Time  `json:"delivered_at,omitempty"`
}

type OrderItem struct {
	ID       string `json:"id"`
	OrderID  string `json:"order_id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Notes    string `json:"notes,omitempty"`
}

func ValidateOrderTransition(from, to OrderStatus, overrideReason string) error {
	if from == to {
		return nil
	}
	if validOrderTransition(from, to) {
		return nil
	}
	if strings.TrimSpace(overrideReason) != "" {
		return nil
	}
	return fmt.Errorf("invalid order transition from %s to %s without override", from, to)
}

func validOrderTransition(from, to OrderStatus) bool {
	switch from {
	case "":
		return to == OrderReceived
	case OrderReceived:
		return to == OrderPreparing || to == OrderCancelled
	case OrderPreparing:
		return to == OrderDelayRisk || to == OrderDelayed || to == OrderReady || to == OrderCancelled
	case OrderDelayRisk:
		return to == OrderDelayed || to == OrderReady || to == OrderCancelled
	case OrderDelayed:
		return to == OrderReady || to == OrderCancelled
	case OrderReady:
		return to == OrderPickedUp || to == OrderCancelled
	case OrderPickedUp:
		return to == OrderDelivered
	default:
		return false
	}
}

type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

func (s Severity) Rank() int {
	switch s {
	case SeverityCritical:
		return 4
	case SeverityHigh:
		return 3
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 1
	default:
		return 0
	}
}

type TaskType string

const (
	TaskWaiterCall   TaskType = "waiter_call"
	TaskOrderStale   TaskType = "order_stale"
	TaskOrderDelayed TaskType = "order_delayed"
	TaskOrderReady   TaskType = "order_ready_pickup"
	TaskComplaint    TaskType = "complaint"
	TaskBillHandoff  TaskType = "bill_handoff"
)

type TaskStatus string

const (
	TaskOpen       TaskStatus = "open"
	TaskClaimed    TaskStatus = "claimed"
	TaskInProgress TaskStatus = "in_progress"
	TaskResolved   TaskStatus = "resolved"
	TaskCancelled  TaskStatus = "cancelled"
)

type FloorTask struct {
	ID                   string     `json:"id"`
	RestaurantID         string     `json:"restaurant_id"`
	ShiftID              string     `json:"shift_id"`
	TableID              string     `json:"table_id,omitempty"`
	TableSessionID       string     `json:"table_session_id,omitempty"`
	Type                 TaskType   `json:"type"`
	Status               TaskStatus `json:"status"`
	Severity             Severity   `json:"severity"`
	PriorityReason       string     `json:"priority_reason"`
	ResponsibleID        string     `json:"responsible_id,omitempty"`
	RelatedOrderID       string     `json:"related_order_id,omitempty"`
	RelatedComplaintID   string     `json:"related_complaint_id,omitempty"`
	RelatedBillHandoffID string     `json:"related_bill_handoff_id,omitempty"`
	DueAt                *time.Time `json:"due_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	ClaimedAt            *time.Time `json:"claimed_at,omitempty"`
	StartedAt            *time.Time `json:"started_at,omitempty"`
	ResolvedAt           *time.Time `json:"resolved_at,omitempty"`
	SourceEventID        string     `json:"source_event_id,omitempty"`
}

func SortFloorTasks(tasks []FloorTask, now time.Time) {
	sort.SliceStable(tasks, func(i, j int) bool {
		left, right := tasks[i], tasks[j]
		if left.Severity.Rank() != right.Severity.Rank() {
			return left.Severity.Rank() > right.Severity.Rank()
		}
		leftOverdue := left.DueAt != nil && !left.DueAt.After(now)
		rightOverdue := right.DueAt != nil && !right.DueAt.After(now)
		if leftOverdue != rightOverdue {
			return leftOverdue
		}
		if left.DueAt != nil && right.DueAt != nil && !left.DueAt.Equal(*right.DueAt) {
			return left.DueAt.Before(*right.DueAt)
		}
		if left.DueAt != nil && right.DueAt == nil {
			return true
		}
		if left.DueAt == nil && right.DueAt != nil {
			return false
		}
		return left.CreatedAt.Before(right.CreatedAt)
	})
}

type ComplaintStatus string

const (
	ComplaintOpen       ComplaintStatus = "open"
	ComplaintClassified ComplaintStatus = "classified"
	ComplaintAssigned   ComplaintStatus = "assigned"
	ComplaintInProgress ComplaintStatus = "in_progress"
	ComplaintResolved   ComplaintStatus = "resolved"
	ComplaintReopened   ComplaintStatus = "reopened"
)

type Complaint struct {
	ID                  string          `json:"id"`
	RestaurantID        string          `json:"restaurant_id"`
	ShiftID             string          `json:"shift_id"`
	TableID             string          `json:"table_id,omitempty"`
	TableSessionID      string          `json:"table_session_id,omitempty"`
	OrderID             string          `json:"order_id,omitempty"`
	Reason              string          `json:"reason"`
	Severity            Severity        `json:"severity"`
	Status              ComplaintStatus `json:"status"`
	ResponsibleID       string          `json:"responsible_id,omitempty"`
	OpenedAt            time.Time       `json:"opened_at"`
	FirstResponseAt     *time.Time      `json:"first_response_at,omitempty"`
	ResolvedAt          *time.Time      `json:"resolved_at,omitempty"`
	ResolutionSummary   string          `json:"resolution_summary,omitempty"`
	EscalatedToLeaderAt *time.Time      `json:"escalated_to_leader_at,omitempty"`
}

func ComplaintNeedsEscalation(c Complaint, now time.Time, limit time.Duration) bool {
	if c.ResponsibleID != "" || c.Status == ComplaintResolved {
		return false
	}
	return !c.OpenedAt.Add(limit).After(now)
}

type BillHandoffStatus string

const (
	BillRequested             BillHandoffStatus = "requested"
	BillInReview              BillHandoffStatus = "in_review"
	BillSentToExistingSystem  BillHandoffStatus = "sent_to_existing_system"
	BillAwaitingCashierAction BillHandoffStatus = "awaiting_cashier_action"
	BillBlocked               BillHandoffStatus = "blocked"
	BillClosed                BillHandoffStatus = "closed"
)

type BillHandoff struct {
	ID             string            `json:"id"`
	RestaurantID   string            `json:"restaurant_id"`
	ShiftID        string            `json:"shift_id"`
	TableID        string            `json:"table_id"`
	TableSessionID string            `json:"table_session_id,omitempty"`
	Status         BillHandoffStatus `json:"status"`
	ResponsibleID  string            `json:"responsible_id,omitempty"`
	BlockReason    string            `json:"block_reason,omitempty"`
	RequestedAt    time.Time         `json:"requested_at"`
	AcceptedAt     *time.Time        `json:"accepted_at,omitempty"`
	ClosedAt       *time.Time        `json:"closed_at,omitempty"`
}

type SLAPolicies struct {
	OrderStaleAfter          time.Duration `json:"order_stale_after"`
	OrderDelayRiskAfter      time.Duration `json:"order_delay_risk_after"`
	OrderDelayedAfter        time.Duration `json:"order_delayed_after"`
	ReadyPickupAfter         time.Duration `json:"ready_pickup_after"`
	ComplaintAssignAfter     time.Duration `json:"complaint_assign_after"`
	BillHandoffAcceptedAfter time.Duration `json:"bill_handoff_accepted_after"`
}

func DefaultSLAPolicies() SLAPolicies {
	return SLAPolicies{
		OrderStaleAfter:          8 * time.Minute,
		OrderDelayRiskAfter:      10 * time.Minute,
		OrderDelayedAfter:        15 * time.Minute,
		ReadyPickupAfter:         2 * time.Minute,
		ComplaintAssignAfter:     2 * time.Minute,
		BillHandoffAcceptedAfter: 3 * time.Minute,
	}
}

type StaffMember struct {
	ID           string `json:"id"`
	RestaurantID string `json:"restaurant_id"`
	Name         string `json:"name"`
	Role         string `json:"role"`
}

type ShiftMetrics struct {
	ShiftID                         string `json:"shift_id"`
	TotalEvents                     int    `json:"total_events"`
	OpenTasks                       int    `json:"open_tasks"`
	OverdueTasks                    int    `json:"overdue_tasks"`
	OrdersWithoutRecentUpdate       int    `json:"orders_without_recent_update"`
	DelayedOrders                   int    `json:"delayed_orders"`
	ReadyOrdersAwaitingPickup       int    `json:"ready_orders_awaiting_pickup"`
	ComplaintsWithoutOwnerOverSLA   int    `json:"complaints_without_owner_over_sla"`
	BillHandoffsWithoutOwnerOverSLA int    `json:"bill_handoffs_without_owner_over_sla"`
	Closed                          bool   `json:"closed"`
}
