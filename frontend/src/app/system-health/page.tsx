import { notFound } from 'next/navigation';
import SystemHealthDashboard from '@/components/system-health/SystemHealthDashboard';

/** QA/design-system inspector; intentionally available only in development. */
export default function SystemHealthPage() {
  if (process.env.NODE_ENV !== 'development') {
    notFound();
  }

  return <SystemHealthDashboard />;
}
