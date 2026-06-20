package domain

import (
	"testing"
	"time"
)

func TestOperationalEventValidation(t *testing.T) {
	event := OperationalEvent{
		ID:            "evt_1",
		RestaurantID:  "rest_1",
		EventType:     EventOrderCreated,
		OccurredAt:    time.Date(2026, 6, 20, 18, 0, 0, 0, time.UTC),
		Source:        SourceCustomer,
		CorrelationID: "corr_1",
	}
	if err := event.Validate(); err != nil {
		t.Fatalf("valid event rejected: %v", err)
	}

	event.Source = "chatty_free_text"
	if err := event.Validate(); err == nil {
		t.Fatal("invalid source accepted")
	}
}

func TestOrderTransitionRequiresOverrideForDeliveredJump(t *testing.T) {
	if err := ValidateOrderTransition(OrderReceived, OrderDelivered, ""); err == nil {
		t.Fatal("received -> delivered without override was accepted")
	}
	if err := ValidateOrderTransition(OrderReceived, OrderDelivered, "lider confirmou entrega manualmente"); err != nil {
		t.Fatalf("override transition rejected: %v", err)
	}
}

func TestSortFloorTasksSeveritySLAThenFIFO(t *testing.T) {
	now := time.Date(2026, 6, 20, 18, 0, 0, 0, time.UTC)
	older := now.Add(-2 * time.Minute)
	newer := now.Add(-1 * time.Minute)
	due := now.Add(-30 * time.Second)
	tasks := []FloorTask{
		{ID: "medium_newer", Severity: SeverityMedium, CreatedAt: newer},
		{ID: "critical", Severity: SeverityCritical, CreatedAt: newer},
		{ID: "medium_older", Severity: SeverityMedium, CreatedAt: older},
		{ID: "medium_due", Severity: SeverityMedium, DueAt: &due, CreatedAt: newer},
	}
	SortFloorTasks(tasks, now)

	got := []string{tasks[0].ID, tasks[1].ID, tasks[2].ID, tasks[3].ID}
	want := []string{"critical", "medium_due", "medium_older", "medium_newer"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order mismatch at %d: got %v want %v", i, got, want)
		}
	}
}

func TestComplaintNeedsEscalation(t *testing.T) {
	opened := time.Date(2026, 6, 20, 18, 0, 0, 0, time.UTC)
	complaint := Complaint{OpenedAt: opened, Status: ComplaintClassified}
	if ComplaintNeedsEscalation(complaint, opened.Add(time.Minute), 2*time.Minute) {
		t.Fatal("complaint escalated before SLA")
	}
	if !ComplaintNeedsEscalation(complaint, opened.Add(3*time.Minute), 2*time.Minute) {
		t.Fatal("complaint did not escalate after SLA")
	}
	complaint.ResponsibleID = "staff_waiter"
	if ComplaintNeedsEscalation(complaint, opened.Add(3*time.Minute), 2*time.Minute) {
		t.Fatal("assigned complaint escalated")
	}
}
