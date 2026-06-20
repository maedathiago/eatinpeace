# Fase 006: linha do tempo visivel da mesa

## Objetivo

Permitir que cliente e equipe vejam a sessao de mesa e o progresso operacional do pedido.

## Escopo

- `POST /v1/table-sessions`.
- `POST /v1/orders`.
- `PATCH /v1/orders/{id}/status`.
- `GET /v1/table-sessions/{id}/timeline`.
- Mensagens curtas para cliente sem promessa de tempo preciso.
- Evento de pedido sem atualizacao detectado pelo avaliador de SLA.

## Branch

```text
phase/006-table-timeline
```

Nesta primeira entrega executavel, o fluxo foi implementado junto da fundacao P0 para evitar telas ou endpoints sobre estados frageis.

## Testes

Coberto por:

```bash
GOCACHE=/tmp/eatinpeace-go-build go test ./...
make test-e2e
```

Testes relevantes:

- transicao de status do pedido;
- bloqueio de salto indevido para `delivered`;
- abertura de sessao, criacao de pedido, atualizacao e consulta de timeline no E2E.

## Escopo fora

Nao ha cardapio completo, preco, estoque, POS, pagamento ou promessa automatica de tempo exato.
