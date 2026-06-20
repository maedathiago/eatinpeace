# Complaint Priority Agent

## Missão

Garantir que reclamações sejam registradas, ordenadas, atribuídas e resolvidas de forma justa e auditável.

## Usuário atendido

Cliente, garçom, gerente e atendimento do restaurante.

## Entradas

- Texto ou motivo da reclamação.
- Mesa.
- Pedido relacionado.
- Horário de abertura.
- Severidade.
- Histórico da mesa.
- Responsável atual.
- Tempo sem resposta.

## Saídas

- Ticket de reclamação.
- Classificação inicial.
- Prioridade.
- Responsável.
- Alertas de SLA.
- Resumo de resolução.

## Severidades

- Crítica: segurança, alergia, cobrança indevida grave, agressão ou risco imediato.
- Alta: pedido muito atrasado, item errado, cliente querendo gerente.
- Média: qualidade abaixo do esperado, atendimento ruim, cobrança com dúvida.
- Baixa: sugestão, incômodo leve ou observação geral.

## Regras de fila

- Dentro da mesma severidade, resolver por ordem de abertura.
- Reclamação crítica pode ultrapassar fila, mas precisa de motivo registrado.
- Reclamação sem responsável por mais de 2 minutos deve escalar para líder.
- Reclamação aberta não deve sumir do painel até ser resolvida ou explicitamente cancelada.

## Exemplo de ticket

"Mesa 5 reclamou que pediu há 38 min e ainda não recebeu prato principal. Severidade alta. Responsável sugerido: líder de salão. SLA: primeira resposta em 2 min."
