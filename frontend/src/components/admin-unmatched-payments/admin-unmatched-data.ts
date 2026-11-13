export type UnmatchedTransaction = {
  id: string;
  received: string;
  amount: string;
  reference: string;
  note: string;
  phone: string;
};

export type UnmatchedMetric = {
  id: string;
  label: string;
  value: string | null;
  note: string | null;
};

export type DirectoryCandidate = {
  initials: string;
  name: string;
  admission: string;
  grade: string;
  boarding: string;
  guardianName: string;
  guardianPhone: string;
};

export type LedgerPreview = {
  projectionNote: string;
  cleared: string;
  total: string;
  progress: number;
};

export type AdminUnmatchedData = {
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
    queueHealth: number | null;
  };
  tabs: {
    unmatched: number | null;
    matched: number | null;
    all: number | null;
  };
  filter: {
    termLabel: string | null;
  };
  transactions: UnmatchedTransaction[];
  metrics: UnmatchedMetric[];
  inspector: {
    student: DirectoryCandidate | null;
    ledger: LedgerPreview | null;
    overpaymentWarning: string | null;
    auditReasonRequired: boolean;
  } | null;
  queue: {
    status: string | null;
    allClear: boolean;
  } | null;
};
