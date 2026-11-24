import { act, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import VerifyPage from "../verify/page";

const navigation = vi.hoisted(() => ({ replace: vi.fn() }));

vi.mock("next/navigation", () => ({
  useRouter: () => navigation,
}));

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "content-type": "application/json" },
  });
}

function setChallenge(resendInMs: number) {
  sessionStorage.setItem(
    "pending_otp_challenge",
    JSON.stringify({
      phone: "+254712345678",
      resend_at: Date.now() + resendInMs,
    }),
  );
}

function verifyResponse(memberships: unknown[]) {
  return jsonResponse({
    status: "success",
    data: {
      token: "token-123",
      expires_at: new Date(Date.now() + 3_600_000).toISOString(),
      user: {
        id: "user-1",
        phone_number: "+254712345678",
        full_name: "Test User",
      },
      memberships,
    },
  });
}

function enterCode() {
  for (let index = 1; index <= 6; index += 1) {
    fireEvent.change(screen.getByLabelText(`Digit ${index} of 6`), {
      target: { value: String(index) },
    });
  }
}

async function flushPromises() {
  for (let i = 0; i < 5; i += 1) {
    await act(async () => {
      await Promise.resolve();
    });
  }
}

describe("VerifyPage", () => {
  beforeEach(() => {
    sessionStorage.clear();
    navigation.replace.mockReset();
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-01-01T00:00:00Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it("counts down the stored challenge and formats cooldowns over a minute", () => {
    setChallenge(120_000);

    render(<VerifyPage />);

    expect(screen.getByText(/Send a new code in/)).toHaveTextContent(
      "Send a new code in 2:00",
    );

    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(screen.getByText(/Send a new code in/)).toHaveTextContent(
      "Send a new code in 1:59",
    );
  });

  it("restarts the countdown from a successful resend response", async () => {
    setChallenge(-1000);
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        jsonResponse({
          status: "success",
          data: { message: "OTP sent", cooldown_seconds: 90 },
        }),
      ),
    );

    render(<VerifyPage />);
    fireEvent.click(screen.getByRole("button", { name: /resend code/i }));
    await flushPromises();

    expect(screen.getByText(/Send a new code in/)).toHaveTextContent(
      "Send a new code in 1:30",
    );

    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(screen.getByText(/Send a new code in/)).toHaveTextContent(
      "Send a new code in 1:29",
    );
  });

  it("uses retry_after_seconds when a resend is rate limited", async () => {
    setChallenge(-1000);
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        jsonResponse(
          {
            status: "error",
            error: {
              code: "SENDER_BLOCKED",
              message: "Too many requests. Try again later.",
              details: { retry_after_seconds: 5 },
            },
          },
          429,
        ),
      ),
    );

    render(<VerifyPage />);
    fireEvent.click(screen.getByRole("button", { name: /resend code/i }));
    await flushPromises();

    expect(screen.getByRole("alert")).toHaveTextContent(
      "Please wait 5 seconds",
    );
    expect(screen.getByText(/Send a new code in/)).toHaveTextContent(
      "Send a new code in 0:05",
    );

    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(screen.getByText(/Send a new code in/)).toHaveTextContent(
      "Send a new code in 0:04",
    );
  });

  it("shows school names and routes the selected membership", async () => {
    setChallenge(60_000);
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        verifyResponse([
          {
            id: "membership-1",
            school_id: "school-1",
            school_name: "Nairobi High",
            role: "teacher",
            active: true,
          },
          {
            id: "membership-2",
            school_id: "school-2",
            school_name: "Mombasa Girls",
            role: "parent",
            active: true,
          },
          {
            id: "membership-3",
            school_id: "school-3",
            school_name: null,
            role: "bursar",
            active: true,
          },
        ]),
      ),
    );

    render(<VerifyPage />);
    enterCode();
    fireEvent.click(screen.getByRole("button", { name: /verify/i }));
    await flushPromises();

    expect(
      screen.getByRole("heading", { name: "Choose a school role" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Nairobi High")).toBeInTheDocument();
    expect(screen.getByText("Mombasa Girls")).toBeInTheDocument();
    expect(screen.getByText("school-3")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /Nairobi High/ }));

    expect(navigation.replace).toHaveBeenCalledWith("/teacher");
    expect(
      JSON.parse(sessionStorage.getItem("auth_membership") || "{}"),
    ).toMatchObject({ role: "teacher", school_name: "Nairobi High" });
  });

  it("shows picker errors and returns to sign in", async () => {
    setChallenge(60_000);
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        verifyResponse([
          {
            id: "membership-9",
            school_id: "school-9",
            school_name: "Unknown School",
            role: "driver",
            active: true,
          },
          {
            id: "membership-10",
            school_id: "school-10",
            school_name: "Second School",
            role: "security",
            active: true,
          },
        ]),
      ),
    );

    render(<VerifyPage />);
    enterCode();
    fireEvent.click(screen.getByRole("button", { name: /verify/i }));
    await flushPromises();

    expect(screen.getByRole("alert")).toHaveTextContent(
      "No portal is configured for the roles on this account.",
    );

    fireEvent.click(screen.getByRole("button", { name: /back to sign in/i }));

    expect(sessionStorage.getItem("pending_otp_challenge")).toBeNull();
    expect(navigation.replace).toHaveBeenCalledWith("/login");
  });
});
