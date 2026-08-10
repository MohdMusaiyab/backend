"use client";
import { useEffect, useRef } from "react";
import { Terminal } from "lucide-react";

interface Log {
  time?: string;
  level?: string;
  msg?: string;
  worker?: string;
  queue?: string;
  send_at?: string;
  request_id?: string;
}

export default function LiveTerminal({ logs }: { logs: Log[] }) {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [logs]);

  return (
    <div className="bg-[#0a0a0a] border border-white/10 rounded-2xl overflow-hidden shadow-[0_10px_40px_rgba(0,0,0,0.5)]">
      <div className="bg-neutral-900/80 border-b border-white/5 px-4 py-3 flex items-center">
        <Terminal className="w-4 h-4 text-neutral-400 mr-2" />
        <span className="text-xs font-mono text-neutral-400">telemetry_logs.sh</span>
        <div className="ml-auto flex space-x-2">
          <div className="w-3 h-3 rounded-full bg-red-500/20 border border-red-500/50"></div>
          <div className="w-3 h-3 rounded-full bg-yellow-500/20 border border-yellow-500/50"></div>
          <div className="w-3 h-3 rounded-full bg-green-500/20 border border-green-500/50"></div>
        </div>
      </div>

      <div className="p-5 h-72 overflow-y-auto font-mono text-[11px] leading-relaxed space-y-2 custom-scrollbar">
        {logs.length === 0 ? (
          <div className="text-neutral-600 italic flex items-center animate-pulse">
            <span className="mr-2">Listening to ws://localhost:8080/ws</span>
            <span className="flex space-x-1">
              <span className="animate-bounce">.</span>
              <span className="animate-bounce" style={{ animationDelay: "0.2s" }}>.</span>
              <span className="animate-bounce" style={{ animationDelay: "0.4s" }}>.</span>
            </span>
          </div>
        ) : (
          logs.map((log, i) => (
            <div key={i} className="flex hover:bg-white/5 p-1 -mx-1 rounded transition-colors break-all">
              <span className="text-emerald-500 mr-3 shrink-0">
                [{log.time ? new Date(log.time).toLocaleTimeString() : "??:??:??"}]
              </span>
              <span className="text-cyan-400 mr-2 shrink-0">[{log.level}]</span>
              <span className="text-neutral-300">
                <span className="text-white font-semibold">{log.msg}</span>
                {log.worker && (
                  <span className="text-purple-400 ml-3">
                    worker=<span className="text-purple-300">&quot;{log.worker}&quot;</span>
                  </span>
                )}
                {log.queue && (
                  <span className="text-yellow-400 ml-3">
                    queue=<span className="text-yellow-300">&quot;{log.queue}&quot;</span>
                  </span>
                )}
                {log.send_at && (
                  <span className="text-rose-400 ml-3">
                    send_at=<span className="text-rose-300">&quot;{log.send_at}&quot;</span>
                  </span>
                )}
                <span className="text-neutral-500 ml-3 italic">
                  req_id=&quot;{log.request_id?.slice(0, 8)}&quot;
                </span>
              </span>
            </div>
          ))
        )}
        <div ref={bottomRef} />
      </div>
    </div>
  );
}
