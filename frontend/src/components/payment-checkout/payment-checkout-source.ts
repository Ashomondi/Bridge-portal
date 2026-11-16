import type { PaymentCheckoutData } from "./payment-checkout-data";

/**
 * Loads the fee balance and verified M-Pesa payment context.
 *
 * Integration point:
 *  - Resolve the guardian, selected child and fee account from the session.
 *  - Load the available amount options and verified M-Pesa number.
 *  - Map the real response into `PaymentCheckoutData`.
 *
 * The loader returns `null` until that context is available. The checkout keeps
 * its layout and renders structured empty states instead of sample amounts or
 * phone numbers.
 */
export async function loadPaymentCheckout(): Promise<PaymentCheckoutData | null> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;

  if (!apiUrl) {
    return null;
  }

  // No authenticated payment request context is wired up yet.
  return null;
}
