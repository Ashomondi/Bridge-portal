export type AuditSecretsStatus = "masked" | "rotated" | "verified";

export type ConnectionsAuditRow = {
  id: string;
  timestamp: string;
  operator: string;
  initials: string;
  role: string;
  event: string;
  status: AuditSecretsStatus;
  statusLabel: string;
};

export type AdminConnectionsData = {
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
    regionLabel: string | null;
    accountId: string | null;
  };
  whatsapp: {
    connected: boolean;
    receiving: boolean;
    displayName: string | null;
    displayNumber: string | null;
    webhookUrl: string | null;
    tokenRotatedAt: string | null;
    lastTestStatus: string | null;
  } | null;
  daraja: {
    paymentType: "paybill" | "till";
    environment: "sandbox" | "live";
    shortcode: string | null;
    passkeySet: boolean;
    consumerKeySet: boolean;
    consumerSecretSet: boolean;
    lastTestStatus: string | null;
  } | null;
  auditRows: ConnectionsAuditRow[];
};
