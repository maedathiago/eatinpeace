# Regras do repositório

## Fonte de verdade

Este repositório é a fonte de verdade do Eat in Peace. Decisões de produto, arquitetura, stack, testes e operação precisam estar documentadas aqui.

Conversas podem orientar decisões, mas a decisão final precisa virar arquivo versionado.

## Fluxo com branches sincronizadas

O fluxo padrão do repositório é trabalhar em branches. `main` é a base de integração e precisa permanecer sincronizada com `origin/main`.

Regras:

- Antes de começar trabalho relevante, buscar o estado remoto e atualizar a `main` local a partir de `origin/main`.
- Toda feature, correção ou etapa deve ter branch própria criada a partir da `main` atualizada.
- A branch local deve ser publicada no remoto quando houver entrega ou colaboração relevante.
- Se `origin/main` avançar durante o trabalho, atualizar a branch com a `main` mais recente antes de finalizar.
- Antes de integrar, confirmar que a branch está validada, documentada e alinhada com a `main` atual.
- Ao concluir a integração, fazer push de `main` para `origin/main` e confirmar que local e remoto estão sem divergência.

## Entregas e documentos de fase

As fases continuam existindo como unidades de planejamento e documentação, e devem ser implementadas em branch própria.

Regras:

- Uma entrega deve representar uma fase ou mudança clara de trabalho.
- Não misturar mudanças independentes no mesmo commit quando isso dificultar revisão ou revert.
- Cada fase relevante deve ter um documento em `docs/phases/`.
- O documento da fase deve registrar objetivo, escopo, entregáveis, documentação alterada e testes executados.
- A branch da fase deve ser nomeada no formato `phase/<numero>-<slug>`, exceto quando a mudança for uma correção menor com outro prefixo claro.

## Histórico e commits

O histórico de `main` deve continuar fácil de ler e reverter.

Regra:

- Preferir commits que representem entregas completas e tenham mensagem clara.
- Correções intermediárias, checkpoints locais e commits de tentativa não devem ser publicados em `main` como ruído permanente.
- Se a branch tiver vários commits locais de tentativa, condensar antes de integrar em `main` ou documentar a exceção.
- Integrar a branch em `main` com histórico simples e depois sincronizar o remoto.

Motivo:

- Facilitar revert de uma entrega inteira.
- Manter `main` legível por marcos de produto e engenharia.
- Reduzir ruído de commits intermediários que não representam entregas estáveis.

Exceção:

- Se uma entrega precisar preservar commits separados por motivo técnico forte, o motivo deve ser documentado.

## Stack oficial

A stack oficial é:

- Go para API, domínio, serviços, integrações, workers, CLIs internos e testes.
- Supabase para Postgres e serviços gerenciados definidos explicitamente por documentação de fase.

Diretrizes:

- Código Go deve seguir a organização definida pelo projeto no momento em que o backend for criado.
- Migrations e políticas de banco devem ser versionadas no repo quando Supabase for introduzido.
- Configuração de ambiente deve ter exemplo versionado, sem segredos reais.
- Integrações com sistemas externos devem ser isoladas por interfaces e documentadas.
- Não construir POS, gateway de pagamento, fiscal, conciliação financeira ou ERP. O produto é uma camada embedded de orquestração de atendimento.

## Documentação obrigatória

Tudo que for criado precisa ter documentação.

Alterações que exigem atualização de documentação:

- Novo fluxo de produto.
- Novo agente.
- Nova API.
- Nova tabela, migration, política ou função Supabase.
- Novo serviço Go.
- Nova dependência relevante.
- Nova variável de ambiente.
- Novo comando operacional.
- Novo teste ou mudança na estratégia de testes.
- Mudança de escopo, arquitetura ou regra de negócio.

Regras:

- Se o comportamento muda, a documentação precisa mudar junto.
- Se uma entidade nova é criada, deve existir documentação suficiente para outra pessoa entender por que ela existe e como usá-la.
- README deve apontar para os documentos centrais.
- Documentos de fase devem registrar quais docs foram alterados.

## Testes obrigatórios

Mudanças executáveis precisam de testes.

Camadas mínimas:

- Testes unitários para regras de domínio, validação e priorização.
- Testes de integração para fronteiras com banco, Supabase e adaptadores externos.
- Testes end-to-end para fluxos críticos do restaurante.

Fluxos end-to-end mínimos quando o produto existir:

- Abrir sessão de mesa.
- Criar ou receber pedido.
- Atualizar status do pedido.
- Abrir chamado de garçom.
- Abrir reclamação e preservar prioridade.
- Solicitar conta e fazer handoff para caixa ou sistema existente.

Uma fase com código executável não está pronta se:

- Não houver testes para o comportamento criado.
- O fluxo end-to-end afetado não tiver cobertura.
- Os comandos de teste não estiverem documentados.
- Algum teste obrigatório estiver quebrado.

## Critério de pronto

Uma fase só está pronta quando:

- Está em branch própria criada a partir de `main` atualizada.
- A branch foi mantida em sync com a `main` mais recente antes da integração.
- O escopo da fase está documentado.
- O código, se houver, está implementado.
- A documentação afetada foi atualizada.
- Testes unitários, integração e end-to-end aplicáveis foram criados ou atualizados.
- Os testes foram executados e o resultado foi registrado no documento da fase ou na descrição do PR.
- A entrega foi integrada em `main`.
- `main` local e `origin/main` foram sincronizadas após a integração.
