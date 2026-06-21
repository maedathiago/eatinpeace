# Fase 012: console operacional P0

## Objetivo

Criar uma primeira UI local em React para visualizar e acionar os fluxos P0 ja existentes na API.

## Escopo

- Tela servida em `/` pelo backend Go.
- Fonte React em `web/`.
- Assets estaticos gerados em `internal/httpapi/static`.
- Acoes de sessao, pedido, fila, reclamacao, conta, fechamento de turno e metricas.
- Uso dos fixtures locais `rest_pilot`, `shift_pilot_open` e `table_01`.

## Branch

```text
phase/012-operational-console-ui
```

## Entregaveis

- Console operacional P0 em React, TypeScript e Vite.
- Build frontend embutido no backend Go.
- Testes para rota raiz e assets estaticos.
- Documentacao de desenvolvimento local atualizada.

## Testes

Executados:

```bash
GOCACHE=/tmp/eatinpeace-go-build go test ./...
make test-e2e
GOCACHE=/tmp/eatinpeace-go-build go vet ./...
cd web && npm run build
```

## Escopo fora

Esta UI nao cria pagamento, POS, fiscal, estoque, delivery, cardapio completo ou autentificacao. Ela e uma console local para operar os eventos P0.
