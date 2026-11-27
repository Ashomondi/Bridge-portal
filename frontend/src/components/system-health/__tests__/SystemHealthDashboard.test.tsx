import type { ComponentProps } from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import SystemHealthDashboard from '../SystemHealthDashboard';

vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
  usePathname: () => '/system-health',
}));

vi.mock('next/link', () => ({
  default: ({ href, children, ...props }: ComponentProps<'a'>) => (
    <a href={typeof href === 'string' ? href : '#'} {...props}>
      {children}
    </a>
  ),
}));

describe('SystemHealthDashboard', () => {
  it('switches previews from the accessible state matrix buttons', async () => {
    const user = userEvent.setup();
    render(<SystemHealthDashboard />);

    expect(
      screen.getByRole('heading', { name: "We can't find that page" }),
    ).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: /403 Forbidden/ }));

    expect(
      screen.getByRole('heading', { level: 3, name: '403 Forbidden' }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole('tab', { name: 'Frame 2 • 403' }),
    ).toHaveAttribute('aria-selected', 'true');
  });

  it('exposes the frame switcher as an accessible tab list', async () => {
    const user = userEvent.setup();
    render(<SystemHealthDashboard />);

    expect(screen.getByRole('tablist', { name: 'System state frames' })).toBeInTheDocument();

    await user.click(screen.getByRole('tab', { name: 'Frame 4 • Offline' }));

    expect(screen.getByRole('heading', { level: 3, name: 'Network Disconnect' })).toBeInTheDocument();
  });
});
