export type OnboardingStepKind = "completed" | "api" | "progress" | "pending";

export type OnboardingStep = {
  id: string;
  kind: OnboardingStepKind;
  title: string;
  meta: string | null;
  timestamp: string | null;
  action: string;
};

export type MetricKind = "latency" | "health" | "velocity";

export type AdminMetric = {
  id: string;
  label: string;
  value: string | null;
  note: string | null;
  kind: MetricKind;
};

export type RegistryStatus = "completed" | "live" | "initiated";

export type AdminRegistryRow = {
  id: string;
  school: string;
  county: string;
  type: string;
  principal: string;
  status: RegistryStatus;
  statusLabel: string;
  progress: string | null;
  created: string;
};

export type AdminOnboardingData = {
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
    syncId: string | null;
    sessionLabel: string | null;
  } | null;
  activeEntity: string | null;
  counts: {
    schools: number | null;
    connections: number | null;
    unmatchedPayments: number | null;
  };
  header: {
    enrolledCount: number | null;
    regionOptions: { value: string; label: string }[];
  };
  onboarding: {
    streamLabel: string;
    entityName: string;
    entityMeta: string | null;
    steps: OnboardingStep[];
  } | null;
  metrics: AdminMetric[];
  registry: {
    rows: AdminRegistryRow[];
    visibleCount: number;
    totalCount: number;
    countyCount: number;
    totalPages: number;
    currentPage: number;
  };
};
