# AGENTS — Mundo Invest Pipefy

Instruções para agentes AI que trabalham neste projeto.

## Comandos

```bash
# build
go build -buildvcs=false -o bin/server ./cmd/server

# testes (um módulo por vez — go.work não suporta ./... direto da raiz)
(cd cmd/server      && go test -buildvcs=false -count=1 ./...) || true
(cd modules/cliente && go test -buildvcs=false -count=1 ./...)
(cd modules/webhook && go test -buildvcs=false -count=1 ./...)
(cd pkg/pipefy      && go test -buildvcs=false -count=1 ./...) || true
(cd pkg/shared      && go test -buildvcs=false -count=1 ./...) || true

# formato
gofmt -w .

# lint (por módulo — go.work não suporta ./... direto da raiz)
cd modules/cliente && golangci-lint run ./... && cd ../..
cd modules/webhook && golangci-lint run ./... && cd ../..
cd cmd/server && golangci-lint run ./...
```

## Estrutura do Projeto

Go Workspace com 5 módulos independentes (`go.work`):

```
cmd/server/         → github.com/mundoinvest/server     (composition root, main)
modules/cliente/    → github.com/mundoinvest/cliente     (bounded context: Cliente)
modules/webhook/    → github.com/mundoinvest/webhook     (bounded context: Webhook)
pkg/pipefy/         → github.com/mundoinvest/pipefy      (ACL — mutations GraphQL)
pkg/shared/         → github.com/mundoinvest/shared      (APIResponse, APIError)
```

## Convenções

### Arquitetura (DDD + CQRS + Port/Adapter)

Cada bounded context segue esta estrutura:

```
modules/<context>/
├── domain/              ← modelos, erros, domain events — zero dependências externas
├── application/         ← commands + queries — define ports (interfaces), orquestra fluxo
└── infrastructure/
    ├── persistence/     ← adapters que implementam as ports de application
    └── http/            ← adapters HTTP (Gin handlers)
```

**Regras:**
- `domain/` não importa `application/` nem `infrastructure/`
- `application/` define interfaces (ports) — NUNCA importa `infrastructure/`
- `infrastructure/` implementa as ports definidas em `application/`
- Nomes de módulo e caminhos de import seguem `github.com/mundoinvest/<modulo>/...`

### Nomenclatura

- **Português para domínio:** `Cliente`, `CriarClienteHandler`, `ValorPatrimonio`, `TipoSolicitacao`
- **Inglês para infra/tech:** `Handler`, `Repository`, `PipefyClient`, `APIResponse`
- **Português para HTTP JSON fields:** `cliente_nome`, `cliente_email`, `valor_patrimonio`

### Domain Events

Dois eventos existem como structs Go, preparados para evolução futura para mensageria:
- `ClienteCriado` — emitido ao persistir um novo cliente
- `ClienteProcessado` — emitido ao atualizar status após webhook

Hoje são criados mas não publicados. Um futuro adapter de mensageria (SQS/SNS) consumiria esses structs.

### Testes

- **Application layer:** testes unitários com mocks (implementações fake das ports)
- **HTTP layer:** testes de integração com SQLite in-memory (`:memory:`) + `httptest`
- **Domain/Infrastructure:** sem testes próprios (cobertos pelos testes de application e HTTP)
- Nomes em português para fixtures/valores de teste: `"João Silva"`, `"Atualização cadastral"`

### Go Workspace

- `go.work` declara 5 módulos. O comando `go test ./...` não funciona da raiz porque o workspace Go não reconhece `.` como um módulo. Execute os testes dentro de cada diretório de módulo.
- O Dockerfile usa `go work sync` para resolver dependências antes do build.

## Fluxo de Mudanças

1. Leia `CONTEXT.md` para entender o domínio e os bounded contexts
2. Leia `docs/local-architecture.md` para visualizar o fluxo de dados
3. Alterações em `domain/` → verifique impacto em `application/` e testes
4. Novas ports (interfaces) → defina-as em `application/`, implemente em `infrastructure/`
5. Toda alteração com mudança de comportamento → adicione ou atualize testes
6. Antes de considerar pronto: `gofmt -w .` e `go test` em cada módulo afetado

## Referências

- [CONTEXT.md](CONTEXT.md) — glossário de domínio, bounded contexts, regras de negócio
- [docs/local-architecture.md](docs/local-architecture.md) — diagrama Mermaid do fluxo de dados local
- [docs/aws-production-architecture.md](docs/aws-production-architecture.md) — arquitetura AWS, trade-offs, capacidade
- [Pipefy API Docs](https://developers.pipefy.com/reference) — documentação oficial das mutations GraphQL
