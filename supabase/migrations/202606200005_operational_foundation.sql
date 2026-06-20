create extension if not exists pgcrypto;

create table if not exists restaurants (
    id uuid primary key default gen_random_uuid(),
    name text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists service_shifts (
    id uuid primary key default gen_random_uuid(),
    restaurant_id uuid not null references restaurants(id),
    name text not null,
    status text not null default 'open' check (status in ('open', 'closed')),
    opened_at timestamptz not null default now(),
    closed_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists tables (
    id uuid primary key default gen_random_uuid(),
    restaurant_id uuid not null references restaurants(id),
    label text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists staff_members (
    id uuid primary key default gen_random_uuid(),
    restaurant_id uuid not null references restaurants(id),
    name text not null,
    role text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists table_sessions (
    id uuid primary key default gen_random_uuid(),
    restaurant_id uuid not null references restaurants(id),
    shift_id uuid not null references service_shifts(id),
    table_id uuid not null references tables(id),
    status text not null check (status in ('active', 'closed')),
    opened_at timestamptz not null default now(),
    closed_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists orders (
    id uuid primary key default gen_random_uuid(),
    restaurant_id uuid not null references restaurants(id),
    shift_id uuid not null references service_shifts(id),
    table_id uuid not null references tables(id),
    table_session_id uuid not null references table_sessions(id),
    status text not null check (status in ('received', 'preparing', 'delay_risk', 'delayed', 'ready', 'picked_up', 'delivered', 'cancelled')),
    ready_at timestamptz,
    delivered_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists order_items (
    id uuid primary key default gen_random_uuid(),
    order_id uuid not null references orders(id),
    name text not null,
    quantity integer not null default 1 check (quantity > 0),
    notes text,
    created_at timestamptz not null default now()
);

create table if not exists operational_events (
    id uuid primary key default gen_random_uuid(),
    restaurant_id uuid not null references restaurants(id),
    shift_id uuid references service_shifts(id),
    table_id uuid references tables(id),
    table_session_id uuid references table_sessions(id),
    event_type text not null,
    occurred_at timestamptz not null,
    source text not null check (source in ('customer', 'staff', 'kitchen', 'cashier', 'system', 'integration')),
    actor_id uuid references staff_members(id),
    payload jsonb not null default '{}'::jsonb,
    correlation_id text not null,
    created_at timestamptz not null default now()
);

create table if not exists floor_tasks (
    id uuid primary key default gen_random_uuid(),
    restaurant_id uuid not null references restaurants(id),
    shift_id uuid not null references service_shifts(id),
    table_id uuid references tables(id),
    table_session_id uuid references table_sessions(id),
    type text not null check (type in ('waiter_call', 'order_stale', 'order_delayed', 'order_ready_pickup', 'complaint', 'bill_handoff')),
    status text not null check (status in ('open', 'claimed', 'in_progress', 'resolved', 'cancelled')),
    severity text not null check (severity in ('low', 'medium', 'high', 'critical')),
    priority_reason text not null,
    responsible_id uuid references staff_members(id),
    related_order_id uuid references orders(id),
    related_complaint_id uuid,
    related_bill_handoff_id uuid,
    due_at timestamptz,
    claimed_at timestamptz,
    started_at timestamptz,
    resolved_at timestamptz,
    source_event_id uuid references operational_events(id),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists complaints (
    id uuid primary key default gen_random_uuid(),
    restaurant_id uuid not null references restaurants(id),
    shift_id uuid not null references service_shifts(id),
    table_id uuid references tables(id),
    table_session_id uuid references table_sessions(id),
    order_id uuid references orders(id),
    reason text not null,
    severity text not null check (severity in ('low', 'medium', 'high', 'critical')),
    status text not null check (status in ('open', 'classified', 'assigned', 'in_progress', 'resolved', 'reopened')),
    responsible_id uuid references staff_members(id),
    opened_at timestamptz not null default now(),
    first_response_at timestamptz,
    resolved_at timestamptz,
    resolution_summary text,
    escalated_to_leader_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

alter table floor_tasks
    add constraint floor_tasks_related_complaint_id_fkey
    foreign key (related_complaint_id) references complaints(id)
    not valid;

create table if not exists bill_handoffs (
    id uuid primary key default gen_random_uuid(),
    restaurant_id uuid not null references restaurants(id),
    shift_id uuid not null references service_shifts(id),
    table_id uuid not null references tables(id),
    table_session_id uuid references table_sessions(id),
    status text not null check (status in ('requested', 'in_review', 'sent_to_existing_system', 'awaiting_cashier_action', 'blocked', 'closed')),
    responsible_id uuid references staff_members(id),
    block_reason text,
    requested_at timestamptz not null default now(),
    accepted_at timestamptz,
    closed_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

alter table floor_tasks
    add constraint floor_tasks_related_bill_handoff_id_fkey
    foreign key (related_bill_handoff_id) references bill_handoffs(id)
    not valid;

create table if not exists sla_policies (
    id uuid primary key default gen_random_uuid(),
    restaurant_id uuid not null references restaurants(id),
    target_type text not null,
    limit_seconds integer not null check (limit_seconds > 0),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (restaurant_id, target_type)
);

create index if not exists service_shifts_restaurant_status_idx on service_shifts (restaurant_id, status);
create index if not exists tables_restaurant_idx on tables (restaurant_id);
create index if not exists table_sessions_shift_status_idx on table_sessions (shift_id, status);
create index if not exists orders_shift_status_idx on orders (shift_id, status, updated_at);
create index if not exists operational_events_shift_timeline_idx on operational_events (shift_id, table_session_id, occurred_at);
create index if not exists operational_events_correlation_idx on operational_events (correlation_id);
create index if not exists floor_tasks_shift_queue_idx on floor_tasks (shift_id, status, severity, due_at, created_at);
create index if not exists floor_tasks_related_order_idx on floor_tasks (related_order_id);
create index if not exists complaints_shift_status_idx on complaints (shift_id, status, opened_at);
create index if not exists bill_handoffs_shift_status_idx on bill_handoffs (shift_id, status, requested_at);
