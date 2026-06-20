# Fase 007: fila operacional unica do salao

## Objetivo

Transformar chamados, pedidos prontos, atrasos, reclamacoes e conta em uma fila priorizada e auditavel.

## Escopo

- `floor_tasks` com tipo, severidade, status, timestamps, responsavel e motivo de prioridade.
- `POST /v1/floor-tasks`.
- `GET /v1/floor-tasks`.
- `PATCH /v1/floor-tasks/{id}`.
- Ordenacao por severidade, vencimento de SLA e ordem de criacao.
- Claim, inicio, reatribuicao com justificativa, resolucao e cancelamento com motivo.

## Branch

```text
phase/007-floor-queue
```

Nesta primeira entrega executavel, a fila foi implementada junto da fundacao P0 porque os fluxos de SLA, reclamacao e conta dependem dela.

## Testes

Coberto por:

```bash
GOCACHE=/tmp/eatinpeace-go-build go test ./...
make test-e2e
```

Testes relevantes:

- ordenacao por severidade, SLA e FIFO;
- reatribuicao exigindo motivo;
- criacao automatica de tarefas a partir de pedido pronto, SLA, reclamacao e conta;
- E2E com fila sendo criada e resolvida nos fluxos P0.

## Escopo fora

Nao ha automacao rigida de equipe nem rebalanceamento completo por setor. Humanos continuam no controle.
