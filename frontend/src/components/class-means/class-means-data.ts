export type ClassMeansIcon = "book" | "calculator" | "language" | "science";

export type CbcLevelCode = "EE" | "ME" | "AE" | "BE";

export type CbcLevel = {
  code: CbcLevelCode;
  label: string;
  description: string;
  tone: "mint" | "teal" | "gold" | "rose";
};

/** CBC performance levels used by Kenyan competency-based assessments. */
export const CBC_LEVELS: CbcLevel[] = [
  {
    code: "EE",
    label: "Exceeding Expectation",
    description: "Demonstrates competency beyond the expected level.",
    tone: "mint",
  },
  {
    code: "ME",
    label: "Meeting Expectation",
    description: "Demonstrates the expected competency.",
    tone: "teal",
  },
  {
    code: "AE",
    label: "Approaching Expectation",
    description: "Approaching the expected competency.",
    tone: "gold",
  },
  {
    code: "BE",
    label: "Below Expectation",
    description: "Requires targeted support to reach the expected level.",
    tone: "rose",
  },
];

export type ClassMeansOption = {
  value: string;
  label: string;
};

export type ClassMeansStrand = {
  id: string;
  subject: string;
  icon: ClassMeansIcon;
  score: number | null;
  cbcLevel: CbcLevelCode | null;
  studentCount: number | null;
  trendLabel: string | null;
  trendPoints: number[];
};

export type ClassMeansRow = {
  id: string;
  subject: string;
  assessment: string;
  cbcLevel: CbcLevelCode | null;
  term1: number | null;
  term2: number | null;
  term3: number | null;
  studentCount: number | null;
};

export type ClassMeansData = {
  portal: {
    name: string | null;
    academicTerm: string | null;
  } | null;
  session: {
    fullName: string;
    role: string;
    subject: string | null;
    initials: string;
  } | null;
  frameworkStandard: string | null;
  filters: {
    classStreams: ClassMeansOption[];
    academicPeriods: ClassMeansOption[];
    assessmentSeries: ClassMeansOption[];
    selectedClassStream: string | null;
    selectedAcademicPeriod: string | null;
    selectedAssessmentSeries: string | null;
  };
  sortLabel: string | null;
  strands: ClassMeansStrand[];
  rows: ClassMeansRow[];
};

export function formatMetric(value: number | null) {
  return value === null || value === undefined
    ? "—"
    : value.toLocaleString("en-GB");
}

/** Returns the display metadata for a CBC level code. */
export function getCbcLevel(code: CbcLevelCode | null) {
  return code ? CBC_LEVELS.find((level) => level.code === code) ?? null : null;
}

export function buildSparklinePoints(values: number[]) {
  if (values.length < 2) {
    return null;
  }

  const min = Math.min(...values);
  const max = Math.max(...values);
  const range = max - min || 1;
  const step = 112 / (values.length - 1);

  return values
    .map((value, index) => {
      const x = 4 + step * index;
      const y = 28 - ((value - min) / range) * 24;
      return `${x.toFixed(1)},${y.toFixed(1)}`;
    })
    .join(" ");
}
