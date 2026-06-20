# Estratégias de produto a partir de reclamações 2025+

## Base usada

Esta lista consolida a análise dos agentes especializados do Eat in Peace sobre o corpus em `research/restaurant-complaints-2025`.

- Fonte principal: RestaurantGuru.
- Recorte: 1.000 avaliações com reclamação, de 2025 em diante.
- Restaurantes consultados: 135.
- Principais temas encontrados: nota baixa ou insatisfação geral, atendimento, qualidade da comida, demora, fila/reserva, delivery, porção, preço, cobrança, pedido errado e limpeza.

## Leitura estratégica

O Eat in Peace não deve tentar resolver diretamente "comida ruim", preço, estoque, delivery ou financeiro. A oportunidade de produto está no que aparece por trás de boa parte das reclamações: o cliente não sabe o que está acontecendo, o salão não sabe o que priorizar, a cozinha não comunica risco a tempo, a conta vira uma espera invisível e a reclamação não tem dono.

Diretriz: o MVP não precisa resolver o restaurante inteiro; precisa impedir que uma mesa, um pedido pronto, um atraso, uma reclamação ou uma conta fiquem invisíveis.

## P0: pacote inicial do MVP

### 1. Linha do tempo visível da mesa

- Agentes responsáveis: Customer Concierge Agent e Customer Trust Agent.
- Ator: cliente final.
- Dor: o cliente não sabe se o pedido foi recebido, esquecido, atrasado, pronto ou entregue.
- Evento capturado: sessão da mesa aberta, pedido criado, status alterado, status sem atualização além do SLA.
- Mudança de estado: `pedido recebido -> em preparo -> pronto -> entregue`; quando parado, `em preparo -> atualização necessária`.
- Métrica de sucesso: redução de chamados de status, tempo médio sem atualização visível, reclamações de atraso por 100 mesas.
- Por que é MVP: usa eventos mínimos e pode funcionar com atualização manual pelo salão ou cozinha.

### 2. Fila operacional única do salão

- Agente responsável: Floor Coordinator Agent.
- Ator: garçom, cumim e líder de salão.
- Dor: o salão atende por memória, insistência do cliente ou urgência verbal, não por prioridade operacional.
- Evento capturado: garçom chamado, pedido pronto, pedido atrasado, reclamação aberta, conta solicitada, mesa sem atualização recente.
- Mudança de estado: `tarefa aberta -> assumida -> em atendimento -> resolvida`, sempre com mesa, horário e responsável.
- Métrica de sucesso: tempo até primeira ação do salão, tarefas dentro do SLA, mesas com mais de uma chamada.
- Por que é MVP: um painel simples com timestamps já reduz esquecimento e polling manual.

### 3. Alerta de atraso antes da reclamação

- Agentes responsáveis: Kitchen Expeditor Agent e Floor Coordinator Agent.
- Ator: cozinha, passador, garçom e líder de salão.
- Dor: o cliente só descobre atraso quando pergunta; o salão só descobre quando a reclamação já existe.
- Evento capturado: pedido recebido, aceito pela cozinha, início de preparo, tempo padrão por item, item sem atualização, volume da fila.
- Mudança de estado: `em preparo -> risco de atraso -> atrasado -> salão avisado -> cliente informado`.
- Métrica de sucesso: atrasos avisados antes da reclamação, tempo entre estouro de SLA e primeira ação, reclamações de demora por turno.
- Por que é MVP: começa com SLAs simples por etapa e não depende de previsão sofisticada.

### 4. Pedido pronto não pode ficar parado

- Agentes responsáveis: Kitchen Expeditor Agent e Floor Coordinator Agent.
- Ator: passador, garçom e cumim.
- Dor: comida pronta perde qualidade, o cliente percebe demora e o garçom não sabe que precisa retirar.
- Evento capturado: item pronto, pedido pronto, tempo desde pronto, setor ou responsável pela mesa.
- Mudança de estado: `pronto -> aguardando retirada -> retirado para entrega -> entregue`; após SLA, vira tarefa prioritária.
- Métrica de sucesso: tempo de pronto até entregue, pedidos prontos acima do SLA, reclamações de comida fria ou demora entre etapas.
- Por que é MVP: exige apenas um botão de pronto e uma fila de retirada.

### 5. Reclamação com ticket, responsável e SLA

- Agente responsável: Complaint Priority Agent.
- Ator: cliente, garçom, líder de salão e gerente.
- Dor: a reclamação some, fica sem dono ou é resolvida por quem insiste mais.
- Evento capturado: reclamação aberta, motivo, mesa, pedido relacionado, severidade, responsável, primeira resposta e resolução.
- Mudança de estado: `aberta -> classificada -> atribuída -> em atendimento -> resolvida` ou `reaberta`.
- Métrica de sucesso: tempo até primeira resposta, reclamações sem responsável por mais de 2 minutos, tempo até resolução, reincidência na mesma mesa.
- Por que é MVP: cria uma fila justa e auditável sem entrar em desconto automático, pagamento ou CRM.

### 6. Conta solicitada com handoff rastreável

- Agente responsável: Bill Handoff Agent.
- Ator: cliente, garçom, caixa e POS existente.
- Dor: o fim da refeição vira espera invisível; o cliente quer sair e não sabe se a conta foi vista.
- Evento capturado: conta solicitada, mesa ou comanda identificada, caixa ou garçom notificado, aceite do handoff, confirmação de fechamento.
- Mudança de estado: `conta solicitada -> em revisão -> enviada ao sistema existente -> aguardando ação do caixa -> fechamento confirmado -> mesa liberada`.
- Métrica de sucesso: tempo até caixa ou garçom assumir, contas acima do SLA, tempo até fechamento confirmado.
- Por que é MVP: coordena a ação sem processar pagamento, emitir nota ou reconciliar caixa.

## P1: segunda camada de produto

### 7. Pedido errado ou incompleto com correção estruturada

- Agente responsável: Complaint Priority Agent.
- Ator: cliente e salão.
- Dor: item errado ou faltando vira conversa repetida e conflito no salão.
- Evento capturado: motivo estruturado como item errado, faltou item, veio incompleto ou não era o pedido, ligado ao pedido original.
- Mudança de estado: `entregue -> problema reportado -> correção atribuída -> corrigido` ou `resolvido com justificativa`.
- Métrica de sucesso: tempo até reconhecimento do erro, tempo até correção, tickets de pedido errado por 100 pedidos.
- Por que é MVP: estrutura uma reclamação de alta severidade sem redesenhar cozinha, cardápio ou financeiro.

### 8. Confirmação simples do pedido antes do preparo

- Agente responsável: Customer Concierge Agent.
- Ator: cliente e garçom.
- Dor: parte dos erros nasce porque o cliente não vê o que foi registrado antes de o pedido seguir.
- Evento capturado: pedido criado pelo cliente ou garçom, observação, alteração, cancelamento antes do aceite.
- Mudança de estado: `registrado -> confirmado pelo cliente -> aceito pela cozinha`.
- Métrica de sucesso: correções antes do preparo, taxa de confirmação, queda em pedidos errados.
- Por que é MVP: funciona como espelho do pedido e não precisa substituir o cardápio digital ou POS.

### 9. Mesa em risco

- Agentes responsáveis: Manager Insights Agent, Complaint Priority Agent e Floor Coordinator Agent.
- Ator: líder de salão e gerente de turno.
- Dor: muitas avaliações ruins surgem de acúmulo: espera, múltiplos chamados, pedido pronto parado, conta esquecida e reclamação sem resposta.
- Evento capturado: atraso, chamados repetidos, reclamação aberta, pedido pronto acima do SLA, conta solicitada sem ação.
- Mudança de estado: `mesa normal -> atenção necessária -> mesa em risco -> escalada para líder`.
- Métrica de sucesso: mesas resgatadas antes de nova reclamação, tempo até intervenção, mesas com múltiplos chamados.
- Por que é MVP: é uma regra simples sobre eventos já capturados.

### 10. Primeiro atendimento e fila de mesa visíveis

- Agente responsável: Floor Coordinator Agent.
- Ator: recepção, garçom, líder de salão e cliente aguardando.
- Dor: espera, reserva confusa e mesa sentada sem primeiro contato geram irritação antes mesmo do pedido.
- Evento capturado: entrada na fila, mesa liberada, mesa ocupada, sessão QR iniciada, tempo sem primeiro atendimento.
- Mudança de estado: `aguardando mesa -> mesa chamada -> sentado -> aguardando primeiro atendimento -> atendido`.
- Métrica de sucesso: espera informada versus real, abandono antes de sentar, tempo sentado até primeiro contato.
- Por que é MVP: pode começar manualmente e não exige construir sistema completo de reservas.

### 11. Conta bloqueada por pendência operacional

- Agente responsável: Bill Handoff Agent, com sinal do Complaint Priority Agent.
- Ator: caixa, garçom, gerente e cliente.
- Dor: cobrança vira conflito quando existe item contestado, reclamação aberta ou correção pendente.
- Evento capturado: conta solicitada com reclamação aberta, item não entregue, dúvida de cobrança, pedido de revisão.
- Mudança de estado: `em revisão -> pendência detectada -> correção solicitada -> liberada para fechamento`.
- Métrica de sucesso: contas fechadas com pendência aberta, tempo de correção, reclamações de cobrança por turno.
- Por que é MVP: protege o handoff sem calcular taxa, imposto, split, pagamento ou fiscal.

### 12. Relatório pós-turno de gargalos e valor entregue

- Agentes responsáveis: Manager Insights Agent e Restaurant Success Agent.
- Ator: gestor, dono e operador.
- Dor: o turno foi caótico, mas o gestor não sabe se o problema principal foi salão, cozinha, conta ou recepção.
- Evento capturado: tempos por etapa, chamados, atrasos, reclamações, contas solicitadas, tarefas vencidas.
- Mudança de estado: `turno em andamento -> turno fechado -> gargalos rankeados -> ação recomendada`.
- Métrica de sucesso: gargalos recorrentes reduzidos, reclamações por turno, tempo médio por etapa.
- Por que é MVP: é leitura dos eventos já capturados e ajuda a vender prova de valor.

## P2: oportunidades depois do fluxo mínimo

### 13. Item bloqueador e pedido parcialmente pronto

- Agente responsável: Kitchen Expeditor Agent.
- Ator: cozinha, passador e garçom.
- Dor: parte do pedido chega, outra atrasa, acompanhamentos esfriam e o cliente sente desorganização.
- Evento capturado: status por grupo simples como bebida, entrada, principal e sobremesa; item atrasado bloqueando pedido.
- Mudança de estado: `em preparo -> parcialmente pronto -> bloqueado por item -> pronto completo` ou `entrega parcial autorizada`.
- Métrica de sucesso: pedidos com item bloqueador acima do SLA, tempo para resolver bloqueio, reclamações de pedido incompleto.
- Por que é MVP depois: exige mais granularidade de cozinha, então deve vir após os estados básicos estarem estáveis.

### 14. Rebalanceamento de tarefas por garçom ou setor

- Agente responsável: Floor Coordinator Agent.
- Ator: líder de salão e garçons.
- Dor: equipe sobrecarregada em um setor gera chamados vencidos e sensação de abandono.
- Evento capturado: tarefas abertas por garçom, tarefas vencidas, mesas em risco por setor.
- Mudança de estado: `atribuída -> sugerida para reatribuição -> reatribuída com justificativa`.
- Métrica de sucesso: tarefas vencidas por garçom, tempo médio de resposta por setor, concentração de atrasos.
- Por que é MVP depois: deve começar como sugestão, mantendo o humano no controle.

### 15. Integration readiness score para parceiros

- Agentes responsáveis: Partner App Experience Agent e Scope-Neutral Value Agent.
- Ator: dono de aplicação parceira, PM ou time técnico.
- Dor: parceiro quer vender valor novo sem reescrever POS, cardápio ou comanda.
- Evento capturado: disponibilidade dos eventos mínimos: mesa iniciada, status de pedido, chamado, reclamação, conta solicitada e handoff.
- Mudança de estado: `parceiro interessado -> eventos mapeados -> prontidão calculada -> piloto manual ou API -> integração estável`.
- Métrica de sucesso: cobertura de eventos mínimos, tempo até piloto, eventos ausentes ou duplicados, restaurantes ativados pelo parceiro.
- Por que é MVP depois: ajuda distribuição e embedding, mas depende de clareza do produto operacional primeiro.

### 16. Radar de qualidade, porção e valor percebido

- Agentes responsáveis: Feedback-to-Insight Agent e Manager Insights Agent.
- Ator: gestor e líder de operação.
- Dor: comida, porção e preço aparecem no corpus, mas não devem virar domínio operacional do Eat in Peace.
- Evento capturado: reclamação categorizada por item, etapa, horário, mesa e resolução.
- Mudança de estado: `feedback isolado -> padrão recorrente -> causa provável -> revisão operacional`.
- Métrica de sucesso: recorrência por item ou horário, reclamações de qualidade por 100 mesas, reincidência após ação do gestor.
- Por que é MVP depois: agrega insight sem controlar cardápio, preço ou estoque.

## Não priorizar agora

- Delivery como operação própria. Pode ser registrado como reclamação ou status vindo de parceiro, mas não deve puxar logística.
- Pagamento, split, nota fiscal, imposto e conciliação. A conta é handoff operacional, não financeiro.
- Gestão de cardápio, estoque, preço ou compras. Reclamações podem virar insight, não workflow primário.
- Chat livre com o cliente. O atendimento deve ser estruturado por eventos e estados operacionais.
- Automação rígida de equipe. O sistema pode sugerir prioridade e reatribuição, mas o restaurante mantém controle.

## Ordem recomendada

1. Implementar o P0 como primeira experiência ponta a ponta: timeline da mesa, fila do salão, atraso, pedido pronto, reclamação e conta.
2. Validar em piloto se diminuem chamados repetidos, reclamações sem responsável, pedido pronto parado e espera invisível de conta.
3. Só então adicionar P1, priorizando pedido errado, mesa em risco e relatório pós-turno.
4. Usar P2 para criar distribuição e aprendizado, sem aumentar escopo para POS, financeiro, fiscal, estoque ou logística.
