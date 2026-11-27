import type { ComponentProps } from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import SettingsPage from '../page';

const nav = vi.hoisted(() => ({ push: vi.fn(), replace: vi.fn() }));

vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: nav.push, replace: nav.replace }),
  usePathname: () => '/settings',
}));

vi.mock('next/link', () => ({
  default: ({ href, children, ...props }: ComponentProps<'a'>) => (
    <a href={typeof href === 'string' ? href : '#'} {...props}>
      {children}
    </a>
  ),
}));

describe('SettingsPage', () => {
  beforeEach(() => {
    sessionStorage.clear();
    nav.push.mockReset();
    nav.replace.mockReset();
    sessionStorage.setItem(
      'auth_user',
      JSON.stringify({
        id: 'user-1',
        full_name: 'Grace Mwangi',
        phone_number: '+254712345678',
      }),
    );
    sessionStorage.setItem(
      'auth_membership',
      JSON.stringify({
        id: 'membership-teacher',
        school_id: 'school-1',
        school_name: 'Lakeview Secondary',
        role: 'teacher',
        badges: [],
      }),
    );
  });

  it('renders the signed-in account instead of a demo banner', () => {
    render(<SettingsPage />);

    expect(screen.getAllByText('Grace Mwangi').length).toBeGreaterThan(0);
    expect(screen.getByText('Lakeview Secondary')).toBeInTheDocument();
    expect(screen.getByText('+254 7•• ••• 678')).toBeInTheDocument();
    expect(screen.queryByText(/Demo environment/i)).not.toBeInTheDocument();
  });

  it('links to the role switcher from settings', () => {
    render(<SettingsPage />);

    expect(
      screen.getByRole('link', { name: 'Choose another role' }),
    ).toHaveAttribute('href', '/select-role');
  });

  it('clears the stored session on sign out', async () => {
    const user = userEvent.setup();
    render(<SettingsPage />);

    await user.click(screen.getByRole('button', { name: 'Sign out' }));

    expect(sessionStorage.getItem('auth_user')).toBeNull();
    expect(sessionStorage.getItem('auth_membership')).toBeNull();
    expect(nav.push).toHaveBeenCalledWith('/login');
  });
});
