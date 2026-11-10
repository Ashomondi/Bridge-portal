export type RollCallStatus = "present" | "absent";

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
    streamName: string | null;
    dateLabel: string | null;
    termLabel: string | null;
    registryRef: string | null;
  };
  students: RollCallStudent[];
};

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
