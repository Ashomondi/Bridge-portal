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

export type SubjectResult = {
  id: string;
  subject: string;
  performanceLevel: string;
  corrected: boolean;
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
  verified: boolean;
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
  adjustment: AdjustmentNotice | null;
  teacherComment: TeacherComment | null;
  levels: PerformanceLevel[];
};
