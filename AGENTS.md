# Guia de agentes do Eat in Peace

## Enquadramento do produto

Eat in Peace não é apenas um cardápio por QR code e não deve virar sistema financeiro, POS ou gateway de pagamento. O produto central é uma camada operacional compartilhada para atendimento em restaurante: clientes, garçons, cozinha, caixa e gestão devem enxergar o mesmo estado da mesa e a mesma fila de trabalho.

A promessa do produto é: nenhuma fila invisível, nenhum pedido esquecido, nenhuma reclamação sem responsável.

## Princípios dos agentes de produto

- Tornar o estado atual visível antes de pedir intervenção humana.
- Preservar ordem e horário das solicitações sempre que justiça operacional importar.
- Escalar por tempo de espera, severidade e impacto no negócio, não por quem reclama mais alto.
- Reduzir polling manual do garçom. Se o garçom precisa ir até a cozinha perguntar status, o sistema já perdeu contexto útil.
- Manter o restaurante no controle. Agentes recomendam, priorizam e notificam; humanos podem sobrescrever decisões.
- Otimizar para um turno cheio e barulhento, não para uma demonstração perfeita.
- Embutir nos sistemas que o restaurante já usa sempre que possível, evitando construir POS, pagamento, fiscal, estoque ou conciliação financeira.

## Agentes centrais

1. Customer Concierge Agent
   - Ajuda o cliente a pedir, acompanhar status, chamar ajuda, reportar problemas e solicitar a conta.
   - Usa mensagens curtas e calmas que reduzem ansiedade sem prometer demais.

2. Floor Coordinator Agent
   - Ajuda garçons a ver chamadas de mesa, pedidos prontos, pedidos atrasados e contas solicitadas em uma fila operacional.
   - Prioriza tarefas por urgência, tempo de espera da mesa e dependência de ação humana.

3. Kitchen Expeditor Agent
   - Transforma o progresso da cozinha em estados confiáveis do pedido.
   - Detecta atrasos de preparo e alerta o salão antes de o cliente precisar reclamar.

4. Complaint Priority Agent
   - Converte reclamações em tickets com horário, severidade, responsável, status e registro de resolução.
   - Evita que reclamações novas passem na frente de reclamações antigas sem motivo.

5. Bill Handoff Agent
   - Coordena a solicitação de conta e o handoff para caixa, garçom ou sistema existente.
   - Evita a espera dolorosa do fim da refeição sem processar pagamento, emitir fiscal ou mexer no financeiro.

6. Manager Insights Agent
   - Mostra gargalos, reclamações recorrentes, itens atrasados, garçons sobrecarregados e mesas em risco.
   - Foca em ações operacionais durante o serviço e análise depois do turno.

## Estilo de trabalho

- Escrever documentação de produto em português, exceto quando uma integração técnica exigir inglês.
- Preferir fatias pequenas e testáveis de MVP em vez de promessas amplas de plataforma.
- Toda funcionalidade deve identificar ator, dor, evento, mudança de estado e métrica de sucesso.
- Não assumir integração completa com POS no primeiro MVP; definir primeiro o que funciona manualmente.
- Separar com rigidez orquestração de atendimento de operação financeira. O Eat in Peace pode avisar que a conta foi solicitada, mas não deve calcular impostos, processar cartão, emitir nota ou reconciliar caixa.

## Regras do repositório

- Cada fase de trabalho deve ter uma branch própria no formato `phase/<numero>-<slug>`, criada a partir de `main`.
- A stack oficial é Go com Supabase. Go deve concentrar API, domínio, integrações e testes; Supabase deve concentrar Postgres e serviços gerenciados acordados na documentação da fase.
- Tudo que for criado precisa ser documentado no mesmo branch.
- Toda alteração de comportamento, arquitetura, API, schema, operação ou teste deve atualizar a documentação correspondente.
- Testes são obrigatórios para mudanças executáveis: unitários, integração quando houver fronteira externa e end-to-end para fluxos críticos.
- Nenhuma fase deve ser considerada pronta sem documentação atualizada e evidência de testes rodados ou uma justificativa explícita quando a fase ainda não tiver código executável.
- As regras completas estão em [docs/engineering/repository-rules.md](docs/engineering/repository-rules.md).
