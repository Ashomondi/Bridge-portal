import type { SVGProps } from "react";

type NavIconProps = SVGProps<SVGSVGElement>;

const iconProps = {
  fill: "none",
  stroke: "currentColor",
  strokeWidth: 1.7,
  strokeLinecap: "round" as const,
  strokeLinejoin: "round" as const,
  "aria-hidden": true,
};

export function DashboardNavIcon(props: NavIconProps) {
  return (
    <svg viewBox="0 0 24 24" {...iconProps} {...props}>
      <path d="M4 10.5 12 4l8 6.5" />
      <path d="M6.5 9.5V19a1 1 0 0 0 1 1H10v-5h4v5h2.5a1 1 0 0 0 1-1V9.5" />
    </svg>
  );
}

export function TimetableNavIcon(props: NavIconProps) {
  return (
    <svg viewBox="0 0 24 24" {...iconProps} {...props}>
      <rect x="3.5" y="5" width="17" height="15" rx="2.5" />
      <path d="M8 3.5v3M16 3.5v3M3.5 9.5h17" />
      <path d="M8 13h2M14 13h2M8 16.5h2M14 16.5h2" />
    </svg>
  );
}

export function ClassesNavIcon(props: NavIconProps) {
  return (
    <svg viewBox="0 0 24 24" {...iconProps} {...props}>
      <circle cx="9" cy="8.5" r="3" />
      <path d="M3.5 19.5c.4-3.2 2.6-5 5.5-5s5.1 1.8 5.5 5" />
      <path d="M15.5 6.2a2.8 2.8 0 0 1 0 5.4M17 14.8c2 .4 3.2 1.9 3.5 4.2" />
    </svg>
  );
}

export function ResourcesNavIcon(props: NavIconProps) {
  return (
    <svg viewBox="0 0 24 24" {...iconProps} {...props}>
      <path d="M4 5.5A2.5 2.5 0 0 1 6.5 3H19v16.5H6.5A2.5 2.5 0 0 0 4 22V5.5Z" />
      <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H19" />
      <path d="M8 7h7M8 10h5" />
    </svg>
  );
}

export function MoreNavIcon(props: NavIconProps) {
  return (
    <svg viewBox="0 0 24 24" {...iconProps} {...props}>
      <circle cx="6" cy="12" r="1" fill="currentColor" stroke="none" />
      <circle cx="12" cy="12" r="1" fill="currentColor" stroke="none" />
      <circle cx="18" cy="12" r="1" fill="currentColor" stroke="none" />
      <circle cx="6" cy="6" r="1.4" />
      <circle cx="12" cy="6" r="1.4" />
      <circle cx="18" cy="6" r="1.4" />
      <circle cx="6" cy="18" r="1.4" />
      <circle cx="12" cy="18" r="1.4" />
      <circle cx="18" cy="18" r="1.4" />
    </svg>
  );
}
