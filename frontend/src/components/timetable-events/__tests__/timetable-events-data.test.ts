import { describe, expect, it } from "vitest";
import { buildWeek, SCHOOL_TIME_ZONE } from "../timetable-events-data";

describe("buildWeek", () => {
  it("uses the Nairobi calendar date at the UTC midnight boundary", () => {
    // 21:30 UTC on Sunday 15 November is 00:30 on Monday 16 November in Nairobi.
    const week = buildWeek(new Date("2026-11-15T21:30:00.000Z"));

    expect(week.map((day) => day.id)).toEqual([
      "2026-11-16",
      "2026-11-17",
      "2026-11-18",
      "2026-11-19",
      "2026-11-20",
    ]);
    expect(week[0].isToday).toBe(true);
    expect(week.filter((day) => day.isToday)).toHaveLength(1);
  });

  it("keeps a late-night Nairobi date on the same school day", () => {
    // 20:59 UTC is 23:59 on Monday 16 November in Nairobi.
    const week = buildWeek(new Date("2026-11-16T20:59:00.000Z"));

    expect(week[0].id).toBe("2026-11-16");
    expect(week[0].isToday).toBe(true);
  });

  it("rolls over to Tuesday at Nairobi midnight", () => {
    // 21:00 UTC on Monday is 00:00 on Tuesday 17 November in Nairobi.
    const week = buildWeek(new Date("2026-11-16T21:00:00.000Z"));

    expect(week[1].id).toBe("2026-11-17");
    expect(week[1].isToday).toBe(true);
  });

  it("maps Sunday back to the preceding Monday-to-Friday week", () => {
    const week = buildWeek(new Date("2026-11-22T10:00:00.000Z"));

    expect(week.map((day) => day.id)).toEqual([
      "2026-11-16",
      "2026-11-17",
      "2026-11-18",
      "2026-11-19",
      "2026-11-20",
    ]);
    expect(week.some((day) => day.isToday)).toBe(false);
  });

  it("returns an empty week for an invalid date", () => {
    expect(buildWeek(new Date(Number.NaN))).toEqual([]);
  });

  it("falls back to the school timezone for an invalid timezone", () => {
    expect(() =>
      buildWeek(new Date("2026-11-15T21:30:00.000Z"), "Not/A-Timezone"),
    ).not.toThrow();
    expect(buildWeek(new Date("2026-11-15T21:30:00.000Z"), "")[0]?.id).toBe(
      "2026-11-16",
    );
    expect(SCHOOL_TIME_ZONE).toBe("Africa/Nairobi");
  });
});
