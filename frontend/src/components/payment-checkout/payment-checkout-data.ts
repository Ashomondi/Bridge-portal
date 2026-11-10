export type PaymentOption = {
  id: string;
  label: string;
  amount: number | null;
  currency: string;
  recommended: boolean;
  custom: boolean;
};

export type PaymentPhone = {
  countryCode: string;
  number: string;
  verified: boolean;
};

export type PaymentCheckoutData = {
  portal: {
    name: string | null;
  } | null;
  session: {
    fullName: string;
    relationship: string;
    initials: string;
  } | null;
  student: {
    name: string;
    admissionNumber: string;
  } | null;
  options: PaymentOption[];
  defaultOptionId: string | null;
  phone: PaymentPhone | null;
  overpaymentNote: string | null;
};

export function formatAmount(
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
 * Structural fallback options so the checkout layout remains visible before
 * real fee options are loaded. These are placeholders, not learner data.
 */
export const PLACEHOLDER_PAYMENT_OPTIONS: PaymentOption[] = [
  {
    id: "placeholder-full",
    label: "[Option]",
    amount: null,
    currency: "KES",
    recommended: false,
    custom: false,
  },
  {
    id: "placeholder-recommended",
    label: "[Option]",
    amount: null,
    currency: "KES",
    recommended: true,
    custom: false,
  },
  {
    id: "placeholder-custom",
    label: "[Custom Amount]",
    amount: null,
    currency: "KES",
    recommended: false,
    custom: true,
  },
];
