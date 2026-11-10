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
