"use client";
import { motion, AnimatePresence } from "framer-motion";
import {
  Server, Database, Route, Mail, MessageSquare, LucideProps, AlertTriangle, Clock
} from "lucide-react";
import {
  useEffect, useRef, ForwardRefExoticComponent, RefAttributes, useReducer,
} from "react";

// ─── Types ───────────────────────────────────────────────────────────────────

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

type NodeStatus = "idle" | "active" | "scheduled" | "success" | "failed";

interface SystemState {
  api: NodeStatus;
  redis: NodeStatus;
  router: NodeStatus;
  email: NodeStatus;
  sms: NodeStatus;
  dlq: NodeStatus;
  scheduledUntil: string | null; // ISO timestamp
  activeRequestId: string | null;
  lastEventType: string | null;
}

type SystemAction =
  | { type: "RESET" }
  | { type: "SET"; node: keyof Omit<SystemState, "scheduledUntil" | "activeRequestId" | "lastEventType">; status: NodeStatus }
  | { type: "SCHEDULE"; until: string; requestId: string }
  | { type: "ACTIVATE"; node: keyof Omit<SystemState, "scheduledUntil" | "activeRequestId" | "lastEventType">; requestId?: string }
  | { type: "SUCCESS"; node: keyof Omit<SystemState, "scheduledUntil" | "activeRequestId" | "lastEventType"> }
  | { type: "FAIL"; requestId?: string }
  | { type: "DELIVER_SCHEDULED" };

const initialState: SystemState = {
  api: "idle", redis: "idle", router: "idle",
  email: "idle", sms: "idle", dlq: "idle",
  scheduledUntil: null, activeRequestId: null, lastEventType: null,
};

function systemReducer(state: SystemState, action: SystemAction): SystemState {
  switch (action.type) {
    case "RESET":
      return { ...initialState };
    case "SET":
      return { ...state, [action.node]: action.status };
    case "SCHEDULE":
      return { ...state, api: "success", redis: "scheduled", scheduledUntil: action.until, activeRequestId: action.requestId, lastEventType: "scheduled" };
    case "ACTIVATE":
      return { ...state, [action.node]: "active", activeRequestId: action.requestId ?? state.activeRequestId };
    case "SUCCESS":
      return { ...state, [action.node]: "success", lastEventType: "success" };
    case "FAIL":
      return { ...state, dlq: "failed", email: state.email === "active" ? "failed" : state.email, sms: state.sms === "active" ? "failed" : state.sms, lastEventType: "failed" };
    case "DELIVER_SCHEDULED":
      return { ...state, redis: "active", scheduledUntil: null };
    default:
      return state;
  }
}

// ─── Sub-Components ───────────────────────────────────────────────────────────

interface NodeProps {
  id: string;
  label: string;
  icon: ForwardRefExoticComponent<Omit<LucideProps, "ref"> & RefAttributes<SVGSVGElement>>;
  status: NodeStatus;
  badge?: string | null;
}

const STATUS_COLORS: Record<NodeStatus, string> = {
  idle: "rgba(255,255,255,0.04)",
  active: "#22d3ee",
  scheduled: "#f59e0b",
  success: "#10b981",
  failed: "#ef4444",
};

const STATUS_GLOWS: Record<NodeStatus, string> = {
  idle: "0 0 0px rgba(0,0,0,0)",
  active: "0 0 22px #22d3ee60",
  scheduled: "0 0 22px #f59e0b60",
  success: "0 0 22px #10b98160",
  failed: "0 0 22px #ef444460",
};

const STATUS_TEXT: Record<NodeStatus, string> = {
  idle: "#525252",
  active: "#22d3ee",
  scheduled: "#f59e0b",
  success: "#10b981",
  failed: "#ef4444",
};

const BADGE_STYLES: Record<string, string> = {
  HELD: "bg-amber-500/20 text-amber-400 border border-amber-500/40",
  FAILED: "bg-red-500/20 text-red-400 border border-red-500/40",
  SENT: "bg-emerald-500/20 text-emerald-400 border border-emerald-500/40",
};

function GraphNode({ label, icon: Icon, status, badge }: NodeProps) {
  return (
    <div className="flex flex-col items-center relative">
      <motion.div
        animate={{
          scale: status === "active" ? 1.12 : status === "scheduled" ? [1, 1.05, 1] : 1,
          borderColor: STATUS_COLORS[status],
          boxShadow: STATUS_GLOWS[status],
        }}
        transition={{
          scale: status === "scheduled" ? { repeat: Infinity, duration: 1.2 } : { duration: 0.3 },
          borderColor: { duration: 0.3 },
          boxShadow: { duration: 0.3 },
        }}
        className="w-16 h-16 rounded-2xl bg-neutral-950 border-2 flex items-center justify-center relative z-10"
      >
        <Icon className="w-7 h-7" style={{ color: STATUS_TEXT[status] }} />
      </motion.div>

      <AnimatePresence>
        {badge && (
          <motion.span
            key={badge}
            initial={{ opacity: 0, y: -6, scale: 0.8 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, scale: 0.8 }}
            className={`absolute -top-5 text-[9px] font-bold px-1.5 py-0.5 rounded-full uppercase tracking-wider ${BADGE_STYLES[badge] ?? "bg-white/10 text-white/50"}`}
          >
            {badge}
          </motion.span>
        )}
      </AnimatePresence>

      <span className="text-[10px] font-bold text-neutral-500 uppercase tracking-wider mt-3 text-center w-20">{label}</span>
    </div>
  );
}

interface FlowLineProps {
  active: boolean;
  color: string;
  vertical?: boolean;
}

function FlowLine({ active, color, vertical = false }: FlowLineProps) {
  return (
    <div className={`${vertical ? "h-full w-px" : "flex-1 h-px"} bg-white/5 relative overflow-hidden mx-2`}>
      <AnimatePresence>
        {active && (
          <motion.div
            key="pulse"
            initial={vertical ? { y: "-100%" } : { x: "-100%" }}
            animate={vertical ? { y: "100%" } : { x: "100%" }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.55, ease: "linear" }}
            className="absolute inset-0 w-full h-full"
            style={{
              background: vertical
                ? `linear-gradient(to bottom, transparent, ${color}, transparent)`
                : `linear-gradient(to right, transparent, ${color}, transparent)`,
            }}
          />
        )}
      </AnimatePresence>
    </div>
  );
}

// ─── Main Component ───────────────────────────────────────────────────────────

export default function ArchitectureGraph({ logs }: { logs: Log[] }) {
  const [sys, dispatch] = useReducer(systemReducer, initialState);
  const resetTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const prevLogLen = useRef(0);

  useEffect(() => {
    if (logs.length <= prevLogLen.current) return;
    const newLogs = logs.slice(prevLogLen.current);
    prevLogLen.current = logs.length;

    for (const log of newLogs) {
      // ── Scheduled (Time-Travel) ──────────────────
      if (log.msg?.includes("Time-Travel engaged") && log.send_at) {
        dispatch({ type: "SCHEDULE", until: log.send_at, requestId: log.request_id ?? "" });
        continue;
      }

      // ── Enqueued (instant) ──────────────────────
      if (log.msg?.includes("Enqueued task")) {
        dispatch({ type: "ACTIVATE", node: "api", requestId: log.request_id });
        dispatch({ type: "SUCCESS", node: "api" });
        dispatch({ type: "ACTIVATE", node: "redis" });
        continue;
      }

      // ── Router picks up (scheduled task delivered) ─
      if (log.worker === "router" && log.msg?.includes("Pulled Event")) {
        if (sys.scheduledUntil) dispatch({ type: "DELIVER_SCHEDULED" });
        dispatch({ type: "ACTIVATE", node: "router" });
        dispatch({ type: "SUCCESS", node: "redis" });
        continue;
      }

      // ── Router fans out ─────────────────────────
      if (log.worker === "router" && log.msg?.includes("Routed task")) {
        dispatch({ type: "SUCCESS", node: "router" });
        if (log.queue === "email") dispatch({ type: "ACTIVATE", node: "email" });
        if (log.queue === "sms") dispatch({ type: "ACTIVATE", node: "sms" });
        continue;
      }

      // ── Delivery success ────────────────────────
      if (log.msg?.includes("Successfully sent")) {
        if (log.worker === "Email") dispatch({ type: "SUCCESS", node: "email" });
        if (log.worker === "SMS") dispatch({ type: "SUCCESS", node: "sms" });
        continue;
      }

      // ── Error → DLQ ────────────────────────────
      if (log.level === "ERROR") {
        dispatch({ type: "FAIL", requestId: log.request_id });
        continue;
      }
    }

    // Auto-reset to idle after 6s of silence
    if (resetTimer.current) clearTimeout(resetTimer.current);
    resetTimer.current = setTimeout(() => dispatch({ type: "RESET" }), 6000);
  }, [logs, sys.scheduledUntil]);

  const isFlowing = (from: NodeStatus, to: NodeStatus) =>
    (from === "active" || from === "success") && to === "active";

  return (
    <div className="bg-neutral-900/50 border border-white/10 rounded-2xl p-8 relative overflow-hidden backdrop-blur-md">
      <div className="absolute inset-0 opacity-30"
        style={{ background: "radial-gradient(ellipse at top right, #0891b220 0%, transparent 60%)" }}
      />

      <div className="relative z-10">
        <div className="flex items-center justify-between mb-10">
          <h2 className="text-xs font-bold text-neutral-500 uppercase tracking-widest flex items-center">
            <span className="w-2 h-2 rounded-full bg-cyan-500 mr-2 animate-pulse" />
            Live Distributed Architecture
          </h2>
          {sys.scheduledUntil && (
            <div className="flex items-center text-amber-400 text-xs bg-amber-500/10 border border-amber-500/30 px-3 py-1 rounded-full">
              <Clock className="w-3 h-3 mr-1.5" />
              Task held in ZSET until {new Date(sys.scheduledUntil).toLocaleTimeString()}
            </div>
          )}
          {sys.lastEventType === "failed" && (
            <div className="flex items-center text-red-400 text-xs bg-red-500/10 border border-red-500/30 px-3 py-1 rounded-full">
              <AlertTriangle className="w-3 h-3 mr-1.5" />
              Task failed → routed to Dead Letter Queue
            </div>
          )}
        </div>

        {/* Main Flow */}
        <div className="flex items-center justify-between px-2">
          <GraphNode id="api" label="API Gateway" icon={Server} status={sys.api} />
          <FlowLine active={isFlowing(sys.api, sys.redis) || sys.redis === "active"} color="#f43f5e" />
          <GraphNode
            id="redis"
            label="Redis Queue"
            icon={Database}
            status={sys.redis}
            badge={sys.redis === "scheduled" ? "HELD" : sys.redis === "success" ? "SENT" : null}
          />
          <FlowLine active={isFlowing(sys.redis, sys.router) || sys.router === "active"} color="#a855f7" />
          <GraphNode id="router" label="Router Worker" icon={Route} status={sys.router} />

          {/* Fan-Out Lines */}
          <div className="flex-1 flex flex-col justify-center h-24 mx-2 gap-8">
            <FlowLine active={isFlowing(sys.router, sys.email) || sys.email === "active"} color="#10b981" />
            <FlowLine active={isFlowing(sys.router, sys.sms) || sys.sms === "active"} color="#3b82f6" />
          </div>

          <div className="flex flex-col gap-8">
            <GraphNode
              id="email"
              label="Email Node"
              icon={Mail}
              status={sys.email}
              badge={sys.email === "success" ? "SENT" : sys.email === "failed" ? "FAILED" : null}
            />
            <GraphNode
              id="sms"
              label="SMS Node"
              icon={MessageSquare}
              status={sys.sms}
              badge={sys.sms === "success" ? "SENT" : sys.sms === "failed" ? "FAILED" : null}
            />
          </div>
        </div>

        {/* Dead Letter Queue Row */}
        <div className="mt-8 flex items-center justify-end px-2">
          <div className="flex-1 h-px border-t border-dashed border-red-500/20 mx-2" />
          <div className="text-[10px] text-red-500/50 uppercase tracking-widest mr-4">Failure Path</div>
          <GraphNode
            id="dlq"
            label="Dead Letter Queue"
            icon={AlertTriangle}
            status={sys.dlq}
            badge={sys.dlq === "failed" ? "FAILED" : null}
          />
        </div>
      </div>
    </div>
  );
}
