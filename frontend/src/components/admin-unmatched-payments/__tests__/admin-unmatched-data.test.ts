import { describe, expect, it } from "vitest";
import {
  filterUnmatchedTransactions,
  type UnmatchedTransaction,
} from "../admin-unmatched-data";

const transactions: UnmatchedTransaction[] = [
  {
    id: "txn-1",
    status: "unmatched",
    received: "17 Nov 2026 09:00",
    amount: "KES 12,500",
    reference: "QWE123",
    note: "Amina fees",
    phone: "+254700000001",
  },
  {
    id: "txn-2",
    status: "matched",
    received: "17 Nov 2026 09:15",
    amount: "KES 8,000",
    reference: "RTY456",
    note: "Brian boarding",
    phone: "+254700000002",
  },
  {
    id: "txn-3",
    status: "unmatched",
    received: "17 Nov 2026 09:30",
    amount: "KES 3,500",
    reference: "UIO789",
    note: "Unknown payer",
    phone: "+254700000003",
  },
];

describe("filterUnmatchedTransactions", () => {
  it("filters by status tab", () => {
    expect(
      filterUnmatchedTransactions(transactions, "unmatched", "").map(
        (row) => row.id,
      ),
    ).toEqual(["txn-1", "txn-3"]);
    expect(
      filterUnmatchedTransactions(transactions, "matched", "").map(
        (row) => row.id,
      ),
    ).toEqual(["txn-2"]);
    expect(
      filterUnmatchedTransactions(transactions, "all", "").map(
        (row) => row.id,
      ),
    ).toEqual(["txn-1", "txn-2", "txn-3"]);
  });

  it("searches reference, note, phone and amount case-insensitively", () => {
    expect(
      filterUnmatchedTransactions(transactions, "all", "qwe").map(
        (row) => row.id,
      ),
    ).toEqual(["txn-1"]);
    expect(
      filterUnmatchedTransactions(transactions, "all", "BOARDING").map(
        (row) => row.id,
      ),
    ).toEqual(["txn-2"]);
    expect(
      filterUnmatchedTransactions(transactions, "all", "700000003").map(
        (row) => row.id,
      ),
    ).toEqual(["txn-3"]);
    expect(
      filterUnmatchedTransactions(transactions, "all", "3,500").map(
        (row) => row.id,
      ),
    ).toEqual(["txn-3"]);
  });

  it("returns an empty list when nothing matches", () => {
    expect(filterUnmatchedTransactions(transactions, "unmatched", "zzz")).toEqual(
      [],
    );
  });
});
