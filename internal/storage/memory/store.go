package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/maedathiago/eatinpeace/internal/application"
	"github.com/maedathiago/eatinpeace/internal/domain"
)

type Store struct {
	mu           sync.RWMutex
	restaurants  map[string]domain.Restaurant
	shifts       map[string]domain.ServiceShift
	tables       map[string]domain.Table
	staff        map[string]domain.StaffMember
	events       []domain.OperationalEvent
	sessions     map[string]domain.TableSession
	orders       map[string]domain.Order
	tasks        map[string]domain.FloorTask
	complaints   map[string]domain.Complaint
	billHandoffs map[string]domain.BillHandoff
}

func NewStore() *Store {
	return &Store{
		restaurants:  map[string]domain.Restaurant{},
		shifts:       map[string]domain.ServiceShift{},
		tables:       map[string]domain.Table{},
		staff:        map[string]domain.StaffMember{},
		events:       []domain.OperationalEvent{},
		sessions:     map[string]domain.TableSession{},
		orders:       map[string]domain.Order{},
		tasks:        map[string]domain.FloorTask{},
		complaints:   map[string]domain.Complaint{},
		billHandoffs: map[string]domain.BillHandoff{},
	}
}

func (s *Store) SaveRestaurant(_ context.Context, restaurant domain.Restaurant) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.restaurants[restaurant.ID] = restaurant
	return nil
}

func (s *Store) SaveShift(_ context.Context, shift domain.ServiceShift) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shifts[shift.ID] = shift
	return nil
}

func (s *Store) GetShift(_ context.Context, id string) (domain.ServiceShift, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	shift, ok := s.shifts[id]
	if !ok {
		return domain.ServiceShift{}, application.ErrNotFound
	}
	return shift, nil
}

func (s *Store) SaveTable(_ context.Context, table domain.Table) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tables[table.ID] = table
	return nil
}

func (s *Store) GetTable(_ context.Context, id string) (domain.Table, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	table, ok := s.tables[id]
	if !ok {
		return domain.Table{}, application.ErrNotFound
	}
	return table, nil
}

func (s *Store) SaveStaffMember(_ context.Context, staff domain.StaffMember) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.staff[staff.ID] = staff
	return nil
}

func (s *Store) AppendEvent(_ context.Context, event domain.OperationalEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
	return nil
}

func (s *Store) ListEvents(_ context.Context, filter application.EventFilter) ([]domain.OperationalEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	events := make([]domain.OperationalEvent, 0, len(s.events))
	for _, event := range s.events {
		if filter.RestaurantID != "" && event.RestaurantID != filter.RestaurantID {
			continue
		}
		if filter.ShiftID != "" && event.ShiftID != filter.ShiftID {
			continue
		}
		if filter.TableSessionID != "" && event.TableSessionID != filter.TableSessionID {
			continue
		}
		events = append(events, event)
	}
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].OccurredAt.Before(events[j].OccurredAt)
	})
	return events, nil
}

func (s *Store) SaveTableSession(_ context.Context, session domain.TableSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
	return nil
}

func (s *Store) GetTableSession(_ context.Context, id string) (domain.TableSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	if !ok {
		return domain.TableSession{}, application.ErrNotFound
	}
	return session, nil
}

func (s *Store) ListTableSessions(_ context.Context, shiftID string) ([]domain.TableSession, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sessions := make([]domain.TableSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		if shiftID != "" && session.ShiftID != shiftID {
			continue
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (s *Store) SaveOrder(_ context.Context, order domain.Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[order.ID] = order
	return nil
}

func (s *Store) GetOrder(_ context.Context, id string) (domain.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, ok := s.orders[id]
	if !ok {
		return domain.Order{}, application.ErrNotFound
	}
	return order, nil
}

func (s *Store) ListOrders(_ context.Context, filter application.OrderFilter) ([]domain.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	orders := make([]domain.Order, 0, len(s.orders))
	for _, order := range s.orders {
		if filter.ShiftID != "" && order.ShiftID != filter.ShiftID {
			continue
		}
		if filter.TableSessionID != "" && order.TableSessionID != filter.TableSessionID {
			continue
		}
		orders = append(orders, order)
	}
	sort.SliceStable(orders, func(i, j int) bool {
		return orders[i].CreatedAt.Before(orders[j].CreatedAt)
	})
	return orders, nil
}

func (s *Store) SaveFloorTask(_ context.Context, task domain.FloorTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = task
	return nil
}

func (s *Store) GetFloorTask(_ context.Context, id string) (domain.FloorTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[id]
	if !ok {
		return domain.FloorTask{}, application.ErrNotFound
	}
	return task, nil
}

func (s *Store) ListFloorTasks(_ context.Context, filter application.TaskFilter) ([]domain.FloorTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tasks := make([]domain.FloorTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		if filter.ShiftID != "" && task.ShiftID != filter.ShiftID {
			continue
		}
		if filter.TableID != "" && task.TableID != filter.TableID {
			continue
		}
		if filter.Status != "" && task.Status != filter.Status {
			continue
		}
		if filter.ResponsibleID != "" && task.ResponsibleID != filter.ResponsibleID {
			continue
		}
		if filter.RelatedOrderID != "" && task.RelatedOrderID != filter.RelatedOrderID {
			continue
		}
		if filter.RelatedComplaintID != "" && task.RelatedComplaintID != filter.RelatedComplaintID {
			continue
		}
		if filter.RelatedBillHandoffID != "" && task.RelatedBillHandoffID != filter.RelatedBillHandoffID {
			continue
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (s *Store) SaveComplaint(_ context.Context, complaint domain.Complaint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.complaints[complaint.ID] = complaint
	return nil
}

func (s *Store) GetComplaint(_ context.Context, id string) (domain.Complaint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	complaint, ok := s.complaints[id]
	if !ok {
		return domain.Complaint{}, application.ErrNotFound
	}
	return complaint, nil
}

func (s *Store) ListComplaints(_ context.Context, filter application.ComplaintFilter) ([]domain.Complaint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	complaints := make([]domain.Complaint, 0, len(s.complaints))
	for _, complaint := range s.complaints {
		if filter.ShiftID != "" && complaint.ShiftID != filter.ShiftID {
			continue
		}
		if filter.TableSessionID != "" && complaint.TableSessionID != filter.TableSessionID {
			continue
		}
		complaints = append(complaints, complaint)
	}
	return complaints, nil
}

func (s *Store) SaveBillHandoff(_ context.Context, bill domain.BillHandoff) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.billHandoffs[bill.ID] = bill
	return nil
}

func (s *Store) GetBillHandoff(_ context.Context, id string) (domain.BillHandoff, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bill, ok := s.billHandoffs[id]
	if !ok {
		return domain.BillHandoff{}, application.ErrNotFound
	}
	return bill, nil
}

func (s *Store) ListBillHandoffs(_ context.Context, filter application.BillFilter) ([]domain.BillHandoff, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bills := make([]domain.BillHandoff, 0, len(s.billHandoffs))
	for _, bill := range s.billHandoffs {
		if filter.ShiftID != "" && bill.ShiftID != filter.ShiftID {
			continue
		}
		if filter.TableSessionID != "" && bill.TableSessionID != filter.TableSessionID {
			continue
		}
		bills = append(bills, bill)
	}
	return bills, nil
}
