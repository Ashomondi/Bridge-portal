# Ikonex Academy – Release 1 Task Assignment
## Developer 2: Frontend Engineer (**Ashy** — @Ashomondi)

- **Developer Name:** Ashy (@Ashomondi)
- **GitHub Username:** `@Ashomondi`
- **Role:** Frontend Engineer (Next.js 14 App Router / TypeScript / Tailwind CSS / PWA)
- **Primary Directory Ownership:** `frontend/`
- **Release Scope:** Release 1 (English Only; Phone + SMS OTP Auth, School Setup Wizard, Staff & Parent QR Onboarding & In-Person Verification UI, Dean of Studies Portal, Teacher Portal, and Parent Portal with In-App AI Assistant & M-Pesa STK Push)

---

## 1. Role and Objective

As the **Frontend Engineer (Ashy — @Ashomondi)** for Ikonex Academy Release 1, you own the entire Next.js 14 Progressive Web Application (`frontend/`) and all user-facing experiences for **Parents, Teachers (Subject & Class Teachers), Deans of Studies, Delegated Verifiers, Principals/Deputies, and the Ikonex Operations Team**.

Your core objectives are:
1. Deliver a responsive, mobile-first, low-bandwidth-friendly Progressive Web Application (PWA) that satisfies the **Release 1 Pass Criteria (`Docs/Ikonex.md` C53, C54)**:
   - Onboarding works end-to-end (school setup wizard, staff & parent QR generation/scanning, document upload, in-person verification, and badge delegation).
   - Parents and staff can log in smoothly via Phone Number + SMS OTP and switch between school memberships if they belong to more than one school.
   - Role and badge permissions (`class_teacher`, `delegated_verifier`, `dean_of_studies`, assigned subject teacher) dynamically drive UI visibility and actions without friction.
2. Connect every screen strictly to the agreed Go backend REST API (`/api/v1/...`), handling **loading, empty, validation error, API error, success, and offline/reconnecting states** gracefully.
3. Never implement backend business logic, fee calculations, CBC level derivations, or direct AI service calls on the client—all AI interactions (`parent-query` and `report-comments/draft`) go through the Go backend API.

> **Note on Design Documentation (`design.md`):** As documented in `Docs/Ikonex.md` Section 2, a standalone `design.md` file was not provided in the repository. Section 4.2 of this document defines the canonical **Ikonex Academy Tailwind CSS Design System, Color Tokens, Typography, and Accessibility (WCAG 2.1 AA) Specifications** that you must implement in `frontend/tailwind.config.ts` and `frontend/src/components/ui/`.

---

## 2. Assigned Features

You own **all Release 1 frontend screens, components, state management, and API integrations**, organized into 5 logical frontend feature modules:

| Feature ID | Feature Name | Target Routes (`frontend/src/app/`) | Connected Backend Owner |
|---|---|---|---|
| **FE-F1** | **Design System, PWA Shell, SMS OTP Authentication, Membership Switcher & In-App Notification Center** | `/`, `/login`, shared layout, notification bell drawer | Evans (@eojuma — Auth/Badges) & Ian (@frihk — Notifications) |
| **FE-F2** | **Admin, Principal Setup Wizard, Staff/Parent QR Onboarding & In-Person Verification Flows** | `/admin`, `/admin/wizard`, `/admin/onboarding`, `/register/[token]` | Evans (@eojuma — Onboarding/QR/Verification/Admissions) |
| **FE-F3** | **Dean of Studies Workspace (Teacher Assignments, Timetables, School Calendar & Marks Reassignment)** | `/admin/dean` (or role-gated tab in `/admin` & `/teacher`) | Ema (@ejao-000 — Academics/Calendar/Timetables) |
| **FE-F4** | **Teacher Portal (Daily Roll-Call Attendance, Subject Marks Entry, Hard-Gated Results Release/Reopen, AI Comment Drafter & Class Means)** | `/teacher`, `/teacher/attendance`, `/teacher/marks`, `/teacher/results` | Ema (@ejao-000 — Academics) & Sos (@sospeter-57 — AI Drafter via Go) |
| **FE-F5** | **Parent Portal (Released Results, CBC Performance Graph, Attendance, Fee Ledger, STK Push, In-App AI Chat, 2FA Phone Change & R2 Placeholders)** | `/parent`, `/parent/results`, `/parent/fees`, `/parent/chat`, `/parent/settings` | Ema (@ejao-000 — Academics), Ian (@frihk — Billing/Chat), Evans (@eojuma — Phone 2FA), Sos (@sospeter-57 — AI Chat via Go) |

### Explicitly Out of Scope for Release 1 (Placeholder or Disabled Routes)
Per `Docs/Ikonex.md` Section 3 & Section 7:
- `frontend/src/app/bursar/page.tsx`: **No bursar portal in Release 1** (`C48`). Render a clean "Release 2 Module – In Release 1, fee structures are managed by the Ikonex Setup Team" guard screen.
- `frontend/src/app/nurse/page.tsx` and `frontend/src/app/librarian/page.tsx`: **Release 2 modules**. Keep clean informational placeholder screens so routes do not break.
- In the **Parent Portal (`/parent`)**: **Clubs, Books Borrowed, and Student Blogs/Newsletters** must render clean, intentional **"Coming Soon in Release 2"** placeholder cards (`Docs/Ikonex.md` C4, `REQUIREMENTS.md` Section C).
- **Swahili translations (`A18`):** Release 1 UI is **English only**.

---

## 3. Detailed Feature Specifications

### Feature FE-F1: Design System, PWA Shell, SMS OTP Authentication & Notification Center
- **User Workflows:**
  1. **SMS OTP Login (`/login`):**
     - Step 1: User enters their Kenyan phone number (validates `07XXXXXXXX`, `01XXXXXXXX`, or `+2547XXXXXXXX` and normalizes to E.164 `+254...`). Calls `POST /api/v1/auth/otp/request`.
     - Displays a 60-second countdown timer before allowing "Resend OTP" (`Docs/Ikonex.md` C18: 1 OTP request per minute).
     - Step 2: User enters the 6-digit OTP code. Calls `POST /api/v1/auth/otp/verify`.
     - Step 3: If the user has memberships at multiple schools or multiple roles, displays a **School & Role Selector** modal; otherwise, automatically activates their membership and redirects to their role portal (`/admin`, `/teacher`, `/parent`).
  2. **Badge-Aware Navigation & Header:**
     - Displays the active School Name, School Logo, User Name, Primary Role, and active **Badges** (e.g., `Class Teacher: Grade 8 East`, `Verifier (Expires in 9h)`).
     - Includes an **In-App Notification Bell** (`GET /api/v1/notifications`, `PATCH /api/v1/notifications/{id}/read`) showing unread count and a slide-over drawer for absence alerts, timetable/calendar updates, fee-change notices, and result re-release notices (`Docs/Ikonex.md` C58).
  3. **Offline / Low-Connectivity Banner:**
     - Detects `navigator.onLine` status and shows a persistent top banner when offline, disabling mutating actions (like STK push or results release) until connectivity is restored.

### Feature FE-F2: Principal Setup Wizard, QR Onboarding & In-Person Verification
- **User Workflows:**
  1. **Ikonex Team Provisioning View (`/admin/platform`):**
     - Accessible to `ikonex_admin` synthetic test accounts to create a school + principal (`POST /api/v1/admin/schools`), load the school's term fee structure (`POST /api/v1/billing/fee-structures`), apply mid-year fee changes with a mandatory note (`POST /api/v1/fee-changes`), and resolve unmatched Paybill payments (`GET /api/v1/payments/unmatched`, `POST /api/v1/payments/{id}/match`).
  2. **Principal Setup Wizard (`/admin/wizard`):**
     - If a school's setup wizard is incomplete, logging in as Principal redirects directly to `/admin/wizard` (`AC-01`).
     - Multi-step guided wizard (`Docs/Ikonex.md` 6.1, C38, C57):
       - **Step 1:** School Contact Details (Required: Official Phone, Email, Office Hours, Physical Address – used by AI assistant when refusing out-of-scope queries).
       - **Step 2:** Academic Year & Term Dates (Term 1, Term 2, Term 3 start & end dates).
       - **Step 3:** CBC Grade Levels (Grade 7–9 Junior, Grade 10–12 Senior) & Class Streams (e.g., Grade 7 North, Grade 7 South).
       - **Step 4:** Dormitories (Name, Gender, Capacity).
       - **Step 5:** Initial CBC Grading Scale (`EE`, `ME`, `AE`, `BE` min/max score ranges, versioned per term).
       - Submits to `POST /api/v1/schools/setup-wizard`.
  3. **Staff & Parent QR Code Management (`/admin/onboarding`):**
     - **Staff QR:** Principal/Deputy generates a Staff Invite QR (`POST /api/v1/staff-invites`), configurable validity window (default `08:00`–`18:00`), displays scannable QR + shareable link + "Revoke QR" button.
     - **Parent QR:** Principal, Deputy, or Dean generates a multi-use Parent Registration QR (`POST /api/v1/parent-qr`), valid for **12 hours** (`AC-03`), with live countdown and "Revoke Now" button (`POST /api/v1/parent-qr/{id}/revoke`).
     - **Verifier Delegation (`C28, C29, AC-05`):** While a Parent QR is active, Principal/Deputy/Dean can select staff members from a dropdown and grant them the time-bound `delegated_verifier` badge (`POST /api/v1/verifier-grants`), which automatically expires when the QR expires.
  4. **Public QR Scan & Application Form (`/register/[token]`):**
     - Unauthenticated mobile-friendly page opened when a staff applicant or parent scans an active QR code (`GET /api/v1/registrations/qr-info?token=...`).
     - If expired or revoked (`AC-02, AC-03`), displays a clear "This QR code has expired or been revoked by the school" screen.
     - **Parent Form:** Parent full name, phone number (editable during onboarding per `C25`), national ID, relationship, student admission number (if matching existing student) or new student details, and document upload (ID/birth certificate photo).
     - **Staff Form:** Full name, phone, national ID, TSC/professional reg number, document upload.
     - Submits to `POST /api/v1/registrations`. Shows a **"Pending In-Person Verification"** confirmation screen explaining that no student data is accessible until verified in person by a staff verifier (`AC-04`).
  5. **In-Person Verification Queue (`/admin/verifications` — also visible to any staff holding `delegated_verifier`):**
     - Lists pending submissions (`GET /api/v1/registrations?status=pending`). Flags duplicate admission number matches (`A12`) when both class-teacher and opening-day QR routes are used.
     - Verifier views uploaded document preview, checks the person physically in front of them, records the document type verified, and clicks **"Approve & Delete Uploaded Document"** or **"Reject"** (`POST /api/v1/registrations/{id}/verify`).
     - Displays confirmation that the raw document file has been permanently purged while retaining the audit verification metadata (`AC-06`).

### Feature FE-F3: Dean of Studies Workspace
- **User Workflows:**
  1. **Teacher & Class Assignments:** Assigns teachers to `(class_id, subject)` pairs and grants the `class_teacher` badge scoped to a specific `class_id` (`POST /api/v1/academics/assignments`).
  2. **Marks Entry Reassignment (`C32, AC-17`):** Allows Dean/Principal/Deputy to reassign marks entry for a `(class_id, subject_id)` to a replacement teacher (`POST /api/v1/classes/{id}/marks-assignments`).
  3. **Timetables & School Calendar (`C59, C60, C61, AC-22`):**
     - Create/edit class and teacher timetables (`GET/PUT /api/v1/timetables`).
     - Edit school calendar events and post-wizard **term dates** (`GET/PUT /api/v1/calendar`).
     - **Critical Rule (`C61, AC-22`):** If the Dean changes a term's `start_date`, the UI **must require a mandatory Fee Due-Date Change Note** before submission, warning the Dean that shifting the term start date automatically shifts term fee due dates and sends a fee-change notice to parents.
  4. **Student Admissions & Class-Teacher Initial Load (`C39`):** Form for admitting students (`POST /api/v1/students/admit`) with admission number, class, dormitory, and guardian phone number.

### Feature FE-F4: Teacher Portal (`/teacher`)
- **User Workflows:**
  1. **My Timetable, Assigned Classes & Other-Class Anonymised Means:**
     - Shows the teacher's personal timetable and calendar events.
     - For classes the teacher does **not** teach as class teacher, shows `GET /api/v1/classes/{id}/stats`—displaying **class and subject mean scores/levels only, never individual student names** (`Docs/Ikonex.md` Section 4).
  2. **Daily Class Roll-Call Attendance (`C35, AC-14`):**
     - Visible **only** if the teacher holds the `class_teacher` badge for that class.
     - Displays the class roster (`GET /api/v1/classes/{id}/roster`) for the selected date (defaults to today) with quick toggle buttons (`Present` / `Absent` / `Late`) and one-click **"Submit Daily Roll-Call"** (`POST /api/v1/classes/{id}/attendance`).
  3. **Subject Marks Entry (`C30, AC-15`):**
     - Subject teacher selects their assigned `(Class, Subject)` and current term.
     - Enters numeric scores (`0–100`) per student in a clean tabular grid; shows live preview of the derived CBC level (`EE`, `ME`, `AE`, `BE`) based on the term's grading scale. Saves via `PUT /api/v1/classes/{id}/subjects/{sid}/marks`.
  4. **Hard-Gated Class Results Release, AI Comment Drafter & Reopen/Re-Release (`C30–C34, AC-16, AC-18`):**
     - Class teacher views the **Term Results Readiness Matrix** (`GET /api/v1/classes/{id}/results/status`) showing which subjects have complete marks (`N/N students`) and which are missing marks.
     - **Hard Gate (`AC-16`):** The **"Release Results to Parents"** button is strictly disabled (with a clear tooltip listing incomplete subjects) until every subject has marks for every student in the class.
     - **AI Report Comment Drafter (`C5`):** Next to each student's report row, the class teacher can click **"Draft Comment with AI"** (`POST /api/v1/report-comments/draft`), which populates an editable textarea. The teacher can edit the text before saving (`PUT /api/v1/classes/{id}/report-comments`).
     - **Reopen & Re-Release (`C33, C34, AC-18`):**
       - If results are already `released`, the class teacher can click **"Reopen Results for Correction"** (`POST /api/v1/classes/{id}/results/reopen`). A banner informs the teacher that parents continue to see the previously released results while reopened.
       - After the subject teacher edits the mark, the class teacher clicks **"Re-Release Updated Results"** (`POST /api/v1/classes/{id}/results/re-release`). The modal requires a **mandatory Teacher's Correction Note** for each modified student/subject (with an optional **"Help Draft Note with AI"** button) explaining why the mark changed.

### Feature FE-F5: Parent Portal (`/parent`)
- **User Workflows:**
  1. **Child Switcher & Overview Dashboard:**
     - If a parent has multiple children at the school, a top pill selector switches the active `student_id`.
     - Displays summary cards for Attendance (`Present/Absent days`), Current Fee Balance (`KES`), Latest Released Term Results, and Upcoming School Events.
  2. **Released Results & Re-Release Correction Banner (`AC-18`):**
     - Displays term report card (`EE/ME/AE/BE` per subject + class teacher comment).
     - If a subject mark was updated via a re-release, highlights the updated subject with an **"Updated"** badge and displays the **Teacher's Explanation Note** (`C34`).
  3. **CBC Performance Graph (`C42, C43, AC-19`):**
     - Interactive chart (`GET /api/v1/students/{id}/performance-history`) plotting CBC performance levels (`BE=1, AE=2, ME=3, EE=4` on Y-axis with level labels) per subject across terms from the day the child joined.
     - **Never displays class rank, position, or percentile** (`AC-19`).
  4. **Fee Structure, Student Ledger & M-Pesa STK Push (`C24, C46, AC-24`):**
     - Displays the term fee structure breakdown, billed charges, payments received, **credit balance** (if overpaid), and running balance (`GET /api/v1/students/{id}/fee-balance`, `GET /api/v1/students/{id}/ledger`).
     - Shows any mid-year **Fee Change Notes** (`C45, AC-21`).
     - **"Pay Fees via M-Pesa" Modal:** Parent enters any positive amount (`KES`) and the M-Pesa phone number to charge, confirms the summary, and triggers `POST /api/v1/payments/stk-push`, showing a live polling status indicator until Daraja callback completes.
  5. **In-App Parent AI Assistant (`/parent/chat` or floating drawer — `C22, C23, AC-23`):**
     - Chat interface connected to `POST /api/v1/chat/message` (`channel: "in_app"`).
     - Supports natural-language questions about fees, attendance, and results, and renders a clear confirmation card when `payment_state == "AWAITING_CONFIRMATION"` with quick **"YES, Confirm Payment"** and **"NO, Cancel"** buttons.
  6. **2FA Registered Phone Number Change (`/parent/settings` — `C25, C27, AC-26`):**
     - Two-step form: Parent enters new phone number -> requests OTP to the **new number** (`POST /api/v1/guardians/me/phone-change/request-otp`) -> enters OTP (`POST /api/v1/guardians/me/phone-change/confirm`).
     - Displays a prominent notice that the old number has been notified by SMS and **chat-initiated payments are paused for 24 hours**.
  7. **Release 2 Placeholders (`C4`):**
     - Dedicated section/tabs for **Clubs**, **Library Books Borrowed**, and **Student Blogs & Newsletters** showing clean, well-designed empty-state placeholder cards stating these modules launch in Release 2.

---

## 4. Technical Implementation

### 4.1 Directory & File Structure (`frontend/`)
```text
frontend/
├── public/
│   └── manifest.json
├── src/
│   ├── app/
│   │   ├── layout.tsx                   # Root layout, AuthProvider, ToastProvider, OfflineBanner
│   │   ├── page.tsx                     # Role-based router / portal landing
│   │   ├── globals.css                  # Tailwind directives & custom CSS variables
│   │   ├── (auth)/
│   │   │   └── login/page.tsx           # Phone + SMS OTP login & membership selector
│   │   ├── register/
│   │   │   └── [token]/page.tsx         # Public Staff & Parent QR registration & document upload
│   │   ├── admin/
│   │   │   ├── layout.tsx               # Admin/Principal/Dean/Team sidebar shell
│   │   │   ├── page.tsx                 # Principal/Deputy overview & Setup Wizard redirect
│   │   │   ├── wizard/page.tsx          # 5-step Principal Setup Wizard
│   │   │   ├── onboarding/page.tsx      # Staff & Parent QR generation, revocation & verifier delegation
│   │   │   ├── verifications/page.tsx   # In-person document verification queue (Principal/Dean/Delegates)
│   │   │   ├── dean/page.tsx            # Teacher assignments, timetables, calendar, marks reassignment
│   │   │   └── platform/page.tsx        # Ikonex Team school creation, fee loader, fee changes, unmatched Paybill
│   │   ├── teacher/
│   │   │   ├── layout.tsx               # Teacher navigation shell (badge-aware)
│   │   │   ├── page.tsx                 # Timetable, assigned classes & anonymised other-class means
│   │   │   ├── attendance/page.tsx      # Daily class-teacher roll-call
│   │   │   ├── marks/page.tsx           # Subject teacher marks entry grid
│   │   │   └── results/page.tsx         # Class teacher readiness check, AI comment drafter, release/reopen
│   │   ├── parent/
│   │   │   ├── layout.tsx               # Parent navigation shell & child switcher
│   │   │   ├── page.tsx                 # Parent overview, events & R2 placeholders (Clubs/Books/Blogs)
│   │   │   ├── results/page.tsx         # Released report cards, re-release notes & CBC Performance Graph
│   │   │   ├── fees/page.tsx            # Fee structure, ledger, credit balance & M-Pesa STK Push modal
│   │   │   ├── chat/page.tsx            # In-App AI Parent Assistant & 2-step chat payment confirmation
│   │   │   └── settings/page.tsx        # 2FA Registered Phone Number change & Consent Centre preview
│   │   ├── bursar/page.tsx              # R2 informational guard screen
│   │   ├── librarian/page.tsx           # R2 informational guard screen
│   │   └── nurse/page.tsx               # R2 informational guard screen
│   ├── components/
│   │   ├── ui/                          # Button, Input, Select, Modal, Badge, Card, Skeleton, Alert, Tabs
│   │   ├── layout/                      # AppHeader, NotificationDrawer, OfflineBanner, RoleGuard
│   │   ├── charts/                      # CbcPerformanceChart.tsx (SVG/Canvas CBC level trend chart)
│   │   └── chat/                        # AssistantChatBox.tsx
│   ├── lib/
│   │   ├── api-client.ts                # Typed fetch wrapper with JWT header, school_id header & error parser
│   │   ├── auth-context.tsx             # User session, active membership, badges & permission helpers
│   │   ├── validators.ts                # Kenyan phone normalization (+254), admission number & form validators
│   │   └── types.ts                     # Shared TypeScript interfaces matching API_CONTRACTS.md
│   └── __tests__/                       # Component & integration tests
├── package.json
├── tailwind.config.ts
└── tsconfig.json
```

### 4.2 Design System & Accessibility Specifications
Configure `frontend/tailwind.config.ts` and `globals.css` with these tokens:
- **Brand Palette:**
  - `primary`: Deep Academic Emerald (`#0F5132` / `emerald-800`, hover `#0A3622`) — evokes trust, Kenyan academic heritage, and M-Pesa familiarity.
  - `secondary`: Warm Savanna Gold (`#D97706` / `amber-600`) — used for badges, highlights, and pending verification states.
  - `surface`: Clean Slate (`#F8FAFC` background, `#FFFFFF` cards, `#E2E8F0` borders).
  - **CBC Performance Level Semantic Colors:**
    - `EE` (Exceeding Expectations): `bg-emerald-100 text-emerald-800 border-emerald-300`
    - `ME` (Meeting Expectations): `bg-blue-100 text-blue-800 border-blue-300`
    - `AE` (Approaching Expectations): `bg-amber-100 text-amber-800 border-amber-300`
    - `BE` (Below Expectations): `bg-rose-100 text-rose-800 border-rose-300`
- **Accessibility & Mobile Responsiveness:**
  - Minimum touch target size of `44px × 44px` on all mobile buttons (critical for teachers taking roll-call on phones and parents on low-end Android devices).
  - WCAG 2.1 AA contrast ratios (`>= 4.5:1` for body text), visible focus rings (`focus-visible:ring-2 focus-visible:ring-emerald-600`), and `aria-live="polite"` regions for OTP timers, STK push status updates, and AI chat replies.

### 4.3 Standard API Envelope Consumed by Frontend
Every Go backend endpoint returns:
```json
{
  "status": "success",
  "data": { ... },
  "meta": { "page": 1, "per_page": 20, "total": 42 }
}
```
And on error (`4xx` / `5xx`):
```json
{
  "status": "error",
  "error": {
    "code": "RESULTS_INCOMPLETE_MARKS",
    "message": "Cannot release results: 2 subjects still have missing marks.",
    "details": { "missing_subjects": ["Mathematics", "Kiswahili"] }
  }
}
```
Your `src/lib/api-client.ts` must automatically attach `Authorization: Bearer <jwt>` and `X-School-ID: <active_school_id>` to all authenticated requests and parse `error.code` and `error.message` into user-friendly toast/inline alerts.

---

## 5. Exact Deliverables

1. **SMS OTP Login & Multi-Role Routing (`AC-01`):** A user can request an OTP, verify the 6-digit code, select their school membership, and land on the right portal (`/admin/wizard`, `/admin`, `/teacher`, or `/parent`).
2. **Principal Setup Wizard (`AC-01`):** A newly created principal is routed to `/admin/wizard` and can configure school contact details, academic year/terms, CBC grades 7–12, classes, dormitories, and the initial grading scale.
3. **QR Generation, Public Registration & In-Person Verification (`AC-02` to `AC-06`):**
   - Principal/Deputy/Dean can generate & revoke Staff and 12h Parent QR codes and delegate verification to staff (`AC-05`).
   - Scanning `/register/[token]` allows applicants/parents to submit forms and upload documents; expired/revoked tokens show an immediate refusal state (`AC-02`, `AC-03`).
   - Pending users see no student data (`AC-04`). Verifiers can review and approve submissions in `/admin/verifications`.
4. **Dean of Studies Controls (`AC-17`, `AC-22`):** Dean can assign teachers, reassign marks entry, manage timetables, and update term dates (with a mandatory fee-change note when term start dates change).
5. **Teacher Roll-Call, Marks, Hard-Gated Release & Reopen (`AC-14` to `AC-18`):**
   - Class teachers can submit daily attendance (`AC-14`).
   - Subject teachers can enter marks only for their assigned subjects (`AC-15`).
   - Class teachers cannot click "Release Results" until 100% of subjects have marks (`AC-16`), can draft comments with AI, and can reopen & re-release results with a mandatory teacher's note (`AC-18`).
   - Non-class teachers viewing other classes see only class/subject means, never student names.
6. **Parent Portal, Performance Graph, STK Push, AI Chat & 2FA Phone Change (`AC-19`, `AC-23`, `AC-24`, `AC-26`):**
   - Parents view released results (with re-release notes if any), a multi-term CBC Performance Graph with zero class rank (`AC-19`), fee ledger (including credit balances, `AC-24`), and placeholders for Clubs/Books/Blogs (`C4`).
   - Parents can initiate M-Pesa STK push from the Fee screen or via the two-step confirmed In-App AI Assistant (`AC-23`), and change their registered phone number with 2FA OTP (`AC-26`).

---

## 6. Acceptance Criteria

- [ ] **AC-FE-01:** `npm run build` and `npm run lint` inside `frontend/` complete with zero TypeScript or ESLint errors.
- [ ] **AC-FE-02 (aligns with AC-01, AC-02, AC-03, AC-04, AC-05):** Onboarding flows (Wizard, Staff QR, 12h Parent QR, Verifier Delegation, Public Registration, In-Person Approval) work end-to-end against the Go backend.
- [ ] **AC-FE-03 (aligns with AC-14, AC-15, AC-16, AC-17, AC-18):** Teacher portal enforces badge-based UI controls, disables results release when any subject mark is missing, supports AI comment drafting, and requires a teacher's note on re-release.
- [ ] **AC-FE-04 (aligns with AC-19):** The CBC Performance Graph component (`CbcPerformanceChart.tsx`) renders `EE/ME/AE/BE` progression across terms and never displays class rank or position.
- [ ] **AC-FE-05 (aligns with AC-23, AC-24, AC-26):** Parent fee ledger displays credit balances accurately, in-app chat requires explicit confirmation before STK push, and the 2FA phone-change screen displays the 24-hour chat payment pause warning.
- [ ] **AC-FE-06:** Every page implements explicit Loading (skeleton), Empty, Error (retryable), and Offline states.

---

## 7. Testing Requirements

1. **Unit & Component Tests:**
   - Kenyan phone number validator and normalizer (`0712345678` -> `+254712345678`).
   - Role & Badge guard helpers (`hasBadge(membership, 'class_teacher', classId)`, `hasActiveVerifierBadge(membership)`).
   - Results Release readiness button state (disabled when `missing_subjects.length > 0`, enabled when `0`).
   - CBC Performance Graph rendering (`EE/ME/AE/BE` data points, zero rank labels).
   - Chat payment confirmation card rendering when `payment_state === "AWAITING_CONFIRMATION"`.
2. **End-to-End / Integration Verification:**
   - Full walkthrough of the 3 Release 1 Pass Criteria (`C53`): Onboarding flow, Parent OTP Login, and Badge-scoped actions across synthetic accounts.

---

## 8. Integration Responsibilities

- **With Developer 3 (Evans — @eojuma – Backend Engineer A):** Integrate `/api/v1/auth/*`, `/api/v1/schools/setup-wizard`, `/api/v1/staff-invites`, `/api/v1/parent-qr`, `/api/v1/registrations`, `/api/v1/verifier-grants`, `/api/v1/students/admit`, and `/api/v1/guardians/me/phone-change`.
- **With Developer 4 (Ema — @ejao-000 – Backend Engineer B):** Integrate `/api/v1/academics/assignments`, `/api/v1/timetables`, `/api/v1/calendar`, `/api/v1/classes/{id}/roster`, `/api/v1/classes/{id}/stats`, `/api/v1/classes/{id}/attendance`, `/api/v1/classes/{id}/subjects/{sid}/marks`, `/api/v1/classes/{id}/results/*`, `/api/v1/report-comments/draft`, and `/api/v1/students/{id}/performance-history`.
- **With Developer 5 (Ian — @frihk – Backend Engineer C):** Integrate `/api/v1/students/{id}/fee-balance`, `/api/v1/students/{id}/ledger`, `/api/v1/payments/stk-push`, `/api/v1/fee-changes`, `/api/v1/payments/unmatched`, `/api/v1/notifications`, and `/api/v1/chat/message`.
- **With Developer 1 (Sos — @sospeter-57 – AI Engineer):** Render the `reply`, `intent`, and `payment_state` fields from `/api/v1/chat/message` and `draft_comment` from `/api/v1/report-comments/draft`.

---

## 9. File Ownership and Conflict Prevention

| Category | Paths | Rule |
|---|---|---|
| **Exclusively Owned by Ashy (@Ashomondi, Dev 2)** | `frontend/**` (all files in `frontend/src/`, `frontend/public/`, `frontend/package.json`, `frontend/tailwind.config.ts`, `frontend/tsconfig.json`, `frontend/next.config.mjs`) | Full read/write ownership. No backend or AI developer may modify files in `frontend/`. |
| **Shared Read-Only / Coordinate Before Edit** | `Docs/API_CONTRACTS.md`, `.env.example`, `docker-compose.yml` | Coordinate via PR review before changing environment variable names. |
| **Strictly Do Not Modify** | `backend/**`, `ai_service/**`, `migrations/**` | Owned by Sos (@sospeter-57), Evans (@eojuma), Ema (@ejao-000), and Ian (@frihk). |

---

## 10. Definition of Done

- [ ] All Release 1 screens (`/login`, `/register/[token]`, `/admin/*`, `/teacher/*`, `/parent/*`) are implemented and connected to real Go backend APIs (no hardcoded mock data left in production components).
- [ ] Clubs, Books Borrowed, and Blogs show clear Release 2 placeholders on `/parent`, and `/bursar`, `/nurse`, `/librarian` show clean Release 2 module guards.
- [ ] All loading, empty, error, success, and offline states are implemented.
- [ ] Responsive layout works cleanly at `360px` mobile width and `1280px` desktop width with WCAG AA contrast.
- [ ] `npm run build` and `npm run lint` pass with zero errors.

