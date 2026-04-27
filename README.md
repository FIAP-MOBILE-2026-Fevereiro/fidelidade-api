# fidelidade-api

API Go estruturada para programa de fidelidade com Gin, Postgres, sqlc, migrações e Docker Compose.

## Stack

- Go com layout `cmd/` e `internal/`
- Gin para HTTP
- pgx + sqlc para acesso tipado ao Postgres
- golang-migrate para migrações
- Postgres 16 + PostGIS para consultas geoespaciais
- JWT para autenticação

## Estrutura

```text
cmd/api                  # entrada da aplicação
db/migrations            # schema versionado
db/queries               # queries sqlc
db/seeds                 # dados demo para teste local
internal/app             # bootstrap da aplicação
internal/http            # handlers, middleware, router
internal/platform        # postgres e migrações
internal/service         # regras de negócio
internal/store           # store + código gerado pelo sqlc
```

## Configuração

Copie `.env.example` para `.env` e ajuste se necessário.

Variáveis principais:

- `DATABASE_URL`
- `JWT_SECRET`
- `HTTP_PORT`
- `PUBLIC_BASE_URL`
- `AUTO_MIGRATE`

## Rodando localmente com Docker

```bash
docker compose up --build
```

O compose sobe:

- `postgres` com PostGIS
- `migrate` para aplicar as migrações
- `seed` com um programa e um QR code de demonstração
- `api` em `http://localhost:3000`

## Rodando localmente sem Docker

1. Suba um Postgres com PostGIS.
2. Configure `.env`.
3. Rode as migrações com o binário ou deixe `AUTO_MIGRATE=true`.
4. Execute:

```bash
go run ./cmd/api
```

## Geração de código

```bash
sqlc generate
go test ./...
```

## Endpoints implementados

- `POST /v1/usuarios`
- `POST /v1/usuarios/autenticacao`
- `GET /v1/usuarios/:usuarioId`
- `PUT /v1/usuarios/:usuarioId`
- `DELETE /v1/usuarios/:usuarioId/excluir`
- `POST /v1/usuarios/:usuarioId/imagem`
- `GET /v1/programas`
- `GET /v1/programas/:programaId`
- `GET /v1/usuarios/:usuarioId/programas/ativos`
- `GET /v1/usuarios/:usuarioId/programas/finalizados`
- `GET /v1/usuarios/:usuarioId/programas/ultimo-selo`
- `GET /v1/usuarios/:usuarioId/programas/:programaId/selos`
- `POST /v1/usuarios/:usuarioId/programas/:programaId/selos`
- `POST /v1/qr-codes/validar`
- `GET /v1/usuarios/:usuarioId/recompensas`
- `POST /v1/usuarios/:usuarioId/programas/:programaId/resgatar`

## Dados demo

O seed cria:

- Programa `prog_b4c5d6e7`
- QR code compatível com o exemplo do contrato

Payload demo do QR code:

```text
eyJwcm9ncmFtYUlkIjoicHJvZ19iNGM1ZDZlNyIsImVzdGFiZWxlY2ltZW50b0lkIjoiZXN0X2ExYjJjM2Q0IiwiY2FyaW1ibyI6IjEyMzQ1Njc4OTAiLCJ0aW1lc3RhbXAiOiIyMDI2LTA0LTI3VDE0OjMwOjAwWiJ9
```

Fluxo mínimo para testar:

1. Criar usuário.
2. Autenticar e obter o token JWT.
3. Validar o QR code em `/v1/qr-codes/validar`.
4. Ganhar selo em `/v1/usuarios/{usuarioId}/programas/prog_b4c5d6e7/selos`.

## Observações

- O upload de imagem usa armazenamento local em `storage/profiles`.
- O middleware exige que `usuarioId` no path coincida com o sujeito do JWT.
- O esquema de dados foi remodelado para Postgres, independente do `db.sql` de referência.
