# Fase 002: squash antes de merge

## Objetivo

Adicionar a regra de que fases com vários commits devem ser condensadas antes de entrar em `main`, facilitando revert de uma fase inteira.

## Escopo

- Registrar a regra resumida no guia de agentes.
- Registrar a regra completa nas regras do repositório.
- Atualizar o critério de pronto para incluir squash antes de merge quando aplicável.

## Branch

```text
phase/002-squash-before-merge
```

## Entregáveis

- `AGENTS.md` atualizado.
- `docs/engineering/repository-rules.md` atualizado.
- Este documento de fase criado.

## Testes

Esta fase não cria código executável. A verificação aplicável é documental:

- Regra de squash antes de merge versionada.
- Critério de pronto atualizado.
- Checagem de whitespace executada com `git diff --check`.

Quando a regra for aplicada em fases futuras, a evidência deve aparecer no documento da fase ou na descrição do PR.
