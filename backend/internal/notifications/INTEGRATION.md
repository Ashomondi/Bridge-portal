# Notifications integration contract

The server owner mounts `notifications.RegisterRoutes(mux, service, identity)`. `identity` is a `contracts.Identity` supplied by trusted authentication middleware; caller headers must never be used directly as identity. `Service.DB` accepts the application's `WithTenantTx(ctx, schoolID, func(pgx.Tx) error) error` through `contracts.TenantDB`. No route is mounted in the shared server file by this change.

`GET /api/v1/notifications?page=1&per_page=25` lists only the authenticated user's notices in the active school. Page size is limited to 100. Success has `status`, a `data` array, and `meta` with `page`, `per_page`, and `total`. Each notice contains `id`, `category`, `title`, `body`, `metadata`, `is_read`, nullable `read_at`, and UTC `created_at`.

`PATCH /api/v1/notifications/{id}/read` accepts no required body and returns `{"status":"success","data":{"id":"...","is_read":true}}`. Repetition preserves the first `read_at`. Unknown notices, another user's notices, and another school's notices return 404 `NOTIFICATION_NOT_FOUND`; missing identity returns 401 `UNAUTHENTICATED`. Invalid pagination returns 422 `INVALID_PAGINATION`. Responses use the master plan's standard error envelope and `Cache-Control: no-store`.

Billing and academics inject `InAppNotifier`:

```go
CreateInAppNotification(ctx context.Context, tx pgx.Tx, schoolID string, input notifications.CreateNotification) error
```

`CreateNotification` contains `RecipientUserID`, `Category`, `Title`, `Body`, and optional `Metadata`. Call it inside the caller's tenant transaction. It validates an active recipient membership in that school and participates in the caller's commit/rollback. It does not enqueue SMS or create its own transaction. Callers must deduplicate recipients for a domain event and avoid secrets in notice content/metadata.

Auth, onboarding, WhatsApp and workers inject `SMSDispatcher.DispatchSMS(context.Context, SMSMessage) error`. A nil error means accepted by the dispatcher/provider, not proven handset delivery. The supported purposes are `app_login`, `whatsapp_session`, `phone_change`, and `phone_change_notice`. `NewSMSDispatcher` receives `SMSConfig` mapped by the application from Africa's Talking configuration. Username `sandbox` selects local redacted logging. `SMSDeliveryError.Retryable` is consumed by the later queue worker; ambiguous network failures cannot guarantee exactly-once SMS delivery.

Verification uses the temporary dependencies recorded in `../billing/TESTING_RESOURCES.md`. Actual PostgreSQL tests cover rollback, tenant/user ownership, paging, and read idempotency. Local HTTP fixtures cover SMS provider behavior; no live SMS is claimed.
