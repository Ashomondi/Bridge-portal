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
  teacher: string;
  room: string;
  start: string;
  end: string;
  label: string;
  duration: string;
};

export type KeyDateKind = "standard" | "notice" | "general";

export type KeyDate = {
  id: string;
  kind: KeyDateKind;
  month: string;
  day: string;
  title: string;
  meta: string;
  badge: string;
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

export const PERIODS: TimetablePeriod[] = [
  {
    id: "period-1",
    kind: "completed",
    subject: "[Subject Name]",
    teacher: "[Teacher]",
    room: "[Room]",
    start: "[Start]",
    end: "[End]",
    label: "Completed",
    duration: "[X] min",
  },
  {
    id: "break-1",
    kind: "break",
    subject: "",
    teacher: "",
    room: "",
    start: "[Start]",
    end: "[End]",
    label: "[Break Label]",
    duration: "[X] min",
  },
  {
    id: "period-2",
    kind: "active",
    subject: "[Subject Name]",
    teacher: "[Teacher]",
    room: "[Room]",
    start: "[Start]",
    end: "[End]",
    label: "In session",
    duration: "[X] min",
  },
  {
    id: "period-3",
    kind: "upcoming",
    subject: "[Subject Name]",
    teacher: "[Teacher]",
    room: "[Room]",
    start: "[Start]",
    end: "[End]",
    label: "Next up",
    duration: "[X] min",
  },
];

export const KEY_DATES: KeyDate[] = [
  {
    id: "date-1",
    kind: "standard",
    month: "[Mon]",
    day: "[DD]",
    title: "[Event Title]",
    meta: "[Time] · [Venue]",
    badge: "[Category]",
  },
  {
    id: "date-2",
    kind: "notice",
    month: "[Mon]",
    day: "[DD]",
    title: "[Event Title]",
    meta: "[Explanation]",
    badge: "NOTICE FROM DEAN",
  },
  {
    id: "date-3",
    kind: "general",
    month: "[Mon]",
    day: "[DD]",
    title: "[Event Title]",
    meta: "[Audience]",
    badge: "[Everyone]",
  },
];

function toDateId(date: Date) {
  const month = `${date.getMonth() + 1}`.padStart(2, "0");
  const day = `${date.getDate()}`.padStart(2, "0");
  return `${date.getFullYear()}-${month}-${day}`;
}
