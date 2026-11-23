export type UnmatchedTransactionStatus = "unmatched" | "matched";

export type UnmatchedTransaction = {
  id: string;
  status: UnmatchedTransactionStatus;
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

export type PaymentInspectorData = {
  student: DirectoryCandidate | null;
  ledger: LedgerPreview | null;
  overpaymentWarning: string | null;
  auditReasonRequired: boolean;
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
  trust: {
    livePolling: boolean | null;
    signatureVerification: boolean | null;
    assignmentAudit: boolean | null;
  } | null;
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
  /** Inspector data keyed by transaction id so selections never go stale. */
  inspectors: Record<string, PaymentInspectorData>;
  queue: {
    status: string | null;
    allClear: boolean;
  } | null;
};

export type UnmatchedTab = "unmatched" | "matched" | "all";

/** Filters the payment list by tab status and a case-insensitive search query. */
export function filterUnmatchedTransactions(
  transactions: UnmatchedTransaction[],
  tab: UnmatchedTab,
  query: string,
) {
  const needle = query.trim().toLowerCase();

  return transactions.filter((transaction) => {
    const matchesTab = tab === "all" || transaction.status === tab;
    const searchable = [
      transaction.received,
      transaction.amount,
      transaction.reference,
      transaction.note,
      transaction.phone,
    ]
      .join(" ")
      .toLowerCase();

    return matchesTab && (!needle || searchable.includes(needle));
  });
}
