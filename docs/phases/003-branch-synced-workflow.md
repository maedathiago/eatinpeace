# Fase 003: fluxo com branches sincronizadas

## Objetivo

Corrigir a regra de trabalho: toda feature, correção ou etapa deve acontecer em branch própria, mantendo a `main` local sincronizada com `origin/main` e a branch alinhada com a `main` mais recente.

## Escopo

- Reafirmar que ninguém trabalha direto em `main` para features ou etapas.
- Manter documentos de fase como registro de planejamento e entrega.
- Exigir que branches sejam criadas a partir de `main` atualizada.
- Exigir que branches sejam mantidas em sync com `main` quando ela avançar.
- Exigir sincronização entre local e remoto: buscar antes de começar, publicar branch quando houver entrega e empurrar `main` após integração validada.

## Entregáveis

- `AGENTS.md` atualizado.
- `docs/engineering/repository-rules.md` atualizado.
- `docs/phases/README.md` atualizado.
- Este documento de fase criado.

## Testes

Esta fase não cria código executável. A verificação aplicável é documental:

- Regra de fluxo por branch versionada.
- Critério de pronto atualizado para exigir branch própria e sincronização com `origin/main`.
- Nenhuma mudança de comportamento executável foi introduzida.
