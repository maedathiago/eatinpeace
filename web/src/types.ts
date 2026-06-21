export type Source = "customer" | "staff" | "kitchen" | "cashier" | "system" | "integration";

export type Severity = "low" | "medium" | "high" | "critical";

export type TableSessionStatus = "active" | "closed";

export type OrderStatus =
  | "received"
  | "preparing"
  | "delay_risk"
  | "delayed"
  | "ready"
  | "picked_up"
  | "delivered"
  | "cancelled";

export type TaskStatus = "open" | "claimed" | "in_progress" | "resolved" | "cancelled";

export type ComplaintStatus =
  | "open"
  | "classified"
  | "assigned"
  | "in_progress"
  | "resolved"
  | "reopened";

export type BillHandoffStatus =
  | "requested"
  | "in_review"
  | "sent_to_existing_system"
  | "awaiting_cashier_action"
  | "blocked"
  | "closed";

export interface WriteResponse<T> {
  resource: T;
  event_id: string;
  correlation_id: string;
  customer_message?: string;
  floor_task_id?: string;
}

export interface TableSession {
  id: string;
  restaurant_id: string;
  shift_id: string;
  table_id: string;
  status: TableSessionStatus;
  opened_at: string;
  closed_at?: string;
}

export interface Table {
  id: string;
  restaurant_id: string;
  label: string;
}

export interface OrderItem {
  id: string;
  order_id: string;
  name: string;
  quantity: number;
  notes?: string;
}

export interface Order {
  id: string;
  restaurant_id: string;
  shift_id: string;
  table_id: string;
  table_session_id: string;
  status: OrderStatus;
  items?: OrderItem[];
  created_at: string;
  updated_at: string;
  ready_at?: string;
  delivered_at?: string;
}

export interface FloorTask {
  id: string;
  restaurant_id: string;
  shift_id: string;
  table_id?: string;
  table_session_id?: string;
  type: string;
  status: TaskStatus;
  severity: Severity;
  priority_reason: string;
  responsible_id?: string;
  related_order_id?: string;
  related_complaint_id?: string;
  related_bill_handoff_id?: string;
  due_at?: string;
  created_at: string;
  claimed_at?: string;
  started_at?: string;
  resolved_at?: string;
  source_event_id?: string;
}

export interface Complaint {
  id: string;
  restaurant_id: string;
  shift_id: string;
  table_id?: string;
  table_session_id?: string;
  order_id?: string;
  reason: string;
  severity: Severity;
  status: ComplaintStatus;
  responsible_id?: string;
  opened_at: string;
  first_response_at?: string;
  resolved_at?: string;
  resolution_summary?: string;
  escalated_to_leader_at?: string;
}

export interface BillHandoff {
  id: string;
  restaurant_id: string;
  shift_id: string;
  table_id: string;
  table_session_id?: string;
  status: BillHandoffStatus;
  responsible_id?: string;
  block_reason?: string;
  requested_at: string;
  accepted_at?: string;
  closed_at?: string;
}

export interface OperationalEvent {
  id: string;
  restaurant_id: string;
  shift_id?: string;
  table_id?: string;
  table_session_id?: string;
  event_type: string;
  occurred_at: string;
  source: Source;
  actor_id?: string;
  payload?: Record<string, unknown>;
  correlation_id: string;
}

export interface Timeline {
  table_session: TableSession;
  table: Table;
  orders: Order[];
  floor_tasks: FloorTask[];
  complaints: Complaint[];
  bill_handoffs: BillHandoff[];
  events: OperationalEvent[];
}

export interface ShiftMetrics {
  shift_id: string;
  total_events: number;
  open_tasks: number;
  overdue_tasks: number;
  orders_without_recent_update: number;
  delayed_orders: number;
  ready_orders_awaiting_pickup: number;
  complaints_without_owner_over_sla: number;
  bill_handoffs_without_owner_over_sla: number;
  closed: boolean;
}
