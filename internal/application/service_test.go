package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/maedathiago/eatinpeace/internal/application"
	"github.com/maedathiago/eatinpeace/internal/domain"
	"github.com/maedathiago/eatinpeace/internal/testsupport"
)

func TestOrderLifecycleCreatesReadyPickupTask(t *testing.T) {
	ctx := context.Background()
	fixture := testsupport.NewFixture(t)
	session := openSession(t, fixture)
	order := createOrder(t, fixture, session.ID)

	if _, err := fixture.Service.UpdateOrderStatus(ctx, application.UpdateOrderStatusCommand{OrderID: order.ID, Status: domain.OrderPreparing, Source: domain.SourceKitchen, ActorID: "staff_kitchen"}); err != nil {
		t.Fatalf("preparing: %v", err)
	}
	result, err := fixture.Service.UpdateOrderStatus(ctx, application.UpdateOrderStatusCommand{OrderID: order.ID, Status: domain.OrderReady, Source: domain.SourceKitchen, ActorID: "staff_kitchen"})
	if err != nil {
		t.Fatalf("ready: %v", err)
	}
	if result.FloorTaskID == "" {
		t.Fatal("ready order did not create pickup task")
	}

	tasks, err := fixture.Service.ListFloorTasks(ctx, application.TaskFilter{RelatedOrderID: order.ID})
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].Type != domain.TaskOrderReady {
		t.Fatalf("unexpected ready tasks: %#v", tasks)
	}

	if _, err := fixture.Service.UpdateOrderStatus(ctx, application.UpdateOrderStatusCommand{OrderID: order.ID, Status: domain.OrderDelivered, Source: domain.SourceStaff, ActorID: "staff_waiter"}); err == nil {
		t.Fatal("ready -> delivered without pickup should require compatible path or override")
	}
	if _, err := fixture.Service.UpdateOrderStatus(ctx, application.UpdateOrderStatusCommand{OrderID: order.ID, Status: domain.OrderPickedUp, Source: domain.SourceStaff, ActorID: "staff_waiter"}); err != nil {
		t.Fatalf("picked up: %v", err)
	}
	if _, err := fixture.Service.UpdateOrderStatus(ctx, application.UpdateOrderStatusCommand{OrderID: order.ID, Status: domain.OrderDelivered, Source: domain.SourceStaff, ActorID: "staff_waiter"}); err != nil {
		t.Fatalf("delivered: %v", err)
	}
}

func TestSLAEvaluationCreatesOperationalTasks(t *testing.T) {
	ctx := context.Background()
	fixture := testsupport.NewFixture(t)
	session := openSession(t, fixture)
	order := createOrder(t, fixture, session.ID)
	if _, err := fixture.Service.UpdateOrderStatus(ctx, application.UpdateOrderStatusCommand{OrderID: order.ID, Status: domain.OrderPreparing, Source: domain.SourceKitchen, ActorID: "staff_kitchen"}); err != nil {
		t.Fatalf("preparing: %v", err)
	}

	fixture.Advance(11 * time.Minute)
	tasks, err := fixture.Service.EvaluateSLA(ctx, "shift_pilot_open")
	if err != nil {
		t.Fatalf("evaluate sla risk: %v", err)
	}
	if len(tasks) == 0 {
		t.Fatal("SLA evaluation did not create any task")
	}
	updated, err := fixture.Store.GetOrder(ctx, order.ID)
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if updated.Status != domain.OrderDelayRisk {
		t.Fatalf("order status = %s, want delay_risk", updated.Status)
	}

	fixture.Advance(16 * time.Minute)
	if _, err := fixture.Service.EvaluateSLA(ctx, "shift_pilot_open"); err != nil {
		t.Fatalf("evaluate sla delayed: %v", err)
	}
	updated, err = fixture.Store.GetOrder(ctx, order.ID)
	if err != nil {
		t.Fatalf("get order delayed: %v", err)
	}
	if updated.Status != domain.OrderDelayed {
		t.Fatalf("order status = %s, want delayed", updated.Status)
	}
}

func TestComplaintEscalationAndTaskLink(t *testing.T) {
	ctx := context.Background()
	fixture := testsupport.NewFixture(t)
	session := openSession(t, fixture)
	result, err := fixture.Service.OpenComplaint(ctx, application.OpenComplaintCommand{
		RestaurantID:   "rest_pilot",
		ShiftID:        "shift_pilot_open",
		TableID:        session.TableID,
		TableSessionID: session.ID,
		Source:         domain.SourceCustomer,
		Reason:         "service_delay",
		Severity:       domain.SeverityMedium,
	})
	if err != nil {
		t.Fatalf("open complaint: %v", err)
	}
	if result.FloorTaskID == "" {
		t.Fatal("complaint did not create floor task")
	}

	fixture.Advance(3 * time.Minute)
	if _, err := fixture.Service.EvaluateSLA(ctx, "shift_pilot_open"); err != nil {
		t.Fatalf("evaluate complaint sla: %v", err)
	}
	complaint, err := fixture.Store.GetComplaint(ctx, result.Resource.ID)
	if err != nil {
		t.Fatalf("get complaint: %v", err)
	}
	if complaint.EscalatedToLeaderAt == nil {
		t.Fatal("unassigned complaint was not escalated")
	}
}

func TestBillHandoffBlocksOnOperationalPendingThenClosesSession(t *testing.T) {
	ctx := context.Background()
	fixture := testsupport.NewFixture(t)
	session := openSession(t, fixture)
	order := createOrder(t, fixture, session.ID)

	blocked, err := fixture.Service.RequestBillHandoff(ctx, application.RequestBillHandoffCommand{
		RestaurantID:   "rest_pilot",
		ShiftID:        "shift_pilot_open",
		TableID:        session.TableID,
		TableSessionID: session.ID,
		Source:         domain.SourceCustomer,
		HandoffTarget:  "cashier",
	})
	if err != nil {
		t.Fatalf("request blocked bill: %v", err)
	}
	if blocked.Resource.Status != domain.BillBlocked {
		t.Fatalf("bill status = %s, want blocked", blocked.Resource.Status)
	}

	if _, err := fixture.Service.UpdateOrderStatus(ctx, application.UpdateOrderStatusCommand{OrderID: order.ID, Status: domain.OrderPreparing, Source: domain.SourceKitchen, ActorID: "staff_kitchen"}); err != nil {
		t.Fatalf("preparing: %v", err)
	}
	if _, err := fixture.Service.UpdateOrderStatus(ctx, application.UpdateOrderStatusCommand{OrderID: order.ID, Status: domain.OrderReady, Source: domain.SourceKitchen, ActorID: "staff_kitchen"}); err != nil {
		t.Fatalf("ready: %v", err)
	}
	if _, err := fixture.Service.UpdateOrderStatus(ctx, application.UpdateOrderStatusCommand{OrderID: order.ID, Status: domain.OrderPickedUp, Source: domain.SourceStaff, ActorID: "staff_waiter"}); err != nil {
		t.Fatalf("picked up: %v", err)
	}
	if _, err := fixture.Service.UpdateOrderStatus(ctx, application.UpdateOrderStatusCommand{OrderID: order.ID, Status: domain.OrderDelivered, Source: domain.SourceStaff, ActorID: "staff_waiter"}); err != nil {
		t.Fatalf("delivered: %v", err)
	}

	requested, err := fixture.Service.RequestBillHandoff(ctx, application.RequestBillHandoffCommand{
		RestaurantID:   "rest_pilot",
		ShiftID:        "shift_pilot_open",
		TableID:        session.TableID,
		TableSessionID: session.ID,
		Source:         domain.SourceCustomer,
		HandoffTarget:  "cashier",
	})
	if err != nil {
		t.Fatalf("request bill: %v", err)
	}
	if requested.Resource.Status != domain.BillRequested {
		t.Fatalf("bill status = %s, want requested", requested.Resource.Status)
	}
	if _, err := fixture.Service.UpdateBillHandoff(ctx, application.UpdateBillHandoffCommand{BillHandoffID: requested.Resource.ID, Action: "accept", ActorID: "staff_cashier"}); err != nil {
		t.Fatalf("accept bill: %v", err)
	}
	if _, err := fixture.Service.UpdateBillHandoff(ctx, application.UpdateBillHandoffCommand{BillHandoffID: requested.Resource.ID, Action: "confirm_close", ActorID: "staff_cashier"}); err != nil {
		t.Fatalf("close bill: %v", err)
	}
	closed, err := fixture.Store.GetTableSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if closed.Status != domain.TableSessionClosed {
		t.Fatalf("session status = %s, want closed", closed.Status)
	}
}

func TestShiftMetricsAggregatesOperationalState(t *testing.T) {
	ctx := context.Background()
	fixture := testsupport.NewFixture(t)
	session := openSession(t, fixture)
	order := createOrder(t, fixture, session.ID)
	if _, err := fixture.Service.UpdateOrderStatus(ctx, application.UpdateOrderStatusCommand{OrderID: order.ID, Status: domain.OrderPreparing, Source: domain.SourceKitchen, ActorID: "staff_kitchen"}); err != nil {
		t.Fatalf("preparing: %v", err)
	}
	fixture.Advance(11 * time.Minute)
	if _, err := fixture.Service.EvaluateSLA(ctx, "shift_pilot_open"); err != nil {
		t.Fatalf("evaluate sla: %v", err)
	}
	if _, err := fixture.Service.OpenComplaint(ctx, application.OpenComplaintCommand{
		RestaurantID:   "rest_pilot",
		ShiftID:        "shift_pilot_open",
		TableID:        session.TableID,
		TableSessionID: session.ID,
		Source:         domain.SourceCustomer,
		Reason:         "service_delay",
		Severity:       domain.SeverityHigh,
	}); err != nil {
		t.Fatalf("open complaint: %v", err)
	}
	fixture.Advance(3 * time.Minute)
	metrics, err := fixture.Service.Metrics(ctx, "shift_pilot_open")
	if err != nil {
		t.Fatalf("metrics: %v", err)
	}
	if metrics.OpenTasks == 0 || metrics.ComplaintsWithoutOwnerOverSLA == 0 {
		t.Fatalf("metrics did not capture operational backlog: %#v", metrics)
	}
	if _, err := fixture.Service.CloseShift(ctx, "shift_pilot_open", "staff_lead"); err != nil {
		t.Fatalf("close shift: %v", err)
	}
	metrics, err = fixture.Service.Metrics(ctx, "shift_pilot_open")
	if err != nil {
		t.Fatalf("metrics after close: %v", err)
	}
	if !metrics.Closed {
		t.Fatal("metrics did not report closed shift")
	}
}

func openSession(t *testing.T, fixture testsupport.Fixture) domain.TableSession {
	t.Helper()
	result, err := fixture.Service.OpenTableSession(context.Background(), application.OpenTableSessionCommand{
		RestaurantID: "rest_pilot",
		ShiftID:      "shift_pilot_open",
		TableID:      "table_01",
		Source:       domain.SourceCustomer,
	})
	if err != nil {
		t.Fatalf("open session: %v", err)
	}
	return result.Resource
}

func createOrder(t *testing.T, fixture testsupport.Fixture, tableSessionID string) domain.Order {
	t.Helper()
	result, err := fixture.Service.CreateOrder(context.Background(), application.CreateOrderCommand{
		RestaurantID:   "rest_pilot",
		ShiftID:        "shift_pilot_open",
		TableID:        "table_01",
		TableSessionID: tableSessionID,
		Source:         domain.SourceStaff,
		ActorID:        "staff_waiter",
		Items:          []application.CreateOrderItem{{Name: "Prato piloto", Quantity: 1}},
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	return result.Resource
}
