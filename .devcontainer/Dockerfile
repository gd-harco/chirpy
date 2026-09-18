FROM mcr.microsoft.com/devcontainers/go:1-1.23-bookworm

# postgresql-client pour psql/pg_isready, pratique pour debug
RUN apt-get update && apt-get install -y postgresql-client

RUN go install github.com/pressly/goose/v3/cmd/goose@latest \
 && go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest \
 && go install github.com/bootdotdev/bootdev@latest
