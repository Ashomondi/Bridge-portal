export type WeekdayNumber = 1 | 2 | 3 | 4 | 5 | 6 | 7;

export type TimetableTerm = {
  id: string;
  label: string;
  startDate: string;
  endDate: string;
  isCurrent: boolean;
};

export type TimetableBreak = {
  id: string;
  label: string;
  startTime: string;
  endTime: string;
  /** ISO weekday numbers the break applies to. Omit to apply every school day. */
  days?: WeekdayNumber[];
};

export type TimetableSlotCategory =
  | "class"
  | "support"
  | "planning"
  | "pastoral";

export type TimetableSlot = {
  id: string;
  subjectId: string;
  subjectName: string;
  className?: string | null;
  teacherName?: string | null;
  teacherMembershipId?: string | null;
  room?: string | null;
  studentCount?: number | null;
  dayOfWeek: WeekdayNumber;
  startTime: string;
  endTime: string;
  /** Optional category for non-lesson periods. Defaults to a normal class. */
  category?: TimetableSlotCategory;
  note?: string | null;
  tag?: string | null;
};

export type TimetableEvent = {
  id: string;
  title: string;
  date: string;
};

/**
 * School-level schedule settings. Every value is supplied by the school's
 * timetable configuration; the UI must not invent times or days.
 */
export type SchoolTimetableSettings = {
  schoolId: string;
  schoolName: string;
  timezone: string;
  weekStartsOn: WeekdayNumber;
  schoolDays: WeekdayNumber[];
  dayStart: string | null;
  dayEnd: string | null;
  periodDurationMinutes: number | null;
  periodsPerDay: number | null;
  breaks: TimetableBreak[];
  term: TimetableTerm | null;
};

export type MyTimetableData = {
  settings: SchoolTimetableSettings;
  slots: TimetableSlot[];
  isPublished: boolean;
  lastUpdatedAt?: string | null;
  verifiedAt?: string | null;
  nextEvent?: TimetableEvent | null;
};

export type WeekDay = {
  id: string;
  weekday: string;
  weekdayShort: string;
  date: number;
  month: string;
  fullLabel: string;
  isToday: boolean;
  dayOfWeek: WeekdayNumber;
};

export type TimetableEntry = {
  id: string;
  kind: "class" | "support" | "break" | "live" | "planning" | "pastoral";
  start: string;
  end: string;
  duration: string;
  title: string;
  badge: string;
  location?: string;
  detail?: string;
  tag?: string;
};

const WEEKDAY_NAMES: Record<WeekdayNumber, string> = {
  1: "Monday",
  2: "Tuesday",
  3: "Wednesday",
  4: "Thursday",
  5: "Friday",
  6: "Saturday",
  7: "Sunday",
};

const WEEKDAY_SHORT: Record<WeekdayNumber, string> = {
  1: "Mon",
  2: "Tue",
  3: "Wed",
  4: "Thu",
  5: "Fri",
  6: "Sat",
  7: "Sun",
};

const DEFAULT_SCHOOL_DAYS: WeekdayNumber[] = [1, 2, 3, 4, 5];

/**
 * Builds the visible week from the school's configured teaching days.
 * When no settings exist yet, the page keeps a Monday–Friday structure but
 * does not create any timetable slots.
 */
export function buildWeek(
  reference: Date,
  settings: SchoolTimetableSettings | null,
): WeekDay[] {
  const schoolDays = settings
    ? normalizeWeekdays(settings.schoolDays)
    : DEFAULT_SCHOOL_DAYS;

  if (schoolDays.length === 0) {
    return [];
  }

  const weekStartsOn = settings?.weekStartsOn ?? 1;
  const anchor = new Date(
    reference.getFullYear(),
    reference.getMonth(),
    reference.getDate(),
  );
  const anchorDay = anchor.getDay();
  const weekStartDay = weekStartsOn === 7 ? 0 : weekStartsOn;
  let offset = anchorDay - weekStartDay;

  if (offset < 0) {
    offset += 7;
  }

  const weekStart = new Date(anchor);
  weekStart.setDate(anchor.getDate() - offset);
  const weekStartIso = weekStart.getDay() === 0 ? 7 : weekStart.getDay();
  const todayId = toDateId(anchor);

  return schoolDays
    .map((dayOfWeek) => {
      let dayOffset = dayOfWeek - weekStartIso;

      if (dayOffset < 0) {
        dayOffset += 7;
      }

      const date = new Date(weekStart);
      date.setDate(weekStart.getDate() + dayOffset);

      return {
        id: toDateId(date),
        weekday: WEEKDAY_NAMES[dayOfWeek],
        weekdayShort: WEEKDAY_SHORT[dayOfWeek],
        date: date.getDate(),
        month: date.toLocaleDateString("en-GB", { month: "short" }),
        fullLabel: date.toLocaleDateString("en-GB", {
          weekday: "long",
          day: "numeric",
          month: "long",
        }),
        isToday: toDateId(date) === todayId,
        dayOfWeek,
      } satisfies WeekDay;
    })
    .sort((a, b) => a.id.localeCompare(b.id));
}

/**
 * Merges real timetable slots with the school's configured breaks for one day.
 * Nothing is generated when the school has no slots for that day.
 */
export function buildDayEntries({
  slots,
  settings,
  weekday,
  isToday,
  nowMinutes,
}: {
  slots: TimetableSlot[];
  settings: SchoolTimetableSettings | null;
  weekday: WeekdayNumber;
  isToday: boolean;
  nowMinutes: number | null;
}): TimetableEntry[] {
  const entries: TimetableEntry[] = [];

  for (const slot of slots) {
    if (slot.dayOfWeek !== weekday) {
      continue;
    }

    const start = parseTime(slot.startTime);
    const end = parseTime(slot.endTime);

    if (start === null || end === null || end <= start) {
      continue;
    }

    const category = slot.category ?? "class";
    const live =
      isToday &&
      nowMinutes !== null &&
      nowMinutes >= start &&
      nowMinutes < end;
    const kind: TimetableEntry["kind"] = live ? "live" : category;

    entries.push({
      id: slot.id,
      kind,
      start: slot.startTime,
      end: slot.endTime,
      duration: formatDuration(slot.startTime, slot.endTime),
      title: slot.subjectName.trim() || "Lesson",
      badge: badgeForKind(kind),
      location: slot.room?.trim() || undefined,
      detail:
        slot.studentCount !== null && slot.studentCount !== undefined
          ? `${slot.studentCount} students`
          : slot.className?.trim() || undefined,
      tag: slot.tag?.trim() || tagForKind(category),
    });
  }

  for (const item of settings?.breaks ?? []) {
    if (item.days && !item.days.includes(weekday)) {
      continue;
    }

    const start = parseTime(item.startTime);
    const end = parseTime(item.endTime);

    if (start === null || end === null || end <= start) {
      continue;
    }

    entries.push({
      id: item.id,
      kind: "break",
      start: item.startTime,
      end: item.endTime,
      duration: formatDuration(item.startTime, item.endTime),
      title: item.label.trim() || "Break",
      badge: "Break",
    });
  }

  return entries.sort((a, b) => a.start.localeCompare(b.start));
}

export function getScheduleCounts(entries: TimetableEntry[]) {
  const periods = entries.filter((entry) => entry.kind !== "break");
  const teaching = periods.filter((entry) =>
    ["class", "support", "live"].includes(entry.kind),
  );

  return {
    teaching: teaching.length,
    periods: periods.length,
  };
}

export function getWeekSummary(slots: TimetableSlot[]) {
  const teaching = slots.filter(
    (slot) => slot.category !== "planning" && slot.category !== "pastoral",
  );

  return {
    teaching: teaching.length,
    periods: slots.length,
    planning: slots.filter((slot) => slot.category === "planning").length,
    support: slots.filter((slot) => slot.category === "support").length,
  };
}

export function getMinutesInTimeZone(timeZone?: string | null) {
  try {
    const parts = new Intl.DateTimeFormat("en-GB", {
      timeZone: timeZone || undefined,
      hour: "2-digit",
      minute: "2-digit",
      hourCycle: "h23",
    }).formatToParts(new Date());
    const hour = Number(parts.find((part) => part.type === "hour")?.value);
    const minute = Number(parts.find((part) => part.type === "minute")?.value);

    if (Number.isFinite(hour) && Number.isFinite(minute)) {
      return hour * 60 + minute;
    }
  } catch {
    // Fall back to the runtime timezone when the configured one is invalid.
  }

  const now = new Date();
  return now.getHours() * 60 + now.getMinutes();
}

export function formatDuration(start: string, end: string) {
  const startMinutes = parseTime(start);
  const endMinutes = parseTime(end);

  if (startMinutes === null || endMinutes === null || endMinutes <= startMinutes) {
    return "";
  }

  const total = endMinutes - startMinutes;
  const hours = Math.floor(total / 60);
  const minutes = total % 60;

  if (hours === 0) {
    return `${minutes} min`;
  }

  return minutes === 0 ? `${hours} hr` : `${hours}h ${minutes}m`;
}

export function formatWeekdayList(days: WeekdayNumber[]) {
  const normalized = normalizeWeekdays(days);

  if (normalized.length === 0) {
    return "Not configured";
  }

  return normalized.map((day) => WEEKDAY_SHORT[day]).join(", ");
}

export function formatTimeWindow(
  start: string | null,
  end: string | null,
) {
  if (!start || !end) {
    return "Not configured";
  }

  return `${start} – ${end}`;
}

export function formatDateTime(
  value: string | null | undefined,
  timeZone?: string | null,
) {
  if (!value) {
    return null;
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return null;
  }

  try {
    return new Intl.DateTimeFormat("en-GB", {
      timeZone: timeZone || undefined,
      day: "numeric",
      month: "short",
      hour: "2-digit",
      minute: "2-digit",
    }).format(date);
  } catch {
    return new Intl.DateTimeFormat("en-GB", {
      day: "numeric",
      month: "short",
      hour: "2-digit",
      minute: "2-digit",
    }).format(date);
  }
}

export function formatDateLabel(
  value: string | null | undefined,
  timeZone?: string | null,
) {
  if (!value) {
    return null;
  }

  const parsed = /^\d{4}-\d{2}-\d{2}$/.test(value)
    ? new Date(`${value}T12:00:00`)
    : new Date(value);

  if (Number.isNaN(parsed.getTime())) {
    return null;
  }

  try {
    return new Intl.DateTimeFormat("en-GB", {
      timeZone: timeZone || undefined,
      day: "numeric",
      month: "short",
      year: "numeric",
    }).format(parsed);
  } catch {
    return new Intl.DateTimeFormat("en-GB", {
      day: "numeric",
      month: "short",
      year: "numeric",
    }).format(parsed);
  }
}

function badgeForKind(kind: TimetableEntry["kind"]) {
  if (kind === "live") return "Now";
  if (kind === "support") return "Support";
  if (kind === "planning") return "Free";
  if (kind === "pastoral") return "Pastoral";

  return "Scheduled";
}

function tagForKind(category: TimetableSlotCategory) {
  if (category === "support") return "Support session";
  if (category === "planning") return "Checklist";
  if (category === "pastoral") return "Pastoral care";

  return undefined;
}

function normalizeWeekdays(days: WeekdayNumber[]) {
  return Array.from(new Set(days))
    .filter((day): day is WeekdayNumber => day >= 1 && day <= 7)
    .sort((a, b) => a - b);
}

function parseTime(value: string) {
  const match = /^(\d{1,2}):(\d{2})/.exec(value.trim());

  if (!match) {
    return null;
  }

  const hours = Number(match[1]);
  const minutes = Number(match[2]);

  if (
    !Number.isFinite(hours) ||
    !Number.isFinite(minutes) ||
    hours < 0 ||
    hours > 23 ||
    minutes < 0 ||
    minutes > 59
  ) {
    return null;
  }

  return hours * 60 + minutes;
}

function toDateId(date: Date) {
  const month = `${date.getMonth() + 1}`.padStart(2, "0");
  const day = `${date.getDate()}`.padStart(2, "0");
  return `${date.getFullYear()}-${month}-${day}`;
}
