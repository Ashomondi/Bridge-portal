# CBC Academics (`internal/academics`)

Release 1 CBC Academics & Classroom Operations, owned by Ema (@ejao-000). The
package owns teacher assignment, timetables/calendar, versioned grading scales,
daily roll-call attendance, scoped rosters, anonymised class means, marks entry
and reassignment, the hard-gated results release/reopen/re-release state
machine, report-comment drafting, and the non-ranking CBC performance graph.

## Wiring (one line in `cmd/server/main.go`, owned by Evans)

```go
academics.RegisterRoutes(mux, academics.Deps{
    Store:          academics.NewSQLStore(db), // *sql.DB opened by Evans
    Notifier:       notificationService,       // implements academics.InAppNotifier
    FeeSyncer:      billingService,            // implements academics.TermDateFeeSyncer
    CommentDrafter: academics.NewHTTPCommentDrafter(aiBaseURL, internalToken),
    AuthMiddleware: middleware.RequireAuth,
    IdentityExtractor: func(r *http.Request) (academics.Identity, bool) {
        // Map Evans's request context/JWT claims to academics.Identity.
        return academics.Identity{...}, true
    },
    Config: academics.Config{
        AllowPrincipalDeanOverride: cfg.AllowPrincipalDeanOverride, // O16
        InternalServiceToken:       cfg.InternalServiceToken,
        AIServiceBaseURL:           cfg.AIServiceBaseURL,
    },
})
```

`IdentityExtractor` is optional: the core middleware can instead put an identity
on the request context with `academics.WithIdentity(ctx, identity)`.

## Consumer-boundary interfaces

| Interface | Implemented by | Method |
|---|---|---|
| `InAppNotifier` | `internal/notifications` (Ian) | `CreateInAppNotification(ctx, Notification)` |
| `TermDateFeeSyncer` | `internal/billing` (Ian) | `OnTermStartDateChanged(ctx, schoolID, termID, newStart, actorUserID, note)` |
| `CommentDrafter` | HTTP proxy to `ai_service` (Sos) | `DraftReportComment(ctx, DraftContext)` |
| `Store` | `SQLStore` here; `MemoryStore` in tests | tenant-scoped persistence |

Ian's `internalapi` can call these read methods on `*academics.Service`:

```go
GetStudentAttendanceSummary(ctx, schoolID, studentID)
GetStudentMarksSummary(ctx, schoolID, studentID, termID)
```

## Persistence

`SQLStore` uses only `database/sql`; the concrete driver (`pgx` stdlib or
`lib/pq`) is registered by `main.go`, so this package adds no `go.mod`
dependency. Every `WithTx` call runs `SELECT set_config('app.school_id', $1,
true)` before the queries, matching the forced-RLS policy on all academic
tables (AC-27).

`MemoryStore` is the deterministic store used by the tests and local demos.

## Migration

`migrations/000003_r1_academics.up.sql` creates the ten Release 1 academic
tables, their constraints/indexes, and `ENABLE`/`FORCE ROW LEVEL SECURITY` with
a `tenant_isolation` policy on each. `.down.sql` drops them in reverse order.

## Tests

```bash
cd backend && go test -v ./internal/academics/...
```

The suite covers AC-14 (scoped roll-call + absence notifications), AC-15/16/17
(marks scope, hard gate, reassignment), AC-18 (snapshot isolation, mandatory
re-release note, targeted notification), AC-19 (scale versioning, anonymised
stats), AC-22 (term start-date fee sync), roster privacy, the AI draft proxy,
and route/error-envelope wiring.

## Explicitly out of scope

`POST /api/v1/syllabus/coverage` returns `501 Not Implemented` (Release 2).
Biometric attendance is Release 2; `attendance_daily.source` already exists so
it can write into the same daily table. Nurse, library, and clubs remain other
developers' Release 2 modules.
