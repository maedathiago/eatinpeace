import { useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import type { LucideIcon } from "lucide-react";
import {
  Activity,
  AlertTriangle,
  Bell,
  ChefHat,
  Check,
  ClipboardCheck,
  Clock,
  FileText,
  ListChecks,
  Play,
  Plus,
  QrCode,
  RefreshCw,
  Send,
  ShieldAlert,
  Utensils,
  UserCheck,
} from "lucide-react";
import {
  closeShift,
  createOrder,
  evaluateSLA,
  getMetrics,
  getTimeline,
  health,
  listFloorTasks,
  openComplaint,
  openTableSession,
  requestBill,
  updateBill,
  updateComplaint,
  updateFloorTask,
  updateOrderStatus,
} from "./api";
import type {
  BillHandoff,
  Complaint,
  FloorTask,
  Order,
  OrderStatus,
  Severity,
  ShiftMetrics,
  Source,
  Timeline,
  WriteResponse,
} from "./types";

const RESTAURANT_ID = "rest_pilot";
const SHIFT_ID = "shift_pilot_open";

const TABLES = [
  { id: "table_01", label: "Mesa 1" },
  { id: "table_02", label: "Mesa 2" },
  { id: "table_03", label: "Mesa 3" },
];

const orderStatusLabels: Record<OrderStatus, string> = {
  received: "recebido",
  preparing: "em preparo",
  delay_risk: "risco de atraso",
  delayed: "atrasado",
  ready: "pronto",
  picked_up: "retirado",
  delivered: "entregue",
  cancelled: "cancelado",
};

const taskStatusLabels: Record<string, string> = {
  open: "aberta",
  claimed: "assumida",
  in_progress: "em andamento",
  resolved: "resolvida",
  cancelled: "cancelada",
};

const taskTypeLabels: Record<string, string> = {
  waiter_call: "chamada",
  order_stale: "pedido sem update",
  order_delayed: "pedido atrasado",
  order_ready_pickup: "retirada",
  complaint: "reclamacao",
  bill_handoff: "conta",
};

const complaintStatusLabels: Record<string, string> = {
  open: "aberta",
  classified: "classificada",
  assigned: "atribuida",
  in_progress: "em atendimento",
  resolved: "resolvida",
  reopened: "reaberta",
};

const billStatusLabels: Record<string, string> = {
  requested: "solicitada",
  in_review: "em revisao",
  sent_to_existing_system: "enviada ao sistema",
  awaiting_cashier_action: "aguardando caixa",
  blocked: "bloqueada",
  closed: "fechada",
};

const eventLabels: Record<string, string> = {
  table_session_opened: "sessao aberta",
  order_created: "pedido criado",
  order_status_changed: "pedido atualizado",
  order_stale_detected: "pedido sem update",
  order_delay_risk_detected: "risco de atraso",
  order_delayed: "pedido atrasado",
  order_ready: "pedido pronto",
  order_picked_up: "pedido retirado",
  order_delivered: "pedido entregue",
  waiter_called: "garcom chamado",
  floor_task_created: "tarefa criada",
  floor_task_claimed: "tarefa assumida",
  floor_task_resolved: "tarefa resolvida",
  complaint_opened: "reclamacao aberta",
  complaint_classified: "reclamacao classificada",
  complaint_assigned: "reclamacao atribuida",
  complaint_first_response_recorded: "primeira resposta",
  complaint_resolved: "reclamacao resolvida",
  bill_requested: "conta solicitada",
  bill_handoff_accepted: "handoff aceito",
  bill_handoff_blocked: "handoff bloqueado",
  bill_close_confirmed: "mesa fechada",
  table_session_closed: "sessao fechada",
  shift_closed: "turno fechado",
};

const sourceLabels: Record<Source, string> = {
  customer: "cliente",
  staff: "salao",
  kitchen: "cozinha",
  cashier: "caixa",
  system: "sistema",
  integration: "integracao",
};

type BusyKey =
  | "boot"
  | "refresh"
  | "session"
  | "order"
  | "status"
  | "complaint"
  | "bill"
  | "sla"
  | "shift"
  | `task-${string}`;

interface ToastState {
  message: string;
  tone: "ok" | "bad";
}

interface OrderDraft {
  name: string;
  quantity: number;
  notes: string;
}

interface ComplaintDraft {
  reason: string;
  severity: Severity;
  description: string;
}

export function App() {
  const [selectedTableId, setSelectedTableId] = useState(TABLES[0].id);
  const [sessionId, setSessionId] = useState("");
  const [timeline, setTimeline] = useState<Timeline | null>(null);
  const [tasks, setTasks] = useState<FloorTask[]>([]);
  const [metrics, setMetrics] = useState<ShiftMetrics | null>(null);
  const [apiOnline, setApiOnline] = useState<boolean | null>(null);
  const [busy, setBusy] = useState<BusyKey | null>("boot");
  const [toast, setToast] = useState<ToastState | null>(null);
  const [lastWrite, setLastWrite] = useState<Pick<WriteResponse<unknown>, "event_id" | "correlation_id" | "floor_task_id"> | null>(null);
  const [orderDraft, setOrderDraft] = useState<OrderDraft>({
    name: "Prato piloto",
    quantity: 1,
    notes: "",
  });
  const [complaintDraft, setComplaintDraft] = useState<ComplaintDraft>({
    reason: "service_delay",
    severity: "high",
    description: "Cliente pediu acompanhamento do atraso.",
  });

  const selectedTable = TABLES.find((table) => table.id === selectedTableId) ?? TABLES[0];
  const orders = timeline?.orders ?? [];
  const complaints = timeline?.complaints ?? [];
  const bills = timeline?.bill_handoffs ?? [];
  const events = timeline?.events ?? [];
  const session = timeline?.table_session;

  const latestOrder = lastItem(orders);
  const latestComplaint = lastItem(complaints);
  const latestBill = lastItem(bills);
  const activeTasks = tasks.filter((task) => !["resolved", "cancelled"].includes(task.status));
  const tableTasks = activeTasks.filter((task) => task.table_id === selectedTableId || task.table_session_id === sessionId);
  const sessionIsActive = Boolean(sessionId) && session?.status !== "closed";

  const statusSteps = useMemo(
    () => [
      { status: "preparing" as const, label: "Em preparo", icon: ChefHat, source: "kitchen" as Source, actorId: "staff_kitchen" },
      { status: "ready" as const, label: "Pronto", icon: Bell, source: "kitchen" as Source, actorId: "staff_kitchen" },
      { status: "picked_up" as const, label: "Retirado", icon: Send, source: "staff" as Source, actorId: "staff_waiter" },
      { status: "delivered" as const, label: "Entregue", icon: Check, source: "staff" as Source, actorId: "staff_waiter" },
    ],
    [],
  );

  useEffect(() => {
    void refreshAll("boot");
  }, []);

  useEffect(() => {
    setSessionId("");
    setTimeline(null);
    setLastWrite(null);
  }, [selectedTableId]);

  useEffect(() => {
    if (!toast) return undefined;
    const timer = window.setTimeout(() => setToast(null), 3200);
    return () => window.clearTimeout(timer);
  }, [toast]);

  async function refreshAll(key: BusyKey = "refresh", knownSessionId = sessionId) {
    setBusy(key);
    try {
      const online = await health();
      setApiOnline(online);
      if (!online) {
        throw new Error("API offline");
      }

      const [taskResponse, shiftMetrics] = await Promise.all([listFloorTasks(SHIFT_ID), getMetrics(SHIFT_ID)]);
      setTasks(taskResponse.tasks ?? []);
      setMetrics(shiftMetrics);

      if (knownSessionId) {
        const nextTimeline = await getTimeline(knownSessionId);
        setTimeline(nextTimeline);
      }
    } catch (error) {
      setToast({ message: errorMessage(error), tone: "bad" });
    } finally {
      setBusy(null);
    }
  }

  async function runWrite<T>(
    key: BusyKey,
    action: () => Promise<WriteResponse<T> | { tasks: FloorTask[] }>,
    fallbackMessage: string,
    nextSessionId = sessionId,
  ) {
    setBusy(key);
    try {
      const result = await action();
      if ("event_id" in result) {
        setLastWrite({
          event_id: result.event_id,
          correlation_id: result.correlation_id,
          floor_task_id: result.floor_task_id,
        });
        setToast({ message: result.customer_message || fallbackMessage, tone: "ok" });
      } else {
        setToast({ message: `${result.tasks.length} alerta(s) avaliados.`, tone: "ok" });
      }
      await refreshAll("refresh", nextSessionId);
    } catch (error) {
      setToast({ message: errorMessage(error), tone: "bad" });
    } finally {
      setBusy(null);
    }
  }

  async function handleOpenSession() {
    setBusy("session");
    try {
      const result = await openTableSession({
        restaurantId: RESTAURANT_ID,
        shiftId: SHIFT_ID,
        tableId: selectedTableId,
      });
      setSessionId(result.resource.id);
      setLastWrite({
        event_id: result.event_id,
        correlation_id: result.correlation_id,
        floor_task_id: result.floor_task_id,
      });
      setToast({ message: result.customer_message || "Sessao aberta.", tone: "ok" });
      await refreshAll("refresh", result.resource.id);
    } catch (error) {
      setToast({ message: errorMessage(error), tone: "bad" });
    } finally {
      setBusy(null);
    }
  }

  async function handleCreateOrder() {
    if (!sessionId) return;
    await runWrite(
      "order",
      () =>
        createOrder({
          restaurantId: RESTAURANT_ID,
          shiftId: SHIFT_ID,
          tableId: selectedTableId,
          tableSessionId: sessionId,
          items: [
            {
              name: orderDraft.name.trim() || "Prato piloto",
              quantity: Number.isFinite(orderDraft.quantity) ? orderDraft.quantity : 1,
              notes: orderDraft.notes.trim() || undefined,
            },
          ],
        }),
      "Pedido registrado.",
    );
  }

  async function handleOrderStatus(status: OrderStatus, source: Source, actorId: string) {
    if (!latestOrder) return;
    await runWrite(
      "status",
      () =>
        updateOrderStatus({
          orderId: latestOrder.id,
          status,
          source,
          actorId,
        }),
      "Pedido atualizado.",
    );
  }

  async function handleOpenComplaint() {
    if (!sessionId) return;
    await runWrite(
      "complaint",
      () =>
        openComplaint({
          restaurantId: RESTAURANT_ID,
          shiftId: SHIFT_ID,
          tableId: selectedTableId,
          tableSessionId: sessionId,
          relatedOrderId: latestOrder?.id,
          reason: complaintDraft.reason,
          severity: complaintDraft.severity,
          description: complaintDraft.description.trim() || "Cliente pediu acompanhamento.",
        }),
      "Reclamacao registrada.",
    );
  }

  async function handleResolveComplaint() {
    if (!latestComplaint) return;

    setBusy("complaint");
    try {
      let result: WriteResponse<Complaint> | null = null;
      if (!latestComplaint.responsible_id) {
        result = await updateComplaint({
          complaintId: latestComplaint.id,
          action: "assign",
          actorId: "staff_lead",
          assigneeId: "staff_waiter",
        });
      }
      if (!latestComplaint.first_response_at) {
        result = await updateComplaint({
          complaintId: latestComplaint.id,
          action: "record_first_response",
          actorId: "staff_waiter",
          note: "Mesa recebeu retorno da equipe.",
        });
      }
      result = await updateComplaint({
        complaintId: latestComplaint.id,
        action: "resolve",
        actorId: "staff_waiter",
        resolutionCode: "resolvido_no_salao",
      });
      setLastWrite({
        event_id: result.event_id,
        correlation_id: result.correlation_id,
        floor_task_id: result.floor_task_id,
      });
      setToast({ message: result.customer_message || "Reclamacao resolvida.", tone: "ok" });
      await refreshAll("refresh");
    } catch (error) {
      setToast({ message: errorMessage(error), tone: "bad" });
    } finally {
      setBusy(null);
    }
  }

  async function handleRequestBill() {
    if (!sessionId) return;
    await runWrite(
      "bill",
      () =>
        requestBill({
          restaurantId: RESTAURANT_ID,
          shiftId: SHIFT_ID,
          tableId: selectedTableId,
          tableSessionId: sessionId,
        }),
      "Conta solicitada.",
    );
  }

  async function handleBillAction(action: "accept" | "send_to_existing_system" | "await_cashier" | "confirm_close") {
    if (!latestBill) return;
    await runWrite(
      "bill",
      () =>
        updateBill({
          billId: latestBill.id,
          action,
          actorId: "staff_cashier",
          assigneeId: "staff_cashier",
        }),
      "Handoff atualizado.",
    );
  }

  async function handleTaskAction(task: FloorTask, action: "claim" | "start" | "resolve") {
    await runWrite(
      `task-${task.id}`,
      () =>
        updateFloorTask({
          taskId: task.id,
          action,
          actorId: "staff_waiter",
          assigneeId: "staff_waiter",
          resolutionNote: action === "resolve" ? "Tratado no salao." : undefined,
        }),
      "Tarefa atualizada.",
    );
  }

  const nextOrderStatuses = new Set(nextManualOrderStatuses(latestOrder?.status));
  const billCanAccept = Boolean(latestBill) && !["closed"].includes(latestBill?.status ?? "");
  const billCanSend = Boolean(latestBill) && ["in_review", "awaiting_cashier_action", "sent_to_existing_system"].includes(latestBill?.status ?? "");
  const billCanClose = Boolean(latestBill) && ["in_review", "sent_to_existing_system", "awaiting_cashier_action"].includes(latestBill?.status ?? "");
  const complaintCanResolve = Boolean(latestComplaint) && latestComplaint?.status !== "resolved";

  return (
    <>
      <header className="topbar">
        <div>
          <p className="eyebrow">Eat in Peace</p>
          <h1>Console operacional P0</h1>
        </div>
        <div className="status-strip" aria-live="polite">
          <StatusPill tone={apiOnline ? "ok" : apiOnline === false ? "bad" : "neutral"}>
            {apiOnline ? "api online" : apiOnline === false ? "api offline" : "conectando"}
          </StatusPill>
          <StatusPill>{RESTAURANT_ID}</StatusPill>
          <StatusPill tone={metrics?.closed ? "bad" : "ok"}>{metrics?.closed ? "turno fechado" : SHIFT_ID}</StatusPill>
        </div>
      </header>

      <main className="workspace">
        <section className="control-band" aria-label="Operacao da mesa">
          <div className="table-focus">
            <div className="table-focus-head">
              <span className="table-label">{selectedTable.label}</span>
              <span className={`session-dot ${session?.status ?? "idle"}`} />
            </div>
            <strong>{session ? labelFor(session.status, { active: "sessao ativa", closed: "sessao fechada" }) : "sessao nao aberta"}</strong>
            <small>{sessionId || "aguardando QR ou operador"}</small>
            <div className="segmented" aria-label="Mesa piloto">
              {TABLES.map((table) => (
                <button
                  key={table.id}
                  type="button"
                  className={table.id === selectedTableId ? "selected" : ""}
                  onClick={() => setSelectedTableId(table.id)}
                  disabled={busy !== null}
                >
                  {table.label}
                </button>
              ))}
            </div>
          </div>

          <div className="command-surface">
            <div className="quick-inputs">
              <label>
                Item
                <input
                  value={orderDraft.name}
                  onChange={(event) => setOrderDraft((draft) => ({ ...draft, name: event.target.value }))}
                />
              </label>
              <label>
                Qtd
                <input
                  type="number"
                  min="1"
                  value={orderDraft.quantity}
                  onChange={(event) => setOrderDraft((draft) => ({ ...draft, quantity: Number(event.target.value) || 1 }))}
                />
              </label>
              <label>
                Motivo
                <select
                  value={complaintDraft.reason}
                  onChange={(event) => setComplaintDraft((draft) => ({ ...draft, reason: event.target.value }))}
                >
                  <option value="service_delay">atraso no atendimento</option>
                  <option value="wrong_item">item incorreto</option>
                  <option value="quality_issue">qualidade</option>
                  <option value="need_help">precisa de ajuda</option>
                </select>
              </label>
              <label>
                Severidade
                <select
                  value={complaintDraft.severity}
                  onChange={(event) => setComplaintDraft((draft) => ({ ...draft, severity: event.target.value as Severity }))}
                >
                  <option value="medium">media</option>
                  <option value="high">alta</option>
                  <option value="critical">critica</option>
                  <option value="low">baixa</option>
                </select>
              </label>
            </div>

            <div className="action-grid" aria-label="Acoes operacionais">
              <CommandButton icon={QrCode} busy={busy === "session"} disabled={busy !== null} onClick={handleOpenSession}>
                Abrir sessao
              </CommandButton>
              <CommandButton icon={Plus} busy={busy === "order"} disabled={!sessionIsActive || busy !== null} onClick={handleCreateOrder}>
                Criar pedido
              </CommandButton>
              {statusSteps.map((step) => (
                <CommandButton
                  key={step.status}
                  icon={step.icon}
                  busy={busy === "status"}
                  disabled={!nextOrderStatuses.has(step.status) || busy !== null}
                  onClick={() => handleOrderStatus(step.status, step.source, step.actorId)}
                >
                  {step.label}
                </CommandButton>
              ))}
              <CommandButton icon={ShieldAlert} busy={busy === "sla"} disabled={busy !== null} onClick={() => runWrite("sla", () => evaluateSLA(SHIFT_ID), "SLA avaliado.")}>
                Avaliar SLA
              </CommandButton>
              <CommandButton icon={AlertTriangle} busy={busy === "complaint"} disabled={!sessionIsActive || busy !== null} onClick={handleOpenComplaint}>
                Reclamacao
              </CommandButton>
              <CommandButton icon={ClipboardCheck} busy={busy === "complaint"} disabled={!complaintCanResolve || busy !== null} onClick={handleResolveComplaint}>
                Resolver reclamacao
              </CommandButton>
              <CommandButton icon={FileText} busy={busy === "bill"} disabled={!sessionIsActive || busy !== null} onClick={handleRequestBill}>
                Solicitar conta
              </CommandButton>
              <CommandButton icon={UserCheck} busy={busy === "bill"} disabled={!billCanAccept || busy !== null} onClick={() => handleBillAction("accept")}>
                Aceitar conta
              </CommandButton>
              <CommandButton icon={Send} busy={busy === "bill"} disabled={!billCanSend || busy !== null} onClick={() => handleBillAction("send_to_existing_system")}>
                Sistema existente
              </CommandButton>
              <CommandButton icon={Check} busy={busy === "bill"} disabled={!billCanClose || busy !== null} onClick={() => handleBillAction("confirm_close")}>
                Confirmar fechamento
              </CommandButton>
              <CommandButton icon={RefreshCw} busy={busy === "refresh"} disabled={busy !== null} onClick={() => refreshAll("refresh")}>
                Atualizar
              </CommandButton>
              <CommandButton icon={Clock} busy={busy === "shift"} disabled={busy !== null || metrics?.closed} onClick={() => runWrite("shift", () => closeShift(SHIFT_ID), "Turno fechado.")}>
                Fechar turno
              </CommandButton>
            </div>
          </div>
        </section>

        <section className="overview-grid" aria-label="Estado operacional">
          <Metric icon={Activity} label="Eventos" value={metrics?.total_events ?? 0} />
          <Metric icon={ListChecks} label="Tarefas abertas" value={metrics?.open_tasks ?? 0} tone={metrics?.open_tasks ? "watch" : "calm"} />
          <Metric icon={Clock} label="Tarefas vencidas" value={metrics?.overdue_tasks ?? 0} tone={metrics?.overdue_tasks ? "risk" : "calm"} />
          <Metric icon={Utensils} label="Prontos parados" value={metrics?.ready_orders_awaiting_pickup ?? 0} tone={metrics?.ready_orders_awaiting_pickup ? "risk" : "calm"} />
          <Metric icon={AlertTriangle} label="Reclamacoes sem dono" value={metrics?.complaints_without_owner_over_sla ?? 0} tone={metrics?.complaints_without_owner_over_sla ? "risk" : "calm"} />
          <Metric icon={FileText} label="Contas sem dono" value={metrics?.bill_handoffs_without_owner_over_sla ?? 0} tone={metrics?.bill_handoffs_without_owner_over_sla ? "risk" : "calm"} />
        </section>

        <section className="work-grid">
          <Panel title="Fila do salao" meta={`${activeTasks.length} abertas`}>
            <div className="task-list">
              {activeTasks.length ? (
                activeTasks.map((task) => (
                  <TaskCard key={task.id} task={task} busy={busy === `task-${task.id}`} onAction={handleTaskAction} />
                ))
              ) : (
                <EmptyState message="Nenhuma tarefa operacional aberta." />
              )}
            </div>
          </Panel>

          <Panel title="Mesa em foco" meta={`${tableTasks.length} tarefas`}>
            <div className="stack-list">
              <OrderCard order={latestOrder} />
              <ComplaintCard complaint={latestComplaint} />
              <BillCard bill={latestBill} />
              {lastWrite && <AuditStrip write={lastWrite} />}
            </div>
          </Panel>

          <Panel title="Linha do tempo" meta={`${events.length} eventos`} wide>
            <ol className="timeline">
              {events.length ? (
                [...events].reverse().map((event) => (
                  <li key={event.id}>
                    <span className="time">{formatTime(event.occurred_at)}</span>
                    <div>
                      <div className="event-type">{eventLabels[event.event_type] ?? event.event_type}</div>
                      <div className="event-meta">
                        {sourceLabels[event.source] ?? event.source} · {event.correlation_id}
                      </div>
                    </div>
                  </li>
                ))
              ) : (
                <EmptyState message="Sem eventos para esta sessao." />
              )}
            </ol>
          </Panel>

          <Panel title="Pedidos" meta={`${orders.length} pedidos`}>
            <div className="stack-list">
              {orders.length ? orders.map((order) => <OrderCard key={order.id} order={order} />) : <EmptyState message="Sem pedido operacional." />}
            </div>
          </Panel>

          <Panel title="Reclamacoes" meta={`${complaints.length} tickets`}>
            <div className="stack-list">
              {complaints.length ? complaints.map((complaint) => <ComplaintCard key={complaint.id} complaint={complaint} />) : <EmptyState message="Sem reclamacao registrada." />}
            </div>
          </Panel>

          <Panel title="Conta" meta={latestBill ? billStatusLabels[latestBill.status] : "sem handoff"}>
            <div className="stack-list">
              {bills.length ? bills.map((bill) => <BillCard key={bill.id} bill={bill} />) : <EmptyState message="Conta ainda nao solicitada." />}
            </div>
          </Panel>
        </section>
      </main>

      <aside className={`toast ${toast ? "show" : ""} ${toast?.tone ?? "ok"}`} aria-live="polite">
        {toast?.message}
      </aside>
    </>
  );
}

function CommandButton({
  icon: Icon,
  children,
  busy,
  disabled,
  onClick,
}: {
  icon: LucideIcon;
  children: string;
  busy?: boolean;
  disabled?: boolean;
  onClick: () => void;
}) {
  return (
    <button type="button" className="command-button" onClick={onClick} disabled={disabled} title={children}>
      <Icon aria-hidden="true" size={18} strokeWidth={2.2} />
      <span>{busy ? "Processando" : children}</span>
    </button>
  );
}

function StatusPill({ children, tone = "neutral" }: { children: string; tone?: "ok" | "bad" | "neutral" }) {
  return <span className={`pill ${tone}`}>{children}</span>;
}

function Metric({ icon: Icon, label, value, tone = "neutral" }: { icon: LucideIcon; label: string; value: number; tone?: "neutral" | "calm" | "watch" | "risk" }) {
  return (
    <article className={`metric ${tone}`}>
      <div>
        <Icon aria-hidden="true" size={18} />
        <span>{label}</span>
      </div>
      <strong>{value}</strong>
    </article>
  );
}

function Panel({ title, meta, wide, children }: { title: string; meta: string; wide?: boolean; children: ReactNode }) {
  return (
    <article className={`panel ${wide ? "wide" : ""}`}>
      <div className="panel-head">
        <h2>{title}</h2>
        <span>{meta}</span>
      </div>
      {children}
    </article>
  );
}

function TaskCard({ task, busy, onAction }: { task: FloorTask; busy: boolean; onAction: (task: FloorTask, action: "claim" | "start" | "resolve") => void }) {
  const isClosed = ["resolved", "cancelled"].includes(task.status);
  return (
    <div className="task">
      <div>
        <div className="item-title">
          {taskTypeLabels[task.type] ?? task.type} · {taskStatusLabels[task.status] ?? task.status}
        </div>
        <div className="item-meta">{task.priority_reason}</div>
        <div className="item-meta">
          {task.table_id || "sem mesa"} · {task.responsible_id || "sem responsavel"} · {task.due_at ? `vence ${formatTime(task.due_at)}` : "sem SLA"}
        </div>
      </div>
      <div className="task-actions">
        <SeverityBadge severity={task.severity} />
        {!isClosed && (
          <div className="mini-actions">
            {task.status === "open" && (
              <button type="button" onClick={() => onAction(task, "claim")} disabled={busy}>
                Assumir
              </button>
            )}
            {task.status !== "in_progress" && task.status !== "resolved" && (
              <button type="button" onClick={() => onAction(task, "start")} disabled={busy}>
                Iniciar
              </button>
            )}
            <button type="button" onClick={() => onAction(task, "resolve")} disabled={busy}>
              Resolver
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

function OrderCard({ order }: { order?: Order }) {
  if (!order) return <EmptyState message="Sem pedido operacional." />;
  return (
    <div className="compact-item">
      <div className="item-title">{orderStatusLabels[order.status]}</div>
      <div className="item-meta">{order.items?.map((item) => `${item.quantity}x ${item.name}`).join(", ") || "sem itens"}</div>
      <StatusTrack current={order.status} />
    </div>
  );
}

function ComplaintCard({ complaint }: { complaint?: Complaint }) {
  if (!complaint) return <EmptyState message="Sem reclamacao registrada." />;
  return (
    <div className="compact-item">
      <div className="card-row">
        <div>
          <div className="item-title">
            {complaint.reason} · {complaintStatusLabels[complaint.status] ?? complaint.status}
          </div>
          <div className="item-meta">{complaint.responsible_id || "sem responsavel"}</div>
        </div>
        <SeverityBadge severity={complaint.severity} />
      </div>
    </div>
  );
}

function BillCard({ bill }: { bill?: BillHandoff }) {
  if (!bill) return <EmptyState message="Conta ainda nao solicitada." />;
  return (
    <div className="compact-item">
      <div className="item-title">{billStatusLabels[bill.status] ?? bill.status}</div>
      <div className="item-meta">{bill.block_reason || bill.responsible_id || "aguardando responsavel"}</div>
    </div>
  );
}

function AuditStrip({ write }: { write: Pick<WriteResponse<unknown>, "event_id" | "correlation_id" | "floor_task_id"> }) {
  return (
    <div className="audit-strip">
      <span>{write.event_id}</span>
      <span>{write.correlation_id}</span>
      {write.floor_task_id && <span>{write.floor_task_id}</span>}
    </div>
  );
}

function StatusTrack({ current }: { current: OrderStatus }) {
  const steps: OrderStatus[] = ["received", "preparing", "ready", "picked_up", "delivered"];
  const normalizedCurrent: OrderStatus = current === "delay_risk" || current === "delayed" ? "preparing" : current;
  const currentIndex = steps.indexOf(normalizedCurrent);
  return (
    <div className={`status-track ${current === "delay_risk" || current === "delayed" ? "risk" : ""} ${current === "cancelled" ? "cancelled" : ""}`} aria-label="Status do pedido">
      {steps.map((step, index) => (
        <span key={step} className={index <= currentIndex ? "done" : ""} title={orderStatusLabels[step]} />
      ))}
    </div>
  );
}

function SeverityBadge({ severity }: { severity: Severity }) {
  return <span className={`severity ${severity}`}>{labelFor(severity, { low: "baixa", medium: "media", high: "alta", critical: "critica" })}</span>;
}

function EmptyState({ message }: { message: string }) {
  return (
    <div className="empty-state">
      <span>{message}</span>
    </div>
  );
}

function labelFor<T extends string>(value: T, labels: Partial<Record<T, string>>) {
  return labels[value] ?? value;
}

function lastItem<T>(items: T[]): T | undefined {
  return items.length ? items[items.length - 1] : undefined;
}

function nextManualOrderStatuses(status?: OrderStatus): OrderStatus[] {
  switch (status) {
    case "received":
      return ["preparing"];
    case "preparing":
    case "delay_risk":
    case "delayed":
      return ["ready"];
    case "ready":
      return ["picked_up"];
    case "picked_up":
      return ["delivered"];
    default:
      return [];
  }
}

function formatTime(value?: string) {
  if (!value) return "--:--";
  return new Intl.DateTimeFormat("pt-BR", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).format(new Date(value));
}

function errorMessage(error: unknown) {
  return error instanceof Error ? error.message : "Erro inesperado.";
}
