# Fase 008: atraso de cozinha e pedido pronto parado

## Objetivo

Avisar o salao antes da reclamacao e impedir que pedido pronto fique parado sem tarefa.

## Escopo

- Politica inicial de SLA em dominio.
- `POST /v1/sla/evaluate`.
- `order_stale_detected`.
- `order_delay_risk_detected`.
- `order_delayed`.
- Tarefa `order_ready_pickup` quando pedido fica pronto.
- Resolucao da tarefa de retirada quando pedido vai para `picked_up` ou `delivered`.

## Branch

```text
phase/008-kitchen-sla-ready-orders
```

Nesta primeira entrega executavel, o avaliador de SLA foi implementado junto da fundacao P0 para validar a fila operacional com eventos reais.

## Testes

Coberto por:

```bash
GOCACHE=/tmp/eatinpeace-go-build go test ./...
make test-e2e
```

Testes relevantes:

- progressao `preparing -> delay_risk -> delayed`;
- tarefa de pedido pronto;
- pedido pronto, retirada e entrega no E2E.

## Escopo fora

Nao ha previsao sofisticada de cozinha, estoque, compra, custo de item ou gestao de cardapio.
