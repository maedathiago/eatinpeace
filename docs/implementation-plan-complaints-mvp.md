# Plano de implementação para saneamento das reclamações 2025+

## Propósito

Transformar os achados em `research/restaurant-complaints-2025` e as estratégias em `docs/product-strategy-complaints-2025.md` em uma sequência implementável de MVP.

O objetivo não é resolver todo problema de restaurante. O objetivo é sanear as causas operacionais que aparecem repetidamente nas reclamações e que o Eat in Peace pode tratar sem virar POS, financeiro, fiscal, estoque, delivery ou ERP:

- mesa sem estado visível;
- pedido sem atualização confiável;
- pedido pronto parado;
- reclamação sem dono;
- fila do salão guiada por memória ou insistência;
- atraso descoberto tarde demais;
- conta solicitada sem handoff rastreável;
- gestor sem leitura do gargalo real depois do turno.

## Critério de sucesso do saneamento

Uma reclamação operacional só é considerada saneada quando o produto consegue provar cinco coisas:

1. o evento foi registrado com mesa, horário e origem;
2. o estado atual ficou visível para quem precisa agir;
3. existe uma próxima ação clara ou uma justificativa explícita para aguardar;
4. a fila preserva prioridade, severidade e ordem quando justiça operacional importar;
5. a métrica do problema pode ser medida por turno.

Sem esses cinco pontos, a funcionalidade pode parecer pronta em demonstração, mas ainda não reduz fila invisível em um salão cheio.

## Escopo que entra

O MVP deve começar como uma camada operacional de eventos, tarefas e handoffs.

Entram no escopo inicial:

- sessão de mesa aberta por QR code ou por operador;
- linha do tempo visível da mesa;
- status simples de pedido;
- fila operacional única do salão;
- regras simples de SLA;
- alerta de atraso;
- tarefa para pedido pronto aguardando retirada;
- ticket de reclamação com severidade, responsável e SLA;
- handoff de conta para caixa, garçom ou sistema existente;
- relatório mínimo de gargalos do turno.

## Escopo que fica fora

Ficam fora do MVP:

- processamento de pagamento;
- dados de cartão;
- split de conta;
- imposto, taxa, nota fiscal ou conciliação;
- cálculo financeiro da conta;
- gestão completa de cardápio, estoque, compra ou preço;
- logística de delivery;
- POS próprio;
- chat livre como canal primário de atendimento.

Quando algum desses temas aparecer em reclamação, o Eat in Peace pode registrar, classificar, priorizar e encaminhar. Não deve assumir o workflow principal desse domínio.

## Base técnica inicial

### Stack

- Go para API, domínio, serviços, agentes, integrações e testes.
- Supabase para Postgres e serviços gerenciados definidos nas fases futuras.
- Banco relacional como fonte de verdade dos estados operacionais.
- Eventos operacionais append-only para auditoria, métricas e reconstrução de timeline.

### Módulos Go esperados

A primeira implementação deve separar pelo menos:

- domínio de mesas, sessões, pedidos, tarefas, reclamações e handoff de conta;
- regras de prioridade, SLA, severidade e transição de estado;
- repositórios Postgres/Supabase;
- handlers HTTP;
- workers ou serviços de avaliação de SLA;
- adaptadores externos opcionais, sempre atrás de interfaces.

### Entidades mínimas

As entidades abaixo são ponto de partida para migrations e contratos de domínio:

| Entidade | Finalidade |
| --- | --- |
| `restaurants` | Identificar o restaurante piloto e parametrizações básicas |
| `service_shifts` | Agrupar eventos, métricas e relatórios por turno |
| `tables` | Representar mesas ou comandas operacionais |
| `table_sessions` | Representar a visita ativa do cliente naquela mesa |
| `orders` | Acompanhar pedido em nível operacional, sem assumir financeiro |
| `order_items` | Permitir granularidade quando necessária, especialmente em P1 |
| `operational_events` | Registrar eventos append-only com horário, origem e payload |
| `floor_tasks` | Unificar chamadas, atrasos, pedidos prontos, reclamações e conta |
| `complaints` | Guardar ticket, severidade, responsável, status e resolução |
| `bill_handoffs` | Coordenar solicitação de conta sem processar pagamento |
| `sla_policies` | Parametrizar limites por restaurante, turno ou tipo de tarefa |
| `staff_members` | Identificar responsáveis humanos e permitir override auditável |

O modelo deve aceitar operação manual primeiro. Integrações com POS, cardápio ou cozinha entram depois como fontes ou consumidores de eventos, não como pré-requisito do MVP.

## Contrato de eventos mínimos

Eventos precisam ter pelo menos:

- `id`;
- `restaurant_id`;
- `shift_id`, quando houver turno aberto;
- `table_id` ou `table_session_id`, quando aplicável;
- `event_type`;
- `occurred_at`;
- `source`, por exemplo `customer`, `staff`, `kitchen`, `cashier`, `system` ou `integration`;
- `actor_id`, quando houver usuário identificado;
- `payload` validado;
- `correlation_id` para juntar eventos do mesmo fluxo.

Tipos iniciais:

- `table_session_opened`;
- `order_created`;
- `order_status_changed`;
- `order_stale_detected`;
- `order_delay_risk_detected`;
- `order_delayed`;
- `order_ready`;
- `order_picked_up`;
- `order_delivered`;
- `waiter_called`;
- `floor_task_created`;
- `floor_task_claimed`;
- `floor_task_resolved`;
- `complaint_opened`;
- `complaint_classified`;
- `complaint_assigned`;
- `complaint_first_response_recorded`;
- `complaint_resolved`;
- `bill_requested`;
- `bill_handoff_accepted`;
- `bill_handoff_blocked`;
- `bill_close_confirmed`;
- `table_session_closed`;
- `shift_closed`.

Cada evento que gera trabalho humano deve produzir ou atualizar uma `floor_task`.

## Fases de implementação recomendadas

### Fase 005: fundação operacional Go + Supabase

Objetivo: criar a base executável para registrar eventos, persistir estados e testar regras de domínio.

- Ator: equipe técnica e operador piloto.
- Dor: sem uma fonte única de verdade, cada tela vira uma interpretação diferente do salão.
- Evento: qualquer comando operacional aceito pela API gera evento persistido e estado derivado.
- Mudança de estado: `comando recebido -> evento validado -> estado atualizado -> tarefa gerada quando aplicável`.
- Métrica de sucesso: 100% dos fluxos P0 conseguem registrar evento com mesa, horário, origem e correlação.

Entregáveis:

- módulo Go inicial;
- migrations Supabase/Postgres;
- repositórios para entidades mínimas;
- validações de transição de estado;
- endpoint de health e contrato base de eventos;
- fixtures de restaurante, turno, mesa e equipe;
- documentação de setup local e comandos de teste.

Testes:

- unitários para validação de eventos e transições;
- integração de migrations e repositórios;
- teste de API para criar evento e consultar estado derivado.

### Fase 006: linha do tempo visível da mesa

Objetivo: permitir que cliente e equipe vejam a sessão e o progresso básico do pedido.

- Ator: cliente final e garçom.
- Dor: o cliente não sabe se o pedido foi recebido, esquecido, atrasado, pronto ou entregue.
- Evento: sessão aberta, pedido criado, status alterado e status sem atualização além do SLA.
- Mudança de estado: `mesa sem sessão -> sessão ativa`; `pedido recebido -> em preparo -> pronto -> entregue`; quando parado, `em preparo -> atualização necessária`.
- Métrica de sucesso: tempo médio sem atualização visível e quantidade de pedidos com status vencido por turno.

Entregáveis:

- endpoints para abrir sessão, criar pedido operacional e consultar timeline;
- endpoint ou comando para atualizar status manualmente;
- mensagens curtas para cliente, sem promessa de tempo preciso quando o restaurante não informar;
- registro de status velho como evento e tarefa;
- documentação do fluxo da mesa.

Testes:

- unitários de transição de status;
- integração de persistência da timeline;
- end-to-end de abertura de sessão, pedido, atualização e consulta de status.

### Fase 007: fila operacional única do salão

Objetivo: transformar chamados, pedidos prontos, atrasos, reclamações e conta em uma fila priorizada.

- Ator: garçom, cumim e líder de salão.
- Dor: o atendimento acontece por memória, insistência ou grito mais recente.
- Evento: garçom chamado, pedido pronto, pedido atrasado, reclamação aberta, conta solicitada e mesa sem atualização recente.
- Mudança de estado: `tarefa aberta -> assumida -> em atendimento -> resolvida`; sempre com mesa, horário, responsável e motivo de prioridade.
- Métrica de sucesso: tempo até primeira ação, tarefas dentro do SLA e mesas com mais de uma chamada.

Entregáveis:

- `floor_tasks` com tipo, prioridade, severidade, timestamps e responsável;
- regra de ordenação preservando severidade e ordem de chegada;
- ações de claim, reatribuição com justificativa e resolução;
- API de listagem filtrável por turno, mesa, responsável e status;
- documentação do algoritmo de prioridade.

Testes:

- unitários de ordenação e desempate;
- unitários para override humano com justificativa;
- integração de criação automática de tarefa a partir de eventos;
- end-to-end de chamado de mesa entrando na fila e sendo resolvido.

### Fase 008: atraso de cozinha e pedido pronto parado

Objetivo: avisar o salão antes da reclamação e impedir que pedido pronto perca qualidade por falta de retirada.

- Ator: cozinha, passador, garçom e líder de salão.
- Dor: atraso e pedido pronto parado só aparecem quando o cliente reclama.
- Evento: pedido aceito, início de preparo, tempo padrão por item, item ou pedido pronto, tempo acima do SLA e volume de fila.
- Mudança de estado: `em preparo -> risco de atraso -> atrasado -> salão avisado -> cliente informado`; `pronto -> aguardando retirada -> retirado para entrega -> entregue`.
- Métrica de sucesso: atrasos avisados antes da reclamação, tempo entre estouro de SLA e primeira ação, tempo de pronto até entregue.

Entregáveis:

- política inicial de SLA por etapa;
- worker ou serviço de avaliação periódica de SLA;
- tarefa prioritária para pedido pronto aguardando retirada;
- evento de risco de atraso separado de atraso confirmado;
- mensagens internas para salão e mensagens simples para cliente.

Testes:

- unitários de cálculo de SLA;
- unitários de progressão de risco para atraso;
- integração do avaliador de SLA com `floor_tasks`;
- end-to-end de pedido pronto que vira tarefa e depois entrega.

### Fase 009: reclamação com ticket, responsável e SLA

Objetivo: garantir que reclamações sejam registradas, classificadas, atribuídas e resolvidas de forma justa.

- Ator: cliente, garçom, líder de salão e gerente.
- Dor: reclamação some, fica sem dono ou é atendida por quem insiste mais.
- Evento: reclamação aberta, motivo, mesa, pedido relacionado, severidade, responsável, primeira resposta e resolução.
- Mudança de estado: `aberta -> classificada -> atribuída -> em atendimento -> resolvida` ou `reaberta`.
- Métrica de sucesso: tempo até primeira resposta, reclamações sem responsável por mais de 2 minutos, tempo até resolução e reincidência na mesma mesa.

Entregáveis:

- `complaints` com severidade, motivo estruturado, responsável e status;
- classificação inicial por motivo estruturado, com override humano;
- criação automática de tarefa do salão;
- escalonamento para líder quando sem responsável acima do SLA;
- registro de resolução e reabertura.

Testes:

- unitários de severidade e prioridade;
- unitários para preservar ordem dentro da mesma severidade;
- integração de reclamação criando tarefa;
- end-to-end de reclamação aberta, assumida, respondida e resolvida.

### Fase 010: conta solicitada com handoff rastreável

Objetivo: coordenar pedido de conta sem assumir operação financeira.

- Ator: cliente, garçom, caixa e sistema existente.
- Dor: o fim da refeição vira espera invisível e o cliente não sabe se a conta foi vista.
- Evento: conta solicitada, mesa ou comanda identificada, caixa ou garçom notificado, aceite do handoff e confirmação de fechamento.
- Mudança de estado: `conta solicitada -> em revisão -> enviada ao sistema existente -> aguardando ação do caixa -> fechamento confirmado -> mesa liberada`.
- Métrica de sucesso: tempo até caixa ou garçom assumir, contas acima do SLA e tempo até fechamento confirmado.

Entregáveis:

- `bill_handoffs` sem cálculo financeiro;
- tarefa do salão ou caixa ao solicitar conta;
- bloqueio operacional quando houver reclamação aberta ou item não entregue;
- confirmação manual de fechamento;
- adaptador opcional para enviar evento a sistema existente, sem depender dele.

Testes:

- unitários de transição de handoff;
- unitários de bloqueio por pendência operacional;
- integração de conta solicitada criando tarefa;
- end-to-end de conta solicitada, assumida e confirmada.

### Fase 011: piloto P0 e métricas de turno

Objetivo: validar o pacote P0 em um fluxo único e gerar evidência para decisão de produto.

- Ator: gestor, líder de salão e equipe piloto.
- Dor: sem métricas por turno, o restaurante não sabe se o caos caiu ou só mudou de lugar.
- Evento: turno fechado com eventos de sessão, pedido, tarefa, reclamação e conta.
- Mudança de estado: `turno em andamento -> turno fechado -> gargalos rankeados -> ação recomendada`.
- Métrica de sucesso: redução de pedidos sem atualização, tarefas vencidas, pedidos prontos parados, reclamações sem responsável e contas sem handoff.

Entregáveis:

- relatório mínimo pós-turno;
- painel ou endpoint de métricas P0;
- fixture de turno cheio com eventos variados;
- roteiro de piloto manual;
- documento de decisão para avançar a P1.

Testes:

- unitários de agregação de métricas;
- integração de relatório a partir de eventos reais de teste;
- end-to-end cobrindo sessão, pedido, atraso, pedido pronto, reclamação, conta e fechamento de turno.

## P1 depois do P0 estável

P1 só deve começar depois que as fases P0 tiverem teste end-to-end e evidência mínima de piloto.

Ordem recomendada:

1. pedido errado ou incompleto com correção estruturada;
2. confirmação simples do pedido antes do preparo;
3. mesa em risco;
4. primeiro atendimento e fila de mesa visíveis;
5. conta bloqueada por pendência operacional;
6. relatório pós-turno mais completo.

Cada item deve reaproveitar eventos e tarefas já existentes. Se exigir domínio financeiro, estoque, POS completo ou workflow primário novo, deve voltar para análise de escopo.

## P2 como expansão controlada

P2 deve ser tratado como oportunidade, não como requisito de MVP.

Entram depois:

- item bloqueador e pedido parcialmente pronto;
- rebalanceamento de tarefas por garçom ou setor;
- integration readiness score para parceiros;
- radar de qualidade, porção e valor percebido.

Esses itens só devem avançar se aumentarem visibilidade, priorização, onboarding, integração ou aprendizado usando eventos já existentes.

## Endpoints iniciais sugeridos

Os nomes abaixo são sugestão de contrato, não obrigação final. A fase técnica deve consolidar o padrão antes de implementar.

```text
POST   /v1/table-sessions
GET    /v1/table-sessions/{id}/timeline
POST   /v1/orders
PATCH  /v1/orders/{id}/status
POST   /v1/floor-tasks
GET    /v1/floor-tasks
PATCH  /v1/floor-tasks/{id}
POST   /v1/complaints
PATCH  /v1/complaints/{id}
POST   /v1/bill-handoffs
PATCH  /v1/bill-handoffs/{id}
POST   /v1/service-shifts/{id}/close
GET    /v1/service-shifts/{id}/metrics
```

## Regras de produto que precisam virar testes

- Pedido não pode saltar para `entregue` sem passar por um estado operacional compatível, exceto com override registrado.
- Tarefa de mesma severidade preserva ordem de criação.
- Reclamação crítica pode passar na frente, mas precisa registrar motivo.
- Reclamação sem responsável acima do SLA escala para líder.
- Pedido pronto acima do SLA vira tarefa prioritária.
- Conta solicitada com reclamação aberta não fecha automaticamente.
- Cliente sempre recebe confirmação de registro da ação.
- Mensagens ao cliente não prometem tempo preciso sem dado informado pelo restaurante.
- Nenhum fluxo calcula imposto, processa pagamento, guarda cartão ou emite fiscal.

## Estratégia de teste acumulada

Toda fase executável deve atualizar `docs/engineering/testing-strategy.md` quando introduzir novo comando, camada ou fluxo.

Comandos-alvo:

```bash
go test ./...
```

Quando houver E2E:

```bash
make test-e2e
```

Cobertura mínima por tipo de problema:

| Problema saneado | Teste obrigatório |
| --- | --- |
| Status invisível | E2E de sessão, pedido e timeline |
| Fila do salão injusta | Unitário de prioridade e E2E de tarefa |
| Atraso descoberto tarde | Unitário de SLA e integração do avaliador |
| Pedido pronto parado | E2E de pronto, tarefa, retirada e entrega |
| Reclamação sem dono | E2E de ticket, responsável e resolução |
| Conta esquecida | E2E de handoff e confirmação |
| Gargalo sem leitura | Integração de relatório pós-turno |

## Critério de pronto para cada fase futura

Cada fase de implementação deve cumprir:

- branch própria criada a partir de `main` atualizada;
- documento de fase em `docs/phases/`;
- documentação do fluxo, API, schema ou regra alterada;
- migrations versionadas quando houver banco;
- testes unitários para regra de domínio;
- testes de integração para persistência ou fronteira externa;
- end-to-end quando o fluxo tocar cliente, salão, cozinha, reclamação ou conta;
- evidência dos comandos executados registrada no documento da fase ou na descrição do PR;
- confirmação explícita de que não houve entrada indevida em POS, financeiro, fiscal, estoque ou delivery.

## Decisão operacional

A próxima fase recomendada é `phase/005-operational-foundation`.

Ela deve criar a base Go + Supabase, contratos de eventos e testes mínimos antes de qualquer tela. Sem essa base, as próximas experiências correm o risco de virar telas bonitas sobre estados frágeis, que é exatamente o problema que o Eat in Peace está tentando sanear.
