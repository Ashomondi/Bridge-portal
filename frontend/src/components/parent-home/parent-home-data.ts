export type ParentHomeStudent = {
  id: string;
  name: string;
  initials: string;
};

export type FeeSummary = {
  termLabel: string;
  status: string;
  amountDue: number | null;
  currency: string;
  dueDate: string | null;
  visibilityNote: string | null;
  paidAmount: number | null;
  totalAmount: number | null;
};

export type AttendanceSnapshot = {
  status: string;
  present: boolean;
};

export type NextLesson = {
  subject: string;
  time: string | null;
  room: string | null;
};

export type ResultRow = {
  id: string;
  subject: string;
  performanceLevel: string;
  levelCode: string;
  masteryPercent: number | null;
};

export type LevelKey = {
  code: string;
  label: string;
  tone: "slate" | "lavender" | "purple" | "indigo";
};

export type CorrectionNotice = {
  note: string;
  hasNote: boolean;
};

export type ResultsSummary = {
  termLabel: string;
  framework: string | null;
  releaseDate: string | null;
  rows: ResultRow[];
  levels: LevelKey[];
  correction: CorrectionNotice | null;
  disclaimer: string | null;
};

export type HomeEvent = {
  id: string;
  month: string;
  day: string;
  title: string;
  meta: string | null;
};

export type ParentHomeData = {
  portal: {
    name: string | null;
  } | null;
  session: {
    fullName: string;
    relationship: string;
    initials: string;
  } | null;
  students: ParentHomeStudent[];
  selectedStudentId: string | null;
  fee: FeeSummary | null;
  attendance: AttendanceSnapshot | null;
  nextLesson: NextLesson | null;
  results: ResultsSummary | null;
  events: HomeEvent[];
};

/**
 * Formats a nullable amount using the school's currency, falling back to KES.
 * Returns an em dash when the amount is unknown.
 */
export function formatCurrency(
  amount: number | null,
  currency: string | null | undefined,
) {
  if (amount === null || amount === undefined) {
    return "—";
  }

  try {
    return new Intl.NumberFormat("en-GB", {
      style: "currency",
      currency: currency || "KES",
      maximumFractionDigits: 0,
    }).format(amount);
  } catch {
    return amount.toLocaleString("en-GB");
  }
}

/**
 * Calculates the paid percentage of a fee balance.
 *
 * Returns `null` when either amount is unknown so the UI can preserve an
 * unknown-progress state. Nonpositive totals report zero progress.
 */
export function getFeeProgress(
  paid: number | null,
  total: number | null,
): number | null {
  if (paid === null || total === null) {
    return null;
  }

  if (total <= 0) {
    return 0;
  }

  return Math.min(100, Math.max(0, Math.round((paid / total) * 100)));
}

export const LEVEL_TONES: Record<LevelKey["tone"], string> = {
  slate: "bg-slate-200",
  lavender: "bg-[#C7B8E8]",
  purple: "bg-[#9B7ED0]",
  indigo: "bg-[#4E3A8A]",
};
