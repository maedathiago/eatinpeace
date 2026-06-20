# Fases do projeto

Cada fase relevante deve ter uma branch própria e um documento nesta pasta. A `main` local deve estar sincronizada com `origin/main` antes de criar a branch, e a branch deve ser mantida alinhada com a `main` enquanto o trabalho avança.

| Fase | Entrega | Objetivo |
| --- | --- | --- |
| [000: regras do repositório](000-repo-rules.md) | Histórica: `phase/000-repo-rules` | Estabelecer as primeiras regras do repositório, stack Go + Supabase, documentação obrigatória e testes obrigatórios |
| [001: agentes de experiência e valor](001-experience-value-agents.md) | Histórica: `phase/001-experience-value-agents` | Definir agentes para restaurantes, donos de aplicações parceiras e clientes finais, com valor adicional sem aumento de escopo |
| [002: squash antes de merge](002-squash-before-merge.md) | Histórica: `phase/002-squash-before-merge` | Exigir commit único de fase antes do merge quando houver vários commits de trabalho |
| [003: fluxo com branches sincronizadas](003-branch-synced-workflow.md) | `phase/003-branch-sync-workflow` | Corrigir a regra operacional: toda feature ou etapa usa branch própria, mantendo branch, `main` local e remoto sincronizados |
| [004: plano de implementação das reclamações](004-complaints-implementation-plan.md) | `phase/004-complaints-implementation-plan` | Converter achados e estratégia de reclamações 2025+ em plano de implementação, fases futuras, eventos mínimos e testes obrigatórios |
| [005: fundação operacional Go + Supabase](005-operational-foundation.md) | `phase/005-operational-foundation` | Criar a base executável de eventos, domínio, API, migrations, fixtures e testes P0 |
| [006: linha do tempo visível da mesa](006-table-timeline.md) | `phase/006-table-timeline` | Abrir sessão, criar pedido, atualizar status e consultar timeline da mesa |
| [007: fila operacional única do salão](007-floor-queue.md) | `phase/007-floor-queue` | Priorizar tarefas por severidade, SLA e ordem de criação com ações auditáveis |
| [008: atraso de cozinha e pedido pronto parado](008-kitchen-sla-ready-orders.md) | `phase/008-kitchen-sla-ready-orders` | Avaliar SLA, detectar risco/atraso e gerar tarefa para pedido pronto |
| [009: reclamação com ticket, responsável e SLA](009-complaints-ticket-sla.md) | `phase/009-complaints-ticket-sla` | Registrar, classificar, atribuir, responder, resolver e reabrir reclamações |
| [010: conta solicitada com handoff rastreável](010-bill-handoff.md) | `phase/010-bill-handoff` | Coordenar conta como handoff operacional sem entrar em financeiro |
| [011: piloto P0 e métricas de turno](011-p0-pilot-metrics.md) | `phase/011-p0-pilot-metrics` | Fechar turno e agregar métricas P0 a partir dos eventos e estados |
