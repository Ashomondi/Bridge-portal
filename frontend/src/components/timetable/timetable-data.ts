export type WeekDay = {
  id: string;
  weekday: string;
  weekdayShort: string;
  date: number;
  month: string;
  fullLabel: string;
  isToday: boolean;
};

export type ScheduleKind =
  | "class"
  | "support"
  | "break"
  | "live"
  | "planning"
  | "pastoral";

export type TimetableEntry = {
  id: string;
  kind: ScheduleKind;
  start: string;
  end: string;
  duration: string;
  title: string;
  badge?: string;
  location?: string;
  detail?: string;
  tag?: string;
};

const SLOT_TIMES = [
  { start: "08:00", end: "08:40", duration: "40 min" },
  { start: "08:45", end: "09:25", duration: "40 min" },
  { start: "09:25", end: "09:45", duration: "20 min" },
  { start: "09:45", end: "10:25", duration: "40 min" },
  { start: "10:30", end: "11:10", duration: "40 min" },
  { start: "11:15", end: "11:55", duration: "40 min" },
] as const;

const DAY_PATTERNS: ScheduleKind[][] = [
  ["class", "class", "break", "support", "planning", "pastoral"],
  ["class", "support", "break", "live", "planning", "pastoral"],
  ["class", "class", "break", "class", "planning", "pastoral"],
  ["support", "class", "break", "class", "planning"],
  ["class", "class", "break", "pastoral", "planning"],
];

const CONTENT: Record<
  ScheduleKind,
  Pick<TimetableEntry, "title" | "badge" | "location" | "detail" | "tag">
> = {
  class: {
    title: "[Lesson Title]",
    badge: "Assigned",
    location: "[Location]",
    detail: "[Student Count]",
  },
  support: {
    title: "[Support Session Title]",
    badge: "Support",
    location: "[Library]",
    detail: "[Group focus]",
    tag: "Group focus",
  },
  break: {
    title: "[Break Name]",
    badge: "Break",
  },
  live: {
    title: "[Current Lesson Title]",
    badge: "Now",
    location: "[Location]",
    detail: "[Student Count]",
  },
  planning: {
    title: "[Planning Block]",
    badge: "Free",
    location: "[Desk / Room]",
    detail: "[Task checklist item]",
    tag: "Checklist",
  },
  pastoral: {
    title: "[Pastoral Session]",
    badge: "Pastoral",
    location: "[Office]",
    detail: "[Student Count]",
    tag: "Pastoral care",
  },
};

export function buildWeek(reference: Date): WeekDay[] {
  const monday = new Date(
    reference.getFullYear(),
    reference.getMonth(),
    reference.getDate(),
  );
  const dayOfWeek = monday.getDay();
  monday.setDate(monday.getDate() + (dayOfWeek === 0 ? -6 : 1 - dayOfWeek));

  const todayId = toDateId(reference);

  return Array.from({ length: 5 }, (_, index) => {
    const date = new Date(monday);
    date.setDate(monday.getDate() + index);

    return {
      id: toDateId(date),
      weekday: date.toLocaleDateString("en-GB", { weekday: "long" }),
      weekdayShort: date.toLocaleDateString("en-GB", { weekday: "short" }),
      date: date.getDate(),
      month: date.toLocaleDateString("en-GB", { month: "short" }),
      fullLabel: date.toLocaleDateString("en-GB", {
        weekday: "long",
        day: "numeric",
        month: "long",
      }),
      isToday: toDateId(date) === todayId,
    };
  });
}

export function getScheduleForDay(dayIndex: number): TimetableEntry[] {
  const pattern = DAY_PATTERNS[dayIndex % DAY_PATTERNS.length];

  return pattern.map((kind, slotIndex) => {
    const time = SLOT_TIMES[slotIndex];

    return {
      id: `day-${dayIndex}-slot-${slotIndex}-${kind}`,
      kind,
      ...time,
      ...CONTENT[kind],
    };
  });
}

export function getWeekSummary() {
  const entries = DAY_PATTERNS.flatMap((_, dayIndex) =>
    getScheduleForDay(dayIndex),
  );
  const periods = entries.filter((entry) => entry.kind !== "break");
  const teaching = periods.filter((entry) =>
    ["class", "support", "live"].includes(entry.kind),
  );
  const planning = periods.filter((entry) => entry.kind === "planning");

  return {
    teaching: teaching.length,
    periods: periods.length,
    planning: planning.length,
    support: periods.filter((entry) => entry.kind === "support").length,
  };
}

function toDateId(date: Date) {
  const month = `${date.getMonth() + 1}`.padStart(2, "0");
  const day = `${date.getDate()}`.padStart(2, "0");
  return `${date.getFullYear()}-${month}-${day}`;
}
