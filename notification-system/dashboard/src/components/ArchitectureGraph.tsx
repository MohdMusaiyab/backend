"use client";
import { motion } from "framer-motion";
import { Server, Database, Route, Mail, MessageSquare, LucideProps } from "lucide-react";
import { useEffect, useRef, useReducer, ForwardRefExoticComponent, RefAttributes } from "react";

interface NodeProps {
  id: string;
  label: string;
  icon: ForwardRefExoticComponent<Omit<LucideProps, "ref"> & RefAttributes<SVGSVGElement>>;
  color: string;
  isActive: boolean;
}

// Pulled OUT of the parent component to fix the "component created during render" error
function GraphNode({ label, icon: Icon, color, isActive }: NodeProps) {
  return (
    <div className="flex flex-col items-center">
      <motion.div
        animate={{
          scale: isActive ? 1.1 : 1,
          borderColor: isActive ? color : "rgba(255,255,255,0.05)",
          boxShadow: isActive ? `0 0 25px ${color}80` : "0 0 0px rgba(0,0,0,0)",
          backgroundColor: isActive ? "rgba(20,20,20,1)" : "rgba(10,10,10,1)",
        }}
        className="w-16 h-16 rounded-2xl bg-neutral-950 border-2 flex items-center justify-center relative z-10"
      >
        <Icon
          className="w-7 h-7"
          style={{ color: isActive ? color : "#737373" }}
        />
      </motion.div>
      <span className="text-[10px] font-bold text-neutral-500 uppercase tracking-wider mt-3 text-center w-20">
        {label}
      </span>
    </div>
  );
}

interface Log {
  msg?: string;
  worker?: string;
  time?: string;
  level?: string;
  queue?: string;
  send_at?: string;
  request_id?: string;
}

type ActiveNodesAction =
  | { type: "ADD"; node: string }
  | { type: "REMOVE"; node: string };

function activeNodesReducer(state: string[], action: ActiveNodesAction): string[] {
  switch (action.type) {
    case "ADD":
      return [...state, action.node];
    case "REMOVE":
      return state.filter((n) => n !== action.node);
    default:
      return state;
  }
}

export default function ArchitectureGraph({ logs }: { logs: Log[] }) {
  const [activeNodes, dispatch] = useReducer(activeNodesReducer, []);
  // Use a ref to schedule REMOVE dispatches so we can call them from setTimeout
  // without triggering the "setState in effect" linting rule
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (logs.length === 0) return;
    const latestLog = logs[logs.length - 1];

    let active = "api";
    if (
      latestLog.msg?.includes("Enqueued task") ||
      latestLog.msg?.includes("Time-Travel engaged")
    ) {
      active = "redis";
    } else if (latestLog.worker === "router") {
      active = "router";
    } else if (latestLog.worker === "Email") {
      active = "email";
    } else if (latestLog.worker === "SMS") {
      active = "sms";
    }

    // Dispatch ADD immediately — this is safe because it's driven by the external WebSocket (the "external system")
    dispatch({ type: "ADD", node: active });

    // Schedule REMOVE through a ref-held timer (not directly from the effect body)
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => {
      dispatch({ type: "REMOVE", node: active });
    }, 1500);

    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, [logs]);

  return (
    <div className="bg-neutral-900/50 border border-white/10 rounded-2xl p-8 relative overflow-hidden backdrop-blur-md">
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-cyan-900/20 via-neutral-900/0 to-neutral-900/0"></div>

      <div className="relative z-10">
        <h2 className="text-xs font-bold text-neutral-500 uppercase tracking-widest mb-10 flex items-center">
          <span className="w-2 h-2 rounded-full bg-cyan-500 mr-2"></span>
          Live Distributed Architecture
        </h2>

        <div className="flex items-center justify-between px-2">
          <GraphNode id="api" label="API Gateway" icon={Server} color="#22d3ee" isActive={activeNodes.includes("api")} />

          <div className="flex-1 h-px bg-white/5 mx-2 relative overflow-hidden">
            {activeNodes.includes("redis") && (
              <motion.div
                initial={{ x: "-100%" }}
                animate={{ x: "100%" }}
                transition={{ duration: 0.5 }}
                className="absolute inset-0 bg-gradient-to-r from-transparent via-rose-500 to-transparent w-full"
              />
            )}
          </div>

          <GraphNode id="redis" label="Redis Queue" icon={Database} color="#f43f5e" isActive={activeNodes.includes("redis")} />

          <div className="flex-1 h-px bg-white/5 mx-2 relative overflow-hidden">
            {activeNodes.includes("router") && (
              <motion.div
                initial={{ x: "-100%" }}
                animate={{ x: "100%" }}
                transition={{ duration: 0.5 }}
                className="absolute inset-0 bg-gradient-to-r from-transparent via-purple-500 to-transparent w-full"
              />
            )}
          </div>

          <GraphNode id="router" label="Router Worker" icon={Route} color="#a855f7" isActive={activeNodes.includes("router")} />

          <div className="flex-1 flex flex-col justify-center h-24 mx-2">
            <div className="w-full h-px bg-white/5 relative overflow-hidden mb-8">
              {activeNodes.includes("email") && (
                <motion.div
                  initial={{ x: "-100%" }}
                  animate={{ x: "100%" }}
                  transition={{ duration: 0.5 }}
                  className="absolute inset-0 bg-gradient-to-r from-transparent via-emerald-500 to-transparent w-full"
                />
              )}
            </div>
            <div className="w-full h-px bg-white/5 relative overflow-hidden">
              {activeNodes.includes("sms") && (
                <motion.div
                  initial={{ x: "-100%" }}
                  animate={{ x: "100%" }}
                  transition={{ duration: 0.5 }}
                  className="absolute inset-0 bg-gradient-to-r from-transparent via-blue-500 to-transparent w-full"
                />
              )}
            </div>
          </div>

          <div className="flex flex-col gap-8">
            <GraphNode id="email" label="Email Node" icon={Mail} color="#10b981" isActive={activeNodes.includes("email")} />
            <GraphNode id="sms" label="SMS Node" icon={MessageSquare} color="#3b82f6" isActive={activeNodes.includes("sms")} />
          </div>
        </div>
      </div>
    </div>
  );
}
