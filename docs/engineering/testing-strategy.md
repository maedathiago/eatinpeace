# Estratégia de testes

## Princípio

Testes são parte obrigatória do produto. Eat in Peace lida com fila, prioridade, reclamação e estado operacional em tempo real; bugs nessas áreas geram atrito direto no salão.

## Camadas

### Testes unitários

Devem cobrir regras puras de domínio:

- Ordenação e prioridade de tarefas.
- SLA de reclamação.
- Transições de status de pedido.
- Regras de handoff de conta.
- Validação de eventos.

### Testes de integração

Devem cobrir fronteiras reais:

- Persistência em Postgres/Supabase.
- Migrations.
- Políticas e permissões quando Supabase Auth ou RLS forem usados.
- Adaptadores para sistemas externos.
- Webhooks recebidos e enviados.

### Testes end-to-end

Devem cobrir fluxos completos do ponto de vista da operação:

1. Cliente abre sessão da mesa.
2. Pedido entra no sistema.
3. Cozinha ou salão atualiza status.
4. Garçom recebe tarefa priorizada.
5. Cliente abre reclamação.
6. Reclamação entra na fila correta e recebe responsável.
7. Cliente solicita conta.
8. Sistema faz handoff para caixa ou POS existente.

## Supabase

Quando Supabase entrar no projeto:

- Migrations devem ser testáveis localmente ou em ambiente dedicado de teste.
- Testes não devem depender de dados manuais.
- Dados de teste devem ser criados e limpos pelo próprio teste.
- Segredos reais nunca devem ser usados em teste versionado.

## Comandos

Os comandos oficiais da fundação P0 são:

```bash
make test
```

O comando roda:

```bash
GOCACHE=/tmp/eatinpeace-go-build go test ./...
```

O end-to-end P0 roda com:

```bash
make test-e2e
```

O E2E cobre sessão de mesa, pedido, SLA, pedido pronto, reclamação, conta como handoff e fechamento de turno.

Se o comando oficial mudar, este documento e o documento da fase precisam ser atualizados no mesmo branch.

## Contrato atual

A fundação executável usa store em memória para API e E2E. A migration Supabase/Postgres fica versionada e tem teste de contrato para garantir que as entidades mínimas existem e que o schema não entrou em POS, financeiro, fiscal, estoque ou delivery.

Quando um repositório Postgres real for implementado, ele deve adicionar teste de integração que aplique as migrations contra Postgres local ou ambiente dedicado.

## Regra de falha

Falha de teste obrigatório bloqueia a fase.

Se uma fase ainda não tem código executável, o documento da fase deve dizer isso explicitamente. A partir da primeira fase com implementação, não há entrega pronta sem teste automatizado e end-to-end aplicável.
