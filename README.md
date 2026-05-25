# Mundo Invest — Client Management & Pipefy Integration

API de gerenciamento de clientes com mapeamento de cards para o Pipefy, desenvolvida como teste técnico para backend.

## Stack

| Camada | Tecnologia |
|--------|-----------|
| Linguagem | Go 1.26 |
| HTTP | [Gin](https://github.com/gin-gonic/gin) |
| Banco | SQLite via [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) (Go puro, zero CGO) |
| Testes | `testing` + `httptest` + `gin.TestMode` |

## Arquitetura

Monólito Modular com **DDD Estratégico + CQRS** via Go Workspace. Cada bounded context é um módulo Go separado, deployável isoladamente.

```
mundo-invest-pipefy/
├── go.work                          # workspace file
├── cmd/server/                      # composition root (entry point)
├── modules/
│   ├── cliente/                     # bounded context: Cliente
│   │   ├── domain/                  # aggregate root, domain events, errors
│   │   ├── application/             # commands (CriarCliente) + queries (ObterClientePorEmail)
│   │   └── infrastructure/          # persistence (SQLite), HTTP handlers
│   └── webhook/                     # bounded context: Webhook
│       ├── domain/                  # ProcessedEvent, CardUpdatedInput, errors
│       ├── application/             # commands (ProcessarCardUpdated)
│       └── infrastructure/          # persistence (SQLite), HTTP handlers
├── pkg/
│   ├── shared/                      # shared kernel (APIResponse, APIError)
│   └── pipefy/                      # anti-corruption layer (GraphQL mutations)
├── Dockerfile                       # multi-stage build
├── docker-compose.yml               # dev setup
├── lefthook.yml                     # pre-commit hooks
├── .github/workflows/ci.yml         # CI/CD pipeline
└── Makefile
```

**Princípios:**
- **DDD Estratégico:** Bounded contexts `cliente` e `webhook` com modelos de domínio próprios
- **CQRS:** Separação de Commands (mutações) e Queries (leituras) na camada de application
- **Port/Adapter:** Application define interfaces (ports), infrastructure implementa (adapters)
- **Domain Events:** `ClienteCriado`, `ClienteProcessado` — preparados para evolução para mensageria (SQS/SNS)
- **Anti-Corruption Layer:** `pkg/pipefy/` isola as mutations GraphQL do domínio

## Execução Local

```bash
# build
go build -buildvcs=false -o bin/server ./cmd/server

# rodar servidor
./bin/server
# Server started on :8080

# rodar testes
go test -count=1 ./cmd/server/... ./modules/... ./pkg/...

# docker
make docker-up
make docker-down

# lint
golangci-lint run ./...

# format
gofmt -w .
```

## Exemplos de Requisição

### 1. Criar Cliente

```bash
curl -X POST http://localhost:8080/clientes \
  -H "Content-Type: application/json" \
  -d '{
    "cliente_nome": "João Silva",
    "cliente_email": "joao.silva@example.com",
    "tipo_solicitacao": "Atualização cadastral",
    "valor_patrimonio": 250000
  }'
```

**Resposta (201):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "cliente_nome": "João Silva",
    "cliente_email": "joao.silva@example.com",
    "tipo_solicitacao": "Atualização cadastral",
    "valor_patrimonio": 250000,
    "status": "Aguardando Análise",
    "card_id": "card_sim_...",
    "created_at": "2026-05-24T18:00:00Z"
  }
}
```

### 2. Webhook — Card Atualizado

```bash
curl -X POST http://localhost:8080/webhooks/pipefy/card-updated \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "evt_123",
    "card_id": "card_456",
    "cliente_email": "joao.silva@example.com",
    "timestamp": "2026-05-18T12:00:00Z"
  }'
```

**Resposta (200):**
```json
{
  "success": true,
  "data": {
    "message": "event processed successfully"
  }
}
```

**Evento duplicado (409):**
```json
{
  "success": false,
  "error": {
    "code": "EVENT_ALREADY_PROCESSED",
    "message": "event already processed"
  }
}
```

**Cliente não encontrado (404):**
```json
{
  "success": false,
  "error": {
    "code": "CLIENT_NOT_FOUND",
    "message": "client not found"
  }
}
```

**Payload inválido (400):**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "invalid request body"
  }
}
```

## Regras de Negócio

| Condição | Prioridade |
|----------|-----------|
| `valor_patrimonio >= 200.000` | `prioridade_alta` |
| `valor_patrimonio < 200.000` | `prioridade_normal` |

## Mapeamento Pipefy (GraphQL)

O pacote `pkg/pipefy/` contém as mutations seguindo a [documentação oficial](https://developers.pipefy.com/reference):

### createCard ([docs](https://developers.pipefy.com/reference/cards#card-mutations))
```graphql
mutation($input: CreateCardInput!) {
  createCard(input: $input) {
    card { id title }
  }
}
```

### updateCardField ([docs](https://developers.pipefy.com/reference/fields#updating-fields-values))
```graphql
mutation($input: UpdateCardFieldInput!) {
  updateCardField(input: $input) {
    card { id }
    success
  }
}
```

O envio é simulado — o card_id é logado no console. Em produção, bastaria trocar `SimulateSend` por uma chamada HTTP `POST https://api.pipefy.com/graphql` com `Authorization: Bearer <token>`.

## Visão de Produção (AWS)

Em ambiente produtivo, a arquitetura escalaria da seguinte forma:

- **API Gateway + Lambda (Go):** Substitui o servidor Gin local. Cada endpoint vira uma função Lambda separada, com API Gateway roteando as requisições. Escala automaticamente com o volume de chamadas.
- **DynamoDB:** Substitui SQLite. Tabela `clientes` com chave primária `email` + GSI por `status` para queries. Tabela `eventos_processados` com TTL para expurgo automático de eventos antigos. DynamoDB Streams pode disparar processamento adicional em tempo real.
- **SQS + Lambda (Webhook):** O endpoint de webhook publica o evento em uma fila SQS; uma segunda Lambda consome a fila e processa (idempotência + regra de prioridade). Isso desacopla a ingestão do processamento e garante retry em caso de falha.
- **Secrets Manager:** Token de autenticação do Pipefy armazenado como secret, injetado na Lambda via variável de ambiente.
- **CloudWatch:** Logs estruturados de cada execução para tracing e alertas.

### Diagrama de fluxo

```
POST /clientes → API Gateway → Lambda CriarCliente → DynamoDB
                                                      ↓
POST /webhooks  → API Gateway → Lambda Ingestão → SQS → Lambda Processar → DynamoDB
                                                                           ↓
                                                                  Envia updateCardField → Pipefy API
```

## CI/CD

Pipeline no GitHub Actions: lint → test → security scan → docker build + push → deploy fake.

## Git Hooks

Lefthook configurado com pre-commit: `gofmt` + `golangci-lint` + `go test`. Instalar com:

```bash
go install github.com/evilmartians/lefthook@latest
lefthook install
```
