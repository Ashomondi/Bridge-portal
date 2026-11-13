import PaymentCheckoutDashboard from "@/components/payment-checkout/PaymentCheckoutDashboard";
import { loadPaymentCheckout } from "@/components/payment-checkout/payment-checkout-source";

export const dynamic = "force-dynamic";

export default async function PaymentCheckoutPage() {
  const data = await loadPaymentCheckout();

  return <PaymentCheckoutDashboard data={data} />;
}
