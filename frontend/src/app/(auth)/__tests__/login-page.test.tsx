import { act, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import LoginPage from "../login/page";

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

async function flushPromises() {
  for (let i = 0; i < 5; i += 1) {
    await act(async () => {
      await Promise.resolve();
    });
  }
}

describe("LoginPage", () => {
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

  it("starts a backend retry-after cooldown after a 429 and counts it down", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        jsonResponse(
          {
            status: "error",
            error: {
              code: "OTP_COOLDOWN",
              message: "Please wait before requesting another OTP.",
              details: { retry_after_seconds: 2 },
            },
          },
          429,
        ),
      ),
    );

    render(<LoginPage />);
    fireEvent.change(screen.getByLabelText(/mobile number/i), {
      target: { value: "712345678" },
    });
    fireEvent.click(screen.getByRole("button", { name: /send code/i }));
    await flushPromises();

    expect(screen.getByRole("alert")).toHaveTextContent(
      "Please wait 2 seconds",
    );
    expect(screen.getByText("Resend available in 2s")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /send code/i })).toBeDisabled();

    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(screen.getByText("Resend available in 1s")).toBeInTheDocument();

    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(screen.getByText("You can ask for a new code once a minute.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /send code/i })).toBeEnabled();
  });

  it("uses the cooldown returned by a successful OTP request", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        jsonResponse({
          status: "success",
          data: { message: "OTP sent", cooldown_seconds: 90 },
        }),
      ),
    );

    render(<LoginPage />);
    fireEvent.change(screen.getByLabelText(/mobile number/i), {
      target: { value: "712345678" },
    });
    fireEvent.click(screen.getByRole("button", { name: /send code/i }));
    await flushPromises();

    expect(navigation.replace).toHaveBeenCalledWith("/verify");
    const challenge = JSON.parse(
      sessionStorage.getItem("pending_otp_challenge") || "{}",
    );
    expect(challenge).toEqual({
      phone: "+254712345678",
      resend_at: Date.now() + 90_000,
    });
  });
});
