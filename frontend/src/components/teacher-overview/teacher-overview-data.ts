export type TodoKind = "warning" | "task" | "progress" | "done";

export type TodoItem = {
  id: string;
  title: string;
  meta: string | null;
  status: string;
  kind: TodoKind;
};

export type TimetableRowKind = "completed" | "active" | "break" | "upcoming";

export type TimetableRow = {
  id: string;
  kind: TimetableRowKind;
  startTime: string | null;
  endTime: string | null;
  subject: string | null;
  location: string | null;
  label: string | null;
};

export type AnnouncementKind = "check" | "edit" | "calendar";

export type Announcement = {
  id: string;
  title: string;
  description: string;
  kind: AnnouncementKind;
};

export type NextLesson = {
  title: string;
  group: string | null;
  startTime: string | null;
  /** ISO timestamp used for a live countdown. */
  startAt: string | null;
  countdownMinutes: number | null;
  unitTopic: string | null;
  planHref: string | null;
};

export type TeacherOverviewData = {
  portal: {
    organizationName: string | null;
    dateLabel: string | null;
    academicTerm: string | null;
    weekLabel: string | null;
  } | null;
  teacher: {
    name: string | null;
    greeting: string | null;
  } | null;
  session: {
    fullName: string;
    role: string;
    subject: string | null;
    initials: string;
  } | null;
  nextLesson: NextLesson | null;
  todo: {
    pendingCount: number | null;
    items: TodoItem[];
  };
  timetable: {
    dayLabel: string | null;
    rows: TimetableRow[];
  };
  announcements: Announcement[];
};

export const TODO_TONE_BY_KIND = {
  warning: "gold",
  task: "lavender",
  progress: "slate",
  done: "mint",
} as const;

export const ANNOUNCEMENT_TONE_BY_KIND = {
  check: "mint",
  edit: "lavender",
  calendar: "slate",
} as const;
