"use client";
import { useState } from "react";
import { Send, Clock, AlertCircle, Info } from "lucide-react";

interface Payload {
  user_id: string;
  template_name: string;
  data: { name: string };
  send_at?: string;
}

interface AppError {
  message: string;
  hint?: string;
  type: "network" | "template" | "user" | "rate_limit" | "server";
}

function parseError(rawError: string, statusCode: number): AppError {
  const raw = rawError.toLowerCase();

  if (raw.includes("template") && (raw.includes("not found") || raw.includes("record not found"))) {
    const templateName = rawError.match(/'([^']+)'/)?.[1] ?? "selected template";
    return {
      type: "template",
      message: "Template not found in the database.",
      hint: `'${templateName}' has not been seeded into Postgres. Only 'welcome_email' is seeded by default.`,
    };
  }
  if (raw.includes("user") && raw.includes("not found")) {
    return {
      type: "user",
      message: "User not found in the database.",
      hint: "The demo user ID does not exist in your local Postgres. Check your DB seeds.",
    };
  }
  if (statusCode === 429) {
    return {
      type: "rate_limit",
      message: "Rate limit exceeded.",
      hint: "Your Go API enforces 5 requests/second. Wait a moment and try again.",
    };
  }
  if (raw.includes("idempotency") || raw.includes("duplicate")) {
    return {
      type: "server",
      message: "Duplicate request blocked by idempotency guard.",
      hint: "Each request needs a unique Idempotency-Key. This is handled automatically — try again.",
    };
  }
  return {
    type: "server",
    message: rawError || "An unexpected server error occurred.",
    hint: "Check your Go server terminal for the full stack trace.",
  };
}

export default function ControlPanel({ serverOnline }: { serverOnline: boolean }) {
  const [loading, setLoading] = useState(false);
  const [cooldown, setCooldown] = useState(false);
  const [delay, setDelay] = useState(0);
  const [template, setTemplate] = useState("welcome_email");
  const [appError, setAppError] = useState<AppError | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault();

    if (cooldown || loading) return;

    if (!serverOnline) {
      setAppError({
        type: "network",
        message: "Cannot send: Server is currently offline.",
        hint: "Start your Go backend with: go run cmd/api/main.go",
      });
      return;
    }

    setLoading(true);
    setAppError(null);
    setSuccessMsg(null);

    const payload: Payload = {
      user_id: "11111111-1111-1111-1111-111111111111",
      template_name: template,
      data: { name: "Jane Doe" },
    };

    if (delay > 0) {
      payload.send_at = new Date(Date.now() + delay * 1000).toISOString();
    }

    try {
      const res = await fetch("http://localhost:8080/notification", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": crypto.randomUUID(),
        },
        body: JSON.stringify(payload),
      });

      if (!res.ok) {
        const errData = await res.json().catch(() => ({ error: res.statusText }));
        setAppError(parseError(errData.error ?? "", res.status));
        setCooldown(true);
        setTimeout(() => setCooldown(false), 3000);
      } else {
        setSuccessMsg(
          delay > 0
            ? `Scheduled! Task will fire in ${delay}s via Redis ZSET.`
            : "Notification fired successfully!"
        );
        setCooldown(true);
        setTimeout(() => {
          setSuccessMsg(null);
          setCooldown(false);
        }, 4000);
      }
    } catch {
      setAppError({
        type: "network",
        message: "Network error: Could not reach the Go API.",
        hint: "Ensure your Go server is running on port 8080.",
      });
      setCooldown(true);
      setTimeout(() => setCooldown(false), 3000);
    }

    setLoading(false);
  };

  return (
    <div className="bg-neutral-900/50 border border-white/10 rounded-2xl p-6 backdrop-blur-xl">
      <h2 className="text-xl font-semibold mb-6 flex items-center">
        <Send className="w-5 h-5 mr-2 text-cyan-400" />
        Trigger Event
      </h2>

      {/* Server offline warning */}
      {!serverOnline && !appError && (
        <div className="mb-4 p-3 rounded-lg bg-yellow-500/10 border border-yellow-500/50 flex items-start text-yellow-400 text-sm">
          <AlertCircle className="w-4 h-4 mr-2 mt-0.5 shrink-0" />
          <span>Waiting for Go Server connection...</span>
        </div>
      )}

      {/* Smart error display with hint */}
      {appError && (
        <div className="mb-4 rounded-lg bg-red-500/10 border border-red-500/40 overflow-hidden">
          <div className="flex items-start p-3 text-red-400 text-sm">
            <AlertCircle className="w-4 h-4 mr-2 mt-0.5 shrink-0" />
            <span className="font-medium">{appError.message}</span>
          </div>
          {appError.hint && (
            <div className="flex items-start px-3 pb-3 text-neutral-400 text-xs border-t border-red-500/20 pt-2">
              <Info className="w-3 h-3 mr-1.5 mt-0.5 shrink-0 text-neutral-500" />
              <span>{appError.hint}</span>
            </div>
          )}
        </div>
      )}

      {/* Success message */}
      {successMsg && (
        <div className="mb-4 p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/50 text-emerald-400 text-sm">
          {successMsg}
        </div>
      )}

      <form onSubmit={handleSend} className="space-y-5">
        <div>
          <label className="block text-sm font-medium text-neutral-400 mb-1">
            Target Database User
          </label>
          <input
            disabled
            value="11111111-1111-1111... (John Doe)"
            className="w-full bg-neutral-950 border border-white/5 rounded-xl p-3 text-sm text-neutral-500 font-mono"
            readOnly
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-neutral-400 mb-1">
            Select Template
          </label>
          <select
            value={template}
            onChange={(e) => setTemplate(e.target.value)}
            className="w-full bg-neutral-950 border border-white/10 rounded-xl p-3 text-sm text-white focus:ring-2 focus:ring-cyan-500 outline-none transition-all"
          >
            <option value="welcome_email">Welcome Email (Routes to Email)</option>
            <option value="password_reset">Password Reset (Routes to Email)</option>
            <option value="order_shipped">Order Shipped (Routes to Email)</option>
            <option value="abandoned_cart">Abandoned Cart (Routes to Email & SMS)</option>
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-neutral-400 mb-2 flex items-center">
            <Clock className="w-4 h-4 mr-1 text-rose-400" />
            Time Travel Delay (Redis ZSET)
          </label>
          <input
            type="range"
            min="0"
            max="60"
            value={delay}
            onChange={(e) => setDelay(parseInt(e.target.value))}
            className="w-full accent-cyan-500"
          />
          <div className="text-right text-xs text-neutral-500 mt-1">
            {delay === 0 ? "Execute Instantly" : `Hold in Redis ZSET for ${delay}s`}
          </div>
        </div>

        <button
          type="submit"
          disabled={loading || cooldown || !serverOnline}
          className={`w-full mt-4 font-medium py-3 rounded-xl transition-all ${
            !serverOnline || cooldown || loading
              ? "bg-neutral-800 text-neutral-500 cursor-not-allowed border border-white/5"
              : "bg-gradient-to-r from-cyan-500 to-blue-500 hover:from-cyan-400 hover:to-blue-400 text-white shadow-[0_0_20px_rgba(6,182,212,0.15)] hover:shadow-[0_0_30px_rgba(6,182,212,0.3)] active:scale-95"
          }`}
        >
          {loading ? "Firing..." : cooldown ? "Cooldown..." : !serverOnline ? "Server Offline" : "Fire Notification"}
        </button>
      </form>
    </div>
  );
}
