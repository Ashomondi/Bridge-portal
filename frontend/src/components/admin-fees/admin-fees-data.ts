export type FeeMatrixRow = {
  id: string;
  title: string;
  term1: string;
  term2: string;
  term3: string;
  total: string;
};

export type FeeTotals = {
  term1: string;
  term2: string;
  term3: string;
  total: string;
};

export type DistributionSegment = {
  id: string;
  label: string;
  percent: string;
  value: number;
  color: string;
};

export type AdminFeesData = {
  portal: {
    name: string | null;
  } | null;
  session: {
    operatorName: string;
    tier: string;
    initials: string;
  } | null;
  console: {
    version: string;
    environment: string;
    gatewayRef: string | null;
    gatewayLive: boolean;
  } | null;
  activeEntity: string | null;
  counts: {
    schools: number | null;
    connections: number | null;
    unmatchedPayments: number | null;
  };
  header: {
    entityName: string | null;
    description: string | null;
    syncLabel: string | null;
    auditVersion: string | null;
  };
  changesLogged: number | null;
  filters: {
    academicYears: { value: string; label: string }[];
    curriculumTiers: { value: string; label: string }[];
    selectedAcademicYear: string | null;
    selectedCurriculumTier: string | null;
    published: boolean;
    publishedAt: string | null;
  };
  grades: string[];
  selectedGrade: string | null;
  cohort: {
    gradeName: string;
    studentCount: number | null;
    description: string | null;
    ledgerBase: string | null;
  } | null;
  matrix: {
    currency: string;
    rows: FeeMatrixRow[];
    totals: FeeTotals | null;
    dueDates: { term: string; date: string }[];
    verificationNote: string | null;
  };
  distribution: {
    gradeTag: string | null;
    total: string | null;
    segments: DistributionSegment[];
  };
  adjustment: {
    policyNote: string;
  } | null;
};
