# Fase 000: regras do repositório

## Objetivo

Estabelecer as regras básicas de trabalho do repositório antes da implementação do produto.

## Escopo

- Definir branch por fase.
- Registrar stack oficial: Go com Supabase.
- Exigir documentação para tudo que for criado.
- Exigir que alterações reflitam na documentação.
- Definir testes obrigatórios, incluindo end-to-end.
- Registrar que Eat in Peace é camada embedded de orquestração, sem escopo financeiro.

## Branch

```text
phase/000-repo-rules
```

## Entregáveis

- `AGENTS.md` atualizado com regras operacionais do repo.
- `docs/engineering/repository-rules.md` criado.
- `docs/engineering/testing-strategy.md` criado.
- `README.md` atualizado para apontar para as regras.
- Este documento de fase criado.

## Testes

Esta fase não cria código executável. A verificação aplicável é documental:

- Regras do repo versionadas.
- Estratégia de testes versionada.
- Branch da fase criada.

A partir da primeira fase com código Go, testes automatizados passam a ser critério obrigatório de pronto, incluindo cobertura end-to-end dos fluxos críticos afetados.
