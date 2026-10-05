BIN     := app
PKG     := example.com/app
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X '$(PKG)/internal/pkg/buildutil.Version=$(VERSION)'
IMAGE   ?= $(BIN):$(VERSION)

.PHONY: dev serve build run migrate migrate-down migrate-status migrate-new seed test vet lint tidy docker-build clean

# All targets read ./.env (loaded by config.LoadDotenv). Copy .env.example first.

# --- local dev -------------------------------------------------------------

# Migrate + seed (both idempotent), then serve.
dev: migrate seed serve

serve:
	go run ./cmd/$(BIN) serve

# --- database --------------------------------------------------------------

migrate:
	go run ./cmd/$(BIN) migrate

migrate-down:
	go run ./cmd/$(BIN) migrate down

migrate-status:
	go run ./cmd/$(BIN) migrate status

# make migrate-new name=add_users
migrate-new:
	@test -n "$(name)" || (echo "usage: make migrate-new name=<snake_case_name>" && exit 1)
	go run ./cmd/$(BIN) migrate create $(name)

seed:
	go run ./cmd/$(BIN) seed

# --- build / checks --------------------------------------------------------

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o bin/$(BIN) ./cmd/$(BIN)

run: build
	./bin/$(BIN) serve

test:
	go test -race ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

docker-build:
	docker build --load --build-arg BUILD_VERSION=$(VERSION) -t $(IMAGE) .

clean:
	rm -rf bin
