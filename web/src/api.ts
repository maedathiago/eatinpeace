import type {
  BillHandoff,
  Complaint,
  FloorTask,
  Order,
  OrderStatus,
  Severity,
  ShiftMetrics,
  Source,
  TableSession,
  Timeline,
  WriteResponse,
} from "./types";

export interface CreateOrderItem {
  name: string;
  quantity: number;
  notes?: string;
}

interface RequestOptions {
  method?: "GET" | "POST" | "PATCH";
  body?: unknown;
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const init: RequestInit = {
    method: options.method ?? "GET",
    headers: {},
  };

  if (options.body !== undefined) {
    init.headers = { "Content-Type": "application/json" };
    init.body = JSON.stringify(options.body);
  }

  const response = await fetch(path, init);
  const text = await response.text();
  const data = text ? JSON.parse(text) : {};

  if (!response.ok) {
    const message = typeof data.error === "string" ? data.error : `Erro HTTP ${response.status}`;
    throw new Error(message);
  }

  return data as T;
}

export async function health(): Promise<boolean> {
  try {
    await request<{ status: string }>("/healthz");
    return true;
  } catch {
    return false;
  }
}

export function openTableSession(input: {
  restaurantId: string;
  shiftId: string;
  tableId: string;
}): Promise<WriteResponse<TableSession>> {
  return request("/v1/table-sessions", {
    method: "POST",
    body: {
      restaurant_id: input.restaurantId,
      shift_id: input.shiftId,
      table_id: input.tableId,
      source: "customer",
      qr_token: "react-console",
    },
  });
}

export function getTimeline(sessionId: string): Promise<Timeline> {
  return request(`/v1/table-sessions/${sessionId}/timeline`);
}

export function createOrder(input: {
  restaurantId: string;
  shiftId: string;
  tableId: string;
  tableSessionId: string;
  items: CreateOrderItem[];
}): Promise<WriteResponse<Order>> {
  return request("/v1/orders", {
    method: "POST",
    body: {
      restaurant_id: input.restaurantId,
      shift_id: input.shiftId,
      table_id: input.tableId,
      table_session_id: input.tableSessionId,
      source: "staff",
      actor_id: "staff_waiter",
      items: input.items,
    },
  });
}

export function updateOrderStatus(input: {
  orderId: string;
  status: OrderStatus;
  source: Source;
  actorId: string;
}): Promise<WriteResponse<Order>> {
  return request(`/v1/orders/${input.orderId}/status`, {
    method: "PATCH",
    body: {
      status: input.status,
      source: input.source,
      actor_id: input.actorId,
    },
  });
}

export function listFloorTasks(shiftId: string): Promise<{ tasks: FloorTask[] }> {
  return request(`/v1/floor-tasks?shift_id=${encodeURIComponent(shiftId)}`);
}

export function updateFloorTask(input: {
  taskId: string;
  action: "claim" | "start" | "resolve";
  actorId: string;
  assigneeId?: string;
  resolutionNote?: string;
}): Promise<WriteResponse<FloorTask>> {
  return request(`/v1/floor-tasks/${input.taskId}`, {
    method: "PATCH",
    body: {
      action: input.action,
      actor_id: input.actorId,
      assignee_id: input.assigneeId,
      resolution_note: input.resolutionNote,
    },
  });
}

export function openComplaint(input: {
  restaurantId: string;
  shiftId: string;
  tableId: string;
  tableSessionId: string;
  relatedOrderId?: string;
  reason: string;
  severity: Severity;
  description: string;
}): Promise<WriteResponse<Complaint>> {
  return request("/v1/complaints", {
    method: "POST",
    body: {
      restaurant_id: input.restaurantId,
      shift_id: input.shiftId,
      table_id: input.tableId,
      table_session_id: input.tableSessionId,
      related_order_id: input.relatedOrderId,
      source: "customer",
      reason_code: input.reason,
      description: input.description,
      severity: input.severity,
    },
  });
}

export function updateComplaint(input: {
  complaintId: string;
  action: "assign" | "record_first_response" | "resolve";
  actorId: string;
  assigneeId?: string;
  note?: string;
  resolutionCode?: string;
}): Promise<WriteResponse<Complaint>> {
  return request(`/v1/complaints/${input.complaintId}`, {
    method: "PATCH",
    body: {
      action: input.action,
      actor_id: input.actorId,
      assignee_id: input.assigneeId,
      note: input.note,
      resolution_code: input.resolutionCode,
    },
  });
}

export function requestBill(input: {
  restaurantId: string;
  shiftId: string;
  tableId: string;
  tableSessionId: string;
}): Promise<WriteResponse<BillHandoff>> {
  return request("/v1/bill-handoffs", {
    method: "POST",
    body: {
      restaurant_id: input.restaurantId,
      shift_id: input.shiftId,
      table_id: input.tableId,
      table_session_id: input.tableSessionId,
      source: "customer",
      handoff_target: "cashier",
    },
  });
}

export function updateBill(input: {
  billId: string;
  action: "accept" | "send_to_existing_system" | "await_cashier" | "confirm_close";
  actorId: string;
  assigneeId?: string;
}): Promise<WriteResponse<BillHandoff>> {
  return request(`/v1/bill-handoffs/${input.billId}`, {
    method: "PATCH",
    body: {
      action: input.action,
      actor_id: input.actorId,
      assignee_id: input.assigneeId,
    },
  });
}

export function evaluateSLA(shiftId: string): Promise<{ tasks: FloorTask[] }> {
  return request("/v1/sla/evaluate", {
    method: "POST",
    body: { shift_id: shiftId },
  });
}

export function closeShift(shiftId: string): Promise<WriteResponse<{ status: string }>> {
  return request(`/v1/service-shifts/${shiftId}/close`, {
    method: "POST",
    body: { actor_id: "staff_lead" },
  });
}

export function getMetrics(shiftId: string): Promise<ShiftMetrics> {
  return request(`/v1/service-shifts/${shiftId}/metrics`);
}
