import { describe, expect, it } from "vitest";
import { sanitizeParentTimetableData } from "../timetable-events-source";

const students = [
  { id: "student-1", name: "Amina Wanjiku", initials: "AW", stream: "Blue" },
  { id: "student-2", name: "Brian Otieno", initials: "BO", stream: "Gold" },
];

function period(overrides: Record<string, unknown> = {}) {
  return {
    id: "period-1",
    studentId: "student-1",
    date: "2026-11-16",
    kind: "upcoming",
    subject: "Mathematics",
    teacher: "Ms. Wanjiru",
    room: "Room 4",
    start: "08:00",
    end: "08:40",
    label: "Period 1",
    duration: "40 min",
    ...overrides,
  };
}

function keyDate(overrides: Record<string, unknown> = {}) {
  return {
    id: "date-1",
    kind: "standard",
    month: "Nov",
    day: "20",
    title: "Sports day",
    meta: "Main field",
    badge: "Event",
    ...overrides,
  };
}

describe("sanitizeParentTimetableData", () => {
  it("rejects invalid payloads", () => {
    expect(sanitizeParentTimetableData(null)).toBeNull();
    expect(sanitizeParentTimetableData(undefined)).toBeNull();
    expect(sanitizeParentTimetableData([])).toBeNull();
    expect(sanitizeParentTimetableData("not-json")).toBeNull();
    expect(sanitizeParentTimetableData({})).toBeNull();
    expect(sanitizeParentTimetableData({ students: [] })).toBeNull();
    expect(sanitizeParentTimetableData({ students: [[]] })).toBeNull();
  });

  it("falls back to the first authorized student when the selection is invalid", () => {
    const selected = sanitizeParentTimetableData({
      students,
      selectedStudentId: "student-2",
    });
    expect(selected?.selectedStudentId).toBe("student-2");

    const fallback = sanitizeParentTimetableData({
      students,
      selectedStudentId: "student-not-linked",
    });
    expect(fallback?.selectedStudentId).toBe("student-1");

    const missing = sanitizeParentTimetableData({ students });
    expect(missing?.selectedStudentId).toBe("student-1");
  });

  it("strips periods and key dates that belong to an unlinked student", () => {
    const result = sanitizeParentTimetableData({
      students,
      selectedStudentId: "student-1",
      periods: [
        period(),
        period({ id: "period-2", studentId: "student-2" }),
        period({ id: "period-3", studentId: "student-3" }),
      ],
      keyDates: [
        keyDate({ id: "date-global" }),
        keyDate({ id: "date-student-2", studentId: "student-2" }),
        keyDate({ id: "date-student-3", studentId: "student-3" }),
      ],
    });

    expect(result?.periods.map((item) => item.id)).toEqual([
      "period-1",
      "period-2",
    ]);
    expect(result?.keyDates.map((item) => item.id)).toEqual([
      "date-global",
      "date-student-2",
    ]);
  });

  it("normalizes optional fields and derives missing student initials", () => {
    const result = sanitizeParentTimetableData({
      students: [
        {
          id: " student-1 ",
          name: "  Amina   Wanjiku ",
          initials: "   ",
          stream: "   ",
        },
      ],
      selectedStudentId: " student-1 ",
      periods: [
        period({
          id: " period-1 ",
          studentId: " student-1 ",
          subject: " Mathematics ",
          teacher: 42,
          room: "   ",
          start: " 08:00 ",
          end: null,
          duration: [],
        }),
      ],
      keyDates: [
        keyDate({
          meta: {},
          badge: "  ",
          studentId: " student-1 ",
        }),
      ],
    });

    expect(result?.students[0]).toEqual({
      id: "student-1",
      name: "Amina   Wanjiku",
      initials: "AW",
      stream: null,
    });
    expect(result?.selectedStudentId).toBe("student-1");
    expect(result?.periods[0]).toMatchObject({
      id: "period-1",
      studentId: "student-1",
      subject: "Mathematics",
      teacher: null,
      room: null,
      start: "08:00",
      end: null,
      duration: null,
    });
    expect(result?.keyDates[0]).toMatchObject({
      meta: null,
      badge: null,
      studentId: "student-1",
    });
  });

  it("drops period and key date records with unusable required fields", () => {
    const result = sanitizeParentTimetableData({
      students,
      periods: [
        period({ id: "  " }),
        period({ id: "period-2", studentId: "" }),
        period({ id: "period-3", kind: "unknown" }),
        period({ id: "period-4", subject: " " }),
      ],
      keyDates: [
        keyDate({ kind: "unknown" }),
        keyDate({ id: "date-2", title: " " }),
      ],
    });

    expect(result?.periods).toEqual([]);
    expect(result?.keyDates).toEqual([]);
  });
});
