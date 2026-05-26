FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY cmd/server/go.mod cmd/server/go.sum* ./cmd/server/
COPY modules/client/go.mod modules/client/go.sum* ./modules/client/
COPY modules/webhook/go.mod modules/webhook/go.sum* ./modules/webhook/
COPY pkg/shared/go.mod pkg/shared/go.sum* ./pkg/shared/
COPY pkg/pipefy/go.mod pkg/pipefy/go.sum* ./pkg/pipefy/

RUN go work sync && cd cmd/server && go mod download

COPY . .

RUN go build -buildvcs=false -o server ./cmd/server

FROM alpine:3.21

ENV GIN_MODE=release

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
