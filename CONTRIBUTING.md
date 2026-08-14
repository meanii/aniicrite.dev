# Contributing

Thanks for your interest! This is a small Go SSR + HTMX + SQLite blog engine.

## Prerequisites

- Go 1.26+
- [`templ`](https://templ.guide) CLI (templates are compiled to Go):

  ```sh
  go install github.com/a-h/templ/cmd/templ@v0.3.1020
  ```

## Local development

```sh
# 1. Generate template code (needed after editing any .templ file)
templ generate

# 2. Create an admin password hash
go run ./cmd/server hash-password 'your-password'

# 3. Run (see the env-var table in the README for all options)
ADMIN_PASSWORD_HASH='<hash from step 2>' \
SESSION_HASH_KEY="$(openssl rand -hex 32)" \
SESSION_BLOCK_KEY="$(openssl rand -hex 32)" \
go run ./cmd/server
```

The server listens on `:8080` by default and stores its SQLite database and
uploads under `./data`.

## Generated code

`web/templates/*_templ.go` are generated from the `.templ` sources by
`templ generate` and **are committed** so that `go build` works without the
`templ` CLI. If you change a `.templ` file, re-run `templ generate` and commit
the result. CI fails if the generated files are stale.

## Before you push

```sh
gofmt -w .
go vet ./...
go test ./...
```

- Go 1.22+ range-over-int is preferred: `for i := range n` over
  `for i := 0; i < n; i++`.
- Keep changes focused; match the surrounding style.

## Reporting bugs / requesting features

Open an issue with clear reproduction steps or a concrete use case.
