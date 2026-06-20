# Fases do projeto

Cada fase relevante deve ter um documento nesta pasta. O fluxo atual usa `main` sincronizada com `origin/main`; branches separadas só são necessárias quando o usuário pedir explicitamente ou quando houver risco técnico forte.

| Fase | Entrega | Objetivo |
| --- | --- | --- |
| [000: regras do repositório](000-repo-rules.md) | Histórica: `phase/000-repo-rules` | Estabelecer as primeiras regras do repositório, stack Go + Supabase, documentação obrigatória e testes obrigatórios |
| [001: agentes de experiência e valor](001-experience-value-agents.md) | Histórica: `phase/001-experience-value-agents` | Definir agentes para restaurantes, donos de aplicações parceiras e clientes finais, com valor adicional sem aumento de escopo |
| [002: squash antes de merge](002-squash-before-merge.md) | Histórica: `phase/002-squash-before-merge` | Exigir commit único de fase antes do merge quando houver vários commits de trabalho |
| [003: fluxo direto na main sincronizada](003-main-synced-workflow.md) | `main` | Trocar o padrão de branch por fase por trabalho direto em `main`, com sincronização obrigatória com o remoto |
