import { describe, expect, it } from "vitest";
import {
  filterRosterStudents,
  formatCountLabel,
  paginateRoster,
  resolveVisibleStudent,
  type RosterStudent,
} from "./class-roster-data";

function makeStudent(
  id: string,
  overrides: Partial<RosterStudent> = {},
): RosterStudent {
  return {
    id,
    admissionNumber: `ADM-${id}`,
    name: `Student ${id}`,
    house: "Kilimanjaro",
    gradeSection: "Form 3A",
    linkedGuardians: 1,
    guardianStatus: "verified",
    guardians: [],
    attendance: {
      overallPercent: 95,
      totalDays: 20,
      daysPresent: 19,
      daysAbsent: 1,
      absenceNoteStatus: null,
    },
    enrollmentType: "Boarder",
    ...overrides,
  };
}

const students = [
  makeStudent("1", { name: "Amina Wanjiku", house: "Kenya" }),
  makeStudent("2", {
    name: "Brian Otieno",
    guardianStatus: "pending",
    attendance: { ...makeStudent("x").attendance, overallPercent: 89 },
  }),
  makeStudent("3", {
    name: "Cynthia Mwangi",
    attendance: { ...makeStudent("x").attendance, overallPercent: null },
  }),
];

describe("filterRosterStudents", () => {
  it("matches name, admission number, house and grade section", () => {
    expect(
      filterRosterStudents(students, {
        query: "amina",
        guardianFilter: "all",
        attendanceFilter: "all",
      }).map((student) => student.id),
    ).toEqual(["1"]);

    expect(
      filterRosterStudents(students, {
        query: "ADM-2",
        guardianFilter: "all",
        attendanceFilter: "all",
      }).map((student) => student.id),
    ).toEqual(["2"]);

    expect(
      filterRosterStudents(students, {
        query: "kenya",
        guardianFilter: "all",
        attendanceFilter: "all",
      }).map((student) => student.id),
    ).toEqual(["1"]);
  });

  it("filters by guardian status", () => {
    expect(
      filterRosterStudents(students, {
        query: "",
        guardianFilter: "pending",
        attendanceFilter: "all",
      }).map((student) => student.id),
    ).toEqual(["2"]);
  });

  it("supports attendance bands including not recorded", () => {
    expect(
      filterRosterStudents(students, {
        query: "",
        guardianFilter: "all",
        attendanceFilter: "above",
      }).map((student) => student.id),
    ).toEqual(["1"]);

    expect(
      filterRosterStudents(students, {
        query: "",
        guardianFilter: "all",
        attendanceFilter: "below",
      }).map((student) => student.id),
    ).toEqual(["2"]);

    expect(
      filterRosterStudents(students, {
        query: "",
        guardianFilter: "all",
        attendanceFilter: "not-recorded",
      }).map((student) => student.id),
    ).toEqual(["3"]);
  });
});

describe("paginateRoster", () => {
  const many = Array.from({ length: 12 }, (_, index) =>
    makeStudent(`${index + 1}`),
  );

  it("returns a bounded page window", () => {
    const page = paginateRoster(many, 2);

    expect(page.totalPages).toBe(3);
    expect(page.safePage).toBe(2);
    expect(page.startIndex).toBe(5);
    expect(page.pageRows.map((student) => student.id)).toEqual([
      "6",
      "7",
      "8",
      "9",
      "10",
    ]);
  });

  it("clamps out-of-range pages", () => {
    expect(paginateRoster(many, 99).safePage).toBe(3);
    expect(paginateRoster(many, -4).safePage).toBe(1);
  });

  it("returns a safe empty page", () => {
    expect(paginateRoster([], 1)).toMatchObject({
      totalPages: 1,
      safePage: 1,
      startIndex: 0,
      pageRows: [],
    });
  });
});

describe("resolveVisibleStudent", () => {
  it("keeps a selected learner when visible", () => {
    expect(resolveVisibleStudent(students, "2")?.id).toBe("2");
  });

  it("falls back to the first visible learner", () => {
    expect(resolveVisibleStudent(students, "missing")?.id).toBe("1");
  });

  it("returns null for an empty page", () => {
    expect(resolveVisibleStudent([], "1")).toBeNull();
  });

  it("keeps details aligned after a filter or page change", () => {
    const many = Array.from({ length: 12 }, (_, index) =>
      makeStudent(`${index + 1}`),
    );
    const firstPage = paginateRoster(many, 1);
    expect(resolveVisibleStudent(firstPage.pageRows, "10")?.id).toBe("1");
  });
});

describe("formatCountLabel", () => {
  it("uses singular and plural nouns", () => {
    expect(formatCountLabel(1, "student", "students")).toBe("1 student");
    expect(formatCountLabel(2, "student", "students")).toBe("2 students");
    expect(formatCountLabel(1, "linked guardian", "linked guardians")).toBe(
      "1 linked guardian",
    );
  });
});
