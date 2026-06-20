# Fases do projeto

Cada fase relevante deve ter uma branch própria e um documento nesta pasta. A `main` local deve estar sincronizada com `origin/main` antes de criar a branch, e a branch deve ser mantida alinhada com a `main` enquanto o trabalho avança.

| Fase | Entrega | Objetivo |
| --- | --- | --- |
| [000: regras do repositório](000-repo-rules.md) | Histórica: `phase/000-repo-rules` | Estabelecer as primeiras regras do repositório, stack Go + Supabase, documentação obrigatória e testes obrigatórios |
| [001: agentes de experiência e valor](001-experience-value-agents.md) | Histórica: `phase/001-experience-value-agents` | Definir agentes para restaurantes, donos de aplicações parceiras e clientes finais, com valor adicional sem aumento de escopo |
| [002: squash antes de merge](002-squash-before-merge.md) | Histórica: `phase/002-squash-before-merge` | Exigir commit único de fase antes do merge quando houver vários commits de trabalho |
| [003: fluxo com branches sincronizadas](003-branch-synced-workflow.md) | `phase/003-branch-sync-workflow` | Corrigir a regra operacional: toda feature ou etapa usa branch própria, mantendo branch, `main` local e remoto sincronizados |
| [004: plano de implementação das reclamações](004-complaints-implementation-plan.md) | `phase/004-complaints-implementation-plan` | Converter achados e estratégia de reclamações 2025+ em plano de implementação, fases futuras, eventos mínimos e testes obrigatórios |
