"use client";

import { useEffect, useState, type FormEvent } from "react";
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

const QUICK_PROMPTS = [
  { id: "fees", label: "Fee balance", icon: CreditCard },
  { id: "attendance", label: "Attendance", icon: CalendarCheck },
  { id: "results", label: "Results", icon: Sparkles },
  { id: "timetable", label: "Timetable", icon: CalendarDays },
];

/**
 * Global assistant launcher and chat shell.
 *
 * The shell is intentionally empty until a real conversation API is connected:
 * it keeps the context header, suggested prompts and input structure without
 * fabricating messages or learner data.
 */
export default function AssistantWidget() {
  const [open, setOpen] = useState(false);
  const [message, setMessage] = useState("");
  const panelRef = useFocusTrap<HTMLDivElement>(open, () => setOpen(false));

  useEffect(() => {
    if (!open) {
      return;
    }

    document.body.style.overflow = "hidden";

    return () => {
      document.body.style.overflow = "";
    };
  }, [open]);

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setMessage("");
  }

  return (
    <>
      {!open && (
        <button
          type="button"
          onClick={() => setOpen(true)}
          aria-label="Open assistant chat"
          aria-expanded={open}
          aria-controls="assistant-panel"
          className="fixed bottom-20 right-4 z-50 grid h-14 w-14 place-items-center rounded-full bg-brand-500 text-white shadow-xl shadow-brand-900/25 transition hover:scale-105 hover:bg-brand-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/50 focus-visible:ring-offset-2 motion-reduce:transition-none lg:bottom-6 lg:right-6 print:hidden"
        >
          <Bot className="h-6 w-6" aria-hidden="true" />
          <span className="absolute -right-0.5 -top-0.5 grid h-5 w-5 place-items-center rounded-full bg-white text-brand-600 shadow-sm">
            <Sparkles className="h-3 w-3" aria-hidden="true" />
          </span>
        </button>
      )}

      {open && (
        <div
          ref={panelRef}
          id="assistant-panel"
          role="dialog"
          aria-modal="true"
          aria-labelledby="assistant-panel-title"
          tabIndex={-1}
          className="fixed inset-x-3 bottom-20 z-50 flex max-h-[78vh] flex-col overflow-hidden rounded-2xl border border-slate-200/80 bg-white shadow-2xl outline-none animate-fade-in motion-reduce:animate-none sm:inset-x-auto sm:right-4 sm:w-[390px] lg:bottom-6 lg:right-6 lg:max-h-[640px] print:hidden"
        >
          <header className="flex items-center justify-between gap-3 bg-gradient-to-br from-[#005E5D] to-[#007A78] px-4 py-3 text-white">
            <div className="flex min-w-0 items-center gap-2.5">
              <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-white/15">
                <Bot className="h-4 w-4" aria-hidden="true" />
              </span>
              <div className="min-w-0">
                <h2
                  id="assistant-panel-title"
                  className="truncate text-sm font-bold"
                >
                  Assistant
                </h2>
                <p className="text-[9px] font-medium text-white/70">
                  Online · ask about your child&apos;s account
                </p>
              </div>
            </div>
            <button
              type="button"
              onClick={() => setOpen(false)}
              data-autofocus
              className="grid h-8 w-8 shrink-0 place-items-center rounded-lg bg-white/10 text-white transition hover:bg-white/20 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/60"
              aria-label="Close assistant chat"
            >
              <X className="h-4 w-4" />
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
                Ask about fees, attendance, results or upcoming school events.
              </p>
            </div>
          </div>

          <div className="border-t border-slate-200/70 bg-white px-3 pb-3 pt-2">
            <div
              className="-mx-1 flex gap-1.5 overflow-x-auto px-1 pb-2"
              aria-label="Suggested prompts"
            >
              {QUICK_PROMPTS.map((prompt) => {
                const Icon = prompt.icon;

                return (
                  <button
                    key={prompt.id}
                    type="button"
                    onClick={() => setMessage(prompt.label)}
                    className="inline-flex shrink-0 items-center gap-1 rounded-full border border-[#E4DBF5] bg-[#F0EBF9] px-2.5 py-1.5 text-[9px] font-semibold text-[#6D4FA3] transition hover:bg-[#E8DDF8] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
                  >
                    <Icon className="h-3 w-3" aria-hidden="true" />
                    {prompt.label}
                  </button>
                );
              })}
            </div>

            <form onSubmit={handleSubmit} className="flex items-center gap-2">
              <div className="flex h-10 min-w-0 flex-1 items-center rounded-xl border border-slate-200/80 bg-slate-50 px-3 focus-within:border-brand-300 focus-within:bg-white focus-within:ring-4 focus-within:ring-brand-500/10">
                <input
                  type="text"
                  value={message}
                  onChange={(event) => setMessage(event.target.value)}
                  placeholder="Ask a question..."
                  className="h-full min-w-0 flex-1 bg-transparent text-xs text-slate-700 outline-none placeholder:text-slate-400"
                />
                <button
                  type="button"
                  className="grid h-7 w-7 shrink-0 place-items-center rounded-lg text-slate-400 transition hover:bg-slate-100 hover:text-brand-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
                  aria-label="Use voice input"
                >
                  <Mic className="h-4 w-4" aria-hidden="true" />
                </button>
              </div>

              <button
                type="submit"
                className="grid h-10 w-10 shrink-0 place-items-center rounded-full bg-brand-500 text-white shadow-lg shadow-brand-500/20 transition hover:bg-brand-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/40"
                aria-label="Send message"
              >
                <Send className="h-4 w-4" aria-hidden="true" />
              </button>
            </form>
          </div>
        </div>
      )}
    </>
  );
}
