create extension if not exists pgcrypto;

create table if not exists restaurants (
    id text primary key default gen_random_uuid()::text,
    name text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists service_shifts (
    id text primary key default gen_random_uuid()::text,
    restaurant_id text not null references restaurants(id),
    name text not null,
    status text not null default 'open' check (status in ('open', 'closed')),
    opened_at timestamptz not null default now(),
    closed_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists tables (
    id text primary key default gen_random_uuid()::text,
    restaurant_id text not null references restaurants(id),
    label text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists staff_members (
    id text primary key default gen_random_uuid()::text,
    restaurant_id text not null references restaurants(id),
    name text not null,
    role text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists table_sessions (
    id text primary key default gen_random_uuid()::text,
    restaurant_id text not null references restaurants(id),
    shift_id text not null references service_shifts(id),
    table_id text not null references tables(id),
    status text not null check (status in ('active', 'closed')),
    opened_at timestamptz not null default now(),
    closed_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists orders (
    id text primary key default gen_random_uuid()::text,
    restaurant_id text not null references restaurants(id),
    shift_id text not null references service_shifts(id),
    table_id text not null references tables(id),
    table_session_id text not null references table_sessions(id),
    status text not null check (status in ('received', 'preparing', 'delay_risk', 'delayed', 'ready', 'picked_up', 'delivered', 'cancelled')),
    ready_at timestamptz,
    delivered_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists order_items (
    id text primary key default gen_random_uuid()::text,
    order_id text not null references orders(id),
    name text not null,
    quantity integer not null default 1 check (quantity > 0),
    notes text,
    created_at timestamptz not null default now()
);

create table if not exists operational_events (
    id text primary key default gen_random_uuid()::text,
    restaurant_id text not null references restaurants(id),
    shift_id text references service_shifts(id),
    table_id text references tables(id),
    table_session_id text references table_sessions(id),
    event_type text not null,
    occurred_at timestamptz not null,
    source text not null check (source in ('customer', 'staff', 'kitchen', 'cashier', 'system', 'integration')),
    actor_id text references staff_members(id),
    payload jsonb not null default '{}'::jsonb,
    correlation_id text not null,
    created_at timestamptz not null default now()
);

create table if not exists floor_tasks (
    id text primary key default gen_random_uuid()::text,
    restaurant_id text not null references restaurants(id),
    shift_id text not null references service_shifts(id),
    table_id text references tables(id),
    table_session_id text references table_sessions(id),
    type text not null check (type in ('waiter_call', 'order_stale', 'order_delayed', 'order_ready_pickup', 'complaint', 'bill_handoff')),
    status text not null check (status in ('open', 'claimed', 'in_progress', 'resolved', 'cancelled')),
    severity text not null check (severity in ('low', 'medium', 'high', 'critical')),
    priority_reason text not null,
    responsible_id text references staff_members(id),
    related_order_id text references orders(id),
    related_complaint_id text,
    related_bill_handoff_id text,
    due_at timestamptz,
    claimed_at timestamptz,
    started_at timestamptz,
    resolved_at timestamptz,
    source_event_id text references operational_events(id),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists complaints (
    id text primary key default gen_random_uuid()::text,
    restaurant_id text not null references restaurants(id),
    shift_id text not null references service_shifts(id),
    table_id text references tables(id),
    table_session_id text references table_sessions(id),
    order_id text references orders(id),
    reason text not null,
    severity text not null check (severity in ('low', 'medium', 'high', 'critical')),
    status text not null check (status in ('open', 'classified', 'assigned', 'in_progress', 'resolved', 'reopened')),
    responsible_id text references staff_members(id),
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
    id text primary key default gen_random_uuid()::text,
    restaurant_id text not null references restaurants(id),
    shift_id text not null references service_shifts(id),
    table_id text not null references tables(id),
    table_session_id text references table_sessions(id),
    status text not null check (status in ('requested', 'in_review', 'sent_to_existing_system', 'awaiting_cashier_action', 'blocked', 'closed')),
    responsible_id text references staff_members(id),
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
    id text primary key default gen_random_uuid()::text,
    restaurant_id text not null references restaurants(id),
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
