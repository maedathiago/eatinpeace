# Fase 003: fluxo direto na main sincronizada

## Objetivo

Atualizar as regras de trabalho para usar `main` como fluxo padrão, mantendo o checkout local sincronizado com `origin/main`.

## Escopo

- Remover a exigência de branch própria por fase.
- Manter documentos de fase como registro de planejamento e entrega.
- Registrar que branches separadas só são necessárias quando o usuário pedir explicitamente ou quando houver risco técnico forte.
- Exigir sincronização com o remoto antes de trabalho relevante e push para `origin/main` ao concluir entregas validadas.

## Entregáveis

- `AGENTS.md` atualizado.
- `docs/engineering/repository-rules.md` atualizado.
- `docs/phases/README.md` atualizado.
- Este documento de fase criado.

## Testes

Esta fase não cria código executável. A verificação aplicável é documental:

- Regra de fluxo direto em `main` versionada.
- Critério de pronto atualizado para exigir sincronização com `origin/main`.
- Nenhuma mudança de comportamento executável foi introduzida.
