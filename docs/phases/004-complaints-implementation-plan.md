# Fase 004: plano de implementação das reclamações

## Objetivo

Transformar os achados de reclamações 2025+ e a estratégia de produto em um plano de implementação sequenciado para sanear os problemas operacionais priorizados pelo Eat in Peace.

## Escopo

- Definir o critério de saneamento para reclamações operacionais.
- Separar o que entra no MVP do que continua fora de escopo.
- Propor a base técnica inicial em Go + Supabase.
- Definir entidades, eventos mínimos e fases futuras de implementação.
- Amarrar cada fatia a ator, dor, evento, mudança de estado e métrica de sucesso.
- Definir testes obrigatórios para cada tipo de problema saneado.

## Branch

```text
phase/004-complaints-implementation-plan
```

## Entregáveis

- `docs/implementation-plan-complaints-mvp.md` criado.
- `docs/product-strategy-complaints-2025.md` atualizado para apontar para o plano.
- `docs/phases/README.md` atualizado.
- `README.md` atualizado.
- Este documento de fase criado.

## Testes

Esta fase não cria código executável. A verificação aplicável é documental:

- o plano cobre o pacote P0 definido na estratégia;
- cada fase futura identifica ator, dor, evento, mudança de estado e métrica;
- o plano preserva a separação entre orquestração de atendimento e operação financeira;
- os testes obrigatórios para futuras mudanças executáveis estão documentados.

Quando a próxima fase criar código Go, banco Supabase ou endpoints, ela deve incluir testes unitários, integração e end-to-end aplicáveis.
