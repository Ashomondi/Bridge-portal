export type RollCallStatus = "present" | "absent";

export type RollCallError =
  | "unauthorized"
  | "forbidden"
  | "not-found"
  | "class-required"
  | "server"
  | "network"
  | "invalid-response";

export type RollCallStudent = {
  id: string;
  name: string;
  initials?: string;
  admissionNumber: string;
};

export type RollCallData = {
  portal: {
    name: string | null;
  } | null;
  session: {
    fullName: string;
    role: string;
    subject: string | null;
    initials: string;
  } | null;
  context: {
    classId: string | null;
    streamName: string | null;
    dateLabel: string | null;
    termLabel: string | null;
    registryRef: string | null;
  };
  students: RollCallStudent[];
};

/** Builds the roll-call link while preserving an explicit class selection. */
export function getRollCallHref(classId: string | null | undefined) {
  const trimmed = classId?.trim();
  return trimmed
    ? `/teacher/roll-call?classId=${encodeURIComponent(trimmed)}`
    : "/teacher/roll-call";
}

/** Returns the stored initials or derives them from the student's name. */
export function getStudentInitials(student: RollCallStudent) {
  if (student.initials?.trim()) {
    return student.initials.trim();
  }

  const parts = student.name
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2);

  return parts.map((part) => part[0]?.toUpperCase()).join("") || "—";
}
