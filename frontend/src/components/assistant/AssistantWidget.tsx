"use client";

import { useEffect, useRef, useState, type FormEvent } from "react";
import { usePathname } from "next/navigation";
import {
  Bot,
  CalendarCheck,
  CalendarDays,
  CreditCard,
  Mic,
  Send,
  Sparkles,
  X,
} from "lucide-react";
import { useFocusTrap } from "@/hooks/useFocusTrap";
import { cn } from "@/utils/cn";

const CHAT_AVAILABLE = false;

const QUICK_PROMPTS = [
  { id: "fees", label: "Fee balance", icon: CreditCard },
  { id: "attendance", label: "Attendance", icon: CalendarCheck },
  { id: "results", label: "Results", icon: Sparkles },
  { id: "timetable", label: "Timetable", icon: CalendarDays },
];

const HIDDEN_ROUTES = new Set(["/", "/login", "/verify", "/select-role"]);

/**
 * Returns true when the assistant must stay hidden for a route.
 *
 * Public/auth screens should not surface an authenticated assistant, and the
 * clinical nurse area is intentionally excluded from global chat launchers.
 */
export function isAssistantHiddenRoute(pathname: string | null) {
  if (!pathname) {
    return false;
  }

  const [path] = pathname.split(/[?#]/);
  const normalized = path.replace(/\/+$/, "") || "/";

  return (
    HIDDEN_ROUTES.has(normalized) ||
    normalized === "/auth" ||
    normalized.startsWith("/auth/") ||
    normalized === "/nurse" ||
    normalized.startsWith("/nurse/")
  );
}

/** Clears a pending timeout stored in a ref. */
function cancelTimeout(ref: { current: number | null }) {
  if (ref.current !== null) {
    window.clearTimeout(ref.current);
    ref.current = null;
  }
}

/**
 * Global assistant launcher and chat shell.
 *
 * The shell is intentionally empty until a real conversation API is connected:
 * it keeps the context header, suggested prompts and input structure without
 * fabricating messages or learner data.
 */
export default function AssistantWidget() {
  const pathname = usePathname();
  const hidden = isAssistantHiddenRoute(pathname);
  const [open, setOpen] = useState(false);
  const [message, setMessage] = useState("");
  const rootRef = useRef<HTMLDivElement>(null);
  const launcherRef = useRef<HTMLButtonElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const launcherFocusTimeoutRef = useRef<number | null>(null);
  const inputFocusTimeoutRef = useRef<number | null>(null);
  const panelOpen = open && !hidden;

  function closePanel() {
    setOpen(false);
    cancelTimeout(inputFocusTimeoutRef);
    cancelTimeout(launcherFocusTimeoutRef);

    // The launcher stays mounted while the panel is open, so focus returns to
    // it when the focus trap cleans up. Focus it explicitly as a fallback.
    launcherFocusTimeoutRef.current = window.setTimeout(() => {
      launcherFocusTimeoutRef.current = null;
      launcherRef.current?.focus();
    }, 0);
  }

  const panelRef = useFocusTrap<HTMLDivElement>(panelOpen, closePanel);

  useEffect(() => {
    return () => {
      cancelTimeout(inputFocusTimeoutRef);
      cancelTimeout(launcherFocusTimeoutRef);
    };
  }, []);

  // Close on route transitions so body overflow and inert state never linger.
  useEffect(() => {
    setOpen(false);
    cancelTimeout(inputFocusTimeoutRef);
    cancelTimeout(launcherFocusTimeoutRef);
  }, [pathname]);

  useEffect(() => {
    if (!panelOpen) {
      return;
    }

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";

    return () => {
      document.body.style.overflow = previousOverflow;
    };
  }, [panelOpen]);

  useEffect(() => {
    if (!panelOpen) {
      return;
    }

    const root = rootRef.current;

    if (!root) {
      return;
    }

    const background = Array.from(document.body.children).filter((element) => {
      if (element === root || element.contains(root)) {
        return false;
      }

      if (["SCRIPT", "STYLE", "LINK"].includes(element.tagName)) {
        return false;
      }

      return !element.hasAttribute("inert");
    });

    background.forEach((element) => element.setAttribute("inert", ""));

    return () => {
      background.forEach((element) => element.removeAttribute("inert"));
    };
  }, [panelOpen]);

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (!CHAT_AVAILABLE) {
      // Keep the draft until the conversation API can actually receive it.
      return;
    }

    setMessage("");
  }

  function focusInputAtEnd() {
    cancelTimeout(inputFocusTimeoutRef);

    inputFocusTimeoutRef.current = window.setTimeout(() => {
      inputFocusTimeoutRef.current = null;
      const input = inputRef.current;

      if (!input) {
        return;
      }

      input.focus();
      const end = input.value.length;
      input.setSelectionRange(end, end);
    }, 0);
  }

  function applyPrompt(label: string) {
    setMessage((current) => {
      const trimmed = current.trim();

      if (trimmed.length === 0) {
        return label;
      }

      // Replace an existing suggestion, but never discard an unrelated draft.
      if (QUICK_PROMPTS.some((prompt) => prompt.label === trimmed)) {
        return label;
      }

      if (trimmed.endsWith(label)) {
        return current;
      }

      return `${current.replace(/\s+$/, "")} ${label}`;
    });

    focusInputAtEnd();
  }

  if (hidden) {
    return null;
  }

  return (
    <div ref={rootRef}>
      <button
        ref={launcherRef}
        type="button"
        onClick={() => setOpen(true)}
        aria-label="Open assistant chat"
        aria-expanded={open}
        aria-controls="assistant-panel"
        aria-hidden={open || undefined}
        tabIndex={open ? -1 : 0}
        className={cn(
          "fixed bottom-[calc(5rem_+_env(safe-area-inset-bottom,0px))] right-4 z-50 grid h-14 w-14 place-items-center rounded-full bg-brand-500 text-white shadow-xl shadow-brand-900/25 transition hover:bg-brand-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-700 focus-visible:ring-offset-2 lg:bottom-6 lg:right-6 print:hidden",
          !open && "animate-assistant-attention motion-reduce:animate-none",
          open && "pointer-events-none opacity-0",
        )}
      >
        <span
          className={cn(
            "text-[15px] font-black tracking-tight",
            !open && "animate-assistant-logo motion-reduce:animate-none",
          )}
          aria-hidden="true"
        >
          AI
        </span>
        <span className="absolute -right-0.5 -top-0.5 grid h-5 w-5 place-items-center rounded-full bg-white text-brand-600 shadow-sm">
          <Sparkles className="h-3 w-3" aria-hidden="true" />
        </span>
      </button>

      {panelOpen && (
        <>
          <div
            className="fixed inset-0 z-40 bg-slate-950/25 backdrop-blur-[1px] animate-fade-in motion-reduce:animate-none print:hidden"
            aria-hidden="true"
            onClick={closePanel}
          />

          <div
            ref={panelRef}
            id="assistant-panel"
            role="dialog"
            aria-modal="true"
            aria-labelledby="assistant-panel-title"
            aria-describedby="assistant-panel-description"
            tabIndex={-1}
            className="fixed inset-x-3 bottom-[calc(5rem_+_env(safe-area-inset-bottom,0px))] z-50 flex max-h-[calc(100dvh_-_6rem_-_env(safe-area-inset-bottom,0px))] flex-col overflow-hidden rounded-2xl border border-slate-200/80 bg-white shadow-2xl outline-none animate-fade-in motion-reduce:animate-none sm:inset-x-auto sm:right-4 sm:w-[390px] sm:max-h-[78vh] lg:bottom-6 lg:right-6 lg:max-h-[640px] print:hidden"
          >
            <header className="flex items-center justify-between gap-3 bg-gradient-to-br from-[#005E5D] to-[#007A78] px-4 py-3 text-white">
              <div className="flex min-w-0 items-center gap-2.5">
                <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-white/15 text-[13px] font-black tracking-tight">
                  AI
                </span>
                <div className="min-w-0">
                  <h2
                    id="assistant-panel-title"
                    className="truncate text-sm font-bold"
                  >
                    Assistant
                  </h2>
                  <p
                    id="assistant-panel-description"
                    className="text-[9px] font-medium text-white/80"
                  >
                    Preview · chat is not connected yet
                  </p>
                </div>
              </div>
              <button
                type="button"
                onClick={closePanel}
                data-autofocus
                className="grid h-8 w-8 shrink-0 place-items-center rounded-lg bg-white/10 text-white transition hover:bg-white/20 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white"
                aria-label="Close assistant chat"
              >
                <X className="h-4 w-4" aria-hidden="true" />
              </button>
            </header>

            <div className="min-h-0 flex-1 overflow-y-auto bg-[#F4F7F6] p-4">
              <div className="flex h-full min-h-[220px] flex-col items-center justify-center text-center">
                <span className="grid h-14 w-14 place-items-center rounded-2xl bg-[#E2F0ED] text-brand-600 shadow-sm">
                  <Bot className="h-6 w-6" aria-hidden="true" />
                </span>
                <p className="mt-3 text-sm font-bold text-slate-800">
                  How can I help?
                </p>
                <p className="mx-auto mt-1 max-w-[250px] text-[10px] leading-relaxed text-slate-500">
                  Ask about fees, attendance, results or upcoming school
                  events.
                </p>
              </div>
            </div>

            <div className="border-t border-slate-200/70 bg-white px-3 pb-3 pt-2">
              <div
                className="-mx-1 flex gap-1.5 overflow-x-auto px-1 pb-2"
                role="group"
                aria-label="Suggested prompts"
              >
                {QUICK_PROMPTS.map((prompt) => {
                  const Icon = prompt.icon;
                  const pressed = message.trim() === prompt.label;

                  return (
                    <button
                      key={prompt.id}
                      type="button"
                      onClick={() => applyPrompt(prompt.label)}
                      aria-pressed={pressed}
                      className="inline-flex shrink-0 items-center gap-1 rounded-full border border-[#E4DBF5] bg-[#F0EBF9] px-2.5 py-1.5 text-[9px] font-semibold text-[#6D4FA3] transition hover:bg-[#E8DDF8] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600 focus-visible:ring-offset-1"
                    >
                      <Icon className="h-3 w-3" aria-hidden="true" />
                      {prompt.label}
                    </button>
                  );
                })}
              </div>

              <form onSubmit={handleSubmit} className="flex items-center gap-2">
                <div className="flex h-10 min-w-0 flex-1 items-center rounded-xl border border-slate-200/80 bg-slate-50 px-3 focus-within:border-brand-600 focus-within:bg-white focus-within:ring-2 focus-within:ring-brand-600">
                  <input
                    ref={inputRef}
                    type="text"
                    value={message}
                    onChange={(event) => setMessage(event.target.value)}
                    placeholder="Ask a question..."
                    aria-label="Message to assistant"
                    aria-describedby={
                      CHAT_AVAILABLE ? undefined : "assistant-chat-unavailable"
                    }
                    className="h-full min-w-0 flex-1 bg-transparent text-xs text-slate-700 outline-none placeholder:text-slate-400"
                  />
                  <button
                    type="button"
                    disabled
                    aria-label="Voice input unavailable"
                    aria-describedby="assistant-voice-unavailable"
                    title="Voice input is not available yet"
                    className="grid h-7 w-7 shrink-0 cursor-not-allowed place-items-center rounded-lg text-slate-400 opacity-40"
                  >
                    <Mic className="h-4 w-4" aria-hidden="true" />
                  </button>
                </div>

                <button
                  type="submit"
                  disabled={!CHAT_AVAILABLE || message.trim().length === 0}
                  aria-label={
                    CHAT_AVAILABLE ? "Send message" : "Send message unavailable"
                  }
                  aria-describedby={
                    CHAT_AVAILABLE ? undefined : "assistant-chat-unavailable"
                  }
                  title={
                    CHAT_AVAILABLE ? "Send message" : "Chat is not connected yet"
                  }
                  className={cn(
                    "grid h-10 w-10 shrink-0 place-items-center rounded-full transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-600",
                    CHAT_AVAILABLE && message.trim().length > 0
                      ? "bg-brand-500 text-white shadow-lg shadow-brand-500/20 hover:bg-brand-600"
                      : "cursor-not-allowed bg-slate-300 text-slate-500 shadow-none",
                  )}
                >
                  <Send className="h-4 w-4" aria-hidden="true" />
                </button>
              </form>

              {!CHAT_AVAILABLE && (
                <p
                  id="assistant-chat-unavailable"
                  className="mt-2 text-[10px] leading-relaxed text-slate-500"
                >
                  Chat isn&apos;t connected yet. Your message will stay in the
                  box until it is.
                </p>
              )}
              <span id="assistant-voice-unavailable" className="sr-only">
                Voice input is not available yet.
              </span>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
