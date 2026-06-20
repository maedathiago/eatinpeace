# Fase 010: conta solicitada com handoff rastreavel

## Objetivo

Coordenar a solicitacao de conta sem assumir operacao financeira.

## Escopo

- `POST /v1/bill-handoffs`.
- `PATCH /v1/bill-handoffs/{id}`.
- Estados `requested`, `in_review`, `sent_to_existing_system`, `awaiting_cashier_action`, `blocked` e `closed`.
- Bloqueio operacional quando ha pedido nao entregue ou reclamacao aberta.
- Criacao de tarefa para caixa ou salao.
- Fechamento manual da sessao de mesa em `confirm_close`.

## Branch

```text
phase/010-bill-handoff
```

Nesta primeira entrega executavel, o handoff de conta foi implementado junto da fundacao P0 para testar bloqueios operacionais antes de qualquer financeiro.

## Testes

Coberto por:

```bash
GOCACHE=/tmp/eatinpeace-go-build go test ./...
make test-e2e
```

Testes relevantes:

- bloqueio por pedido pendente;
- bloqueio por reclamacao aberta;
- aceite de handoff;
- confirmacao manual de fechamento;
- ausencia de campos financeiros no E2E.

## Escopo fora

Nao calcula total, taxa, imposto, split, pagamento, dados de cartao, nota fiscal ou conciliacao.
