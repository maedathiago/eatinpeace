# Desenvolvimento local

## Requisitos

- Go 1.26 ou compatível.
- Docker para rodar Postgres local quando os testes de integração com banco forem ativados.
- Supabase CLI não é requisito para a primeira fundação; as migrations ficam versionadas em `supabase/migrations`.

## Comandos

Rodar todos os testes:

```bash
make test
```

Rodar o E2E P0:

```bash
make test-e2e
```

Subir a API local em memória:

```bash
make run
```

Por padrão a API escuta em `:8080`. Para trocar:

```bash
EATINPEACE_HTTP_ADDR=:8081 make run
```

## Banco

A migration inicial está em:

```text
supabase/migrations/202606200005_operational_foundation.sql
```

Ela define o contrato relacional para Supabase/Postgres. O caminho local executável usa store em memória para permitir desenvolvimento e E2E sem depender de segredos, Supabase CLI ou dados manuais.

O pacote `internal/storage/postgres` fornece o repositório baseado em `database/sql`, sem fixar driver externo no módulo inicial. A aplicação local usa o store em memória até que a fase de ambiente defina o driver e a URL de banco oficiais.

Quando uma fase passar a exigir Postgres real no ciclo padrão, o teste de integração deve aplicar as migrations versionadas contra um banco local ou ambiente dedicado de teste.
