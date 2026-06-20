# Fase 009: reclamacao com ticket, responsavel e SLA

## Objetivo

Garantir que reclamacoes sejam registradas, classificadas, atribuidas e resolvidas com justica operacional.

## Escopo

- `POST /v1/complaints`.
- `PATCH /v1/complaints/{id}`.
- Severidade, motivo estruturado, responsavel, primeira resposta, resolucao e reabertura.
- Criacao automatica de tarefa do salao.
- Escalonamento de reclamacao sem responsavel acima do SLA.

## Branch

```text
phase/009-complaints-ticket-sla
```

Nesta primeira entrega executavel, reclamacoes foram implementadas junto da fundacao P0 para garantir integracao com eventos, fila e metricas.

## Testes

Coberto por:

```bash
GOCACHE=/tmp/eatinpeace-go-build go test ./...
make test-e2e
```

Testes relevantes:

- severidade e ordenacao;
- reclamacao criando tarefa;
- escalonamento sem responsavel;
- fluxo E2E de abrir, atribuir, registrar primeira resposta, resolver, reabrir e resolver.

## Escopo fora

Nao ha desconto automatico, CRM, chat livre como canal primario, pagamento ou conciliacao financeira.
