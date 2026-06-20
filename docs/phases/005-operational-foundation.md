# Fase 005: fundacao operacional Go + Supabase

## Objetivo

Criar a base executavel para registrar eventos, persistir estados operacionais, versionar o schema Supabase/Postgres e testar regras de dominio.

## Escopo

- Modulo Go `github.com/maedathiago/eatinpeace`.
- API HTTP em `cmd/api`.
- Dominio de eventos, sessoes, pedidos, tarefas, reclamacoes, handoff de conta, SLA e metricas.
- Store em memoria para desenvolvimento e E2E local.
- Repositorio Postgres baseado em `database/sql`, pronto para receber driver/configuracao de ambiente.
- Migration Supabase/Postgres com entidades minimas.
- Fixtures locais de restaurante, turno, mesa e equipe.
- Makefile com comandos oficiais.

## Branch

```text
phase/005-operational-foundation
```

## Entregaveis

- `go.mod`, `cmd/api`, `internal/domain`, `internal/application`, `internal/httpapi`, `internal/storage`.
- `supabase/migrations/202606200005_operational_foundation.sql`.
- `docs/api/p0-operational-api.md`.
- `docs/operations/local-development.md`.
- `docs/engineering/testing-strategy.md` atualizado.
- Este documento de fase.

## Testes

Executados:

```bash
GOCACHE=/tmp/eatinpeace-go-build go test ./...
make test-e2e
```

Cobertura criada:

- validacao de evento e fonte;
- transicoes de pedido com override auditado;
- ordenacao de tarefas por severidade, SLA e FIFO;
- avaliacao de SLA;
- migration com entidades minimas;
- E2E P0 sobre API HTTP.

## Excecao de historico

A branch preserva dois commits porque o repositorio Postgres e o ajuste do tipo de IDs da migration foram adicionados como correcao de revisao depois do primeiro pacote executavel ja publicado. Os dois commits pertencem a mesma fase e foram validados juntos antes da integracao final.

## Escopo fora

Esta fase nao calcula impostos, processa pagamento, guarda cartao, emite fiscal, concilia caixa, gerencia estoque, implementa POS ou delivery.
