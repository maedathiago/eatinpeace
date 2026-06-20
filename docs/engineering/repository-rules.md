# Regras do repositório

## Fonte de verdade

Este repositório é a fonte de verdade do Eat in Peace. Decisões de produto, arquitetura, stack, testes e operação precisam estar documentadas aqui.

Conversas podem orientar decisões, mas a decisão final precisa virar arquivo versionado.

## Branch por fase

Cada fase deve ter sua própria branch.

Convenção:

```text
phase/<numero>-<slug>
```

Exemplos:

```text
phase/000-repo-rules
phase/001-product-mvp-scope
phase/002-go-api-foundation
phase/003-supabase-schema
```

Regras:

- Criar a branch a partir de `main`.
- Uma branch deve representar uma fase clara de trabalho.
- Não misturar fases independentes na mesma branch.
- Cada fase deve ter um documento em `docs/phases/`.
- O documento da fase deve registrar objetivo, escopo, entregáveis, documentação alterada e testes executados.

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

- Está em branch própria.
- O escopo da fase está documentado.
- O código, se houver, está implementado.
- A documentação afetada foi atualizada.
- Testes unitários, integração e end-to-end aplicáveis foram criados ou atualizados.
- Os testes foram executados e o resultado foi registrado no documento da fase ou na descrição do PR.
