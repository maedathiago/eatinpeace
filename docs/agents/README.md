# Agentes do produto

Estes agentes descrevem o primeiro modelo operacional do Eat in Peace. Eles são agentes de produto antes de serem agentes técnicos: cada um é dono de uma parte visível da experiência do restaurante e pode virar workflow, assistente de IA, motor de regras, dashboard ou automação com humano no loop.

## Mapa dos agentes

### Agentes operacionais

| Agente | Usuário principal | Experiência melhorada |
| --- | --- | --- |
| [Customer Concierge Agent](customer-concierge-agent.md) | Cliente | Pedido visível, chamado registrado, conta sem caça ao garçom |
| [Floor Coordinator Agent](floor-coordinator-agent.md) | Garçom e salão | Fila operacional clara e priorização de tarefas |
| [Kitchen Expeditor Agent](kitchen-expeditor-agent.md) | Cozinha e passador | Status confiável e atraso antecipado |
| [Complaint Priority Agent](complaint-priority-agent.md) | Cliente e gerente | Reclamação com ordem, dono e SLA |
| [Bill Handoff Agent](bill-handoff-agent.md) | Cliente, garçom e caixa | Solicitação de conta mais rápida sem mexer no financeiro |
| [Manager Insights Agent](manager-insights-agent.md) | Gestor | Gargalos e aprendizado de turno |

### Agentes de experiência e valor

| Agente | Usuário principal | Experiência melhorada |
| --- | --- | --- |
| [Restaurant Success Agent](experience-value-agents.md#restaurant-success-agent) | Restaurante e gestor | Adoção, configuração e prova de valor |
| [Partner App Experience Agent](experience-value-agents.md#partner-app-experience-agent) | Dono de aplicação parceira | Embedding simples e valor para o produto parceiro |
| [Customer Trust Agent](experience-value-agents.md#customer-trust-agent) | Cliente final | Clareza, calma e sensação de justiça |
| [Scope-Neutral Value Agent](experience-value-agents.md#scope-neutral-value-agent) | Produto e parceiros | Novos valores sem aumentar escopo |
| [Feedback-to-Insight Agent](experience-value-agents.md#feedback-to-insight-agent) | Restaurante, parceiro e produto | Aprendizado acionável a partir de eventos já capturados |

## Eventos compartilhados

- Sessão da mesa iniciada.
- Pedido criado.
- Pedido aceito pela cozinha.
- Status do pedido alterado.
- Garçom chamado.
- Reclamação aberta.
- Reclamação atribuída.
- Reclamação resolvida.
- Conta solicitada.
- Conta enviada para o sistema existente.
- Fechamento confirmado pelo caixa.
- Mesa liberada.

## Primeira experiência ponta a ponta

1. Cliente escaneia QR da mesa.
2. Customer Concierge Agent abre a sessão e mostra ações disponíveis.
3. Pedido entra na fila e Kitchen Expeditor Agent acompanha status.
4. Floor Coordinator Agent mostra ao garçom o que precisa de ação.
5. Se houver atraso, o sistema alerta antes da reclamação.
6. Se o cliente reclamar, Complaint Priority Agent cria ticket com ordem e responsável.
7. No fim da refeição, Bill Handoff Agent coordena a solicitação de conta e entrega para o sistema existente.
8. Manager Insights Agent resume gargalos do turno.

## Limite do MVP

A primeira versão não deve depender de integração completa com POS e não deve mexer em financeiro. Atualizações manuais da cozinha, do salão ou do caixa são aceitáveis se provarem o valor central: estado compartilhado e menos filas invisíveis.
