# Ikonex Academy – Release 1 Task Assignment
## Developer 5: Backend Engineer C (**Ian** — @frihk)

- **Developer Name:** Ian (@frihk)
- **GitHub Username:** `@frihk`
- **Role:** Backend Engineer C (Go — Fees & Student Ledgers, Safaricom Daraja M-Pesa STK/C2B, Per-School WhatsApp Bot & 6h OTP Sessions, Notifications, Asynq Workers & Internal AI Bridge)
- **Primary Directory Ownership:** `backend/internal/billing/`, `backend/internal/whatsapp/`, `backend/internal/notifications/`, `backend/internal/jobs/`, `backend/internal/internalapi/`, `backend/cmd/worker/`, and `migrations/000004_r1_billing_whatsapp_notifications.*`
- **Release Scope:** Release 1 (Team-Loaded Fee Structures, Automated Term Charging, Ledger Credit Handling, Mid-Year Fee Changes, Daraja Sandbox STK Push & Direct Paybill Matching, Per-School WhatsApp Webhook & 6-Hour OTP Sessions, In-App & Transactional SMS Notifications, Asynq Background Jobs, and Go-to-AI Internal API)

---

## 1. Role and Objective

As **Backend Engineer C (Ian — @frihk)**, you own the **Financial Ledger & M-Pesa Billing Engine**, the **External Communications & Per-School WhatsApp Channel**, the **Asynq Background Worker (`cmd/worker`)**, and the **Internal AI API Bridge (`internal/internalapi`)**.

Your responsibilities enforce several of Ikonex Academy's most critical security and financial invariants (`Docs/Ikonex.md` Section 5 & Section 6.6–6.9):
1. **No Bursar in Release 1 — Team-Loaded Fee Structures (`C46–C49`):** Fee structures are loaded as database records by the Ikonex team (`fee_periods`, `fee_structures`, `fee_structure_items`), charged automatically once at the start of the year divided into terms (`AC-20`), with due dates tied directly to each term's `start_date` (`C61, AC-22`).
2. **Mid-Year Fee Changes Require a Note (`C45, AC-21`):** Every mid-year fee adjustment requires a mandatory explanation note (`fee_change_notes`), is audit-logged, and sends an in-app notification **only** to guardians of the affected students.
3. **M-Pesa STK Push, Any-Amount Credit & Direct Paybill Matching (`C23, C24, C50, AC-23–AC-26`):**
   - Any positive payment amount is accepted; any overpayment beyond `current_balance` is held as a **credit** on the student's ledger (`AC-24`) and automatically applied to the next term charge (`A22`).
   - Direct Paybill (C2B) payments whose account number matches a valid `admission_number` within that school are auto-matched (`AC-25`); invalid/missing references are stored in `unmatched_payments` for the Ikonex team to resolve.
   - Chat-initiated payments must be blocked for **24 hours** after a guardian changes their registered phone number (`C27, AC-26`).
4. **Per-School WhatsApp Bot & 6-Hour OTP Chat Sessions (`C12–C20, AC-08–AC-12`):**
   - Each school has its own WhatsApp Business number mapped to `school_id` (`school_whatsapp_bots`) so admission numbers never collide across schools (`AC-12`).
   - Any WhatsApp number can start a chat by entering a student's `admission_number` -> the bot displays **masked guardian numbers** (`07•• ••• 482`) -> sends the SMS OTP **only** to the chosen registered number -> establishes a **6-hour verified session** (`AC-08, AC-09, AC-11`).
   - Rate limits: 1 OTP/minute (`AC-10`, plus `A3` per admission number); `> 5 requests in 30 minutes` blocks the **requesting WhatsApp number** for 24 hours without blocking the guardian's account (`AC-10`).
5. **AI Service Boundary Enforcement (`C21, AC-28`):** Your `internalapi` package is the **only** gateway through which Sos's Python AI service can read school/student facts or trigger a confirmed STK push, authenticated via `X-Internal-Service-Token` + short-lived HMAC-signed `X-User-Context`.

---

## 2. Assigned Features

You own **3 cohesive, end-to-end backend features**:

| Feature ID | Feature Name | Owned Go Packages | Why These Belong Together |
|---|---|---|---|
| **BE-C1** | **Team-Loaded Fee Structures, Automated Term Charging, Student Ledgers & Mid-Year Fee Changes** | `internal/billing` (`structures.go`, `ledgers.go`, `fee_changes.go`) | Manages the core accounting ledger, generic billing periods (`fee_periods`), term due-date synchronization, credit carry-forward, and mid-year fee adjustment notes. |
| **BE-C2** | **Safaricom Daraja M-Pesa Integration (STK Push, Direct Paybill C2B Matching & Unmatched Resolution)** | `internal/billing` (`mpesa_stk.go`, `mpesa_c2b.go`, `unmatched.go`) | Connects external M-Pesa money flows directly to student ledgers, enforcing the 24h phone-change payment lock, overpayment credit rules, and unmatched C2B queue. |
| **BE-C3** | **Per-School WhatsApp Bot, 6h OTP Sessions, Notifications (In-App + SMS), Asynq Workers & Internal AI API** | `internal/whatsapp`, `internal/notifications`, `internal/jobs`, `internal/internalapi`, `cmd/worker` | Unifies all external messaging channels (WhatsApp Cloud API + Africa's Talking SMS), background job execution, and the internal bridge to Sos's AI service. |

### Explicitly Out of Scope for Release 1 (Do Not Implement Yet)
- **Bursar Payables, Supplier Vouchers & Signatory Workflows:** Release 2 (`C48`).
- **Payment Plans, Parent Pledges & AI Escalating Default Reminders (`POST /api/v1/payment-plans/{id}/reminder-check`):** Release 2 (`C6–C8`).
- **AI Fuzzy Reconciliation for Unmatched Payments:** Release 2 (in Release 1, the Ikonex team resolves unmatched payments manually via `POST /api/v1/payments/{id}/match`).
- **SMS Event Alerts (Absences/Sickbay):** Release 2 (`C36, C58` — Release 1 event notifications are strictly **in-app only**; SMS is used only for login/WhatsApp/phone-change OTPs and the old-number security notice).

---

## 3. Detailed Feature Specifications

### Feature BE-C1: Fee Structures, Automated Term Charging, Student Ledgers & Mid-Year Fee Changes
1. **Generic Billing Periods & Team-Loaded Fee Structures (`C46–C49`, Section 9 Proposed):**
   - Build `fee_periods` on generic billing periods (`period_type = 'term'` in Release 1, linked to `term_id` and `due_date = terms.start_date`) so Release 2 custom structures need no migration.
   - `POST /api/v1/billing/fee-structures`: Callable by the Ikonex team (`ikonex_admin`). Loads the fee structure and itemized lines (`fee_structure_items`) for a `(school_id, academic_year_id, grade_level_id, term_id)`.
2. **Automated Start-of-Year / Term Charging (`C45, C46, C61, AC-20`):**
   - When a fee structure is loaded for an active year/term (or when a new student is admitted via Evans's `StudentChargeHook`), automatically generate per-term `fee_charges` and double-entry `ledger_entries` (`entry_type = 'charge'`) for every enrolled student in that grade level, divided by term (`AC-20`).
   - Each term charge's `due_date` equals `terms.start_date` (`C61`).
   - If a student has an existing `credit_balance > 0` on their `fee_ledgers` from a prior overpayment, automatically apply that credit against the new term's charge (`A22`).
3. **Mid-Year Fee Changes & Term Start-Date Shifts (`C45, C61, AC-21, AC-22`):**
   - `POST /api/v1/fee-changes`: Callable by the Ikonex team on the school's written request (`A15`).
   - **Mandatory Note Rule (`AC-21`):** Request must include a non-empty `note` explaining the reason for the fee change. If `note` is empty, reject with `422 Unprocessable Entity` (`FEE_CHANGE_NOTE_REQUIRED`).
   - Updates the affected students' `fee_charges` / `fee_ledgers`, writes a row to `fee_change_notes` and `audit_log` (`AC-29`), and sends an in-app notification **only** to guardians of the affected students (`AC-21`).
   - Expose `OnTermStartDateChanged(ctx, tx, schoolID, termID, newStartDate, actorUserID, note)` for Ema's `PUT /api/v1/calendar` handler: updates `fee_periods.due_date` and `fee_charges.due_date` to `newStartDate`, records a `fee_change_notes` entry, and notifies affected parents (`AC-22`).
4. **Parent & Team Fee Balance / Ledger Read (`C10, AC-24`):**
   - `GET /api/v1/students/{id}/fee-balance` and `GET /api/v1/students/{id}/ledger`:
   - Accessible to **any guardian linked to the student** in `student_guardians` (`C10`: all linked guardians see fees) or the Ikonex team.
   - Returns `opening_balance`, `billed_amount`, `paid_amount`, `current_balance` (`>= 0`), `credit_balance` (`>= 0`, representing excess paid over total billed, `AC-24`), `due_date`, item breakdown, and any `fee_change_notes`.

### Feature BE-C2: Safaricom Daraja M-Pesa Integration (STK Push, C2B Matching & Unmatched Queue)
1. **STK Push Initiation (`POST /api/v1/payments/stk-push` — `C23, C24, C27, AC-23, AC-24, AC-26`):**
   - Callable by any linked guardian of `student_id` (`A2`) from the Parent Portal, or by Sos's AI service via `POST /api/v1/internal/payments/stk-push` after explicit parent confirmation.
   - **Validation & Business Rules:**
     1. **24-Hour Phone-Change Security Freeze (`C27, AC-26`):** If the request comes from a chat session (`source == "chat"`) and the guardian's `users.chat_payment_paused_until > NOW()`, reject immediately with `403 Forbidden` (`code: "CHAT_PAYMENT_PAUSED_PHONE_CHANGE"`).
     2. **Explicit Confirmation Check (`AC-23`):** If called via the internal AI endpoint, `parent_confirmed` must be `true`; otherwise reject with `422 Unprocessable Entity`.
     3. **Any Positive Amount Accepted (`C24, AC-23`):** Accept any `amount > 0`. Never reject an amount for being less than or greater than `current_balance`.
     4. **Account Reference:** Set Daraja `AccountReference` to the student's `admission_number` so callback matching is deterministic.
     5. Enqueue an Asynq background job (`jobs:mpesa:stk_poll`) to poll Daraja if the webhook callback does not arrive within 45 seconds.
2. **Daraja STK Callback & Direct Paybill C2B Confirmation (`C24, C50, AC-24, AC-25`):**
   - `POST /api/v1/billing/mpesa/callback` (STK Push result) and `POST /api/v1/billing/mpesa/c2b-confirmation` (Direct Paybill payment):
   - Resolve `school_id` from the receiving Paybill/BusinessShortCode configured in `school_mpesa_configs`.
   - Look up `students` in that `school_id` where `UPPER(TRIM(admission_number)) = UPPER(TRIM(BillRefNumber))`.
   - **If Valid Student Match Found (`AC-25`):**
     - Insert row in `payments` (`matched_status = 'matched'`, `student_id = student.id`).
     - Insert row in `ledger_entries` (`entry_type = 'payment'`).
     - Update `fee_ledgers`: increase `paid_amount`. If `paid_amount > opening_balance + billed_amount`, set `current_balance = 0.00` and store the excess in `credit_balance = paid_amount - (opening_balance + billed_amount)` (`C24, AC-24`).
     - Create an in-app payment receipt notification for the student's guardians.
   - **If Admission Number Missing or Not Found in School (`C50, AC-25`):**
     - Insert row in `payments` (`matched_status = 'unmatched'`, `student_id = NULL`) and insert row in `unmatched_payments` with payer phone, payer name, raw `BillRefNumber`, amount, and M-Pesa receipt number.
3. **Unmatched Payment Resolution by Ikonex Team (`C50, AC-25`):**
   - `GET /api/v1/payments/unmatched`: Lists all unresolved rows in `unmatched_payments` for the school.
   - `POST /api/v1/payments/{id}/match`: Callable by Ikonex team in Release 1. Assigns the payment to a valid `student_id`, credits the student's `fee_ledgers` (including `credit_balance` if overpaid), marks `unmatched_payments.resolved_at = NOW()`, and writes to `audit_log`.

### Feature BE-C3: Per-School WhatsApp Bot, 6h OTP Sessions, Notifications, Asynq Workers & Internal AI API
1. **Per-School WhatsApp Bot & 6-Hour OTP Verification State Machine (`C12–C20, AC-08–AC-12`):**
   - Create package `backend/internal/whatsapp/`.
   - **Tenant Routing (`C16, AC-12`):** `POST /api/v1/whatsapp/webhook` reads the receiving WhatsApp Business phone number (`metadata.display_phone_number` / `phone_number_id`) and resolves `school_id` from `school_whatsapp_bots`. Every subsequent student lookup is strictly scoped to that `school_id` via `WithTenantTx`—a bot for School A can never match an admission number belonging to School B (`AC-12`).
   - **Conversational Identity & OTP Flow (also exposed via REST `POST /api/v1/chat/sessions` & `POST /api/v1/chat/sessions/{id}/otp` for testing):**
     - **State 1 (`AWAITING_ADMISSION_NUMBER`):** Any WhatsApp number (`sender_wa_number`) can start a chat (`AC-08`). Bot prompts for the student's admission number.
     - **State 2 (`AWAITING_GUARDIAN_SELECTION`):** Sender sends an admission number.
       - If no student in `school_id` matches that admission number (`C14`), return a generic message (*"We could not verify that admission number. Please check and try again or contact the school office."*)—send **no data and no OTP**.
       - If matched, fetch all linked guardians in `student_guardians`. If multiple guardians exist (`C15, AC-09`), reply with numbered **masked phone numbers** (e.g., `1. 07•• ••• 482`, `2. 07•• ••• 910`) and ask the sender to pick their registered number. (If only 1 guardian exists, display the masked number and prompt to confirm sending OTP).
     - **State 3 (`AWAITING_OTP`):**
       - Enforce OTP Rate Limits (`C18–C20, A3, AC-10`):
         1. If `sender_wa_number` is currently blocked in `sender_rate_blocks` (`blocked_until > NOW()`), refuse immediately.
         2. If an OTP was requested within the last **60 seconds** by this `sender_wa_number` OR for this `(school_id, admission_number)` (`A3`), refuse the second request (`AC-10`).
         3. Count OTP requests from `sender_wa_number` in the last **30 minutes**. If this is the **6th request** (`> 5 in 30 minutes`), insert a 24-hour block in `sender_rate_blocks` for `sender_wa_number` (`blocked_until = NOW() + INTERVAL '24 hours'`) and refuse. **Crucial (`C20, AC-10`):** Never block the guardian's user account or registered phone number—only block the requesting `sender_wa_number`.
       - Send the 6-digit SMS OTP **strictly to the chosen guardian's registered phone number** in `users.phone_number` (`C12, C13, AC-08, AC-09`), never to `sender_wa_number`.
     - **State 4 (`VERIFIED` — 6-Hour Session `C13, AC-11`):**
       - When the sender replies with the valid OTP, mark `chat_sessions.status = 'verified'` and `expires_at = NOW() + INTERVAL '6 hours'`.
       - For any message arriving when `NOW() < expires_at`, sign `X-User-Context` (`channel = "whatsapp"`) and forward the message to Sos's AI service (`POST /api/v1/ai/parent-query`), then send the reply back via WhatsApp.
       - Once `NOW() >= expires_at` (`AC-11`), require a fresh OTP before processing any query.
2. **In-App Notifications & Transactional SMS (`internal/notifications` — `C58`):**
   - `GET /api/v1/notifications` & `PATCH /api/v1/notifications/{id}/read`: Serves in-app notifications (`in_app_notifications`) to the logged-in user.
   - `SMSDispatcher`: Sends transactional SMS via Africa's Talking API (or deterministic local sandbox logger when `AFRICASTALKING_USERNAME=sandbox`) strictly for: (a) App login OTPs, (b) WhatsApp session OTPs, and (c) Old-number security alerts on phone change (`C12, C27, C58`).
3. **Asynq Background Worker (`internal/jobs`, `cmd/worker/main.go`):**
   - Implements Redis/Asynq queues for:
     - `sms:transactional`: Async SMS delivery with retries.
     - `mpesa:stk_poll`: Polls Daraja STK push query API if callback is delayed.
     - `onboarding:purge_unverified`: Periodic task calling Evans's cleanup logic to delete rejected/unverified `registration_submissions` after the configured retention days (`AC-07`).
4. **Internal AI API (`internal/internalapi` — `C21, AC-28`):**
   - Validates `X-Internal-Service-Token` and verifies the HMAC-SHA256 signature and `exp` of `X-User-Context`.
   - Sets tenant RLS (`WithTenantTx`) from `X-User-Context.school_id` and enforces guardian-child ownership (`student_guardians`) on every student lookup (`Docs/API_CONTRACTS.md`: *"Parent-facing endpoints additionally scope to children this parent is a guardian of"*).
   - Implements all 7 internal endpoints called by Sos (`GET /api/v1/internal/school-profile`, `GET /api/v1/internal/parents/{user_id}/children`, `GET /api/v1/internal/students/{id}/fee-balance`, `GET /api/v1/internal/students/{id}/attendance-summary`, `GET /api/v1/internal/students/{id}/marks-summary`, `POST /api/v1/internal/payments/stk-push`, `POST /api/v1/internal/ai/audit-log`) and the frontend chat proxy `POST /api/v1/chat/message`.

---

## 4. Technical Implementation

### 4.1 Database Migration Owned by Ian (`migrations/000004_r1_billing_whatsapp_notifications.up.sql`)
Create `000004_r1_billing_whatsapp_notifications.up.sql` and `.down.sql` with all Release 1 Billing, WhatsApp, Notification, and AI Audit tables and RLS policies:
- `school_mpesa_configs`: `school_id (PK/FK)`, `shortcode VARCHAR(30) UNIQUE NOT NULL`, `environment VARCHAR(20) DEFAULT 'sandbox'`, `passkey_encrypted TEXT`, `test_payment_verified BOOLEAN DEFAULT FALSE`, `updated_at`.
- `fee_periods`: `id`, `school_id`, `academic_year_id`, `term_id`, `period_type VARCHAR(30) NOT NULL DEFAULT 'term'`, `name VARCHAR(100) NOT NULL`, `due_date DATE NOT NULL`, `created_at`, `UNIQUE(school_id, academic_year_id, term_id)`.
- `fee_charges`: `id`, `school_id`, `student_id`, `fee_period_id`, `fee_structure_id`, `amount NUMERIC(12,2) NOT NULL`, `due_date DATE NOT NULL`, `created_at`, `UNIQUE(school_id, student_id, fee_period_id)`.
- `fee_change_notes`: `id`, `school_id`, `term_id`, `fee_period_id`, `change_type VARCHAR(50) NOT NULL`, `amount_delta NUMERIC(12,2) DEFAULT 0`, `old_due_date DATE`, `new_due_date DATE`, `note TEXT NOT NULL`, `created_by UUID NOT NULL`, `affected_students_count INT NOT NULL`, `created_at`.
- `ledger_entries`: `id`, `school_id`, `student_id`, `term_id`, `entry_type VARCHAR(30) NOT NULL CHECK (entry_type IN ('charge', 'payment', 'credit_applied', 'adjustment'))`, `amount NUMERIC(12,2) NOT NULL`, `reference VARCHAR(100)`, `description TEXT`, `created_at`.
- Alter `fee_ledgers`: Add `credit_balance NUMERIC(12,2) NOT NULL DEFAULT 0.00`, `due_date DATE`.
- `unmatched_payments`: `id`, `school_id`, `payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE`, `transaction_ref VARCHAR(100) NOT NULL`, `amount NUMERIC(12,2) NOT NULL`, `payer_phone VARCHAR(20)`, `payer_name VARCHAR(255)`, `raw_account_ref VARCHAR(100)`, `status VARCHAR(30) NOT NULL DEFAULT 'unmatched'`, `resolved_student_id UUID REFERENCES students(id)`, `resolved_by UUID REFERENCES users(id)`, `resolution_note TEXT`, `resolved_at TIMESTAMPTZ`, `created_at`.
- `school_whatsapp_bots`: `id`, `school_id`, `phone_number_id VARCHAR(100) UNIQUE NOT NULL`, `display_phone_number VARCHAR(30) UNIQUE NOT NULL`, `is_active BOOLEAN DEFAULT TRUE`, `created_at`.
- `chat_sessions`: `id`, `school_id`, `channel VARCHAR(20) NOT NULL CHECK (channel IN ('whatsapp', 'in_app'))`, `sender_wa_number VARCHAR(30)`, `student_id UUID REFERENCES students(id)`, `guardian_user_id UUID REFERENCES users(id)`, `state VARCHAR(40) NOT NULL DEFAULT 'AWAITING_ADMISSION_NUMBER'`, `verified_at TIMESTAMPTZ`, `expires_at TIMESTAMPTZ`, `lookup_count INT NOT NULL DEFAULT 0`, `payment_init_count INT NOT NULL DEFAULT 0`, `created_at`, `updated_at`.
- `in_app_notifications`: `id`, `school_id`, `recipient_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE`, `category VARCHAR(50) NOT NULL`, `title VARCHAR(255) NOT NULL`, `body TEXT NOT NULL`, `metadata JSONB`, `is_read BOOLEAN NOT NULL DEFAULT FALSE`, `read_at TIMESTAMPTZ`, `created_at`.
- `ai_action_logs`: `id`, `school_id`, `user_id UUID`, `session_id UUID`, `channel VARCHAR(20) NOT NULL`, `feature VARCHAR(50) NOT NULL`, `intent VARCHAR(50)`, `endpoints_read JSONB`, `action_taken VARCHAR(100)`, `metadata JSONB`, `created_at`.
- Apply `ENABLE ROW LEVEL SECURITY`, `FORCE ROW LEVEL SECURITY`, and `tenant_isolation` policies on all tenant tables above.

### 4.2 Owned Go Files & Package Structure
```text
backend/
├── cmd/
│   └── worker/
│       └── main.go                      # Asynq worker binary entry point
└── internal/
    ├── billing/
    │   ├── billing.go                   # Service struct, RegisterRoutes(mux, deps), cross-module hooks
    │   ├── structures.go                # Team fee structure loader & auto start-of-year/term charging
    │   ├── ledgers.go                   # Student balance, credit_balance carry-forward, ledger history
    │   ├── fee_changes.go               # Mid-year fee changes (note required) & term start-date hook
    │   ├── mpesa_stk.go                 # Daraja STK push, 24h phone-change lock check, callback handler
    │   ├── mpesa_c2b.go                 # Direct Paybill C2B auto-matching by school + admission number
    │   ├── unmatched.go                 # Unmatched payments list & team manual match resolution
    │   └── billing_test.go
    ├── whatsapp/
    │   ├── whatsapp.go                  # Webhook handler, school_whatsapp_bots lookup, session state machine
    │   ├── masking.go                   # Phone masking helper ("+254712345678" -> "07•• ••• 678")
    │   └── whatsapp_test.go
    ├── notifications/
    │   ├── notifications.go             # In-app notifications CRUD + Africa's Talking SMS client
    │   └── notifications_test.go
    ├── jobs/
    │   └── jobs.go                      # Asynq task definitions & handlers (SMS, STK poll, purge)
    └── internalapi/
        ├── internalapi.go               # Service token + HMAC X-User-Context middleware & internal routes
        ├── context_signer.go            # Signs X-User-Context for outbound calls to Sos's AI service
        └── internalapi_test.go
```

---

## 5. Exact Deliverables

1. **Automated Term Charging & Credit Carry-Forward (`AC-20`, `AC-24`):** Loading a school's fee structure charges enrolled students once at the start of the year divided into terms (`AC-20`), sets due dates to `terms.start_date`, and holds any overpayment as `credit_balance` on `fee_ledgers` (`AC-24`).
2. **Mid-Year Fee Changes & Term-Date Due-Date Shifts (`AC-21`, `AC-22`):**
   - `POST /api/v1/fee-changes` without a `note` fails with `422`; with a `note`, it updates affected ledgers, writes to `fee_change_notes` and `audit_log`, and notifies **only** affected parents (`AC-21`).
   - Invoking `OnTermStartDateChanged` shifts term fee due dates and sends the fee-change notice (`AC-22`).
3. **STK Push, 24h Phone-Change Lock & Direct Paybill Matching (`AC-23`, `AC-25`, `AC-26`):**
   - STK push accepts any positive amount (`AC-23`) and blocks chat-initiated payments for 24 hours after a guardian phone number change (`AC-26`).
   - Direct Paybill C2B payments with a valid school admission number auto-credit the student's ledger (`AC-25`); invalid/missing references land in `unmatched_payments` and can be resolved via `POST /api/v1/payments/{id}/match`.
4. **Per-School WhatsApp Bot & 6-Hour OTP Sessions (`AC-08` to `AC-12`):**
   - Unregistered WhatsApp numbers can start a chat; multiple guardians are shown as masked numbers (`07•• ••• 482`), and the SMS OTP is sent **only** to the selected registered number (`AC-08`, `AC-09`).
   - Rate limiter refuses a 2nd OTP within 1 minute and blocks the requesting WhatsApp number for 24 hours on the 6th request in 30 minutes without locking the guardian's account (`AC-10`).
   - Verified sessions expire after 6 hours (`AC-11`), and School A's bot cannot resolve School B's students (`AC-12`).
5. **Internal AI API & Action Logging (`AC-28`):** Every call from Sos's AI service is authenticated with `X-Internal-Service-Token` + HMAC-signed `X-User-Context`, scoped to the parent's own children, and logged in `ai_action_logs`.

---

## 6. Acceptance Criteria

- [ ] **AC-BE-C01 (aligns with AC-08, AC-09, AC-11, AC-12):** WhatsApp session state machine masks guardian phones, sends OTP only to the chosen registered number, expires after 6 hours, and isolates admission numbers strictly by `school_whatsapp_bots.school_id`.
- [ ] **AC-BE-C02 (aligns with AC-10):** Requesting a 2nd WhatsApp OTP within 60s returns `429`; the 6th request within 30 minutes blocks `sender_wa_number` for 24h while the guardian can still log in from the web app or their own number.
- [ ] **AC-BE-C03 (aligns with AC-20, AC-21, AC-22):** Term charges are created once per year divided into terms; `POST /api/v1/fee-changes` rejects missing notes (`422`) and notifies only affected guardians when valid.
- [ ] **AC-BE-C04 (aligns with AC-23, AC-24, AC-25, AC-26):** Overpaying a fee balance records `current_balance = 0` and `credit_balance = excess`; valid C2B Paybill references auto-match while invalid references create `unmatched_payments` rows; chat payments within 24h of a phone change return `403 CHAT_PAYMENT_PAUSED_PHONE_CHANGE`.
- [ ] **AC-BE-C05 (aligns with AC-28):** `internalapi` rejects requests with missing/invalid `X-Internal-Service-Token` or expired `X-User-Context`, rejects parent queries for a `student_id` not linked to `X-User-Context.user_id`, and writes to `ai_action_logs`.

---

## 7. Testing Requirements

Write automated Go tests across your packages:
1. **Billing & M-Pesa Tests (`internal/billing/billing_test.go`):**
   - Start-of-year term charging (`AC-20`), overpayment credit calculation (`AC-24`), and credit application to next term (`A22`).
   - Mid-year fee change missing note rejection (`422`) and targeted parent notification (`AC-21`).
   - Term start-date change shifting fee due dates (`AC-22`).
   - Direct Paybill C2B valid admission number auto-match vs. invalid reference landing in `unmatched_payments` (`AC-25`).
   - 24h phone-change chat payment block (`AC-26`).
2. **WhatsApp Session & OTP Rate Limit Tests (`internal/whatsapp/whatsapp_test.go`):**
   - Unknown admission number returns generic reply and sends zero OTPs (`C14`).
   - Multiple guardians return masked numbers (`07•• ••• 482`) and OTP dispatches only to chosen registered number (`AC-08`, `AC-09`).
   - 1-minute cooldown and 6-in-30-minutes 24h sender block leaving guardian account unblocked (`AC-10`).
   - 6-hour session expiry (`AC-11`) and cross-school bot isolation (`AC-12`).
3. **Internal AI API Security Tests (`internal/internalapi/internalapi_test.go`):**
   - HMAC signature verification on `X-User-Context`, guardian-child ownership check (`403` when Parent A queries Parent B's child), and AI audit logging (`AC-28`).

---

## 8. Integration Responsibilities

- **Depends on Developer 3 (Evans — @eojuma – Backend Engineer A):**
  - Uses `database.WithTenantTx`, `database.AuditLogger`, and `middleware` helpers.
  - Exposes `notifications.SMSDispatcher` and `billing.StudentChargeHook` interfaces for Evans's (@eojuma) `auth` and `onboarding` packages.
- **Integrates with Developer 4 (Ema — @ejao-000 – Backend Engineer B):**
  - Exposes `notifications.Service.CreateInAppNotification` and `billing.Service.OnTermStartDateChanged` for Ema's (@ejao-000) `academics` package.
  - Calls Ema's (@ejao-000) `academics.Service.GetStudentAttendanceSummary` and `GetStudentMarksSummary` inside `internalapi` to serve Sos's (@sospeter-57) AI service.
- **Integrates with Developer 1 (Sos — @sospeter-57 – AI Engineer):**
  - Signs `X-User-Context` (`context_signer.go`) and forwards verified WhatsApp and in-app chat messages to Sos's (@sospeter-57) `POST /api/v1/ai/parent-query`, and serves `/api/v1/internal/*` read/STK endpoints to Sos (@sospeter-57).

---

## 9. File Ownership and Conflict Prevention

| Category | Paths | Rule |
|---|---|---|
| **Exclusively Owned by Ian (@frihk, Dev 5)** | `backend/cmd/worker/main.go`, `backend/internal/billing/**`, `backend/internal/whatsapp/**`, `backend/internal/notifications/**`, `backend/internal/jobs/**`, `backend/internal/internalapi/**`, `migrations/000004_r1_billing_whatsapp_notifications.*` | Full read/write ownership. No other developer edits these packages or migration files. |
| **Shared Files (Coordinate via PR)** | `backend/cmd/server/main.go` (owned by Evans — @eojuma; Ian exports `RegisterRoutes` in `billing`, `whatsapp`, `notifications`, `internalapi`), `Docs/API_CONTRACTS.md` | Coordinate route mounting with Evans (@eojuma). |
| **Strictly Do Not Modify** | `backend/internal/auth/**`, `backend/internal/onboarding/**`, `backend/internal/middleware/**` (Evans — @eojuma), `backend/internal/academics/**` (Ema — @ejao-000), `ai_service/**` (Sos — @sospeter-57), `frontend/**` (Ashy — @Ashomondi) | Owned by other developers. |

---

## 10. Definition of Done

- [ ] `migrations/000004_r1_billing_whatsapp_notifications.up.sql` and `.down.sql` apply and roll back cleanly with `FORCE ROW LEVEL SECURITY` on all tenant tables.
- [ ] `billing`, `whatsapp`, `notifications`, `jobs`, `internalapi`, and `cmd/worker` compile and pass `go test -v ./...`.
- [ ] Per-school WhatsApp webhook, masked guardian OTP verification, 6-hour session expiry, and sender-only 24h rate blocking work end-to-end.
- [ ] Daraja sandbox STK push, overpayment ledger credit, C2B auto-matching, and unmatched queue resolution work end-to-end.
- [ ] Internal AI API enforces service token + signed user context and logs every AI action in `ai_action_logs`.
