"use client";

import { useMemo, useRef, useState } from "react";
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
  CLASS_PORTAL,
  ROSTER_PAGE_SIZE,
  ROSTER_STUDENTS,
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

const BOTTOM_NAV = [
  { label: "Home", icon: Home },
  { label: "Timetable", icon: CalendarDays },
  { label: "Class", icon: UsersRound, active: true },
  { label: "Roll-call", icon: ClipboardCheck },
  { label: "More", icon: MoreHorizontal },
];


export default function ClassRosterDashboard() {
  const [selectedId, setSelectedId] = useState(ROSTER_STUDENTS[0]?.id ?? "");
  const [page, setPage] = useState(1);
  const [query, setQuery] = useState("");
  const [guardianFilter, setGuardianFilter] = useState("all");
  const [attendanceFilter, setAttendanceFilter] = useState("all");
  const [menuOpen, setMenuOpen] = useState(false);
  const detailRef = useRef<HTMLElement | null>(null);

  const filteredStudents = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();

    return ROSTER_STUDENTS.filter((student) => {
      const matchesQuery =
        normalizedQuery.length === 0 ||
        [student.name, student.admissionNumber, student.house]
          .join(" ")
          .toLowerCase()
          .includes(normalizedQuery);
      const matchesGuardian =
        guardianFilter === "all" || student.guardianStatus === guardianFilter;

      return matchesQuery && matchesGuardian;
    });
  }, [guardianFilter, query]);

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
    ROSTER_STUDENTS.find((student) => student.id === selectedId) ??
    ROSTER_STUDENTS[0];
  const firstVisible = filteredStudents.length === 0 ? 0 : startIndex + 1;
  const lastVisible = Math.min(
    startIndex + ROSTER_PAGE_SIZE,
    filteredStudents.length,
  );

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
    setPage(Math.min(Math.max(nextPage, 1), totalPages));
  }

  return (
    <div className="min-h-screen bg-[#F4F7F6] text-slate-900">
      <aside className="fixed inset-y-0 left-0 z-40 hidden w-[272px] border-r border-slate-200/70 bg-white lg:flex print:hidden">
        <SidebarContent />
      </aside>

      <MobileAppBar onOpenMenu={() => setMenuOpen(true)} />

      <div className="lg:pl-[272px]">
        <DesktopHeader />

        <main className="mx-auto w-full max-w-[1440px] px-4 pb-28 pt-20 sm:px-6 lg:px-8 lg:pb-10 lg:pt-7">
          <PageHeader />

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
                onGuardianChange={(value) => {
                  setGuardianFilter(value);
                  setPage(1);
                }}
                onAttendanceChange={setAttendanceFilter}
              />

              <RosterTable
                students={pageRows}
                selectedId={selectedStudent?.id}
                onSelect={selectStudent}
                page={safePage}
                totalPages={totalPages}
                firstVisible={firstVisible}
                lastVisible={lastVisible}
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
                <div className="rounded-2xl border border-dashed border-slate-300 bg-white p-6 text-center shadow-card">
                  <UsersRound className="mx-auto h-5 w-5 text-slate-400" />
                  <p className="mt-2 text-sm font-semibold text-slate-700">
                    No learner selected
                  </p>
                  <p className="mt-1 text-xs text-slate-500">
                    Select a row to view guardian and attendance details.
                  </p>
                </div>
              )}
            </aside>
          </div>
        </main>
      </div>

      <MobileBottomNav />

      {menuOpen && <MobileDrawer onClose={() => setMenuOpen(false)} />}
    </div>
  );
}

function SidebarContent({ onNavigate }: { onNavigate?: () => void }) {
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
              <p className="text-[9px] font-semibold uppercase tracking-[0.16em] text-slate-400">
                Teacher portal
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
              [AB]
            </span>
            <div className="min-w-0 flex-1">
              <p className="truncate text-xs font-semibold text-slate-800">
                {CLASS_PORTAL.userFullName}
              </p>
              <p className="truncate text-[10px] text-slate-400">
                {CLASS_PORTAL.userRole}
              </p>
              <p className="truncate text-[9px] font-medium text-brand-600">
                {CLASS_PORTAL.userSubject}
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

function DesktopHeader() {
  return (
    <header className="sticky top-0 z-30 hidden h-[68px] items-center border-b border-slate-200/60 bg-[#F4F7F6]/85 backdrop-blur-xl lg:flex print:hidden">
      <div className="mx-auto flex w-full max-w-[1440px] items-center justify-between gap-6 px-8">
        <div className="flex items-center gap-2 text-xs font-medium text-slate-400">
          <span className="text-slate-700">{CLASS_PORTAL.portalName}</span>
          <ChevronRight className="h-3.5 w-3.5" />
          <span>{CLASS_PORTAL.academicTerm}</span>
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
              [AB]
            </span>
            <div className="min-w-0 text-left">
              <p className="max-w-[120px] truncate text-xs font-semibold text-slate-800">
                {CLASS_PORTAL.userFullName}
              </p>
              <p className="text-[10px] text-slate-400">
                {CLASS_PORTAL.userRole}
              </p>
            </div>
            <ChevronDown className="h-3.5 w-3.5 text-slate-400" />
          </button>
        </div>
      </div>
    </header>
  );
}

function MobileAppBar({ onOpenMenu }: { onOpenMenu: () => void }) {
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
            {CLASS_PORTAL.className}
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
            [AB]
          </button>
        </div>
      </div>
    </header>
  );
}

function PageHeader() {
  return (
    <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          <span className="h-2 w-2 rounded-full bg-brand-500 shadow-[0_0_0_4px_rgba(0,122,120,0.10)]" />
          <p className="text-[10px] font-semibold uppercase tracking-[0.18em] text-slate-400">
            My class roster
          </p>
        </div>
        <h1 className="mt-2 text-[26px] font-bold tracking-tight text-slate-900 sm:text-3xl">
          {CLASS_PORTAL.className}
        </h1>
        <div className="mt-1.5 flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-slate-500">
          <span>
            Class teacher:{" "}
            <span className="font-semibold text-slate-700">
              {CLASS_PORTAL.classTeacher}
            </span>
          </span>
          <span className="text-slate-300">•</span>
          <span>
            <span className="font-semibold text-slate-700">
              {CLASS_PORTAL.enrolledCount}
            </span>{" "}
            students enrolled
          </span>
          <span className="text-slate-300">•</span>
          <span>{CLASS_PORTAL.cycle}</span>
        </div>
      </div>

      <div className="flex w-full gap-2 sm:w-auto">
        <button
          type="button"
          className="inline-flex h-10 flex-1 items-center justify-center gap-2 rounded-xl border border-slate-200/80 bg-white px-3 text-xs font-semibold text-slate-600 shadow-sm transition hover:border-brand-200 hover:bg-brand-50/40 hover:text-brand-700 sm:flex-none"
        >
          <UserPlus className="h-4 w-4" />
          Add students
        </button>
        <button
          type="button"
          className="inline-flex h-10 flex-1 items-center justify-center gap-2 rounded-xl bg-brand-500 px-3 text-xs font-semibold text-white shadow-lg shadow-brand-500/20 transition hover:bg-brand-600 sm:flex-none"
        >
          <ClipboardCheck className="h-4 w-4" />
          Take roll-call
        </button>
      </div>
    </div>
  );
}

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
            { value: "above", label: "Above 90%" },
            { value: "below", label: "Below 90%" },
          ]}
        />
      </div>
    </div>
  );
}

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

function RosterTable({
  students,
  selectedId,
  onSelect,
  page,
  totalPages,
  firstVisible,
  lastVisible,
  onPageChange,
}: {
  students: RosterStudent[];
  selectedId: string | undefined;
  onSelect: (student: RosterStudent) => void;
  page: number;
  totalPages: number;
  firstVisible: number;
  lastVisible: number;
  onPageChange: (page: number) => void;
}) {
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
          {CLASS_PORTAL.enrolledCount} students
        </span>
      </div>

      {students.length === 0 ? (
        <div className="px-6 py-12 text-center">
          <Search className="mx-auto h-5 w-5 text-slate-400" />
          <p className="mt-2 text-sm font-semibold text-slate-700">
            No learners match these filters
          </p>
          <p className="mt-1 text-xs text-slate-500">
            Clear the search or choose a different guardian filter.
          </p>
        </div>
      ) : (
        <>
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
                {students.map((student) => {
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
                          <Avatar student={student} size="sm" />
                          <div className="min-w-0">
                            <p className="truncate text-xs font-semibold text-slate-800">
                              {student.name}
                            </p>
                            <p className="text-[10px] text-slate-400">
                              {student.gradeSection}
                            </p>
                          </div>
                        </div>
                      </td>
                      <td className="px-4 py-3.5 text-[11px] text-slate-500 sm:px-5">
                        {student.house}
                      </td>
                      <td className="px-4 py-3.5 sm:px-5">
                        <div className="flex items-center gap-2">
                          <span className="text-[11px] text-slate-500">
                            {student.linkedGuardians} linked
                          </span>
                          <GuardianBadge status={student.guardianStatus} />
                        </div>
                      </td>
                      <td className="px-4 py-3.5 sm:px-5">
                        <span className="inline-flex rounded-full bg-slate-100 px-2 py-1 text-[10px] font-semibold text-slate-600">
                          {student.overallAttendance}
                        </span>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          <div className="divide-y divide-slate-100 md:hidden">
            {students.map((student) => {
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
                      <Avatar student={student} size="md" />
                      <div className="min-w-0">
                        <p className="truncate text-xs font-semibold text-slate-800">
                          {student.name}
                        </p>
                        <p className="text-[10px] text-brand-600">
                          {student.admissionNumber}
                        </p>
                        <p className="truncate text-[10px] text-slate-400">
                          {student.house} · {student.gradeSection}
                        </p>
                      </div>
                    </div>
                    <span className="shrink-0 rounded-full bg-slate-100 px-2 py-1 text-[10px] font-semibold text-slate-600">
                      {student.overallAttendance}
                    </span>
                  </div>
                  <div className="mt-3 flex items-center justify-between gap-2">
                    <span className="text-[10px] text-slate-500">
                      {student.linkedGuardians} linked guardians
                    </span>
                    <GuardianBadge status={student.guardianStatus} />
                  </div>
                </button>
              );
            })}
          </div>
        </>
      )}

      <div className="flex flex-col gap-3 border-t border-slate-100 bg-slate-50/30 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
        <p className="text-[10px] text-slate-500">
          Showing{" "}
          <span className="font-semibold text-slate-700">
            {firstVisible}–{lastVisible}
          </span>{" "}
          of <span className="font-semibold text-slate-700">{CLASS_PORTAL.totalRecords}</span>{" "}
          enrolled learners
        </p>

        <div className="flex items-center gap-1">
          <button
            type="button"
            onClick={() => onPageChange(page - 1)}
            disabled={page <= 1}
            className="inline-flex h-8 items-center gap-1 rounded-lg px-2 text-[10px] font-semibold text-slate-500 transition hover:bg-white disabled:cursor-not-allowed disabled:opacity-40"
          >
            <ChevronLeft className="h-3.5 w-3.5" />
            Previous
          </button>

          {Array.from({ length: totalPages }, (_, index) => index + 1).map(
            (pageNumber) => (
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
            ),
          )}

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

function TableHeading({ children }: { children: React.ReactNode }) {
  return (
    <th className="whitespace-nowrap px-4 py-2.5 text-[9px] font-semibold uppercase tracking-[0.12em] text-slate-400 sm:px-5">
      {children}
    </th>
  );
}

function Avatar({
  student,
  size = "md",
}: {
  student: RosterStudent;
  size?: "sm" | "md" | "lg";
}) {
  return (
    <span
      className={cn(
        "grid shrink-0 place-items-center rounded-full font-bold ring-1 ring-white",
        AVATAR_TONE_CLASSES[student.avatarTone],
        size === "sm" && "h-8 w-8 text-[9px]",
        size === "md" && "h-9 w-9 text-[10px]",
        size === "lg" && "h-12 w-12 text-xs",
      )}
    >
      {student.initials}
    </span>
  );
}

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

function StudentDetail({ student }: { student: RosterStudent }) {
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
            {student.house}
          </span>
        </div>

        <p className="mt-3 text-[11px] text-slate-500">
          {student.gradeSection} · {student.enrollmentType}
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
                {student.guardianRelationship}
              </p>
              <span className="rounded-full bg-white px-2 py-0.5 text-[9px] font-semibold text-brand-700 ring-1 ring-brand-100">
                Primary
              </span>
            </div>
            <p className="mt-1 text-[11px] text-slate-500">
              {student.guardianPhone}
            </p>
            <p
              className={cn(
                "mt-2 inline-flex items-center gap-1 text-[9px] font-medium",
                student.guardianStatus === "verified"
                  ? "text-brand-700"
                  : "text-[#8A6A25]",
              )}
            >
              {student.guardianStatus === "verified" ? (
                <ShieldCheck className="h-3 w-3" />
              ) : (
                <Clock3 className="h-3 w-3" />
              )}
              {student.guardianVerificationNote}
            </p>
          </div>
        </div>

        <div className="mt-5 border-t border-slate-100 pt-5">
          <div className="flex items-center justify-between gap-2">
            <h3 className="text-xs font-bold text-slate-800">
              Attendance summary
            </h3>
            <span className="rounded-full bg-brand-50 px-2 py-1 text-[9px] font-semibold text-brand-700 ring-1 ring-brand-100">
              {student.overallAttendance}
            </span>
          </div>

          <div className="mt-3 grid grid-cols-3 gap-2">
            <MetricCard
              label="Total days"
              value={student.totalDays}
              icon={CalendarDays}
              tone="teal"
            />
            <MetricCard
              label="Present"
              value={student.daysPresent}
              icon={CalendarCheck}
              tone="green"
            />
            <MetricCard
              label="Absent"
              value={student.daysAbsent}
              icon={CalendarX}
              tone="gold"
            />
          </div>

          <div className="mt-3 flex items-center gap-2 rounded-xl bg-slate-50 px-3 py-2.5">
            <Info className="h-3.5 w-3.5 shrink-0 text-slate-400" />
            <p className="min-w-0 flex-1 truncate text-[10px] text-slate-500">
              Absence note status:{" "}
              <span className="font-semibold text-slate-700">
                {student.absenceNoteStatus}
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

      <button
        type="button"
        className="flex w-full items-center justify-between gap-3 rounded-2xl border border-brand-100 bg-[#E8F5F3] px-4 py-3.5 text-left text-[11px] font-semibold text-brand-800 shadow-sm transition hover:bg-brand-50"
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

function MobileBottomNav() {
  return (
    <nav
      aria-label="Mobile navigation"
      className="fixed inset-x-0 bottom-0 z-40 border-t border-slate-200/70 bg-white/95 backdrop-blur-xl lg:hidden print:hidden"
    >
      <div className="grid grid-cols-5 px-1 pb-[max(0.35rem,env(safe-area-inset-bottom))] pt-1">
        {BOTTOM_NAV.map((item) => {
          const Icon = item.icon;

          return (
            <a
              key={item.label}
              href={item.active ? "/teacher/class" : "#"}
              aria-current={item.active ? "page" : undefined}
              className={cn(
                "relative flex flex-col items-center gap-1 rounded-xl py-2 text-[9px] font-semibold transition",
                item.active
                  ? "text-brand-600"
                  : "text-slate-400 hover:text-slate-600",
              )}
            >
              {item.active && (
                <span className="absolute -top-1 h-0.5 w-8 rounded-full bg-brand-500" />
              )}
              <Icon className="h-5 w-5" />
              <span>{item.label}</span>
            </a>
          );
        })}
      </div>
    </nav>
  );
}

function MobileDrawer({ onClose }: { onClose: () => void }) {
  return (
    <div className="fixed inset-0 z-50 lg:hidden print:hidden">
      <button
        type="button"
        onClick={onClose}
        className="absolute inset-0 cursor-default bg-slate-900/40 backdrop-blur-sm animate-fade-in"
        aria-label="Close navigation menu"
      />
      <div className="absolute inset-y-0 left-0 flex w-[84%] max-w-[320px] flex-col bg-white shadow-2xl animate-drawer-in">
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
          <SidebarContent onNavigate={onClose} />
        </div>
      </div>
    </div>
  );
}
