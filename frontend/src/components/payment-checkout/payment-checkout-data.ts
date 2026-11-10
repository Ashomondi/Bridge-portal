export type PaymentOption = {
  id: string;
  label: string;
  amount: string;
  recommended: boolean;
  custom: boolean;
};

/**
 * Structural checkout choices. Replace these with the amount options returned
 * by the fee ledger once the payment API is connected.
 */
export const PAYMENT_OPTIONS: PaymentOption[] = [
  {
    id: "option-full",
    label: "[Option]",
    amount: "[Amount]",
    recommended: false,
    custom: false,
  },
  {
    id: "option-recommended",
    label: "[Option]",
    amount: "[Amount]",
    recommended: true,
    custom: false,
  },
  {
    id: "option-custom",
    label: "[Custom Amount]",
    amount: "",
    recommended: false,
    custom: true,
  },
];

export const DEFAULT_PAYMENT_OPTION = "option-recommended";
