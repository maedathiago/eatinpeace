# Estratégia embedded e comercial

## Direção

Eat in Peace deve ser uma camada embedded de orquestração de atendimento, não um sistema financeiro e não um novo POS.

A estratégia é entrar onde o restaurante já opera: cardápio digital, POS, KDS, sistema de salão, comanda eletrônica, QR code existente, tablet do garçom ou painel de gestão. O produto deve tirar build do caminho: menos sistema próprio para substituir tudo, mais módulo plugável para melhorar a experiência de atendimento.

## O que não vamos construir

- Gateway de pagamento.
- Carteira digital.
- Split de conta.
- Conciliação financeira.
- Emissão fiscal.
- Gestão tributária.
- Maquininha.
- POS completo.
- ERP de restaurante.
- Estoque completo.

Esses sistemas já existem, são difíceis de substituir e carregam responsabilidade operacional, fiscal e regulatória. Eat in Peace deve se acoplar a eles.

## O que vamos vender

O restaurante paga para reduzir caos operacional:

- Menos cliente irritado por falta de status.
- Menos garçom fazendo polling na cozinha.
- Menos pedido pronto parado.
- Menos reclamação sem dono.
- Menos mesa esperando conta sem visibilidade.
- Mais previsibilidade no salão.
- Mais dados sobre gargalos reais do turno.

## Formas de embedding

1. Widget de mesa
   - Um link ou QR code que abre a experiência do cliente.
   - Pode coexistir com o QR code/cardápio que o restaurante já usa.

2. Painel operacional
   - Tela para salão, cozinha ou caixa acompanharem filas.
   - Pode começar manual e depois integrar por API.

3. Webhooks e API
   - Recebe eventos de pedido, status, mesa e conta.
   - Envia eventos de chamado, reclamação, atraso e solicitação de conta.

4. White-label ou módulo parceiro
   - Parceiros de software para restaurante podem vender Eat in Peace como módulo de atendimento.
   - O parceiro mantém POS, financeiro e relacionamento principal; Eat in Peace entrega a inteligência operacional.

## Modelo comercial

Opções iniciais a testar:

- Mensalidade por unidade.
- Mensalidade por número de mesas.
- Mensalidade por volume de atendimentos.
- Add-on vendido por parceiros com revenue share.
- Plano piloto pago com setup leve e mensalidade reduzida.

O preço deve se ancorar em perda operacional evitada, não em número de features. A conversa comercial é: quantas mesas ficam paradas, quantas reclamações viram desconto, quanto tempo o garçom perde perguntando status e quantas avaliações ruins nascem de falta de visibilidade.

## Wedge de entrada

Começar por uma dor estreita e pagável:

1. Status do pedido e fila de chamados para salão.
2. Reclamação com ordem, responsável e SLA.
3. Solicitação de conta com handoff para caixa ou POS existente.

Esse wedge evita o build pesado e cria valor sem pedir que o restaurante troque seu sistema principal.

## Métrica de valor

- Tempo médio de resposta a chamado.
- Tempo de pedido pronto até entregue.
- Tempo de primeira resposta a reclamação.
- Tempo de conta solicitada até caixa agir.
- Percentual de mesas com mais de uma chamada.
- Reclamações por turno.
- Avaliação do cliente sobre clareza do atendimento.
