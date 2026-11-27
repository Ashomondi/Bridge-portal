import type { ComponentProps } from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import SelectRolePage from '../page';
import type { Membership } from '@/lib/auth';

const nav = vi.hoisted(() => {
  const push = vi.fn();
  const replace = vi.fn();
  return {
    push,
    replace,
    router: { push, replace },
    pathname: '/select-role',
  };
});

vi.mock('next/navigation', () => ({
  useRouter: () => nav.router,
  usePathname: () => nav.pathname,
  useSearchParams: () => new URLSearchParams(),
}));

vi.mock('next/link', () => ({
  default: ({ href, children, ...props }: ComponentProps<'a'>) => (
    <a href={typeof href === 'string' ? href : '#'} {...props}>
      {children}
    </a>
  ),
}));

const parent: Membership = {
  id: 'membership-parent',
  school_id: 'school-1',
  school_name: 'Lakeview Secondary',
  role: 'parent',
  badges: ['guardian'],
};

const teacher: Membership = {
  id: 'membership-teacher',
  school_id: 'school-1',
  school_name: 'Lakeview Secondary',
  role: 'teacher',
  badges: [],
};

describe('SelectRolePage', () => {
  beforeEach(() => {
    sessionStorage.clear();
    nav.push.mockReset();
    nav.replace.mockReset();
  });

  it('renders one card per membership id and switches to the chosen portal', async () => {
    sessionStorage.setItem('auth_memberships', JSON.stringify([parent, teacher]));
    const user = userEvent.setup();

    render(<SelectRolePage />);

    const teacherCard = await screen.findByRole('button', {
      name: 'Lakeview Secondary, Teacher',
    });
    const parentCard = screen.getByRole('button', {
      name: 'Lakeview Secondary, Parent',
    });

    expect(teacherCard).toBeVisible();
    expect(parentCard).toBeVisible();

    await user.click(teacherCard);

    expect(nav.push).toHaveBeenCalledWith('/teacher');
    expect(JSON.parse(sessionStorage.getItem('auth_membership') || '{}')).toEqual(teacher);
  });

  it('falls back to the current membership when the list is missing', async () => {
    sessionStorage.setItem('auth_membership', JSON.stringify(parent));
    const user = userEvent.setup();

    render(<SelectRolePage />);

    const parentCard = await screen.findByRole('button', {
      name: 'Lakeview Secondary, Parent',
    });
    await user.click(parentCard);

    expect(nav.push).toHaveBeenCalledWith('/parent');
    expect(nav.replace).not.toHaveBeenCalledWith('/login');
  });

  it('sends signed-out users back to login', async () => {
    render(<SelectRolePage />);

    await waitFor(() => expect(nav.replace).toHaveBeenCalledWith('/login'));
  });
});
