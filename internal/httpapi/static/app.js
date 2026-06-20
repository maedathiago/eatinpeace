const state = {
  restaurantId: "rest_pilot",
  shiftId: "shift_pilot_open",
  tableId: "table_01",
  sessionId: "",
  orderId: "",
  complaintId: "",
  billId: "",
};

const $ = (selector) => document.querySelector(selector);
const buttons = document.querySelectorAll("[data-action]");

buttons.forEach((button) => {
  button.addEventListener("click", () => runAction(button.dataset.action));
});

boot();

async function boot() {
  await checkHealth();
  await refreshAll();
}

async function runAction(action) {
  try {
    setBusy(true);
    if (action === "open-session") await openSession();
    if (action === "create-order") await createOrder();
    if (action === "preparing") await updateOrder("preparing", "kitchen", "staff_kitchen");
    if (action === "ready") await updateOrder("ready", "kitchen", "staff_kitchen");
    if (action === "picked-up") await updateOrder("picked_up", "staff", "staff_waiter");
    if (action === "delivered") await updateOrder("delivered", "staff", "staff_waiter");
    if (action === "complaint") await openComplaint();
    if (action === "resolve-complaint") await resolveComplaint();
    if (action === "bill") await requestBill();
    if (action === "accept-bill") await updateBill("accept");
    if (action === "close-bill") await updateBill("confirm_close");
    if (action === "close-shift") await closeShift();
    if (action === "refresh") await refreshAll();
    await refreshAll();
  } catch (error) {
    showToast(error.message);
  } finally {
    setBusy(false);
  }
}

async function checkHealth() {
  try {
    await api("/healthz");
    $("#api-status").textContent = "api online";
    $("#api-status").className = "pill ok";
  } catch {
    $("#api-status").textContent = "api offline";
    $("#api-status").className = "pill bad";
  }
}

async function openSession() {
  const data = await api("/v1/table-sessions", {
    method: "POST",
    body: {
      restaurant_id: state.restaurantId,
      shift_id: state.shiftId,
      table_id: state.tableId,
      source: "customer",
      qr_token: "local-ui",
    },
  });
  state.sessionId = data.resource.id;
  showToast(data.customer_message || "Sessao aberta.");
}

async function createOrder() {
  requireSession();
  const data = await api("/v1/orders", {
    method: "POST",
    body: {
      restaurant_id: state.restaurantId,
      shift_id: state.shiftId,
      table_id: state.tableId,
      table_session_id: state.sessionId,
      source: "staff",
      actor_id: "staff_waiter",
      items: [{ name: "Prato piloto", quantity: 1 }],
    },
  });
  state.orderId = data.resource.id;
  showToast(data.customer_message || "Pedido criado.");
}

async function updateOrder(status, source, actorId) {
  requireOrder();
  const data = await api(`/v1/orders/${state.orderId}/status`, {
    method: "PATCH",
    body: { status, source, actor_id: actorId },
  });
  showToast(data.customer_message || "Pedido atualizado.");
}

async function openComplaint() {
  requireSession();
  const data = await api("/v1/complaints", {
    method: "POST",
    body: {
      restaurant_id: state.restaurantId,
      shift_id: state.shiftId,
      table_id: state.tableId,
      table_session_id: state.sessionId,
      related_order_id: state.orderId,
      source: "customer",
      reason_code: "service_delay",
      description: "Cliente pediu acompanhamento do atraso.",
      severity: "high",
    },
  });
  state.complaintId = data.resource.id;
  showToast(data.customer_message || "Reclamacao registrada.");
}

async function resolveComplaint() {
  requireComplaint();
  await api(`/v1/complaints/${state.complaintId}`, {
    method: "PATCH",
    body: {
      action: "assign",
      actor_id: "staff_lead",
      assignee_id: "staff_waiter",
    },
  });
  await api(`/v1/complaints/${state.complaintId}`, {
    method: "PATCH",
    body: {
      action: "record_first_response",
      actor_id: "staff_waiter",
      note: "Mesa recebeu retorno da equipe.",
    },
  });
  const data = await api(`/v1/complaints/${state.complaintId}`, {
    method: "PATCH",
    body: {
      action: "resolve",
      actor_id: "staff_waiter",
      resolution_code: "resolvido_no_salao",
    },
  });
  showToast(data.customer_message || "Reclamacao resolvida.");
}

async function requestBill() {
  requireSession();
  const data = await api("/v1/bill-handoffs", {
    method: "POST",
    body: {
      restaurant_id: state.restaurantId,
      shift_id: state.shiftId,
      table_id: state.tableId,
      table_session_id: state.sessionId,
      source: "customer",
      handoff_target: "cashier",
    },
  });
  state.billId = data.resource.id;
  showToast(data.customer_message || "Conta solicitada.");
}

async function updateBill(action) {
  requireBill();
  const data = await api(`/v1/bill-handoffs/${state.billId}`, {
    method: "PATCH",
    body: {
      action,
      actor_id: "staff_cashier",
      assignee_id: "staff_cashier",
    },
  });
  showToast(data.customer_message || "Handoff atualizado.");
}

async function closeShift() {
  const data = await api(`/v1/service-shifts/${state.shiftId}/close`, {
    method: "POST",
    body: { actor_id: "staff_lead" },
  });
  showToast(`Turno ${data.resource.status}.`);
}

async function refreshAll() {
  await checkHealth();
  if (state.sessionId) {
    const timeline = await api(`/v1/table-sessions/${state.sessionId}/timeline`);
    renderTimeline(timeline);
    renderOrders(timeline.orders || []);
    renderComplaints(timeline.complaints || []);
    renderBills(timeline.bill_handoffs || []);
    const latestOrder = timeline.orders?.at(-1);
    const latestComplaint = timeline.complaints?.at(-1);
    const latestBill = timeline.bill_handoffs?.at(-1);
    if (latestOrder) state.orderId = latestOrder.id;
    if (latestComplaint) state.complaintId = latestComplaint.id;
    if (latestBill) state.billId = latestBill.id;
    $("#session-state").textContent = timeline.table_session.status;
    $("#session-id").textContent = timeline.table_session.id;
  }
  const tasks = await api(`/v1/floor-tasks?shift_id=${state.shiftId}`);
  renderTasks(tasks.tasks || []);
  const metrics = await api(`/v1/service-shifts/${state.shiftId}/metrics`);
  renderMetrics(metrics);
  updateButtons();
}

function renderTimeline(data) {
  const events = data.events || [];
  $("#timeline-count").textContent = `${events.length} eventos`;
  $("#timeline").innerHTML = events.map((event) => `
    <li>
      <span class="time">${formatTime(event.occurred_at)}</span>
      <div>
        <div class="event-type">${event.event_type}</div>
        <div class="event-meta">${event.source} · ${event.correlation_id}</div>
      </div>
    </li>
  `).join("");
}

function renderTasks(tasks) {
  const openTasks = tasks.filter((task) => !["resolved", "cancelled"].includes(task.status));
  $("#task-count").textContent = `${openTasks.length} abertas`;
  $("#tasks").innerHTML = tasks.length ? tasks.map((task) => `
    <div class="task">
      <div>
        <div class="item-title">${task.type} · ${task.status}</div>
        <div class="item-meta">${task.priority_reason}</div>
        <div class="item-meta">${task.responsible_id || "sem responsavel"}</div>
      </div>
      <span class="severity ${task.severity}">${task.severity}</span>
    </div>
  `).join("") : emptyState("Nenhuma tarefa criada ainda.");
}

function renderOrders(orders) {
  $("#order-count").textContent = `${orders.length} pedidos`;
  $("#orders").innerHTML = orders.length ? orders.map((order) => `
    <div class="compact-item">
      <div class="item-title">${order.status}</div>
      <div class="item-meta">${order.items?.map((item) => item.name).join(", ") || "sem itens"}</div>
    </div>
  `).join("") : emptyState("Sem pedido operacional.");
}

function renderComplaints(complaints) {
  $("#complaint-count").textContent = `${complaints.length} tickets`;
  $("#complaints").innerHTML = complaints.length ? complaints.map((complaint) => `
    <div class="compact-item">
      <div class="item-title">${complaint.reason} · ${complaint.status}</div>
      <div class="item-meta">${complaint.responsible_id || "sem responsavel"} · ${complaint.severity}</div>
    </div>
  `).join("") : emptyState("Sem reclamacao registrada.");
}

function renderBills(bills) {
  const latest = bills.at(-1);
  $("#bill-state").textContent = latest ? latest.status : "sem handoff";
  $("#bills").innerHTML = bills.length ? bills.map((bill) => `
    <div class="compact-item">
      <div class="item-title">${bill.status}</div>
      <div class="item-meta">${bill.block_reason || bill.responsible_id || "aguardando responsavel"}</div>
    </div>
  `).join("") : emptyState("Conta ainda nao solicitada.");
}

function renderMetrics(metrics) {
  $("#metric-events").textContent = metrics.total_events ?? 0;
  $("#metric-open-tasks").textContent = metrics.open_tasks ?? 0;
  $("#metric-overdue").textContent = metrics.overdue_tasks ?? 0;
  $("#metric-delayed").textContent = metrics.delayed_orders ?? 0;
  $("#metric-ready").textContent = metrics.ready_orders_awaiting_pickup ?? 0;
  $("#metric-complaints").textContent = metrics.complaints_without_owner_over_sla ?? 0;
}

function updateButtons() {
  const hasSession = Boolean(state.sessionId);
  const hasOrder = Boolean(state.orderId);
  const hasComplaint = Boolean(state.complaintId);
  const hasBill = Boolean(state.billId);
  setDisabled("create-order", !hasSession);
  setDisabled("preparing", !hasOrder);
  setDisabled("ready", !hasOrder);
  setDisabled("picked-up", !hasOrder);
  setDisabled("delivered", !hasOrder);
  setDisabled("complaint", !hasSession);
  setDisabled("resolve-complaint", !hasComplaint);
  setDisabled("bill", !hasSession);
  setDisabled("accept-bill", !hasBill);
  setDisabled("close-bill", !hasBill);
}

function setBusy(isBusy) {
  buttons.forEach((button) => {
    if (isBusy) button.dataset.wasDisabled = String(button.disabled);
    button.disabled = isBusy || button.dataset.wasDisabled === "true";
    if (!isBusy) delete button.dataset.wasDisabled;
  });
  if (!isBusy) updateButtons();
}

function setDisabled(action, disabled) {
  const button = document.querySelector(`[data-action="${action}"]`);
  if (button) button.disabled = disabled;
}

async function api(path, options = {}) {
  const init = {
    method: options.method || "GET",
    headers: {},
  };
  if (options.body) {
    init.headers["Content-Type"] = "application/json";
    init.body = JSON.stringify(options.body);
  }
  const response = await fetch(path, init);
  const text = await response.text();
  const data = text ? JSON.parse(text) : {};
  if (!response.ok) {
    throw new Error(data.error || `Erro HTTP ${response.status}`);
  }
  return data;
}

function requireSession() {
  if (!state.sessionId) throw new Error("Abra a sessao da mesa primeiro.");
}

function requireOrder() {
  if (!state.orderId) throw new Error("Crie um pedido primeiro.");
}

function requireComplaint() {
  if (!state.complaintId) throw new Error("Abra uma reclamacao primeiro.");
}

function requireBill() {
  if (!state.billId) throw new Error("Solicite a conta primeiro.");
}

function emptyState(message) {
  return `<div class="compact-item"><div class="item-meta">${message}</div></div>`;
}

function formatTime(value) {
  if (!value) return "--:--";
  return new Intl.DateTimeFormat("pt-BR", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).format(new Date(value));
}

function showToast(message) {
  const toast = $("#toast");
  toast.textContent = message;
  toast.classList.add("show");
  window.clearTimeout(showToast.timeout);
  showToast.timeout = window.setTimeout(() => toast.classList.remove("show"), 2800);
}
