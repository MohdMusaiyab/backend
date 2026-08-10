"use client";
import { useState } from "react";
import { Send, Clock, AlertCircle } from "lucide-react";

interface Payload {
  user_id: string;
  template_name: string;
  data: { name: string };
  send_at?: string;
}

export default function ControlPanel({ serverOnline }: { serverOnline: boolean }) {
  const [loading, setLoading] = useState(false);
  const [delay, setDelay] = useState(0);
  const [template, setTemplate] = useState("welcome_email");
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!serverOnline) {
      setErrorMsg("Cannot send: Server is currently offline.");
      return;
    }
    
    setLoading(true);
    setErrorMsg(null);
    setSuccessMsg(null);

    const payload: Payload = {
      user_id: "11111111-1111-1111-1111-111111111111",
      template_name: template,
      data: { name: "Jane Doe" },
    };

    if (delay > 0) {
      const futureDate = new Date(Date.now() + delay * 1000);
      payload.send_at = futureDate.toISOString();
    }

    try {
      const res = await fetch("http://localhost:8080/notification", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": "ui-req-" + Date.now(),
        },
        body: JSON.stringify(payload),
      });
      
      if (!res.ok) {
        // Handle server errors (e.g. Postgres DB connection issues)
        const errData = await res.json().catch(() => ({}));
        setErrorMsg(`Server Error: ${errData.error || res.statusText || "Database or internal error"}`);
      } else {
        setSuccessMsg("Event triggered successfully!");
        setTimeout(() => setSuccessMsg(null), 3000);
      }
    } catch (err) {
      setErrorMsg("Network error: Could not connect to the Go API.");
      console.error(err);
    }

    setLoading(false);
  };

  return (
    <div className="bg-neutral-900/50 border border-white/10 rounded-2xl p-6 backdrop-blur-xl">
      <h2 className="text-xl font-semibold mb-6 flex items-center">
        <Send className="w-5 h-5 mr-2 text-cyan-400" />
        Trigger Event
      </h2>

      {errorMsg && (
        <div className="mb-4 p-3 rounded-lg bg-red-500/10 border border-red-500/50 flex items-start text-red-400 text-sm">
          <AlertCircle className="w-4 h-4 mr-2 mt-0.5 shrink-0" />
          <span>{errorMsg}</span>
        </div>
      )}
      
      {successMsg && (
        <div className="mb-4 p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/50 text-emerald-400 text-sm">
          {successMsg}
        </div>
      )}
      
      {!serverOnline && !errorMsg && (
        <div className="mb-4 p-3 rounded-lg bg-yellow-500/10 border border-yellow-500/50 flex items-start text-yellow-400 text-sm">
          <AlertCircle className="w-4 h-4 mr-2 mt-0.5 shrink-0" />
          <span>Waiting for Go Server connection...</span>
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
            <option value="abandoned_cart">Abandoned Cart (Routes to SMS &amp; Email)</option>
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
            {delay === 0 ? "Execute Instantly" : `Hold in Queue for ${delay} seconds`}
          </div>
        </div>

        <button
          type="submit"
          disabled={loading || !serverOnline}
          className={`w-full mt-4 font-medium py-3 rounded-xl transition-all shadow-[0_0_20px_rgba(6,182,212,0.15)] ${
            !serverOnline 
              ? "bg-neutral-800 text-neutral-500 cursor-not-allowed border border-white/5" 
              : "bg-gradient-to-r from-cyan-500 to-blue-500 hover:from-cyan-400 hover:to-blue-400 text-white hover:shadow-[0_0_30px_rgba(6,182,212,0.3)] active:scale-95"
          }`}
        >
          {loading ? "Firing..." : !serverOnline ? "Server Offline" : "Fire Notification"}
        </button>
      </form>
    </div>
  );
}
