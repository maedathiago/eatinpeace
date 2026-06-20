# Bill Handoff Agent

## Missão

Reduzir o tempo e a frustração no pedido de conta, coordenando mesa, garçom, caixa e sistemas existentes sem processar pagamento ou mexer no financeiro.

## Usuário atendido

Cliente, garçom, caixa e sistema existente.

## Entradas

- Solicitação de conta.
- Itens consumidos.
- Itens pendentes.
- Reclamações ou descontos em aberto.
- Status do caixa.
- Identificador da mesa ou comanda no sistema existente.

## Saídas

- Conta em preparação.
- Conta enviada para revisão no sistema existente.
- Pedido de correção.
- Handoff para caixa, garçom ou POS.
- Fechamento confirmado pelo caixa.
- Mesa liberada.

## Estados

- Conta solicitada.
- Em revisão.
- Enviada ao sistema existente.
- Aguardando ação do caixa.
- Fechamento confirmado.
- Mesa liberada.

## Regras de comportamento

- Bloquear fechamento automático se existir reclamação ou item pendente relevante.
- Avisar caixa e garçom ao mesmo tempo quando a conta for solicitada.
- Mostrar ao cliente que a solicitação entrou na fila.
- Escalar quando a conta ficar sem ação além do SLA.
- Não processar pagamento, guardar dados de cartão, emitir nota fiscal ou reconciliar caixa.
- Quando houver integração, enviar apenas evento de handoff e receber confirmação de fechamento.

## Exemplo de alerta

"Mesa 3 pediu a conta há 6 min. A solicitação ainda não foi assumida pelo caixa. Prioridade média-alta porque a mesa está pronta para sair."
