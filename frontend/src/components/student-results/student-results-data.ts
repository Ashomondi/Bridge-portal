export type ResultsTerm = {
  id: string;
  label: string;
  released: boolean;
  releaseDate: string | null;
};

export type ResultsStudent = {
  id: string;
  name: string;
  initials: string;
};

export type SubjectResultTerm = "T1" | "T2" | "T3";

export type CbcLevelCode = "EE" | "ME" | "AE" | "BE";

/** CBC levels are ordered from 1 (BE) to 4 (EE) for trend plotting. */
export const CBC_LEVEL_VALUE: Record<CbcLevelCode, number> = {
  BE: 1,
  AE: 2,
  ME: 3,
  EE: 4,
};

export type CbcTrendPoint = {
  term: SubjectResultTerm;
  label: string;
  level: CbcLevelCode | null;
  value: number | null;
  released: boolean;
};

export type SubjectResult = {
  id: string;
  subject: string;
  performanceLevel: string;
  cbcLevel: CbcLevelCode | null;
  corrected: boolean;
  amendmentNote: string | null;
  activeTerms: SubjectResultTerm[];
};

export type AdjustmentNotice = {
  id: string;
  title: string;
  description: string;
  timestamp: string;
  noteHref: string | null;
};

export type TeacherComment = {
  comment: string;
  teacherName: string;
  role: string;
  initials: string;
  attribution: string | null;
  verified: boolean;
  verifiedAt: string | null;
};

export type PerformanceLevel = {
  id: string;
  level: string;
  title: string;
  description: string;
  tone: "slate" | "lavender" | "purple" | "indigo";
};

export type StudentResultsData = {
  portal: {
    name: string | null;
  } | null;
  session: {
    fullName: string;
    relationship: string;
    contact: string | null;
    initials: string;
  } | null;
  students: ResultsStudent[];
  selectedStudentId: string | null;
  terms: ResultsTerm[];
  selectedTermId: string | null;
  results: SubjectResult[];
  trend: CbcTrendPoint[];
  adjustment: AdjustmentNotice | null;
  teacherComment: TeacherComment | null;
  levels: PerformanceLevel[];
};
