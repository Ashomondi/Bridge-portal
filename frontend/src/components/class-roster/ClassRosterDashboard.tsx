"use client";

import { useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import {
  Award,
  BarChart3,
  Bell,
  CalendarCheck,
  CalendarDays,
  CalendarX,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  ClipboardCheck,
  ClipboardList,
  Clock3,
  ExternalLink,
  GraduationCap,
  Home,
  Info,
  LayoutDashboard,
  Lock,
  LogOut,
  Menu,
  MessageSquare,
  MoreHorizontal,
  Search,
  ShieldCheck,
  UserPlus,
  UsersRound,
  X,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/utils/cn";
import {
  AVATAR_TONE_CLASSES,
  ROSTER_PAGE_SIZE,
  formatCount,
  formatPercent,
  getPrimaryGuardian,
  getStudentInitials,
  type ClassRosterData,
  type GuardianStatus,
  type RosterStudent,
} from "./class-roster-data";

type NavItem = {
  label: string;
  icon: LucideIcon;
  href: string;
  active?: boolean;
  badge?: string;
  dot?: boolean;
};

const NAV_ITEMS: NavItem[] = [
  { label: "Home", icon: Home, href: "#" },
  { label: "My timetable", icon: CalendarDays, href: "/teacher" },
  {
    label: "My class",
    icon: UsersRound,
    href: "/teacher/class",
    active: true,
    badge: "Class",
  },
  { label: "Class means", icon: BarChart3, href: "#" },
  { label: "Roll-call", icon: ClipboardCheck, href: "#" },
  { label: "Marks", icon: ClipboardList, href: "#" },
  { label: "Results", icon: Award, href: "#" },
  { label: "Comments", icon: MessageSquare, href: "#" },
  { label: "Add students", icon: UserPlus, href: "#" },
  { label: "Notifications", icon: Bell, href: "#", dot: true },
];

type BottomNavItem = {
  label: string;
  icon: LucideIcon;
  href?: string;
  active?: boolean;
  opensDrawer?: boolean;
};

const BOTTOM_NAV: BottomNavItem[] = [
  { label: "Home", icon: Home },
  { label: "Timetable", icon: CalendarDays, href: "/teacher" },
  { label: "Class", icon: UsersRound, active: true, href: "/teacher/class" },
  { label: "Roll-call", icon: ClipboardCheck },
  { label: "More", icon: MoreHorizontal, opensDrawer: true },
];

/**
 * Teacher roster dashboard. Renders the class roster, filters, pagination and
 * the contextual learner detail panel for the authenticated teacher.
 */
export default function ClassRosterDashboard({
  roster,
  isPreview = false,
}: {
  roster: ClassRosterData | null;
  isPreview?: boolean;
}) {
  const students = roster?.students ?? [];
  const portalName = roster?.portal?.name?.trim() || "Teacher portal";
  const academicTerm =
    roster?.portal?.academicTerm?.trim() || "No active term";
  const classInfo = roster?.class ?? null;
  const className = classInfo?.name?.trim() || "My Class Roster";
  const classTeacher = classInfo?.classTeacherName?.trim() || "Not assigned";
  const enrolledCount = classInfo?.enrolledCount ?? students.length;
  const cycleLabel = classInfo?.cycleLabel?.trim() || "No cycle set";
  const session = roster?.session ?? null;
  const userFullName = session?.userFullName?.trim() || "Teacher";
  const userRole = session?.userRole?.trim() || "Staff";
  const userSubject = session?.userSubject?.trim() || "No subject assigned";
  const userInitials = session?.initials?.trim() || initialsFromName(userFullName);

  const [selectedId, setSelectedId] = useState(students[0]?.id ?? "");
  const [page, setPage] = useState(1);
  const [query, setQuery] = useState("");
  const [guardianFilter, setGuardianFilter] = useState("all");
  const [attendanceFilter, setAttendanceFilter] = useState("all");
  const [menuOpen, setMenuOpen] = useState(false);
  const detailRef = useRef<HTMLElement | null>(null);

  const filteredStudents = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();

    return students.filter((student) => {
      const matchesQuery =
        normalizedQuery.length === 0 ||
        [
          student.name,
          student.admissionNumber,
          student.house ?? "",
          student.gradeSection ?? "",
        ]
          .join(" ")
          .toLowerCase()
          .includes(normalizedQuery);
      const matchesGuardian =
        guardianFilter === "all" || student.guardianStatus === guardianFilter;
      const percentage = student.attendance.overallPercent;
      const matchesAttendance =
        attendanceFilter === "all" ||
        (attendanceFilter === "above" &&
          percentage !== null &&
          percentage >= 90) ||
        (attendanceFilter === "below" &&
          percentage !== null &&
          percentage < 90);

      return matchesQuery && matchesGuardian && matchesAttendance;
    });
  }, [attendanceFilter, guardianFilter, query, students]);

  const totalPages = Math.max(
    1,
    Math.ceil(filteredStudents.length / ROSTER_PAGE_SIZE),
  );
  const safePage = Math.min(page, totalPages);
  const startIndex = (safePage - 1) * ROSTER_PAGE_SIZE;
  const pageRows = filteredStudents.slice(
    startIndex,
    startIndex + ROSTER_PAGE_SIZE,
  );
  const selectedStudent =
    pageRows.find((student) => student.id === selectedId) ??
    pageRows[0] ??
    null;
  const firstVisible = filteredStudents.length === 0 ? 0 : startIndex + 1;
  const lastVisible = Math.min(
    startIndex + ROSTER_PAGE_SIZE,
    filteredStudents.length,
  );
  const filtersApplied =
    query.trim().length > 0 ||
    guardianFilter !== "all" ||
    attendanceFilter !== "all";

  // Keep the selected learner aligned with a row that is visible on the
  // current page whenever filters or pagination change.
  useEffect(() => {
    if (pageRows.length === 0) {
      if (selectedId !== "") {
        setSelectedId("");
      }
      return;
    }

    if (!pageRows.some((student) => student.id === selectedId)) {
      setSelectedId(pageRows[0].id);
    }
  }, [pageRows, selectedId]);

  function selectStudent(student: RosterStudent) {
    setSelectedId(student.id);

    if (
      typeof window !== "undefined" &&
      window.matchMedia("(max-width: 1023px)").matches
    ) {
      window.setTimeout(() => {
        detailRef.current?.scrollIntoView({
          behavior: "smooth",
          block: "start",
        });
      }, 50);
    }
  }

  function goToPage(nextPage: number) {
    const target = Math.min(Math.max(nextPage, 1), totalPages);
    setPage(target);
    const firstOnPage = filteredStudents[(target - 1) * ROSTER_PAGE_SIZE];

    if (firstOnPage) {
      setSelectedId(firstOnPage.id);
    }
  }

  function changeGuardianFilter(value: string) {
    setGuardianFilter(value);
    setPage(1);
  }

  function changeAttendanceFilter(value: string) {
    setAttendanceFilter(value);
    setPage(1);
  }

  return (
    <div className="min-h-screen bg-[#F4F7F6] text-slate-900">
      <aside className="fixed inset-y-0 left-0 z-40 hidden w-[272px] border-r border-slate-200/70 bg-white lg:flex print:hidden">
        <SidebarContent
          portalName={portalName}
          userFullName={userFullName}
          userRole={userRole}
          userSubject={userSubject}
          userInitials={userInitials}
        />
      </aside>

      <MobileAppBar
        className={className}
        userInitials={userInitials}
        onOpenMenu={() => setMenuOpen(true)}
      />

      <div className="lg:pl-[272px]">
        <DesktopHeader
          portalName={portalName}
          academicTerm={academicTerm}
          userFullName={userFullName}
          userRole={userRole}
          userInitials={userInitials}
        />

        <main className="mx-auto w-full max-w-[1440px] px-4 pb-28 pt-20 sm:px-6 lg:px-8 lg:pb-10 lg:pt-7">
          <PageHeader
            className={className}
            classTeacher={classTeacher}
            enrolledCount={enrolledCount}
            cycleLabel={cycleLabel}
            isPreview={isPreview}
          />

          <div className="mt-6 grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1.6fr)_minmax(340px,1fr)]">
            <div className="min-w-0">
              <FilterBar
                query={query}
                guardianFilter={guardianFilter}
                attendanceFilter={attendanceFilter}
                onQueryChange={(value) => {
                  setQuery(value);
                  setPage(1);
                }}
                onGuardianChange={changeGuardianFilter}
                onAttendanceChange={changeAttendanceFilter}
              />

              <RosterTable
                students={pageRows}
                selectedId={selectedStudent?.id}
                onSelect={selectStudent}
                page={safePage}
                totalPages={totalPages}
                firstVisible={firstVisible}
                lastVisible={lastVisible}
                totalVisible={filteredStudents.length}
                filtersApplied={filtersApplied}
                hasRosterStudents={students.length > 0}
                onPageChange={goToPage}
              />
            </div>

            <aside
              ref={detailRef}
              className="scroll-mt-24 space-y-4 xl:sticky xl:top-[84px] xl:self-start"
            >
              {selectedStudent ? (
                <StudentDetail student={selectedStudent} />
              ) : (
                <EmptyStudentDetail />
              )}
            </aside>
          </div>
        </main>
      </div>

      <MobileBottomNav onOpenMenu={() => setMenuOpen(true)} />

      {menuOpen && (
        <MobileDrawer
          portalName={portalName}
          userFullName={userFullName}
          userRole={userRole}
          userSubject={userSubject}
          userInitials={userInitials}
          onClose={() => setMenuOpen(false)}
        />
      )}
    </div>
  );
}

/** Sidebar branding, staff navigation and profile/logout footer. */
function SidebarContent({
  portalName,
  userFullName,
  userRole,
  userSubject,
  userInitials,
  onNavigate,
}: {
  portalName: string;
  userFullName: string;
  userRole: string;
  userSubject: string;
  userInitials: string;
  onNavigate?: () => void;
}) {
  return (
    <div className="flex h-full w-full flex-col overflow-y-auto px-4 py-5">
      <div className="flex items-start justify-between gap-3 px-1">
        <div className="flex items-center gap-2.5">
          <span className="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-brand-500 text-white shadow-lg shadow-brand-500/20">
            <GraduationCap className="h-[22px] w-[22px]" />
          </span>
          <div>
            <p className="text-[15px] font-bold tracking-tight text-slate-900">
              Ikonex <span className="text-brand-600">Academy</span>
            </p>
            <div className="mt-0.5 flex items-center gap-1.5">
              <p className="max-w-[112px] truncate text-[9px] font-semibold uppercase tracking-[0.16em] text-slate-400">
                {portalName}
              </p>
              <span className="rounded-full bg-brand-50 px-1.5 py-0.5 text-[8px] font-bold uppercase tracking-wide text-brand-700 ring-1 ring-brand-100">
                Staff
              </span>
            </div>
          </div>
        </div>
      </div>

      <nav className="mt-6 space-y-0.5" aria-label="Primary">
        {NAV_ITEMS.map((item) => {
          const Icon = item.icon;

          return (
            <a
              key={item.label}
              href={item.href}
              onClick={onNavigate}
              aria-current={item.active ? "page" : undefined}
              className={cn(
                "group flex items-center gap-3 rounded-xl px-3 py-2.5 text-[12px] font-medium transition",
                item.active
                  ? "bg-[#E8F5F3] font-semibold text-brand-700 ring-1 ring-brand-100"
                  : "text-slate-500 hover:bg-slate-50 hover:text-slate-800",
              )}
            >
              <Icon
                className={cn(
                  "h-[17px] w-[17px] shrink-0",
                  item.active
                    ? "text-brand-600"
                    : "text-slate-400 group-hover:text-brand-600",
                )}
              />
              <span className="flex-1">{item.label}</span>
              {item.badge && (
                <span className="rounded-full bg-white px-1.5 py-0.5 text-[8px] font-bold text-brand-700 ring-1 ring-brand-100">
                  {item.badge}
                </span>
              )}
              {item.dot && (
                <span className="h-1.5 w-1.5 rounded-full bg-rose-500" />
              )}
            </a>
          );
        })}
      </nav>

      <div className="mt-auto pt-6">
        <div className="rounded-2xl border border-slate-200/80 bg-slate-50/70 p-3">
          <div className="flex items-center gap-2.5">
            <span className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-slate-900 text-[10px] font-bold text-white">
              {userInitials}
            </span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-xs font-semibold text-slate-800">
                {userFullName}
              </p>
              <p className="truncate text-[10px] text-slate-400">{userRole}</p>
              <p className="truncate text-[9px] font-medium text-brand-600">
                {userSubject}
              </p>
            </div>
          </div>
          <button
            type="button"
            className="mt-3 inline-flex h-9 w-full items-center justify-center gap-2 rounded-xl border border-slate-200/80 bg-white text-[11px] font-semibold text-slate-600 transition hover:border-brand-200 hover:bg-brand-50/40 hover:text-brand-700"
          >
            <LogOut className="h-3.5 w-3.5" />
            Sign out
          </button>
        </div>
      </div>
    </div>
  );
}

/** Sticky desktop header with breadcrumb, notifications and profile menu. */
function DesktopHeader({
  portalName,
  academicTerm,
  userFullName,
  userRole,
  userInitials,
}: {
  portalName: string;
  academicTerm: string;
  userFullName: string;
  userRole: string;
  userInitials: string;
}) {
  return (
    <header className="sticky top-0 z-30 hidden h-[68px] items-center border-b border-slate-200/60 bg-[#F4F7F6]/85 backdrop-blur-xl lg:flex print:hidden">
      <div className="mx-auto flex w-full max-w-[1440px] items-center justify-between gap-6 px-8">
        <div className="flex items-center gap-2 text-xs font-medium text-slate-400">
          <span className="max-w-[220px] truncate text-slate-700">
            {portalName}
          </span>
          <ChevronRight className="h-3.5 w-3.5" />
          <span className="max-w-[220px] truncate">{academicTerm}</span>
        </div>

        <div className="flex items-center gap-2.5">
          <button
            type="button"
            className="relative grid h-10 w-10 place-items-center rounded-xl border border-slate-200/80 bg-white text-slate-500 shadow-sm transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
            aria-label="Notifications"
          >
            <Bell className="h-[18px] w-[18px]" />
            <span className="absolute right-2.5 top-2.5 h-2 w-2 rounded-full bg-rose-500 ring-2 ring-white" />
          </button>

          <button
            type="button"
            className="flex items-center gap-2.5 rounded-xl border border-slate-200/80 bg-white py-1.5 pl-1.5 pr-3 shadow-sm transition hover:bg-slate-50"
          >
            <span className="grid h-8 w-8 place-items-center rounded-lg bg-brand-500 text-[10px] font-bold text-white">
              {userInitials}
            </span>
            <div className="min-w-0 text-left">
              <p className="max-w-[120px] truncate text-xs font-semibold text-slate-800">
                {userFullName}
              </p>
              <p className="text-[10px] text-slate-400">{userRole}</p>
            </div>
            <ChevronDown className="h-3.5 w-3.5 text-slate-400" />
          </button>
        </div>
      </div>
    </header>
  );
}

/** Fixed mobile app bar with the navigation drawer trigger. */
function MobileAppBar({
  className,
  userInitials,
  onOpenMenu,
}: {
  className: string;
  userInitials: string;
  onOpenMenu: () => void;
}) {
  return (
    <header className="fixed inset-x-0 top-0 z-40 border-b border-slate-200/70 bg-white/90 backdrop-blur-xl lg:hidden print:hidden">
      <div className="flex h-16 items-center justify-between gap-3 px-4">
        <button
          type="button"
          onClick={onOpenMenu}
          className="grid h-10 w-10 shrink-0 place-items-center rounded-xl border border-slate-200/80 bg-white text-slate-600 shadow-sm transition hover:bg-slate-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
          aria-label="Open navigation menu"
        >
          <Menu className="h-5 w-5" />
        </button>

        <div className="min-w-0 flex-1 text-center">
          <p className="truncate text-sm font-bold tracking-tight text-slate-900">
            {className}
          </p>
          <p className="truncate text-[9px] font-semibold uppercase tracking-[0.16em] text-slate-400">
            My class roster
          </p>
        </div>

        <div className="flex shrink-0 items-center gap-1.5">
          <button
            type="button"
            className="relative grid h-10 w-10 place-items-center rounded-xl text-slate-500 transition hover:bg-slate-50"
            aria-label="Notifications"
          >
            <Bell className="h-5 w-5" />
            <span className="absolute right-2 top-2 h-2 w-2 rounded-full bg-rose-500 ring-2 ring-white" />
          </button>
          <button
            type="button"
            className="grid h-9 w-9 place-items-center rounded-full bg-brand-500 text-[10px] font-bold text-white shadow-sm ring-2 ring-white"
            aria-label="Open user profile"
          >
            {userInitials || <UsersRound className="h-4 w-4" />}
          </button>
        </div>
      </div>
    </header>
  );
}

/**
 * Page heading with class metadata and roster actions. The roster actions stay
 * disabled until their handlers are implemented.
 */
function PageHeader({
  className,
  classTeacher,
  enrolledCount,
  cycleLabel,
  isPreview,
}: {
  className: string;
  classTeacher: string;
  enrolledCount: number;
  cycleLabel: string;
  isPreview: boolean;
}) {
  return (
    <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span className="h-2 w-2 rounded-full bg-brand-500 shadow-[0_0_0_4px_rgba(0,122,120,0.10)]" />
          <p className="text-[10px] font-semibold uppercase tracking-[0.18em] text-slate-400">
            My class roster
          </p>
          {isPreview && (
            <span className="rounded-full bg-[#FEF3D6] px-2 py-0.5 text-[9px] font-semibold text-[#7A5A16] ring-1 ring-[#F3E2B8]">
              Preview
            </span>
          )}
        </div>
        <h1 className="mt-2 text-[26px] font-bold tracking-tight text-slate-900 sm:text-3xl">
          {className}
        </h1>
        <div className="mt-1.5 flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-slate-500">
          <span>
            Class teacher:{" "}
            <span className="font-semibold text-slate-700">{classTeacher}</span>
          </span>
          <span className="text-slate-300">•</span>
          <span>
            <span className="font-semibold text-slate-700">
              {enrolledCount.toLocaleString("en-GB")}
            </span>{" "}
            students enrolled
          </span>
          <span className="text-slate-300">•</span>
          <span>{cycleLabel}</span>
        </div>
        {isPreview && (
          <p className="mt-2 max-w-2xl text-[11px] leading-relaxed text-[#7A5A16]">
            Preview: live roster sync is not connected yet. Learner, guardian
            and attendance data will appear once the authenticated roster
            endpoint is available.
          </p>
        )}
      </div>

      <div className="flex w-full gap-2 sm:w-auto">
        <button
          type="button"
          disabled
          aria-disabled="true"
          title="Add students is not available yet"
          className="inline-flex h-10 flex-1 items-center justify-center gap-2 rounded-xl border border-slate-200/80 bg-white px-3 text-xs font-semibold text-slate-600 shadow-sm transition hover:border-brand-200 hover:bg-brand-50/40 hover:text-brand-700 disabled:cursor-not-allowed disabled:opacity-50 sm:flex-none"
        >
          <UserPlus className="h-4 w-4" />
          Add students
        </button>
        <button
          type="button"
          disabled
          aria-disabled="true"
          title="Roll-call is not available yet"
          className="inline-flex h-10 flex-1 items-center justify-center gap-2 rounded-xl bg-brand-500 px-3 text-xs font-semibold text-white shadow-lg shadow-brand-500/20 transition hover:bg-brand-600 disabled:cursor-not-allowed disabled:opacity-50 sm:flex-none"
        >
          <ClipboardCheck className="h-4 w-4" />
          Take roll-call
        </button>
      </div>
    </div>
  );
}

/**
 * Filter controls for the roster: search, guardian status and attendance.
 */
function FilterBar({
  query,
  guardianFilter,
  attendanceFilter,
  onQueryChange,
  onGuardianChange,
  onAttendanceChange,
}: {
  query: string;
  guardianFilter: string;
  attendanceFilter: string;
  onQueryChange: (value: string) => void;
  onGuardianChange: (value: string) => void;
  onAttendanceChange: (value: string) => void;
}) {
  return (
    <div className="rounded-2xl border border-slate-200/80 bg-white p-3 shadow-card">
      <div className="flex flex-col gap-2.5 md:flex-row">
        <label className="relative min-w-0 flex-1">
          <span className="sr-only">Search roster</span>
          <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
          <input
            type="search"
            value={query}
            onChange={(event) => onQueryChange(event.target.value)}
            placeholder="Search by name, admission number or house..."
            className="h-11 w-full rounded-xl border border-slate-200/80 bg-slate-50/60 pl-9 pr-3 text-xs text-slate-700 outline-none transition placeholder:text-slate-400 focus:border-brand-300 focus:bg-white focus:ring-4 focus:ring-brand-500/10"
          />
        </label>

        <FilterSelect
          label="Guardians"
          value={guardianFilter}
          onChange={onGuardianChange}
          options={[
            { value: "all", label: "All" },
            { value: "verified", label: "Verified" },
            { value: "pending", label: "Pending" },
          ]}
        />
        <FilterSelect
          label="Attendance"
          value={attendanceFilter}
          onChange={onAttendanceChange}
          options={[
            { value: "all", label: "All" },
            { value: "above", label: "90% or above" },
            { value: "below", label: "Below 90%" },
          ]}
        />
      </div>
    </div>
  );
}

/** Labelled select used by the roster filter bar. */
function FilterSelect({
  label,
  value,
  options,
  onChange,
}: {
  label: string;
  value: string;
  options: { value: string; label: string }[];
  onChange: (value: string) => void;
}) {
  return (
    <label className="relative shrink-0 md:w-[148px]">
      <span className="sr-only">{label}</span>
      <select
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className="h-11 w-full appearance-none rounded-xl border border-slate-200/80 bg-slate-50/60 pl-3 pr-8 text-[11px] font-semibold text-slate-600 outline-none transition focus:border-brand-300 focus:bg-white focus:ring-4 focus:ring-brand-500/10"
      >
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {label}: {option.label}
          </option>
        ))}
      </select>
      <ChevronDown className="pointer-events-none absolute right-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-slate-400" />
    </label>
  );
}

/**
 * Returns a bounded window of page numbers so pagination never overflows the
 * available width, even for long rosters.
 */
function getVisiblePages(
  currentPage: number,
  totalPages: number,
  maxButtons = 5,
) {
  if (totalPages <= maxButtons) {
    return Array.from({ length: totalPages }, (_, index) => index + 1);
  }

  const half = Math.floor(maxButtons / 2);
  let start = Math.max(1, currentPage - half);
  let end = start + maxButtons - 1;

  if (end > totalPages) {
    end = totalPages;
    start = totalPages - maxButtons + 1;
  }

  return Array.from({ length: end - start + 1 }, (_, index) => start + index);
}

/** Roster table with row selection, empty state and bounded pagination. */
function RosterTable({
  students,
  selectedId,
  onSelect,
  page,
  totalPages,
  firstVisible,
  lastVisible,
  totalVisible,
  filtersApplied,
  hasRosterStudents,
  onPageChange,
}: {
  students: RosterStudent[];
  selectedId: string | undefined;
  onSelect: (student: RosterStudent) => void;
  page: number;
  totalPages: number;
  firstVisible: number;
  lastVisible: number;
  totalVisible: number;
  filtersApplied: boolean;
  hasRosterStudents: boolean;
  onPageChange: (page: number) => void;
}) {
  const emptyTitle = hasRosterStudents
    ? "No learners match these filters"
    : "No learners available";
  const emptyBody = hasRosterStudents
    ? "Clear the search or choose a different guardian or attendance filter."
    : "The class roster will appear here once the school provides it.";

  return (
    <section className="mt-4 overflow-hidden rounded-2xl border border-slate-200/80 bg-white shadow-card">
      <div className="flex items-center justify-between gap-3 border-b border-slate-100 px-4 py-4 sm:px-5">
        <div>
          <p className="text-[9px] font-semibold uppercase tracking-[0.18em] text-slate-400">
            Class list
          </p>
          <h2 className="mt-0.5 text-sm font-bold tracking-tight text-slate-900">
            Enrolled learners
          </h2>
        </div>
        <span className="rounded-full bg-brand-50 px-2.5 py-1 text-[10px] font-semibold text-brand-700 ring-1 ring-brand-100">
          {totalVisible.toLocaleString("en-GB")} students
        </span>
      </div>

      <div className="hidden md:block">
        <table className="w-full border-collapse text-left">
          <thead>
            <tr className="border-b border-slate-100 bg-slate-50/60">
              <TableHeading>Admission #</TableHeading>
              <TableHeading>Student name</TableHeading>
              <TableHeading>Dormitory / House</TableHeading>
              <TableHeading>Linked guardians</TableHeading>
              <TableHeading>Term attendance</TableHeading>
            </tr>
          </thead>
          <tbody>
            {students.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-5 py-12">
                  <EmptyTableState title={emptyTitle} body={emptyBody} />
                </td>
              </tr>
            ) : (
              students.map((student, index) => {
                const active = student.id === selectedId;

                return (
                  <tr
                    key={student.id}
                    tabIndex={0}
                    aria-selected={active}
                    onClick={() => onSelect(student)}
                    onKeyDown={(event) => {
                      if (event.key === "Enter" || event.key === " ") {
                        event.preventDefault();
                        onSelect(student);
                      }
                    }}
                    className={cn(
                      "cursor-pointer border-b border-slate-100 transition last:border-0 focus-visible:outline-none",
                      active
                        ? "bg-[#F1FAF8]"
                        : "hover:bg-slate-50/70 focus-visible:bg-slate-50",
                    )}
                  >
                    <td className="px-4 py-3.5 sm:px-5">
                      <button
                        type="button"
                        className="text-xs font-semibold text-brand-600 hover:underline"
                        onClick={(event) => {
                          event.stopPropagation();
                          onSelect(student);
                        }}
                      >
                        {student.admissionNumber}
                      </button>
                    </td>
                    <td className="px-4 py-3.5 sm:px-5">
                      <div className="flex items-center gap-2.5">
                        <Avatar student={student} toneIndex={index} size="sm" />
                        <div className="min-w-0">
                          <p className="truncate text-xs font-semibold text-slate-800">
                            {student.name}
                          </p>
                          <p className="text-[10px] text-slate-400">
                            {student.gradeSection || "Not provided"}
                          </p>
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3.5 text-[11px] text-slate-500 sm:px-5">
                      {student.house || "Not assigned"}
                    </td>
                    <td className="px-4 py-3.5 sm:px-5">
                      <div className="flex items-center gap-2">
                        <span className="text-[11px] text-slate-500">
                          {student.linkedGuardians > 0
                            ? `${student.linkedGuardians} linked`
                            : "None linked"}
                        </span>
                        <GuardianBadge status={student.guardianStatus} />
                      </div>
                    </td>
                    <td className="px-4 py-3.5 sm:px-5">
                      <span className="inline-flex rounded-full bg-slate-100 px-2 py-1 text-[10px] font-semibold text-slate-600">
                        {formatPercent(student.attendance.overallPercent)}
                      </span>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      <div className="divide-y divide-slate-100 md:hidden">
        {students.length === 0 ? (
          <div className="px-4 py-10">
            <EmptyTableState title={emptyTitle} body={emptyBody} />
          </div>
        ) : (
          students.map((student, index) => {
            const active = student.id === selectedId;

            return (
              <button
                key={student.id}
                type="button"
                onClick={() => onSelect(student)}
                className={cn(
                  "w-full px-4 py-3.5 text-left transition",
                  active
                    ? "border-l-2 border-brand-500 bg-[#F1FAF8]"
                    : "border-l-2 border-transparent hover:bg-slate-50",
                )}
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="flex min-w-0 items-center gap-2.5">
                    <Avatar student={student} toneIndex={index} size="md" />
                    <div className="min-w-0">
                      <p className="truncate text-xs font-semibold text-slate-800">
                        {student.name}
                      </p>
                      <p className="text-[10px] text-brand-600">
                        {student.admissionNumber}
                      </p>
                      <p className="truncate text-[10px] text-slate-400">
                        {student.house || "Not assigned"} ·{" "}
                        {student.gradeSection || "Not provided"}
                      </p>
                    </div>
                  </div>
                  <span className="shrink-0 rounded-full bg-slate-100 px-2 py-1 text-[10px] font-semibold text-slate-600">
                    {formatPercent(student.attendance.overallPercent)}
                  </span>
                </div>
                <div className="mt-3 flex items-center justify-between gap-2">
                  <span className="text-[10px] text-slate-500">
                    {student.linkedGuardians > 0
                      ? `${student.linkedGuardians} linked guardians`
                      : "No guardians linked"}
                  </span>
                  <GuardianBadge status={student.guardianStatus} />
                </div>
              </button>
            );
          })
        )}
      </div>

      <div className="flex flex-col gap-3 border-t border-slate-100 bg-slate-50/30 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
        <p className="text-[10px] text-slate-500">
          {totalVisible === 0 ? (
            <>
              Showing 0 {filtersApplied ? "matching" : "enrolled"} learners
            </>
          ) : (
            <>
              Showing{" "}
              <span className="font-semibold text-slate-700">
                {firstVisible}–{lastVisible}
              </span>{" "}
              of{" "}
              <span className="font-semibold text-slate-700">
                {totalVisible.toLocaleString("en-GB")}
              </span>{" "}
              {filtersApplied ? "matching" : "enrolled"} learners
            </>
          )}
        </p>

        <div className="flex flex-wrap items-center gap-1">
          <button
            type="button"
            onClick={() => onPageChange(page - 1)}
            disabled={page <= 1}
            className="inline-flex h-8 items-center gap-1 rounded-lg px-2 text-[10px] font-semibold text-slate-500 transition hover:bg-white disabled:cursor-not-allowed disabled:opacity-40"
          >
            <ChevronLeft className="h-3.5 w-3.5" />
            Previous
          </button>

          {getVisiblePages(page, totalPages).map((pageNumber) => (
              <button
                key={pageNumber}
                type="button"
                onClick={() => onPageChange(pageNumber)}
                aria-current={pageNumber === page ? "page" : undefined}
                className={cn(
                  "grid h-8 w-8 place-items-center rounded-lg text-[10px] font-semibold transition",
                  pageNumber === page
                    ? "bg-brand-500 text-white shadow-sm"
                    : "text-slate-500 hover:bg-white",
                )}
              >
                {pageNumber}
              </button>
          ))}

          <button
            type="button"
            onClick={() => onPageChange(page + 1)}
            disabled={page >= totalPages}
            className="inline-flex h-8 items-center gap-1 rounded-lg px-2 text-[10px] font-semibold text-slate-500 transition hover:bg-white disabled:cursor-not-allowed disabled:opacity-40"
          >
            Next
            <ChevronRight className="h-3.5 w-3.5" />
          </button>
        </div>
      </div>
    </section>
  );
}

/** Centred empty-state block used inside the roster table. */
function EmptyTableState({ title, body }: { title: string; body: string }) {
  return (
    <div className="mx-auto max-w-sm text-center">
      <span className="mx-auto grid h-11 w-11 place-items-center rounded-2xl bg-slate-50 text-slate-400">
        <UsersRound className="h-5 w-5" />
      </span>
      <p className="mt-3 text-sm font-semibold text-slate-700">{title}</p>
      <p className="mt-1 text-xs leading-relaxed text-slate-500">{body}</p>
    </div>
  );
}

/** Table column heading cell. */
function TableHeading({ children }: { children: ReactNode }) {
  return (
    <th className="px-4 py-2.5 text-[9px] font-semibold uppercase tracking-[0.12em] text-slate-400 sm:px-5">
      {children}
    </th>
  );
}

/** Circular initials avatar for a learner. */
function Avatar({
  student,
  toneIndex,
  size = "md",
}: {
  student: RosterStudent;
  toneIndex?: number;
  size?: "sm" | "md" | "lg";
}) {
  const tone =
    AVATAR_TONE_CLASSES[
      (student.avatarTone ?? toneIndex ?? 0) % AVATAR_TONE_CLASSES.length
    ];

  return (
    <span
      className={cn(
        "grid shrink-0 place-items-center rounded-full font-bold ring-1 ring-white",
        tone,
        size === "sm" && "h-8 w-8 text-[9px]",
        size === "md" && "h-9 w-9 text-[10px]",
        size === "lg" && "h-12 w-12 text-xs",
      )}
    >
      {getStudentInitials(student)}
    </span>
  );
}

/** Guardian verification status pill. */
function GuardianBadge({ status }: { status: GuardianStatus }) {
  const verified = status === "verified";

  return (
    <span
      className={cn(
        "inline-flex shrink-0 items-center gap-1 rounded-full px-2 py-0.5 text-[9px] font-semibold ring-1",
        verified
          ? "bg-[#E2F0ED] text-brand-700 ring-brand-100"
          : "bg-[#FEF3D6] text-[#8A6A25] ring-[#EBD7A5]",
      )}
    >
      {verified ? (
        <ShieldCheck className="h-3 w-3" />
      ) : (
        <Clock3 className="h-3 w-3" />
      )}
      {verified ? "Verified" : "Pending"}
    </span>
  );
}

/** Contextual learner details: identity, guardians and attendance summary. */
function StudentDetail({ student }: { student: RosterStudent }) {
  const primaryGuardian = getPrimaryGuardian(student);
  const attendance = student.attendance;
  const guardianStatus = primaryGuardian?.status ?? student.guardianStatus;

  return (
    <>
      <section className="rounded-2xl border border-slate-200/80 bg-white p-5 shadow-card">
        <div className="flex items-start justify-between gap-3">
          <div className="flex min-w-0 items-center gap-3">
            <Avatar student={student} size="lg" />
            <div className="min-w-0">
              <h2 className="truncate text-base font-bold tracking-tight text-slate-900">
                {student.name}
              </h2>
              <button
                type="button"
                className="mt-0.5 text-[11px] font-semibold text-brand-600 hover:underline"
              >
                {student.admissionNumber}
              </button>
            </div>
          </div>
          <span className="shrink-0 rounded-full bg-[#F0EBF9] px-2.5 py-1 text-[9px] font-semibold text-[#6D4FA3] ring-1 ring-[#E4DBF5]">
            {student.house || "Not assigned"}
          </span>
        </div>

        <p className="mt-3 text-[11px] text-slate-500">
          {student.gradeSection || "Grade / section not provided"} ·{" "}
          {student.enrollmentType || "Enrollment type not provided"}
        </p>

        <div className="mt-5 border-t border-slate-100 pt-5">
          <div className="flex items-center justify-between gap-2">
            <h3 className="text-xs font-bold text-slate-800">Guardians</h3>
            <span className="inline-flex items-center gap-1 text-[9px] font-semibold text-brand-700">
              <Lock className="h-3 w-3" />
              Data encrypted
            </span>
          </div>

          <div className="mt-3 rounded-xl bg-[#F8FAF9] p-3.5 ring-1 ring-slate-100">
            <div className="flex items-center justify-between gap-2">
              <p className="text-xs font-semibold text-slate-800">
                {primaryGuardian?.relationship || "No guardian linked"}
              </p>
              {primaryGuardian?.isPrimary && (
                <span className="rounded-full bg-white px-2 py-0.5 text-[9px] font-semibold text-brand-700 ring-1 ring-brand-100">
                  Primary
                </span>
              )}
            </div>
            <p className="mt-1 text-[11px] text-slate-500">
              {primaryGuardian?.phoneMasked || "No contact recorded"}
            </p>
            {primaryGuardian ? (
              <p
                className={cn(
                  "mt-2 inline-flex items-center gap-1 text-[9px] font-medium",
                  guardianStatus === "verified"
                    ? "text-brand-700"
                    : "text-[#8A6A25]",
                )}
              >
                {guardianStatus === "verified" ? (
                  <ShieldCheck className="h-3 w-3" />
                ) : (
                  <Clock3 className="h-3 w-3" />
                )}
                {primaryGuardian.verificationNote ||
                  (guardianStatus === "verified"
                    ? "Verification record available"
                    : "Verification pending")}
              </p>
            ) : (
              <p className="mt-2 inline-flex items-center gap-1 text-[9px] font-medium text-slate-400">
                <Info className="h-3 w-3" />
                No verification record
              </p>
            )}
          </div>
        </div>

        <div className="mt-5 border-t border-slate-100 pt-5">
          <div className="flex items-center justify-between gap-2">
            <h3 className="text-xs font-bold text-slate-800">
              Attendance summary
            </h3>
            <span className="rounded-full bg-brand-50 px-2 py-1 text-[9px] font-semibold text-brand-700 ring-1 ring-brand-100">
              {formatPercent(attendance.overallPercent)}
            </span>
          </div>

          <div className="mt-3 grid grid-cols-3 gap-2">
            <MetricCard
              label="Total days"
              value={formatCount(attendance.totalDays)}
              icon={CalendarDays}
              tone="teal"
            />
            <MetricCard
              label="Present"
              value={formatCount(attendance.daysPresent)}
              icon={CalendarCheck}
              tone="green"
            />
            <MetricCard
              label="Absent"
              value={formatCount(attendance.daysAbsent)}
              icon={CalendarX}
              tone="gold"
            />
          </div>

          <div className="mt-3 flex items-center gap-2 rounded-xl bg-slate-50 px-3 py-2.5">
            <Info className="h-3.5 w-3.5 shrink-0 text-slate-400" />
            <p className="min-w-0 flex-1 truncate text-[10px] text-slate-500">
              Absence note status:{" "}
              <span className="font-semibold text-slate-700">
                {attendance.absenceNoteStatus || "Not recorded"}
              </span>
            </p>
            <button
              type="button"
              className="shrink-0 text-[9px] font-semibold text-brand-600 hover:underline"
            >
              View
            </button>
          </div>
        </div>
      </section>

      <ActionAndPolicyCards disabled />
    </>
  );
}

/** Skeleton detail panel shown when no learner is selected. */
function EmptyStudentDetail() {
  return (
    <>
      <section className="rounded-2xl border border-slate-200/80 bg-white p-5 shadow-card">
        <div className="flex items-start justify-between gap-3">
          <div className="flex min-w-0 items-center gap-3">
            <span className="grid h-12 w-12 shrink-0 place-items-center rounded-full bg-slate-100 text-slate-400 ring-1 ring-white">
              <UsersRound className="h-5 w-5" />
            </span>
            <div className="min-w-0">
              <h2 className="truncate text-base font-bold tracking-tight text-slate-900">
                No learner selected
              </h2>
              <p className="mt-0.5 text-[11px] font-semibold text-slate-400">—</p>
            </div>
          </div>
          <span className="shrink-0 rounded-full bg-slate-100 px-2.5 py-1 text-[9px] font-semibold text-slate-500">
            Not assigned
          </span>
        </div>

        <p className="mt-3 text-[11px] text-slate-500">
          Select a learner from the roster to view guardian and attendance
          details.
        </p>

        <div className="mt-5 border-t border-slate-100 pt-5">
          <div className="flex items-center justify-between gap-2">
            <h3 className="text-xs font-bold text-slate-800">Guardians</h3>
            <span className="inline-flex items-center gap-1 text-[9px] font-semibold text-brand-700">
              <Lock className="h-3 w-3" />
              Data encrypted
            </span>
          </div>
          <div className="mt-3 rounded-xl bg-[#F8FAF9] p-3.5 text-center ring-1 ring-slate-100">
            <p className="text-[11px] font-medium text-slate-500">
              No guardian data available
            </p>
          </div>
        </div>

        <div className="mt-5 border-t border-slate-100 pt-5">
          <div className="flex items-center justify-between gap-2">
            <h3 className="text-xs font-bold text-slate-800">
              Attendance summary
            </h3>
            <span className="rounded-full bg-slate-100 px-2 py-1 text-[9px] font-semibold text-slate-500">
              Not recorded
            </span>
          </div>

          <div className="mt-3 grid grid-cols-3 gap-2">
            <MetricCard
              label="Total days"
              value="—"
              icon={CalendarDays}
              tone="teal"
            />
            <MetricCard
              label="Present"
              value="—"
              icon={CalendarCheck}
              tone="green"
            />
            <MetricCard
              label="Absent"
              value="—"
              icon={CalendarX}
              tone="gold"
            />
          </div>

          <div className="mt-3 flex items-center gap-2 rounded-xl bg-slate-50 px-3 py-2.5">
            <Info className="h-3.5 w-3.5 shrink-0 text-slate-400" />
            <p className="text-[10px] text-slate-500">
              Absence note status:{" "}
              <span className="font-semibold text-slate-700">
                Not recorded
              </span>
            </p>
          </div>
        </div>
      </section>

      <ActionAndPolicyCards disabled />
    </>
  );
}

/** Assessment-portfolio action and data-protection policy callout. */
function ActionAndPolicyCards({ disabled = false }: { disabled?: boolean }) {
  return (
    <>
      <button
        type="button"
        disabled={disabled}
        className="flex w-full items-center justify-between gap-3 rounded-2xl border border-brand-100 bg-[#E8F5F3] px-4 py-3.5 text-left text-[11px] font-semibold text-brand-800 shadow-sm transition hover:bg-brand-50 disabled:cursor-not-allowed disabled:opacity-50"
      >
        <span className="inline-flex items-center gap-2">
          <LayoutDashboard className="h-4 w-4 text-brand-600" />
          View assessment portfolio
        </span>
        <ExternalLink className="h-3.5 w-3.5 text-brand-600" />
      </button>

      <div className="rounded-2xl border border-slate-200/80 bg-[#F4F7F6] p-4">
        <div className="flex items-start gap-2.5">
          <span className="grid h-8 w-8 shrink-0 place-items-center rounded-lg bg-white text-slate-500 shadow-sm">
            <Lock className="h-3.5 w-3.5" />
          </span>
          <div>
            <p className="text-[10px] font-bold text-slate-700">
              Data protection
            </p>
            <p className="mt-1 text-[10px] leading-relaxed text-slate-500">
              [Compliance disclaimer] Student and guardian information is
              accessed under school policy and data-protection rules.
            </p>
          </div>
        </div>
      </div>
    </>
  );
}

/** Small attendance metric tile. */
function MetricCard({
  label,
  value,
  icon: Icon,
  tone,
}: {
  label: string;
  value: string;
  icon: LucideIcon;
  tone: "teal" | "green" | "gold";
}) {
  return (
    <div
      className={cn(
        "rounded-xl p-3",
        tone === "teal" && "bg-[#E8F5F3] text-brand-800",
        tone === "green" && "bg-[#E2F0ED] text-brand-800",
        tone === "gold" && "bg-[#FEF3D6] text-[#7A5A16]",
      )}
    >
      <Icon className="h-3.5 w-3.5 opacity-70" />
      <p className="mt-2 text-[15px] font-bold leading-none tracking-tight">
        {value}
      </p>
      <p className="mt-1 text-[8px] font-semibold uppercase tracking-wide opacity-60">
        {label}
      </p>
    </div>
  );
}

/** Mobile bottom navigation. Implemented destinations link out; "More" opens the drawer. */
function MobileBottomNav({ onOpenMenu }: { onOpenMenu: () => void }) {
  return (
    <nav
      aria-label="Mobile navigation"
      className="fixed inset-x-0 bottom-0 z-40 border-t border-slate-200/70 bg-white/95 backdrop-blur-xl lg:hidden print:hidden"
    >
      <div className="grid grid-cols-5 px-1 pb-[max(0.35rem,env(safe-area-inset-bottom))] pt-1">
        {BOTTOM_NAV.map((item) => {
          const Icon = item.icon;
          const itemClassName = cn(
            "relative flex flex-col items-center gap-1 rounded-xl py-2 text-[9px] font-semibold transition",
            item.active
              ? "text-brand-600"
              : "text-slate-400 hover:text-slate-600",
          );
          const content = (
            <>
              {item.active && (
                <span className="absolute -top-1 h-0.5 w-8 rounded-full bg-brand-500" />
              )}
              <Icon className="h-5 w-5" />
              <span>{item.label}</span>
            </>
          );

          if (item.opensDrawer) {
            return (
              <button
                key={item.label}
                type="button"
                onClick={onOpenMenu}
                className={itemClassName}
              >
                {content}
              </button>
            );
          }

          if (item.href) {
            return (
              <a
                key={item.label}
                href={item.href}
                aria-current={item.active ? "page" : undefined}
                className={itemClassName}
              >
                {content}
              </a>
            );
          }

          return (
            <span
              key={item.label}
              aria-disabled="true"
              className={cn(itemClassName, "cursor-default opacity-60")}
            >
              {content}
            </span>
          );
        })}
      </div>
    </nav>
  );
}

/**
 * Slide-in mobile navigation drawer. Moves focus inside on open, contains Tab,
 * closes on Escape and restores focus to the trigger on close.
 */
function MobileDrawer({
  portalName,
  userFullName,
  userRole,
  userSubject,
  userInitials,
  onClose,
}: {
  portalName: string;
  userFullName: string;
  userRole: string;
  userSubject: string;
  userInitials: string;
  onClose: () => void;
}) {
  const drawerRef = useRef<HTMLDivElement | null>(null);
  const onCloseRef = useRef(onClose);
  onCloseRef.current = onClose;

  useEffect(() => {
    const previouslyFocused = document.activeElement as HTMLElement | null;
    const drawer = drawerRef.current;

    const getFocusable = () =>
      drawer
        ? Array.from(
            drawer.querySelectorAll<HTMLElement>(
              'a[href], button:not([disabled]), input, select, textarea, [tabindex]:not([tabindex="-1"])',
            ),
          ).filter((element) => element.offsetParent !== null)
        : [];

    getFocusable()[0]?.focus();

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        event.preventDefault();
        onCloseRef.current();
        return;
      }

      if (event.key !== "Tab") {
        return;
      }

      const focusable = getFocusable();
      if (focusable.length === 0) {
        return;
      }

      const first = focusable[0];
      const last = focusable[focusable.length - 1];

      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    }

    document.addEventListener("keydown", handleKeyDown);

    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      previouslyFocused?.focus();
    };
  }, []);

  return (
    <div className="fixed inset-0 z-50 lg:hidden print:hidden">
      <button
        type="button"
        onClick={onClose}
        tabIndex={-1}
        className="absolute inset-0 cursor-default bg-slate-900/40 backdrop-blur-sm animate-fade-in"
        aria-label="Close navigation menu"
      />
      <div
        ref={drawerRef}
        role="dialog"
        aria-modal="true"
        aria-label="Navigation menu"
        className="absolute inset-y-0 left-0 flex w-[84%] max-w-[320px] flex-col bg-white shadow-2xl animate-drawer-in"
      >
        <div className="flex items-center justify-between px-5 pt-5">
          <p className="text-[10px] font-semibold uppercase tracking-[0.18em] text-slate-400">
            Navigation
          </p>
          <button
            type="button"
            onClick={onClose}
            className="grid h-9 w-9 place-items-center rounded-xl bg-slate-50 text-slate-500 transition hover:bg-slate-100"
            aria-label="Close navigation menu"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
        <div className="min-h-0 flex-1">
          <SidebarContent
            portalName={portalName}
            userFullName={userFullName}
            userRole={userRole}
            userSubject={userSubject}
            userInitials={userInitials}
            onNavigate={onClose}
          />
        </div>
      </div>
    </div>
  );
}

/** Derives up to two uppercase initials from a display name. */
function initialsFromName(name: string) {
  const parts = name
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2);

  return parts.map((part) => part[0]?.toUpperCase()).join("") || "T";
}
