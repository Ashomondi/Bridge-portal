export type WeekDay = {
  id: string;
  label: string;
  date: number;
  isToday: boolean;
};

export type TimetablePeriodKind = "completed" | "break" | "active" | "upcoming";

export type TimetablePeriod = {
  id: string;
  studentId: string;
  date: string;
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

/**
 * School timezone used for week and day boundaries. Kenya does not observe
 * daylight saving, so a fixed IANA zone keeps the parent timetable stable
 * regardless of where the Next.js server is running.
 */
export const SCHOOL_TIME_ZONE = "Africa/Nairobi";

/**
 * Builds the Monday-to-Friday teaching week that contains `reference`.
 *
 * Calendar dates are resolved in the school's timezone first, then the week is
 * calculated in UTC so a server running in another timezone cannot shift the
 * selected day or week near midnight.
 */
export function buildWeek(reference: Date): WeekDay[] {
  const { year, month, day } = toSchoolDateParts(reference);
  const today = Date.UTC(year, month - 1, day);
  const dayOfWeek = new Date(today).getUTCDay();
  const monday = new Date(today);
  monday.setUTCDate(monday.getUTCDate() + (dayOfWeek === 0 ? -6 : 1 - dayOfWeek));
  const todayId = toDateId(new Date(today));

  return WEEKDAY_LABELS.map((label, index) => {
    const date = new Date(monday);
    date.setUTCDate(monday.getUTCDate() + index);
    const id = toDateId(date);

    return {
      id,
      label,
      date: date.getUTCDate(),
      isToday: id === todayId,
    };
  });
}

/** Resolves a date to its year, month and day in the school's timezone. */
function toSchoolDateParts(date: Date) {
  const parts = new Intl.DateTimeFormat("en-CA", {
    timeZone: SCHOOL_TIME_ZONE,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(date);

  const values = Object.fromEntries(
    parts
      .filter((part) => part.type !== "literal")
      .map((part) => [part.type, Number(part.value)]),
  );

  return {
    year: values.year,
    month: values.month,
    day: values.day,
  };
}

/** Formats a UTC calendar date as a `YYYY-MM-DD` identifier. */
function toDateId(date: Date) {
  const month = `${date.getUTCMonth() + 1}`.padStart(2, "0");
  const day = `${date.getUTCDate()}`.padStart(2, "0");
  return `${date.getUTCFullYear()}-${month}-${day}`;
}
