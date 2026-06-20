# Eat in Peace: tese do produto

## Problema

Restaurantes com operação de salão funcionam como uma fila invisível. O cliente não sabe se o pedido foi recebido, se está parado, se a cozinha atrasou, se a reclamação foi vista ou se a conta foi esquecida. O garçom também não tem uma fonte única de verdade e acaba virando mensageiro entre mesa, cozinha e caixa.

## Tese

Eat in Peace cria uma camada embedded de orquestração de atendimento para restaurantes. A experiência melhora quando cliente, garçom, cozinha, caixa e gestão compartilham o mesmo estado operacional em tempo real, mesmo que o restaurante continue usando os sistemas que já existem.

O diferencial não é "pedir pelo QR code". O diferencial é transformar pedidos, reclamações, chamadas e contas em filas visíveis, priorizadas e acompanháveis.

## Promessa

- Para o cliente: acompanhar o pedido, pedir ajuda, reclamar e solicitar a conta sem depender de sorte ou insistência.
- Para o garçom: ver o que precisa de ação agora, sem circular pelo salão fazendo polling manual.
- Para a cozinha: receber pedidos e alertas com contexto, ordem e impacto no salão.
- Para o caixa: saber quais mesas querem a conta e quais dependem de revisão, sem substituir o sistema financeiro.
- Para o gestor: enxergar gargalos reais da operação.

## MVP recomendado

1. Mesa escaneia QR code e abre uma sessão de atendimento.
2. Cliente faz pedido ou registra pedido feito pelo garçom, dependendo da operação do restaurante.
3. Cozinha ou garçom atualiza status: recebido, preparando, pronto, entregue.
4. Cliente vê status do pedido e tempo estimado simples.
5. Cliente pode chamar garçom, reclamar ou pedir conta.
6. Garçom vê uma fila operacional de mesas e tarefas.
7. Reclamações viram tickets com ordem, tempo, status e responsável.
8. Pedido de conta vira handoff para caixa, garçom ou POS existente.
9. Gestor vê atrasos, reclamações e tempos médios.

## Fora de escopo

- Processar pagamento.
- Guardar dados de cartão.
- Emitir nota fiscal.
- Fazer conciliação de caixa.
- Controlar impostos, taxas, split de pagamento ou repasse financeiro.
- Substituir POS, ERP, maquininha, sistema fiscal ou sistema de estoque.

Eat in Peace pode se integrar a esses sistemas, mas o produto deve vender a camada de experiência e operação, não a camada financeira.

## Hipóteses a validar

- Clientes reclamam menos quando o status é visível, mesmo que o tempo total não mude.
- Garçons perdem menos tempo indo até a cozinha perguntar status.
- Reclamações com fila e responsável reduzem conflito no salão.
- Pedido de conta pelo celular reduz tempo de mesa parada no fim da refeição, mesmo quando o pagamento continua fora do Eat in Peace.
- Restaurantes pagam por redução de caos operacional, não apenas por cardápio digital.
- Parceiros que já vendem software para restaurantes podem embutir Eat in Peace como módulo de atendimento, reduzindo custo de aquisição.

## Risco principal

Se o produto tentar construir POS, financeiro, pagamento, fiscal, estoque e cardápio completo, o MVP fica pesado demais e entra em competição direta com sistemas estabelecidos. A primeira versão deve funcionar como camada de orquestração e handoff, provando valor antes de integrar sistemas.
