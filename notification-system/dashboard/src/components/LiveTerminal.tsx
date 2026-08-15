"use client";
import { useEffect, useRef, useState } from "react";
import { Terminal, Activity, X } from "lucide-react";

interface Log {
  time?: string;
  level?: string;
  msg?: string;
  worker?: string;
  queue?: string;
  send_at?: string;
  request_id?: string;
}

const COLORS = ["text-pink-400", "text-cyan-400", "text-yellow-400", "text-emerald-400"];

export default function LiveTerminal({ logs }: { logs: Log[] }) {
  const [closedIds, setClosedIds] = useState<Set<string>>(new Set());

  // Extract unique request IDs in chronological order of appearance
  const allReqIds = Array.from(new Set(logs.map((l) => l.request_id).filter(Boolean))) as string[];
  
  // Show up to the 3 most recent requests that haven't been closed by the user
  const visibleReqIds = allReqIds.filter(id => !closedIds.has(id)).slice(-3);

  // If no requests yet (or all closed), show a system waiting state
  if (visibleReqIds.length === 0) {
    return (
      <div className="bg-[#0a0a0a] border border-white/10 rounded-2xl overflow-hidden shadow-lg h-72">
        <div className="bg-neutral-900/80 border-b border-white/5 px-4 py-3 flex items-center">
          <Terminal className="w-4 h-4 text-neutral-400 mr-2" />
          <span className="text-xs font-mono text-neutral-400">telemetry_logs.sh</span>
        </div>
        <div className="p-5 flex flex-col items-center justify-center h-full text-neutral-600 font-mono text-xs">
          <span className="flex space-x-1 animate-pulse">
            Waiting for notifications to process...
          </span>
        </div>
      </div>
    );
  }

  // Swimlanes
  return (
    <div className={`grid grid-cols-1 md:grid-cols-${visibleReqIds.length} gap-4`} style={{ gridTemplateColumns: `repeat(${visibleReqIds.length}, minmax(0, 1fr))` }}>
      {visibleReqIds.map((id, index) => {
        const reqLogs = logs.filter((l) => l.request_id === id);
        return (
          <TerminalWindow
            key={id}
            id={id}
            logs={reqLogs}
            colorClass={COLORS[index % COLORS.length]}
            onClose={() => setClosedIds(prev => new Set(prev).add(id))}
          />
        );
      })}
    </div>
  );
}

function TerminalWindow({ id, logs, colorClass, onClose }: { id: string; logs: Log[]; colorClass: string; onClose: () => void }) {
  const bottomRef = useRef<HTMLDivElement>(null);
  
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [logs]);

  return (
    <div className="bg-[#0a0a0a] border border-white/10 rounded-xl overflow-hidden shadow-[0_10px_40px_rgba(0,0,0,0.5)] h-72 flex flex-col transition-all">
      <div className="bg-neutral-900/80 border-b border-white/5 px-3 py-2 flex items-center">
        <Activity className={`w-3 h-3 ${colorClass} mr-2`} />
        <span className={`text-[10px] font-mono font-semibold ${colorClass}`}>Req: {id.slice(0, 8)}</span>
        <div className="ml-auto flex space-x-1.5 group/controls">
          <button 
            onClick={onClose}
            className="w-2.5 h-2.5 rounded-full bg-red-500/20 border border-red-500/50 hover:bg-red-500 flex items-center justify-center transition-colors cursor-pointer group"
          >
            <X className="w-2 h-2 text-neutral-900 opacity-0 group-hover:opacity-100 transition-opacity" />
          </button>
          <div className="w-2.5 h-2.5 rounded-full bg-yellow-500/20 border border-yellow-500/50"></div>
          <div className="w-2.5 h-2.5 rounded-full bg-green-500/20 border border-green-500/50"></div>
        </div>
      </div>

      <div className="p-3 overflow-y-auto font-mono text-[10px] leading-relaxed flex-1 space-y-1.5 custom-scrollbar">
        {logs.map((log, i) => (
          <div key={i} className="flex hover:bg-white/5 p-1 -mx-1 rounded transition-colors break-all">
            <span className="text-neutral-500 mr-2 shrink-0">
              [{log.time ? new Date(log.time).toLocaleTimeString([], { hour12: false }) : "??:??"}]
            </span>
            <span className="text-neutral-300">
              <span className={log.level === "ERROR" ? "text-red-400 font-bold" : "text-white"}>{log.msg}</span>
              {log.worker && (
                <span className="text-neutral-400 ml-2">
                  w=<span className="text-purple-300">&quot;{log.worker}&quot;</span>
                </span>
              )}
            </span>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>
    </div>
  );
}
