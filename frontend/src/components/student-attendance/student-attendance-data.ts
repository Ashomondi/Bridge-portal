export type AttendanceStatus = "present" | "absent" | "not-marked" | "no-school";

export type AttendanceStudent = {
  id: string;
  name: string;
  initials: string;
};

export type AbsenceRecord = {
  id: string;
  date: string;
  status: string;
  markedBy: string;
  timestamp: string;
};

export type StudentAttendanceData = {
  portal: {
    name: string | null;
  } | null;
  session: {
    fullName: string;
    relationship: string;
    initials: string;
  } | null;
  termLabel: string | null;
  termAttendancePercent: number | null;
  students: AttendanceStudent[];
  selectedStudentId: string | null;
  summary: {
    presentDays: number | null;
    absentDays: number | null;
    recordedDays: number | null;
    liveSync: boolean;
  };
  statusByDate: Record<string, AttendanceStatus>;
  absenceRecords: AbsenceRecord[];
  complianceNote: string | null;
};

export type CalendarCell = {
  key: string;
  date: Date;
  day: number;
  currentMonth: boolean;
  status: AttendanceStatus | null;
};

export function buildCalendarMonth(
  year: number,
  month: number,
  statusByDate: Record<string, AttendanceStatus>,
) {
  const firstOfMonth = new Date(year, month, 1);
  const mondayOffset = (firstOfMonth.getDay() + 6) % 7;
  const gridStart = new Date(year, month, 1 - mondayOffset);
  const cells: CalendarCell[] = [];

  for (let index = 0; index < 42; index += 1) {
    const date = new Date(gridStart);
    date.setDate(gridStart.getDate() + index);
    const currentMonth = date.getMonth() === month;
    const key = toDateKey(date);
    const dayOfWeek = date.getDay();
    const weekend = dayOfWeek === 0 || dayOfWeek === 6;

    cells.push({
      key,
      date,
      day: date.getDate(),
      currentMonth,
      status: currentMonth
        ? statusByDate[key] ?? (weekend ? "no-school" : "not-marked")
        : null,
    });
  }

  return cells;
}

export function formatMonthLabel(year: number, month: number) {
  return new Date(year, month, 1).toLocaleDateString("en-GB", {
    month: "long",
    year: "numeric",
  });
}

export function toDateKey(date: Date) {
  const month = `${date.getMonth() + 1}`.padStart(2, "0");
  const day = `${date.getDate()}`.padStart(2, "0");
  return `${date.getFullYear()}-${month}-${day}`;
}

export function formatDayCount(value: number | null) {
  return value === null || value === undefined
    ? "—"
    : `${value.toLocaleString("en-GB")} days`;
}

export function formatRecordedDays(value: number | null) {
  return value === null || value === undefined
    ? "No recorded school days"
    : `${value.toLocaleString("en-GB")} recorded school days`;
}

export function formatTermAttendance(value: number | null) {
  return value === null || value === undefined
    ? "No attendance data"
    : `${value.toLocaleString("en-GB")}% Term Attendance`;
}
