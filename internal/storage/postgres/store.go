package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/maedathiago/eatinpeace/internal/application"
	"github.com/maedathiago/eatinpeace/internal/domain"
)

type Store struct {
	db *sql.DB
}

var _ application.Store = (*Store)(nil)

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) SaveRestaurant(ctx context.Context, restaurant domain.Restaurant) error {
	_, err := s.db.ExecContext(ctx, `
insert into restaurants (id, name)
values ($1, $2)
on conflict (id) do update set name = excluded.name, updated_at = now()
`, restaurant.ID, restaurant.Name)
	return err
}

func (s *Store) SaveShift(ctx context.Context, shift domain.ServiceShift) error {
	_, err := s.db.ExecContext(ctx, `
insert into service_shifts (id, restaurant_id, name, status, opened_at, closed_at)
values ($1, $2, $3, $4, $5, $6)
on conflict (id) do update set
    name = excluded.name,
    status = excluded.status,
    closed_at = excluded.closed_at,
    updated_at = now()
`, shift.ID, shift.RestaurantID, shift.Name, shift.Status, shift.OpenedAt, nullableTime(shift.ClosedAt))
	return err
}

func (s *Store) GetShift(ctx context.Context, id string) (domain.ServiceShift, error) {
	row := s.db.QueryRowContext(ctx, `select id, restaurant_id, name, status, opened_at, closed_at from service_shifts where id = $1`, id)
	var shift domain.ServiceShift
	var closedAt sql.NullTime
	if err := row.Scan(&shift.ID, &shift.RestaurantID, &shift.Name, &shift.Status, &shift.OpenedAt, &closedAt); err != nil {
		return domain.ServiceShift{}, mapSQLError(err)
	}
	shift.ClosedAt = timePtr(closedAt)
	return shift, nil
}

func (s *Store) SaveTable(ctx context.Context, table domain.Table) error {
	_, err := s.db.ExecContext(ctx, `
insert into tables (id, restaurant_id, label)
values ($1, $2, $3)
on conflict (id) do update set label = excluded.label, updated_at = now()
`, table.ID, table.RestaurantID, table.Label)
	return err
}

func (s *Store) GetTable(ctx context.Context, id string) (domain.Table, error) {
	row := s.db.QueryRowContext(ctx, `select id, restaurant_id, label from tables where id = $1`, id)
	var table domain.Table
	if err := row.Scan(&table.ID, &table.RestaurantID, &table.Label); err != nil {
		return domain.Table{}, mapSQLError(err)
	}
	return table, nil
}

func (s *Store) SaveStaffMember(ctx context.Context, staff domain.StaffMember) error {
	_, err := s.db.ExecContext(ctx, `
insert into staff_members (id, restaurant_id, name, role)
values ($1, $2, $3, $4)
on conflict (id) do update set name = excluded.name, role = excluded.role, updated_at = now()
`, staff.ID, staff.RestaurantID, staff.Name, staff.Role)
	return err
}

func (s *Store) AppendEvent(ctx context.Context, event domain.OperationalEvent) error {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
insert into operational_events (
    id, restaurant_id, shift_id, table_id, table_session_id, event_type,
    occurred_at, source, actor_id, payload, correlation_id
) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb, $11)
`, event.ID, event.RestaurantID, nullableString(event.ShiftID), nullableString(event.TableID), nullableString(event.TableSessionID),
		event.EventType, event.OccurredAt, event.Source, nullableString(event.ActorID), string(payload), event.CorrelationID)
	return err
}

func (s *Store) ListEvents(ctx context.Context, filter application.EventFilter) ([]domain.OperationalEvent, error) {
	query := `select id, restaurant_id, shift_id, table_id, table_session_id, event_type, occurred_at, source, actor_id, payload, correlation_id from operational_events where 1 = 1`
	args := []any{}
	query, args = addTextFilter(query, args, "restaurant_id", filter.RestaurantID)
	query, args = addTextFilter(query, args, "shift_id", filter.ShiftID)
	query, args = addTextFilter(query, args, "table_session_id", filter.TableSessionID)
	query += " order by occurred_at asc"
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []domain.OperationalEvent
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *Store) SaveTableSession(ctx context.Context, session domain.TableSession) error {
	_, err := s.db.ExecContext(ctx, `
insert into table_sessions (id, restaurant_id, shift_id, table_id, status, opened_at, closed_at)
values ($1, $2, $3, $4, $5, $6, $7)
on conflict (id) do update set status = excluded.status, closed_at = excluded.closed_at, updated_at = now()
`, session.ID, session.RestaurantID, session.ShiftID, session.TableID, session.Status, session.OpenedAt, nullableTime(session.ClosedAt))
	return err
}

func (s *Store) GetTableSession(ctx context.Context, id string) (domain.TableSession, error) {
	row := s.db.QueryRowContext(ctx, `select id, restaurant_id, shift_id, table_id, status, opened_at, closed_at from table_sessions where id = $1`, id)
	session, err := scanTableSession(row)
	if err != nil {
		return domain.TableSession{}, mapSQLError(err)
	}
	return session, nil
}

func (s *Store) ListTableSessions(ctx context.Context, shiftID string) ([]domain.TableSession, error) {
	query := `select id, restaurant_id, shift_id, table_id, status, opened_at, closed_at from table_sessions where 1 = 1`
	args := []any{}
	query, args = addTextFilter(query, args, "shift_id", shiftID)
	query += " order by opened_at asc"
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []domain.TableSession
	for rows.Next() {
		session, err := scanTableSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}

func (s *Store) SaveOrder(ctx context.Context, order domain.Order) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)
	if _, err := tx.ExecContext(ctx, `
insert into orders (id, restaurant_id, shift_id, table_id, table_session_id, status, ready_at, delivered_at, created_at, updated_at)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
on conflict (id) do update set
    status = excluded.status,
    ready_at = excluded.ready_at,
    delivered_at = excluded.delivered_at,
    updated_at = excluded.updated_at
`, order.ID, order.RestaurantID, order.ShiftID, order.TableID, order.TableSessionID, order.Status, nullableTime(order.ReadyAt), nullableTime(order.DeliveredAt), order.CreatedAt, order.UpdatedAt); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `delete from order_items where order_id = $1`, order.ID); err != nil {
		return err
	}
	for _, item := range order.Items {
		if _, err := tx.ExecContext(ctx, `insert into order_items (id, order_id, name, quantity, notes) values ($1, $2, $3, $4, $5)`, item.ID, order.ID, item.Name, item.Quantity, nullableString(item.Notes)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) GetOrder(ctx context.Context, id string) (domain.Order, error) {
	row := s.db.QueryRowContext(ctx, `select id, restaurant_id, shift_id, table_id, table_session_id, status, ready_at, delivered_at, created_at, updated_at from orders where id = $1`, id)
	order, err := scanOrder(row)
	if err != nil {
		return domain.Order{}, mapSQLError(err)
	}
	order.Items, err = s.listOrderItems(ctx, order.ID)
	return order, err
}

func (s *Store) ListOrders(ctx context.Context, filter application.OrderFilter) ([]domain.Order, error) {
	query := `select id, restaurant_id, shift_id, table_id, table_session_id, status, ready_at, delivered_at, created_at, updated_at from orders where 1 = 1`
	args := []any{}
	query, args = addTextFilter(query, args, "shift_id", filter.ShiftID)
	query, args = addTextFilter(query, args, "table_session_id", filter.TableSessionID)
	query += " order by created_at asc"
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []domain.Order
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		order.Items, err = s.listOrderItems(ctx, order.ID)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

func (s *Store) SaveFloorTask(ctx context.Context, task domain.FloorTask) error {
	_, err := s.db.ExecContext(ctx, `
insert into floor_tasks (
    id, restaurant_id, shift_id, table_id, table_session_id, type, status, severity,
    priority_reason, responsible_id, related_order_id, related_complaint_id,
    related_bill_handoff_id, due_at, claimed_at, started_at, resolved_at, source_event_id, created_at
) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
on conflict (id) do update set
    status = excluded.status,
    severity = excluded.severity,
    priority_reason = excluded.priority_reason,
    responsible_id = excluded.responsible_id,
    due_at = excluded.due_at,
    claimed_at = excluded.claimed_at,
    started_at = excluded.started_at,
    resolved_at = excluded.resolved_at,
    updated_at = now()
`, task.ID, task.RestaurantID, task.ShiftID, nullableString(task.TableID), nullableString(task.TableSessionID), task.Type, task.Status, task.Severity,
		task.PriorityReason, nullableString(task.ResponsibleID), nullableString(task.RelatedOrderID), nullableString(task.RelatedComplaintID),
		nullableString(task.RelatedBillHandoffID), nullableTime(task.DueAt), nullableTime(task.ClaimedAt), nullableTime(task.StartedAt),
		nullableTime(task.ResolvedAt), nullableString(task.SourceEventID), task.CreatedAt)
	return err
}

func (s *Store) GetFloorTask(ctx context.Context, id string) (domain.FloorTask, error) {
	row := s.db.QueryRowContext(ctx, floorTaskSelect()+` where id = $1`, id)
	task, err := scanFloorTask(row)
	if err != nil {
		return domain.FloorTask{}, mapSQLError(err)
	}
	return task, nil
}

func (s *Store) ListFloorTasks(ctx context.Context, filter application.TaskFilter) ([]domain.FloorTask, error) {
	query := floorTaskSelect() + ` where 1 = 1`
	args := []any{}
	query, args = addTextFilter(query, args, "shift_id", filter.ShiftID)
	query, args = addTextFilter(query, args, "table_id", filter.TableID)
	query, args = addTextFilter(query, args, "status", string(filter.Status))
	query, args = addTextFilter(query, args, "responsible_id", filter.ResponsibleID)
	query, args = addTextFilter(query, args, "related_order_id", filter.RelatedOrderID)
	query, args = addTextFilter(query, args, "related_complaint_id", filter.RelatedComplaintID)
	query, args = addTextFilter(query, args, "related_bill_handoff_id", filter.RelatedBillHandoffID)
	query += " order by created_at asc"
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []domain.FloorTask
	for rows.Next() {
		task, err := scanFloorTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (s *Store) SaveComplaint(ctx context.Context, complaint domain.Complaint) error {
	_, err := s.db.ExecContext(ctx, `
insert into complaints (
    id, restaurant_id, shift_id, table_id, table_session_id, order_id, reason,
    severity, status, responsible_id, opened_at, first_response_at, resolved_at,
    resolution_summary, escalated_to_leader_at
) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
on conflict (id) do update set
    reason = excluded.reason,
    severity = excluded.severity,
    status = excluded.status,
    responsible_id = excluded.responsible_id,
    first_response_at = excluded.first_response_at,
    resolved_at = excluded.resolved_at,
    resolution_summary = excluded.resolution_summary,
    escalated_to_leader_at = excluded.escalated_to_leader_at,
    updated_at = now()
`, complaint.ID, complaint.RestaurantID, complaint.ShiftID, nullableString(complaint.TableID), nullableString(complaint.TableSessionID),
		nullableString(complaint.OrderID), complaint.Reason, complaint.Severity, complaint.Status, nullableString(complaint.ResponsibleID),
		complaint.OpenedAt, nullableTime(complaint.FirstResponseAt), nullableTime(complaint.ResolvedAt), nullableString(complaint.ResolutionSummary),
		nullableTime(complaint.EscalatedToLeaderAt))
	return err
}

func (s *Store) GetComplaint(ctx context.Context, id string) (domain.Complaint, error) {
	row := s.db.QueryRowContext(ctx, complaintSelect()+` where id = $1`, id)
	complaint, err := scanComplaint(row)
	if err != nil {
		return domain.Complaint{}, mapSQLError(err)
	}
	return complaint, nil
}

func (s *Store) ListComplaints(ctx context.Context, filter application.ComplaintFilter) ([]domain.Complaint, error) {
	query := complaintSelect() + ` where 1 = 1`
	args := []any{}
	query, args = addTextFilter(query, args, "shift_id", filter.ShiftID)
	query, args = addTextFilter(query, args, "table_session_id", filter.TableSessionID)
	query += " order by opened_at asc"
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var complaints []domain.Complaint
	for rows.Next() {
		complaint, err := scanComplaint(rows)
		if err != nil {
			return nil, err
		}
		complaints = append(complaints, complaint)
	}
	return complaints, rows.Err()
}

func (s *Store) SaveBillHandoff(ctx context.Context, bill domain.BillHandoff) error {
	_, err := s.db.ExecContext(ctx, `
insert into bill_handoffs (
    id, restaurant_id, shift_id, table_id, table_session_id, status,
    responsible_id, block_reason, requested_at, accepted_at, closed_at
) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
on conflict (id) do update set
    status = excluded.status,
    responsible_id = excluded.responsible_id,
    block_reason = excluded.block_reason,
    accepted_at = excluded.accepted_at,
    closed_at = excluded.closed_at,
    updated_at = now()
`, bill.ID, bill.RestaurantID, bill.ShiftID, bill.TableID, nullableString(bill.TableSessionID), bill.Status, nullableString(bill.ResponsibleID),
		nullableString(bill.BlockReason), bill.RequestedAt, nullableTime(bill.AcceptedAt), nullableTime(bill.ClosedAt))
	return err
}

func (s *Store) GetBillHandoff(ctx context.Context, id string) (domain.BillHandoff, error) {
	row := s.db.QueryRowContext(ctx, billSelect()+` where id = $1`, id)
	bill, err := scanBill(row)
	if err != nil {
		return domain.BillHandoff{}, mapSQLError(err)
	}
	return bill, nil
}

func (s *Store) ListBillHandoffs(ctx context.Context, filter application.BillFilter) ([]domain.BillHandoff, error) {
	query := billSelect() + ` where 1 = 1`
	args := []any{}
	query, args = addTextFilter(query, args, "shift_id", filter.ShiftID)
	query, args = addTextFilter(query, args, "table_session_id", filter.TableSessionID)
	query += " order by requested_at asc"
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bills []domain.BillHandoff
	for rows.Next() {
		bill, err := scanBill(rows)
		if err != nil {
			return nil, err
		}
		bills = append(bills, bill)
	}
	return bills, rows.Err()
}

func (s *Store) listOrderItems(ctx context.Context, orderID string) ([]domain.OrderItem, error) {
	rows, err := s.db.QueryContext(ctx, `select id, order_id, name, quantity, notes from order_items where order_id = $1 order by created_at asc`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		var notes sql.NullString
		if err := rows.Scan(&item.ID, &item.OrderID, &item.Name, &item.Quantity, &notes); err != nil {
			return nil, err
		}
		item.Notes = stringValue(notes)
		items = append(items, item)
	}
	return items, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanEvent(row scanner) (domain.OperationalEvent, error) {
	var event domain.OperationalEvent
	var shiftID, tableID, tableSessionID, actorID sql.NullString
	var payload []byte
	if err := row.Scan(&event.ID, &event.RestaurantID, &shiftID, &tableID, &tableSessionID, &event.EventType, &event.OccurredAt, &event.Source, &actorID, &payload, &event.CorrelationID); err != nil {
		return domain.OperationalEvent{}, err
	}
	event.ShiftID = stringValue(shiftID)
	event.TableID = stringValue(tableID)
	event.TableSessionID = stringValue(tableSessionID)
	event.ActorID = stringValue(actorID)
	if len(payload) > 0 {
		if err := json.Unmarshal(payload, &event.Payload); err != nil {
			return domain.OperationalEvent{}, err
		}
	}
	return event, nil
}

func scanTableSession(row scanner) (domain.TableSession, error) {
	var session domain.TableSession
	var closedAt sql.NullTime
	if err := row.Scan(&session.ID, &session.RestaurantID, &session.ShiftID, &session.TableID, &session.Status, &session.OpenedAt, &closedAt); err != nil {
		return domain.TableSession{}, err
	}
	session.ClosedAt = timePtr(closedAt)
	return session, nil
}

func scanOrder(row scanner) (domain.Order, error) {
	var order domain.Order
	var readyAt, deliveredAt sql.NullTime
	if err := row.Scan(&order.ID, &order.RestaurantID, &order.ShiftID, &order.TableID, &order.TableSessionID, &order.Status, &readyAt, &deliveredAt, &order.CreatedAt, &order.UpdatedAt); err != nil {
		return domain.Order{}, err
	}
	order.ReadyAt = timePtr(readyAt)
	order.DeliveredAt = timePtr(deliveredAt)
	return order, nil
}

func scanFloorTask(row scanner) (domain.FloorTask, error) {
	var task domain.FloorTask
	var tableID, sessionID, responsibleID, orderID, complaintID, billID, sourceEventID sql.NullString
	var dueAt, claimedAt, startedAt, resolvedAt sql.NullTime
	if err := row.Scan(&task.ID, &task.RestaurantID, &task.ShiftID, &tableID, &sessionID, &task.Type, &task.Status, &task.Severity,
		&task.PriorityReason, &responsibleID, &orderID, &complaintID, &billID, &dueAt, &claimedAt, &startedAt, &resolvedAt,
		&sourceEventID, &task.CreatedAt); err != nil {
		return domain.FloorTask{}, err
	}
	task.TableID = stringValue(tableID)
	task.TableSessionID = stringValue(sessionID)
	task.ResponsibleID = stringValue(responsibleID)
	task.RelatedOrderID = stringValue(orderID)
	task.RelatedComplaintID = stringValue(complaintID)
	task.RelatedBillHandoffID = stringValue(billID)
	task.DueAt = timePtr(dueAt)
	task.ClaimedAt = timePtr(claimedAt)
	task.StartedAt = timePtr(startedAt)
	task.ResolvedAt = timePtr(resolvedAt)
	task.SourceEventID = stringValue(sourceEventID)
	return task, nil
}

func scanComplaint(row scanner) (domain.Complaint, error) {
	var complaint domain.Complaint
	var tableID, sessionID, orderID, responsibleID, resolution sql.NullString
	var firstResponseAt, resolvedAt, escalatedAt sql.NullTime
	if err := row.Scan(&complaint.ID, &complaint.RestaurantID, &complaint.ShiftID, &tableID, &sessionID, &orderID, &complaint.Reason,
		&complaint.Severity, &complaint.Status, &responsibleID, &complaint.OpenedAt, &firstResponseAt, &resolvedAt,
		&resolution, &escalatedAt); err != nil {
		return domain.Complaint{}, err
	}
	complaint.TableID = stringValue(tableID)
	complaint.TableSessionID = stringValue(sessionID)
	complaint.OrderID = stringValue(orderID)
	complaint.ResponsibleID = stringValue(responsibleID)
	complaint.FirstResponseAt = timePtr(firstResponseAt)
	complaint.ResolvedAt = timePtr(resolvedAt)
	complaint.ResolutionSummary = stringValue(resolution)
	complaint.EscalatedToLeaderAt = timePtr(escalatedAt)
	return complaint, nil
}

func scanBill(row scanner) (domain.BillHandoff, error) {
	var bill domain.BillHandoff
	var sessionID, responsibleID, blockReason sql.NullString
	var acceptedAt, closedAt sql.NullTime
	if err := row.Scan(&bill.ID, &bill.RestaurantID, &bill.ShiftID, &bill.TableID, &sessionID, &bill.Status,
		&responsibleID, &blockReason, &bill.RequestedAt, &acceptedAt, &closedAt); err != nil {
		return domain.BillHandoff{}, err
	}
	bill.TableSessionID = stringValue(sessionID)
	bill.ResponsibleID = stringValue(responsibleID)
	bill.BlockReason = stringValue(blockReason)
	bill.AcceptedAt = timePtr(acceptedAt)
	bill.ClosedAt = timePtr(closedAt)
	return bill, nil
}

func floorTaskSelect() string {
	return `select id, restaurant_id, shift_id, table_id, table_session_id, type, status, severity, priority_reason,
responsible_id, related_order_id, related_complaint_id, related_bill_handoff_id, due_at, claimed_at,
started_at, resolved_at, source_event_id, created_at from floor_tasks`
}

func complaintSelect() string {
	return `select id, restaurant_id, shift_id, table_id, table_session_id, order_id, reason, severity, status,
responsible_id, opened_at, first_response_at, resolved_at, resolution_summary, escalated_to_leader_at from complaints`
}

func billSelect() string {
	return `select id, restaurant_id, shift_id, table_id, table_session_id, status, responsible_id, block_reason,
requested_at, accepted_at, closed_at from bill_handoffs`
}

func addTextFilter(query string, args []any, column string, value string) (string, []any) {
	if strings.TrimSpace(value) == "" {
		return query, args
	}
	args = append(args, value)
	return fmt.Sprintf("%s and %s = $%d", query, column, len(args)), args
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return *value
}

func timePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}

func stringValue(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func mapSQLError(err error) error {
	if err == sql.ErrNoRows {
		return application.ErrNotFound
	}
	return err
}

func rollback(tx *sql.Tx) {
	_ = tx.Rollback()
}
