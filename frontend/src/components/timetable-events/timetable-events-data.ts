export type WeekDay = {
  id: string;
  label: string;
  date: number;
  isToday: boolean;
};

export type TimetablePeriodKind = "completed" | "break" | "active" | "upcoming";

export type TimetablePeriod = {
  id: string;
  kind: TimetablePeriodKind;
  subject: string;
  teacher: string | null;
  room: string | null;
  start: string | null;
  end: string | null;
  label: string;
  duration: string | null;
};

export type KeyDateKind = "standard" | "notice" | "general";

export type KeyDate = {
  id: string;
  kind: KeyDateKind;
  month: string;
  day: string;
  title: string;
  meta: string | null;
  badge: string | null;
};

export type TimetableStudent = {
  id: string;
  name: string;
  initials: string;
};

export type AcademicNotice = {
  authority: string;
  note: string;
};

export type ParentTimetableData = {
  portal: {
    name: string | null;
  } | null;
  session: {
    fullName: string;
    relationship: string;
    initials: string;
  } | null;
  students: TimetableStudent[];
  selectedStudentId: string | null;
  weekLabel: string | null;
  termLabel: string | null;
  notice: AcademicNotice | null;
  periods: TimetablePeriod[];
  keyDates: KeyDate[];
};

const WEEKDAY_LABELS = ["Mon", "Tue", "Wed", "Thu", "Fri"];

export function buildWeek(reference: Date): WeekDay[] {
  const monday = new Date(
    reference.getFullYear(),
    reference.getMonth(),
    reference.getDate(),
  );
  const dayOfWeek = monday.getDay();
  monday.setDate(monday.getDate() + (dayOfWeek === 0 ? -6 : 1 - dayOfWeek));
  const todayId = toDateId(reference);

  return WEEKDAY_LABELS.map((label, index) => {
    const date = new Date(monday);
    date.setDate(monday.getDate() + index);

    return {
      id: toDateId(date),
      label,
      date: date.getDate(),
      isToday: toDateId(date) === todayId,
    };
  });
}

function toDateId(date: Date) {
  const month = `${date.getMonth() + 1}`.padStart(2, "0");
  const day = `${date.getDate()}`.padStart(2, "0");
  return `${date.getFullYear()}-${month}-${day}`;
}
