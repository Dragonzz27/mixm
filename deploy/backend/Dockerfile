FROM golang:1.25.0-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY internal ./internal
COPY cmd ./cmd
RUN CGO_ENABLED=0 go build -trimpath -o /out/server ./cmd/server && \
    CGO_ENABLED=0 go build -trimpath -o /out/worker ./cmd/worker && \
    CGO_ENABLED=0 go build -trimpath -o /out/admin ./cmd/admin && \
    CGO_ENABLED=0 go build -trimpath -o /out/initdb ./cmd/initdb && \
    CGO_ENABLED=0 go build -trimpath -o /out/migrate ./cmd/migrate && \
    CGO_ENABLED=0 go build -trimpath -o /out/migrate-sqlite ./cmd/migrate-sqlite

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -g 10001 app && adduser -D -u 10001 -G app app && \
    mkdir -p /app/uploads/assignments /app/static/export /app/data /app/kubeconfig && \
    chown -R app:app /app
WORKDIR /app
COPY --from=build /out/ /usr/local/bin/
USER app
EXPOSE 8002 30022
CMD ["server"]
