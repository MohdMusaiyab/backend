"use client";
import { motion, AnimatePresence } from "framer-motion";
import {
  Server, Database, Route, Mail, MessageSquare, LucideProps, AlertTriangle, Layers
} from "lucide-react";
import {
  useEffect, ForwardRefExoticComponent, RefAttributes, useState, useRef
} from "react";

export interface Log {
  time?: string;
  level?: string;
  msg?: string;
  worker?: string;
  queue?: string;
  send_at?: string;
  request_id?: string;
  error?: string;
}

// ─── State Engine ─────────────────────────────────────────────────────────────

type NodeLocation = "api" | "redis" | "router" | "email" | "sms" | "dlq";
type ReqStatus = "active" | "held" | "error" | "done";

interface RequestTracker {
  id: string;
  location: NodeLocation;
  status: ReqStatus;
  updatedAt: number;
  sendAt?: string;
}

// ─── Sub-Components ───────────────────────────────────────────────────────────

interface NodeProps {
  id: string;
  label: string;
  icon: ForwardRefExoticComponent<Omit<LucideProps, "ref"> & RefAttributes<SVGSVGElement>>;
  active: number;
  held: number;
  error: number;
}

function GraphNode({ label, icon: Icon, active, held, error }: NodeProps) {
  const isHeld = held > 0;
  const isError = error > 0;
  const isActive = active > 0;
  
  let statusColor = "rgba(255,255,255,0.04)";
  let glow = "0 0 0px rgba(0,0,0,0)";
  let textColor = "#525252";
  let scale = 1;

  if (isError) {
    statusColor = "#ef4444"; glow = "0 0 25px #ef444480"; textColor = "#ef4444"; scale = 1.1;
  } else if (isHeld) {
    statusColor = "#f59e0b"; glow = "0 0 25px #f59e0b80"; textColor = "#f59e0b"; scale = 1.05;
  } else if (isActive) {
    statusColor = "#22d3ee"; glow = "0 0 25px #22d3ee80"; textColor = "#22d3ee"; scale = 1.1;
  }

  const total = active + held + error;

  return (
    <div className="flex flex-col items-center relative">
      <motion.div
        animate={{ scale, borderColor: statusColor, boxShadow: glow }}
        transition={{ duration: 0.3 }}
        className="w-16 h-16 rounded-2xl bg-neutral-950 border-2 flex items-center justify-center relative z-10"
      >
        <Icon className="w-7 h-7 transition-colors duration-300" style={{ color: textColor }} />
      </motion.div>

      <AnimatePresence>
        {total > 0 && (
          <motion.div
            initial={{ opacity: 0, scale: 0 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0 }}
            className={`absolute -top-3 -right-2 w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-bold border-2 border-neutral-900 z-20 ${
              isError ? "bg-red-500 text-white" : isHeld ? "bg-amber-500 text-white" : "bg-cyan-500 text-white"
            }`}
          >
            {total}
          </motion.div>
        )}
      </AnimatePresence>

      <span className="text-[10px] font-bold text-neutral-500 uppercase tracking-wider mt-3 text-center w-20">{label}</span>
    </div>
  );
}

function FlowLine({ active, color }: { active: boolean; color: string }) {
  return (
    <div className="flex-1 h-px bg-white/5 relative overflow-hidden mx-2">
      <AnimatePresence>
        {active && (
          <motion.div
            initial={{ x: "-100%" }}
            animate={{ x: "100%" }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.8, ease: "linear", repeat: Infinity }}
            className="absolute inset-0 w-full h-full"
            style={{ background: `linear-gradient(to right, transparent, ${color}, transparent)` }}
          />
        )}
      </AnimatePresence>
    </div>
  );
}

// ─── Main Component ───────────────────────────────────────────────────────────

export default function ArchitectureGraph({ logs }: { logs: Log[] }) {
  const [requests, setRequests] = useState<Record<string, RequestTracker>>({});
  const [nowMs, setNowMs] = useState(0);
  const prevLogLen = useRef(0);

  // Force re-render every second to clear stale/done requests naturally
  useEffect(() => {
    const t = setInterval(() => setNowMs(Date.now()), 1000);
    return () => clearInterval(t);
  }, []);

  useEffect(() => {
    if (logs.length <= prevLogLen.current) return;
    const newLogs = logs.slice(prevLogLen.current);
    prevLogLen.current = logs.length;

    setRequests(prev => {
      const next = { ...prev };
      
      for (const log of newLogs) {
        if (!log.request_id) continue;
        const id = log.request_id;
        const req = next[id] || { id, location: "api", status: "active", updatedAt: Date.now() };
        req.updatedAt = Date.now();

        // ── State Transitions ──
        if (log.msg?.includes("Enqueued task")) {
          req.location = "redis";
          req.status = log.send_at ? "held" : "active";
          if (log.send_at) req.sendAt = log.send_at;
        } else if (log.msg?.includes("Time-Travel engaged")) {
          req.location = "redis";
          req.status = "held";
          if (log.send_at) req.sendAt = log.send_at;
        } else if (log.worker === "router" && log.msg?.includes("Pulled Event")) {
          req.location = "router";
          req.status = "active";
        } else if (log.worker === "router" && log.msg?.includes("Routed task")) {
          req.location = log.queue === "email" ? "email" : "sms";
          req.status = "active";
        } else if (log.msg?.includes("Successfully sent")) {
          req.status = "done";
        } else if (log.level === "ERROR") {
          req.location = "dlq";
          req.status = "error";
        }
        
        next[id] = req;
      }
      return next;
    });
  }, [logs]);

  // ── Compute Heatmap Load ──
  const stats = {
    api: { active: 0, held: 0, error: 0 },
    redis: { active: 0, held: 0, error: 0 },
    router: { active: 0, held: 0, error: 0 },
    email: { active: 0, held: 0, error: 0 },
    sms: { active: 0, held: 0, error: 0 },
    dlq: { active: 0, held: 0, error: 0 },
  };

  let totalActive = 0;

  Object.values(requests).forEach(req => {
    if (req.status === "done") return; // Dropped instantly from visualization
    
    // Auto-expire requests if we missed the success log (e.g. 5s for active, longer for held)
    const isStale = req.status === "active" && (nowMs - req.updatedAt > 6000);
    if (isStale) return;

    // Check if a held task is still held
    if (req.status === "held" && req.sendAt) {
       const fireTime = new Date(req.sendAt).getTime();
       if (nowMs > fireTime + 2000) return; // Expire 2s after it should have fired if router didn't pick it up
    }

    stats[req.location][req.status]++;
    totalActive++;
  });

  const hasFlow = (loc: keyof typeof stats) => stats[loc].active > 0;

  return (
    <div className="bg-neutral-900/50 border border-white/10 rounded-2xl p-8 relative overflow-hidden backdrop-blur-md">
      <div className="absolute inset-0 opacity-30"
        style={{ background: "radial-gradient(ellipse at top right, #0891b220 0%, transparent 60%)" }}
      />

      <div className="relative z-10">
        <div className="flex items-center justify-between mb-10">
          <h2 className="text-xs font-bold text-neutral-500 uppercase tracking-widest flex items-center">
            <Layers className="w-4 h-4 text-cyan-500 mr-2" />
            Distributed Heatmap
          </h2>
          {totalActive > 0 && (
            <div className="flex items-center text-cyan-400 text-xs bg-cyan-500/10 border border-cyan-500/30 px-3 py-1 rounded-full animate-pulse">
              Tracking {totalActive} concurrent request{totalActive > 1 ? 's' : ''} in flight
            </div>
          )}
        </div>

        <div className="flex items-center justify-between px-2">
          <GraphNode id="api" label="API Gateway" icon={Server} {...stats.api} />
          <FlowLine active={hasFlow("api") || hasFlow("redis")} color="#22d3ee" />
          
          <GraphNode id="redis" label="Redis Queue" icon={Database} {...stats.redis} />
          <FlowLine active={hasFlow("redis") || hasFlow("router")} color="#a855f7" />
          
          <GraphNode id="router" label="Router Worker" icon={Route} {...stats.router} />

          <div className="flex-1 flex flex-col justify-center h-24 mx-2 gap-8">
            <FlowLine active={hasFlow("router") || hasFlow("email")} color="#10b981" />
            <FlowLine active={hasFlow("router") || hasFlow("sms")} color="#3b82f6" />
          </div>

          <div className="flex flex-col gap-8">
            <GraphNode id="email" label="Email Node" icon={Mail} {...stats.email} />
            <GraphNode id="sms" label="SMS Node" icon={MessageSquare} {...stats.sms} />
          </div>
        </div>

        {/* Dead Letter Queue Row */}
        <div className="mt-8 flex items-center justify-end px-2">
          <div className="flex-1 h-px border-t border-dashed border-red-500/20 mx-2" />
          <div className="text-[10px] text-red-500/50 uppercase tracking-widest mr-4">Failure Path</div>
          <GraphNode id="dlq" label="Dead Letter Queue" icon={AlertTriangle} {...stats.dlq} />
        </div>
      </div>
    </div>
  );
}
