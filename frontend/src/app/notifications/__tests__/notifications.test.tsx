import type { ComponentProps } from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import NotificationsPage from '../page';

const nav = vi.hoisted(() => ({ pathname: '/notifications' }));

vi.mock('next/navigation', () => ({
  usePathname: () => nav.pathname,
  useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
}));

vi.mock('next/link', () => ({
  default: ({ href, children, ...props }: ComponentProps<'a'>) => (
    <a href={typeof href === 'string' ? href : '#'} {...props}>
      {children}
    </a>
  ),
}));

describe('NotificationsPage', () => {
  beforeEach(() => {
    nav.pathname = '/notifications';
  });

  it('filters unread notifications and hides read ones', async () => {
    const user = userEvent.setup();
    render(<NotificationsPage />);

    expect(screen.getByText('Term 1 results are ready')).toBeInTheDocument();
    expect(screen.getByText('Amani was marked absent')).toBeInTheDocument();

    await user.click(screen.getByRole('tab', { name: /Unread/ }));

    expect(screen.getByText('Term 1 results are ready')).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.queryByText('Amani was marked absent')).not.toBeInTheDocument();
    });
  });

  it('marks everything as read and shows the unread empty state', async () => {
    const user = userEvent.setup();
    render(<NotificationsPage />);

    await user.click(screen.getByRole('button', { name: 'Mark all as read' }));
    expect(screen.getByRole('button', { name: 'Mark all as read' })).toBeDisabled();

    await user.click(screen.getByRole('tab', { name: /Unread/ }));

    expect(screen.getByText('You have no unread notifications.')).toBeInTheDocument();
  });

  it('derives the active bottom navigation item from the pathname', () => {
    render(<NotificationsPage />);

    expect(screen.getByRole('link', { name: 'Alerts' })).toHaveAttribute(
      'aria-current',
      'page',
    );
  });
});
