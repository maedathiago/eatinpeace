# Fase 011: piloto P0 e metricas de turno

## Objetivo

Validar o pacote P0 em fluxo unico e gerar evidencia operacional para decisao de produto.

## Escopo

- `POST /v1/service-shifts/{id}/close`.
- `GET /v1/service-shifts/{id}/metrics`.
- Agregacao de eventos, pedidos, tarefas, reclamacoes e handoffs.
- Fixture E2E de turno com sessao, pedido, atraso, pedido pronto, reclamacao, conta e fechamento.

## Branch

```text
phase/011-p0-pilot-metrics
```

Nesta primeira entrega executavel, metricas P0 foram implementadas junto da fundacao para provar que os eventos ja sustentam leitura de gargalos.

## Testes

Coberto por:

```bash
GOCACHE=/tmp/eatinpeace-go-build go test ./...
make test-e2e
```

Testes relevantes:

- agregacao de backlog operacional;
- turno fechado refletido em metricas;
- E2E completo de P0.

## Escopo fora

Nao ha dashboard visual, analytics comercial amplo, BI externo, financeiro, fiscal, estoque, POS ou delivery.
