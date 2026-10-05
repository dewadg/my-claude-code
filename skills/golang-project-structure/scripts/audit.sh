#!/usr/bin/env bash
# Audit a Go service against the golang-project-structure layout.
# usage: audit.sh [project-dir]   (default: .)
# Prints one line per finding: "<FAIL|WARN> <rule> <path>: <detail>".
# Exit 1 when any FAIL is found, 0 otherwise. Read-only.
set -uo pipefail

ROOT=${1:-.}
cd "$ROOT" || { echo "no such dir: $ROOT" >&2; exit 2; }

# Findings are buffered so the summary can count lines emitted from subshells.
OUT=$(mktemp)
trap 'rm -f "$OUT"' EXIT
exec 3>&1 1>"$OUT"

fail() { echo "FAIL $1 $2: $3"; }
warn() { echo "WARN $1 $2: $3"; }
has() { grep -qE -- "$1" "$2" 2>/dev/null; }

# Non-test Go files under a path, one per line.
gofiles() { find "$1" -type f -name '*.go' ! -name '*_test.go' 2>/dev/null | sort; }

LAYERS='handler|service|repository|routes|module|schemas|errors|deps|doc|service_interfaces|store_interfaces'
EXEMPT_DOMAINS=' pkg infra seeders httpserver '

# ---------------------------------------------------------------- module ---
if [ ! -f go.mod ]; then
  fail module go.mod "missing; run from the Go module root"
  exec 1>&3; cat "$OUT"; exit 1
fi
MODULE=$(awk '/^module /{print $2; exit}' go.mod)

# ------------------------------------------------------------------- cmd ---
BINS=$(find cmd -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort)
[ -n "$BINS" ] || fail cmd cmd/ "no cmd/<binary>/ directory"
for b in $BINS; do
  [ -f "$b/main.go" ] || { fail cmd "$b" "missing main.go"; continue; }
  has 'github.com/spf13/cobra' "$b/main.go" || fail cmd "$b/main.go" "root command must be cobra"
  has 'LoadDotenv\(' "$b/main.go" || warn cmd "$b/main.go" "config.LoadDotenv() not called before Execute"

  SERVE=$(grep -lE 'fx\.New\(' "$b"/*.go 2>/dev/null | head -1)
  if [ -z "$SERVE" ]; then
    warn cmd "$b" "no fx.New in any subcommand (fine for a pure CLI, not for a service)"
  else
    has 'fx\.Supply\(cfg\)' "$SERVE" || fail fx "$SERVE" "config must be supplied as a value: fx.Supply(cfg)"
    has 'config\.Parse\(\)' "$SERVE" || fail config "$SERVE" "config.Parse() must run inside the subcommand, before fx.New"
    has 'AssertRequiredEnv\(' "$SERVE" || fail config "$SERVE" "config.AssertRequiredEnv(cfg) must run before fx.New"
    [ "$(grep -cE 'fx\.New\(' "$b"/*.go | awk -F: '{s+=$2} END{print s}')" -le 1 ] ||
      warn fx "$b" "more than one fx.New; the serve command should be the only registration point"
  fi

  if [ -d build/migrations ]; then
    grep -lqE 'goose\.NewProvider\(' "$b"/*.go 2>/dev/null ||
      fail migrate "$b" "no subcommand builds a goose.NewProvider over migrations.FS"
  fi
done

# ------------------------------------------------------------ migrations ---
if [ -d build/migrations ]; then
  E=build/migrations/embed.go
  [ -f "$E" ] || fail migrate "$E" "missing; migrations must be //go:embed'd"
  [ -f "$E" ] && { has '//go:embed \*\.sql' "$E" || fail migrate "$E" "must embed *.sql"; }
  for f in build/migrations/*.sql; do
    [ -e "$f" ] || continue
    n=$(basename "$f")
    echo "$n" | grep -qE '^[0-9]{5}_[a-z0-9_]+\.sql$' ||
      fail migrate "$f" "name must be NNNNN_snake_case.sql (goose sequential)"
    has '^-- \+goose Up' "$f" || fail migrate "$f" "missing '-- +goose Up'"
  done
  grep -rqE 'pressly/goose' go.mod || fail migrate go.mod "github.com/pressly/goose/v3 not required"
else
  warn migrate build/migrations "absent (fine only if the service has no database)"
fi

# ---------------------------------------------------------------- config ---
C=internal/pkg/config
if [ ! -d "$C" ]; then
  fail config "$C" "missing"
else
  CFILES=$(gofiles "$C")
  # shellcheck disable=SC2086
  grep -qE 'github.com/caarlos0/env' $CFILES || fail config "$C" "env must be parsed with github.com/caarlos0/env"
  # shellcheck disable=SC2086
  grep -qE 'github.com/joho/godotenv' $CFILES || fail config "$C" ".env must be loaded with github.com/joho/godotenv"
  for fn in 'func Parse\(\)' 'func AssertRequiredEnv\(' 'func LoadDotenv\('; do
    # shellcheck disable=SC2086
    grep -qE "$fn" $CFILES || fail config "$C" "missing ${fn//\\/}"
  done
  # Env reads belong in config only.
  for f in $(gofiles internal); do
    case "$f" in "$C"/*|internal/pkg/observability/*) continue ;; esac
    has 'os\.(Getenv|LookupEnv)\(' "$f" && warn config "$f" "reads env directly; add a Config field instead"
  done
fi

# ----------------------------------------------------------------- infra ---
I=internal/infra
if [ ! -f "$I/module.go" ]; then
  fail infra "$I/module.go" "missing"
else
  has 'fx\.Module\("infra"' "$I/module.go" || fail infra "$I/module.go" 'must declare var Module = fx.Module("infra", ...)'
  grep -qE 'name:"base"' $(gofiles "$I") || warn infra "$I" 'no "base"-tagged logger; NewLogger cannot warn after its own pipeline shuts down'
  grep -qE 'slog\.SetDefault\(' $(gofiles "$I") || fail logging "$I" "configured logger never installed with slog.SetDefault"
fi

# --------------------------------------------------------- observability ---
O=internal/pkg/observability
if [ ! -d "$O" ]; then
  fail otel "$O" "missing"
else
  OF=$(gofiles "$O")
  # shellcheck disable=SC2086
  grep -qE 'func Setup(Logs)?\(' $OF || fail otel "$O" "missing log pipeline entry point (Setup/SetupLogs)"
  # shellcheck disable=SC2086
  grep -qE 'func SetupMetrics\(' $OF || fail otel "$O" "missing SetupMetrics"
  # shellcheck disable=SC2086
  grep -qE 'func SetupTracing\(' $OF || fail otel "$O" "missing SetupTracing"
  # shellcheck disable=SC2086
  grep -qE 'func Start\(ctx context\.Context' $OF || fail otel "$O" "missing span helper Start(ctx, name)"
  # shellcheck disable=SC2086
  grep -qE 'func End\(span trace\.Span, err \*error\)' $OF || fail otel "$O" "missing span helper End(span, &err)"
  # shellcheck disable=SC2086
  grep -qE 'noop\.NewTracerProvider|noop\.NewMeterProvider' $OF || warn otel "$O" "disabled pipelines should return no-op providers"
  # shellcheck disable=SC2086
  grep -qE 'SetTextMapPropagator' $OF || fail otel "$O" "no global propagator installed"
fi

# --------------------------------------------------------------- logging ---
L=internal/pkg/logging
if [ ! -d "$L" ]; then
  fail logging "$L" "missing"
else
  grep -qE 'KeySource *= *"source"' $(gofiles "$L") || fail logging "$L" 'missing KeySource = "source"'
  grep -qE 'KeyError *= *"error"' $(gofiles "$L") || fail logging "$L" 'missing KeyError = "error"'
fi
for f in $(gofiles internal) $(gofiles cmd); do
  case "$f" in internal/pkg/observability/*|internal/pkg/logging/*) continue ;; esac
  # Third-party loggers.
  has '"(go\.uber\.org/zap|github\.com/rs/zerolog|github\.com/sirupsen/logrus)"' "$f" &&
    fail logging "$f" "use log/slog, not a third-party logger"
  has '^\s*"log"$' "$f" && warn logging "$f" 'imports stdlib "log"; use log/slog'
  case "$f" in cmd/*) ;; *)
    has '\bfmt\.Print' "$f" && warn logging "$f" "fmt.Print* in library code; log via slog" ;;
  esac
  # Package-level calls skip the source attribute.
  grep -nE '\bslog\.(Debug|Info|Warn|Error)(Context)?\(' "$f" 2>/dev/null | while IFS=: read -r ln _; do
    echo "WARN logging $f:$ln: package-level slog call has no source; use l := slog.With(logging.KeySource, \"<pkg>.<Struct>.<Method>\")"
  done
  grep -nE 'slog\.With\("source"' "$f" 2>/dev/null | while IFS=: read -r ln _; do
    echo "WARN logging $f:$ln: literal \"source\"; use logging.KeySource"
  done
  grep -nE '(Debug|Info|Warn|Error)(Context)?\(([a-zA-Z.()]+, )?"[^"]*", "err",' "$f" 2>/dev/null | while IFS=: read -r ln _; do
    echo "WARN logging $f:$ln: attr key \"err\"; use logging.KeyError (\"error\")"
  done
  grep -nE '\.(Debug|Info|Warn|Error)(Context)?\(([a-zA-Z.()]+, )?"[A-Z][a-z]' "$f" 2>/dev/null | while IFS=: read -r ln _; do
    echo "WARN logging $f:$ln: message starts uppercase; messages are lowercase phrases"
  done
done

# ------------------------------------------------------------ httpserver ---
H=internal/httpserver
if [ -d "$H" ]; then
  has 'fx\.Module\("httpserver"' "$H/module.go" || fail http "$H/module.go" 'must declare var Module = fx.Module("httpserver", ...)'
  grep -qE 'ResultTags\(`name:"' "$H/module.go" || fail http "$H/module.go" "route groups must be provided under name tags"
  grep -qE 'net\.Listen\(' $(gofiles "$H") || warn http "$H" "OnStart should bind with net.Listen synchronously, then Serve on a goroutine"
fi

# --------------------------------------------------------------- domains ---
for d in $(find internal -mindepth 1 -maxdepth 1 -type d | sort); do
  name=$(basename "$d")
  case "$EXEMPT_DOMAINS" in *" $name "*) continue ;; esac

  if [ ! -f "$d/module.go" ]; then
    fail domain "$d" "missing module.go"
  else
    has "fx\.Module\(\"$name\"" "$d/module.go" || fail domain "$d/module.go" "must declare var Module = fx.Module(\"$name\", ...)"
  fi
  # Registration point.
  grep -rqE "\b$name\.Module\b" cmd 2>/dev/null || warn domain "$d" "$name.Module not registered in any cmd/ fx.New list"

  for f in $(find "$d" -maxdepth 1 -type f -name '*.go' ! -name '*_test.go' | sort); do
    base=$(basename "$f" .go)
    if echo "$base" | grep -qE '(^|_)handlers(_|$)'; then
      fail naming "$f" "layer word is 'handler', never 'handlers'"
    fi
    layered=0
    if echo "$base" | grep -qE "^((admin|visitor|shared)_)?($LAYERS)(_[a-z0-9_]+)?$"; then
      layered=1
    elif echo "$base" | grep -qE "(^|_)(handler|service|repository|routes|module|schemas)(_|$)"; then
      fail naming "$f" "layer word must lead: {actor_}{layer}{_flow}.go (e.g. ${name}_repository.go -> repository.go, x_handler.go -> handler_x.go)"
    fi

    layer=$(echo "$base" | sed -E "s/^((admin|visitor|shared)_)?//" | sed -E "s/^(service_interfaces|store_interfaces|[a-z]+).*/\1/")

    # Persistence imports: repository layer, module.go, deps.go only.
    if has '"(gorm\.io/gorm|database/sql|github\.com/jackc/pgx[^"]*)"' "$f"; then
      case "$layer" in repository|module|deps) ;; *)
        fail layering "$f" "persistence import outside repository/module/deps files" ;;
      esac
    fi
    # Transport imports: never in service/repository or concern files.
    if has '"github\.com/(labstack/echo[^"]*|go-chi/chi[^"]*|gin-gonic/gin)"' "$f"; then
      if [ $layered -eq 0 ]; then
        fail layering "$f" "concern file imports the router; it belongs in the handler layer (rename handler_<x>.go)"
      else
        case "$layer" in service|repository|store_interfaces)
          fail layering "$f" "$layer file imports the router" ;;
        esac
      fi
    fi
    # Cross-domain imports.
    grep -oE "\"$MODULE/internal/[a-z0-9_]+" "$f" 2>/dev/null | sed -E 's#.*/internal/##' | sort -u | while read -r other; do
      case "$other" in pkg|"$name") ;; *)
        echo "WARN domain $f: imports internal/$other; depend on a consumer-defined interface wired through fx instead" ;;
      esac
    done
    # Tracing on layered methods that take a context.
    case "$layer" in handler|service|repository)
      if has '^func \([a-z]+ \*?[A-Za-z]+\) [A-Za-z]+\((ctx context\.Context|c echo\.Context)' "$f" && ! has '\.Start\((ctx|c\.Request\(\)\.Context\(\))' "$f"; then
        echo "WARN tracing $f: methods take a context but open no span (tracing.Start + defer tracing.End(span, &err))"
      fi ;;
    esac
    # Sentinels live in errors.go.
    if [ "$layer" != errors ] && has '^\s*(var )?Err[A-Za-z]+ *= *(errors\.New|fmt\.Errorf)' "$f"; then
      fail placement "$f" "error sentinels belong in errors.go"
    fi
  done
done

# ------------------------------------------------------------ Dockerfile ---
if [ ! -f Dockerfile ]; then
  fail docker Dockerfile "missing"
else
  [ "$(grep -cE '^FROM ' Dockerfile)" -ge 2 ] || fail docker Dockerfile "must be multi-stage (builder + runtime)"
  has 'COPY go\.mod go\.sum' Dockerfile || warn docker Dockerfile "copy go.mod/go.sum and 'go mod download' before sources for layer caching"
  has 'CGO_ENABLED=0' Dockerfile || warn docker Dockerfile "build with CGO_ENABLED=0 for a static binary"
  has '\-trimpath' Dockerfile || warn docker Dockerfile "build with -trimpath"
  has 'buildutil\.Version=' Dockerfile || fail docker Dockerfile "inject version via -ldflags -X .../buildutil.Version"
  has '^FROM gcr\.io/distroless/[a-z0-9-]+:nonroot' Dockerfile || warn docker Dockerfile "runtime stage should be gcr.io/distroless/*:nonroot"
  if [ -d build/migrations ]; then
    has '^COPY build' Dockerfile || fail docker Dockerfile "build/ (embedded migrations) is not copied into the builder"
  fi
  has '^ENTRYPOINT \["/' Dockerfile || fail docker Dockerfile 'exec-form ENTRYPOINT ["/<binary>"] required'
fi
[ -f .dockerignore ] || warn docker .dockerignore "missing; .env and .git leak into the build context"

# -------------------------------------------------------------- Makefile ---
if [ ! -f Makefile ]; then
  fail make Makefile "missing"
else
  for t in build dev serve migrate test vet; do
    has "^$t:" Makefile || fail make Makefile "missing target '$t'"
  done
  for t in seed migrate-new migrate-down migrate-status lint docker-build; do
    has "^$t:" Makefile || warn make Makefile "missing target '$t'"
  done
fi
[ -f .env.example ] || fail env .env.example "missing; every env var Config reads must be listed"
if [ -f .env.example ] && [ -d "$C" ]; then
  for v in $(grep -ohE 'env:"[A-Z0-9_]+"' $(gofiles "$C") | sed -E 's/env:"([A-Z0-9_]+)"/\1/' | sort -u); do
    grep -qE "^#? ?$v=" .env.example || warn env .env.example "$v is parsed by config but not documented"
  done
fi

exec 1>&3
cat "$OUT"
FAILS=$(grep -c '^FAIL ' "$OUT")
WARNS=$(grep -c '^WARN ' "$OUT")
echo "---"
echo "summary: $FAILS FAIL, $WARNS WARN"
[ "$FAILS" -eq 0 ]
