import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import AssistantWidget, { isAssistantHiddenRoute } from "../AssistantWidget";

const routeState = vi.hoisted(() => ({ pathname: "/parent" }));

vi.mock("next/navigation", () => ({
  usePathname: () => routeState.pathname,
}));

function renderWidget() {
  return render(<AssistantWidget />);
}

describe("isAssistantHiddenRoute", () => {
  it.each([
    "/",
    "/login",
    "/verify",
    "/select-role",
    "/auth",
    "/auth/callback",
    "/nurse",
    "/nurse/records",
  ])("hides the widget on %s", (pathname) => {
    expect(isAssistantHiddenRoute(pathname)).toBe(true);
  });

  it.each([
    "/parent",
    "/parent/timetable",
    "/teacher/overview",
    "/admin",
    "/bursar/payments",
  ])("shows the widget on %s", (pathname) => {
    expect(isAssistantHiddenRoute(pathname)).toBe(false);
  });

  it("handles trailing slashes, queries and empty values", () => {
    expect(isAssistantHiddenRoute("/login/")).toBe(true);
    expect(isAssistantHiddenRoute("/nurse/records/?tab=vitals")).toBe(true);
    expect(isAssistantHiddenRoute("/parent/")).toBe(false);
    expect(isAssistantHiddenRoute(null)).toBe(false);
  });
});

describe("AssistantWidget", () => {
  beforeEach(() => {
    routeState.pathname = "/parent";
  });

  it("renders no launcher on hidden routes", () => {
    routeState.pathname = "/login";

    renderWidget();

    expect(
      screen.queryByRole("button", { name: "Open assistant chat" }),
    ).not.toBeInTheDocument();
  });

  it("closes the open panel and restores page scrolling on route transitions", async () => {
    const user = userEvent.setup();
    const { rerender } = renderWidget();

    await user.click(
      screen.getByRole("button", { name: "Open assistant chat" }),
    );

    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(document.body.style.overflow).toBe("hidden");

    routeState.pathname = "/login";
    rerender(<AssistantWidget />);

    await waitFor(() => {
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
    await waitFor(() => {
      expect(document.body.style.overflow).not.toBe("hidden");
    });
    expect(
      screen.queryByRole("button", { name: "Open assistant chat" }),
    ).not.toBeInTheDocument();
  });

  it("exposes dialog ARIA and focuses the close control when opened", async () => {
    const user = userEvent.setup();
    renderWidget();

    await user.click(
      screen.getByRole("button", { name: "Open assistant chat" }),
    );

    const dialog = screen.getByRole("dialog");
    const closeButton = screen.getByRole("button", {
      name: "Close assistant chat",
    });

    expect(dialog).toHaveAttribute("aria-modal", "true");
    expect(dialog).toHaveAttribute(
      "aria-labelledby",
      "assistant-panel-title",
    );
    expect(dialog).toHaveAttribute(
      "aria-describedby",
      "assistant-panel-description",
    );
    expect(
      screen.getByRole("textbox", { name: "Message to assistant" }),
    ).toBeInTheDocument();

    await waitFor(() => {
      expect(closeButton).toHaveFocus();
    });
  });

  it("closes on Escape and returns focus to the launcher", async () => {
    const user = userEvent.setup();
    renderWidget();
    const launcher = screen.getByRole("button", {
      name: "Open assistant chat",
    });

    await user.click(launcher);
    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: "Close assistant chat" }),
      ).toHaveFocus();
    });

    await user.keyboard("{Escape}");

    await waitFor(() => {
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
    await waitFor(() => {
      expect(launcher).toHaveFocus();
    });
  });

  it("inerts eligible background elements and restores only what it changed", async () => {
    const sibling = document.createElement("div");
    const alreadyInert = document.createElement("div");
    const script = document.createElement("script");

    alreadyInert.setAttribute("inert", "");
    document.body.append(sibling, alreadyInert, script);

    const user = userEvent.setup();
    renderWidget();

    await user.click(
      screen.getByRole("button", { name: "Open assistant chat" }),
    );

    expect(sibling).toHaveAttribute("inert");
    expect(alreadyInert).toHaveAttribute("inert");
    expect(script).not.toHaveAttribute("inert");

    await user.keyboard("{Escape}");

    await waitFor(() => {
      expect(sibling).not.toHaveAttribute("inert");
    });
    expect(alreadyInert).toHaveAttribute("inert");

    sibling.remove();
    alreadyInert.remove();
    script.remove();
  });

  it("applies a prompt, focuses the input and keeps aria-pressed in sync", async () => {
    const user = userEvent.setup();
    renderWidget();

    await user.click(
      screen.getByRole("button", { name: "Open assistant chat" }),
    );

    const input = screen.getByRole("textbox", {
      name: "Message to assistant",
    }) as HTMLInputElement;
    const prompt = screen.getByRole("button", { name: "Fee balance" });

    await user.click(prompt);

    expect(input).toHaveValue("Fee balance");
    expect(prompt).toHaveAttribute("aria-pressed", "true");
    await waitFor(() => {
      expect(input).toHaveFocus();
    });
    expect(input.selectionStart).toBe("Fee balance".length);
    expect(input.selectionEnd).toBe("Fee balance".length);
  });

  it("preserves an unrelated typed draft when a prompt is selected", async () => {
    const user = userEvent.setup();
    renderWidget();

    await user.click(
      screen.getByRole("button", { name: "Open assistant chat" }),
    );

    const input = screen.getByRole("textbox", {
      name: "Message to assistant",
    }) as HTMLInputElement;

    await user.type(input, "My custom question");
    await user.click(screen.getByRole("button", { name: "Attendance" }));

    expect(input).toHaveValue("My custom question Attendance");
    await waitFor(() => {
      expect(input).toHaveFocus();
    });
    expect(input.selectionStart).toBe(input.value.length);
    expect(input.selectionEnd).toBe(input.value.length);
  });
});
