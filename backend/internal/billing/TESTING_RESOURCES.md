# Developer 05 testing resources and team handoff

## Authorization and cleanup rule

On 2026-11-09 the user authorized installing pgx and Asynq for testing only, required a Markdown record of versions/resources for the group, and required removing test dependency changes from protected documents afterward. This rule is also saved in IMPLEMENTATION_PLAN.md as persistent task memory.

Use a temporary Go `-modfile` outside the source tree for these test dependencies. The application still has one source module. Preserve `backend/go.mod` byte-for-byte and preserve the absence of `backend/go.sum`; never commit testing dependencies to either protected path. Do not modify other developers' sources, shared configuration, or their migrations. Any temporary database/Redis instances must contain only synthetic test data.

## Starting state

- Branch: `feat/ian-backend-engineer-c`.
- Resumed implementation commit: `0ef7d5e8ba65829c146dde8c1844e3c11efd351f`.
- Protected `backend/go.mod`: module `github.com/Ashomondi/ikonex-academy/backend`, Go `1.22.2`, no requirements.
- Protected `backend/go.sum`: absent.
- Existing unrelated edit: one leading blank line in `Docs/DECISIONS.md`; preserve unchanged.

## Direct test dependencies

| Resource | Requested version | Purpose | Status |
|---|---|---|---|
| github.com/jackc/pgx/v5 | v5.7.2 | Compatible pgx.Tx contracts, PostgreSQL integration tests | Installed; module checksums verified |
| github.com/hibiken/asynq | v0.25.1 | Actual Redis-backed queue implementation and tests | Installed; module checksums verified |
| Go | go1.22.2 linux/amd64 | Build, vet, unit/integration/race tests | Verified installed |
| Docker | 27.3.1 | Disposable PostgreSQL 16 and Redis 7 resources | Verified available |
| PostgreSQL | postgres:16-alpine, digest sha256:aa6cb33cf4c19c765cefcf48d96176258800870a9a1e930ad126200f9d825139 | Migration UP/DOWN/UP, tenant RLS, financial transaction tests | Image downloaded |
| Redis | 7, exact image digest to be recorded after pull | Queue and worker integration tests | Pending |

PostgreSQL reports `16.15`, x86_64 Linux/musl, GCC 15.2.0. Test container: `ikonex-dev05-postgres`, published only on `127.0.0.1:32768`, with disposable trust authentication and no live data or mounted production volumes. Started from the digest recorded above. Stop/remove it after the test session. The test role will be non-superuser for application/RLS assertions; the bootstrap account is reserved for disposable schema setup and teardown.

Initial dependency-enabled baseline: `go test -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -race ./...` passed. INF-02: the same manifest with `go test -race ./internal/internalapi/contracts` and `go vet ./internal/internalapi/contracts` passed. Protected module files have no diff.

Record the resolved transitive module versions, exact image digests, container names and exposed loopback ports, commands, test results, and cleanup verification below as they become available. Never record live credentials, OTPs, full personal records, or service tokens.

## Reproduction and resolved module graph

The current temporary manifest is `/tmp/ikonex-dev05-testing.xunwQ5/testing.mod` (with its companion testing.sum). It was copied from the unchanged application manifest. From `backend/`:

```sh
go get -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod github.com/jackc/pgx/v5/pgxpool@v5.7.2 github.com/hibiken/asynq@v0.25.1
go list -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -m all
go mod verify -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod
```

Checksum verification returned `all modules verified`. The resolved module graph (including dependencies' own test/tool requirements, which are not all compiled into this application) is:

```text
github.com/bsm/ginkgo/v2 v2.12.0
github.com/bsm/gomega v1.27.10
github.com/cespare/xxhash/v2 v2.2.0
github.com/davecgh/go-spew v1.1.1
github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f
github.com/frankban/quicktest v1.14.6
github.com/golang/protobuf v1.5.0
github.com/google/go-cmp v0.6.0
github.com/google/uuid v1.6.0
github.com/hibiken/asynq v0.25.1
github.com/jackc/pgpassfile v1.0.0
github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761
github.com/jackc/pgx/v5 v5.7.2
github.com/jackc/puddle/v2 v2.2.2
github.com/kr/pretty v0.3.1
github.com/kr/text v0.2.0
github.com/pmezard/go-difflib v1.0.0
github.com/redis/go-redis/v9 v9.7.0
github.com/robfig/cron/v3 v3.0.1
github.com/rogpeppe/go-internal v1.9.0
github.com/spf13/cast v1.7.0
github.com/stretchr/objx v0.1.0
github.com/stretchr/testify v1.8.1
go.uber.org/goleak v1.3.0
golang.org/x/crypto v0.31.0
golang.org/x/mod v0.17.0
golang.org/x/net v0.21.0
golang.org/x/sync v0.10.0
golang.org/x/sys v0.28.0
golang.org/x/term v0.27.0
golang.org/x/text v0.21.0
golang.org/x/time v0.8.0
golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d
golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
google.golang.org/protobuf v1.35.2
gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
gopkg.in/yaml.v3 v3.0.1
```

## Migration verification

```sh
DEV05_TEST_DATABASE_URL='postgres://postgres@127.0.0.1:32768/postgres?sslmode=disable' go test -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -v -race ./internal/billing -run TestMigration
go vet -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod ./internal/billing/...
```

Both migration tests passed (10.530 seconds including race instrumentation). Each test created its own randomly named synthetic database and non-superuser/NOBYPASSRLS role and removed both through test cleanup. UP/DOWN/UP, baseline schema/data preservation, RLS, cross-school reference rejection, missing-note checks and nonnegative balances were exercised. PostgreSQL/Redis versions remain separate from Go dependency versions. One initial invocation failed before testing because an alternate Bash login shell selected an inaccessible `/home/bocal/go/pkg/mod`; rerunning in the normal project shell succeeded without changing dependency versions or protected files.

## Integration responsibility

BILL-01 passed with the temporary manifest: exact-money parsing/JSON, overflow, balance conservation, rejection of rounding/NaN, and PostgreSQL NUMERIC(12,2) round trips including ±9999999999.99. `go test -v -race ./internal/billing -run 'TestMoney|TestBalances'` used the local test database and passed in 3.501 seconds; billing vet passed. Per-term credit transfer will use separate `credit_applied`/`credit_transferred` counters, preserving actual `paid_amount` and preventing aggregate double-counting. The owned migration cycle is repeated after adding these columns.

NOTIF-04/05 verification: `DEV05_TEST_DATABASE_URL='postgres://postgres@127.0.0.1:32768/postgres?sslmode=disable' go test -modfile=/tmp/ikonex-dev05-testing.xunwQ5/testing.mod -v -race ./internal/notifications` passed (9.682 seconds), including real PostgreSQL transaction rollback, recipient/tenant isolation, HTTP pagination, and read timestamp idempotency. `go vet` with the same manifest passed. Test databases and roles were cleaned by the harness; protected files remained unchanged.

Evans must add the recorded module requirements (`github.com/jackc/pgx/v5 v5.7.2` and `github.com/hibiken/asynq v0.25.1`) to the protected application manifest during integration. A successful test run with a temporary manifest is not evidence that the unchanged dependency-free application manifest can resolve newly imported pgx or Asynq packages. Report the exact manifest used with each result.

## Completed Phase Verification Summary

All test suites across all packages passed with `-count=1 -race`:

- `backend/internal/notifications`: PASS (4.981s) - normalization, sandbox redaction, concurrency, transaction rollback, HTTP pagination & read idempotency.
- `backend/internal/jobs`: PASS (2.515s) - task creation/validation, retry policies, purger adapter, and Redis end-to-end delivery (`ikonex-dev05-redis`, `redis:7-alpine`, digest `sha256:7b1753b4c5ad8458906c272a8132b8c15f5557b000379eb115fb239706a384b7`, `127.0.0.1:6379`).
- `backend/internal/whatsapp`: PASS (3.676s) - phone masking, 4-state session machine, OTP cooldown, 24h sender rate block without account lockout, 6h expiry, school isolation.
- `backend/internal/internalapi`: PASS (3.709s) - HMAC-SHA256 UserContext, constant-time signature check, parent ownership guard, marks minimal-facts guard for WhatsApp, confirmed STK push, AI audit logging, in-app chat proxy.
- `backend/internal/internalapi/contracts`: PASS (1.030s) - identity, response envelopes, pagination, error mapping.
- `backend/internal/billing`: PASS (19.362s) - migration UP/DOWN/UP on PostgreSQL 16 (`ikonex-dev05-postgres`, `127.0.0.1:32768`), RLS tenant isolation, fee structure validation & loading, 20-goroutine concurrent term charging, credit carry-forward, mid-year fee changes, term start-date sync, Daraja STK push/query, C2B auto-matching, unmatched queue manual resolution.

Protected files (`backend/go.mod`, `backend/go.sum`, etc.) remain 100% untouched with zero diff against `origin/main`.
