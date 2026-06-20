# API operacional P0

Esta API entrega a primeira camada executável do Eat in Peace para saneamento de reclamações operacionais. Ela não é POS, financeiro, fiscal, estoque, pagamento ou delivery.

Todas as escritas retornam:

- `event_id`;
- `correlation_id`;
- `customer_message`, quando a ação tiver retorno ao cliente;
- `floor_task_id`, quando a ação gerar trabalho humano.

## Fluxos

### Sessão de mesa

```text
POST /v1/table-sessions
GET  /v1/table-sessions/{id}/timeline
```

Abre a sessão operacional da mesa e permite consultar a linha do tempo com sessão, mesa, pedidos, tarefas, reclamações, handoffs de conta e eventos.

### Pedido operacional

```text
POST  /v1/orders
PATCH /v1/orders/{id}/status
```

Estados permitidos:

```text
received -> preparing -> delay_risk -> delayed -> ready -> picked_up -> delivered
```

`cancelled` é terminal. Saltos fora da sequência exigem `override=true` e `reason`.

Quando o pedido fica `ready`, o sistema cria tarefa `order_ready_pickup`. A entrega normal exige `picked_up` antes de `delivered`.

### Fila operacional

```text
POST  /v1/floor-tasks
GET   /v1/floor-tasks
PATCH /v1/floor-tasks/{id}
```

A fila é ordenada por severidade, vencimento de SLA e ordem de criação. Reatribuição e cancelamento exigem justificativa.

Ações:

```text
claim
start
reassign
resolve
cancel
```

### Reclamações

```text
POST  /v1/complaints
PATCH /v1/complaints/{id}
```

Uma reclamação cria ticket e tarefa de salão. Reclamações sem responsável acima do SLA são escaladas pelo avaliador de SLA.

Ações:

```text
classify
assign
record_first_response
resolve
reopen
```

### Conta como handoff

```text
POST  /v1/bill-handoffs
PATCH /v1/bill-handoffs/{id}
```

A conta é apenas handoff operacional. O fluxo não calcula total, taxa, imposto, cartão, nota fiscal ou conciliação.

Se houver pedido não entregue ou reclamação aberta, o handoff fica `blocked` e uma tarefa operacional é criada.

Ações:

```text
accept
send_to_existing_system
await_cashier
block
confirm_close
```

`confirm_close` fecha a sessão de mesa, mas não executa pagamento.

### SLA e métricas

```text
POST /v1/sla/evaluate
POST /v1/service-shifts/{id}/close
GET  /v1/service-shifts/{id}/metrics
```

O avaliador de SLA detecta pedido sem atualização, risco de atraso, atraso, pedido pronto parado e reclamação sem responsável.

As métricas do turno agregam eventos, pedidos, tarefas, reclamações e handoffs para mostrar gargalos operacionais.

## Fixtures locais

Ao rodar `go run ./cmd/api`, a API cria fixtures em memória:

- restaurante `rest_pilot`;
- turno `shift_pilot_open`;
- mesas `table_01`, `table_02`, `table_03`;
- equipe `staff_waiter`, `staff_lead`, `staff_kitchen`, `staff_cashier`.
