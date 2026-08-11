"use client";
import { useEffect, useState } from "react";
import ControlPanel from "@/components/ControlPanel";
import LiveTerminal from "@/components/LiveTerminal";
import ArchitectureGraph, { Log } from "@/components/ArchitectureGraph";

export default function Home() {
  const [logs, setLogs] = useState<Log[]>([]);
  const [wsStatus, setWsStatus] = useState<"connecting" | "connected" | "disconnected">("connecting");

  useEffect(() => {
    let ws: WebSocket;
    let reconnectTimer: NodeJS.Timeout;

    const connect = () => {
      setWsStatus("connecting");
      ws = new WebSocket("ws://localhost:8080/ws");
      
      ws.onopen = () => setWsStatus("connected");
      
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          setLogs((prev) => [...prev, data]);
        } catch (_e) {
          console.error("Failed to parse log", event.data);
        }
      };

      ws.onclose = () => {
        setWsStatus("disconnected");
        reconnectTimer = setTimeout(connect, 3000); // Auto reconnect every 3s
      };

      ws.onerror = () => {
        ws.close();
      };
    };

    connect();

    return () => {
      clearTimeout(reconnectTimer);
      if (ws) ws.close();
    };
  }, []);

  return (
    <main className="min-h-screen bg-neutral-950 text-white p-8 font-sans selection:bg-cyan-500/30">
      <div className="max-w-7xl mx-auto space-y-8">
        
        {/* HEADER */}
        <header className="flex items-center justify-between border-b border-white/10 pb-6">
          <div>
            <h1 className="text-3xl font-bold bg-gradient-to-r from-emerald-400 to-cyan-400 bg-clip-text text-transparent">
              Notification Engine
            </h1>
            <p className="text-neutral-400 mt-2 text-sm">Real-time distributed system visualization</p>
          </div>
          <div className="flex items-center space-x-2 bg-neutral-900 border border-white/5 px-4 py-2 rounded-full shadow-inner">
            <span className="relative flex h-3 w-3">
              {wsStatus === "connected" && <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>}
              <span className={`relative inline-flex rounded-full h-3 w-3 ${wsStatus === "connected" ? "bg-emerald-500" : wsStatus === "connecting" ? "bg-yellow-500" : "bg-red-500"}`}></span>
            </span>
            <span className={`text-sm font-medium ${wsStatus === "connected" ? "text-emerald-400" : wsStatus === "connecting" ? "text-yellow-400" : "text-red-400"}`}>
              {wsStatus === "connected" ? "WebSocket Connected" : wsStatus === "connecting" ? "Connecting..." : "Server Offline"}
            </span>
          </div>
        </header>

        {/* DASHBOARD GRID */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <div className="lg:col-span-1 space-y-8">
            <ControlPanel serverOnline={wsStatus === "connected"} />
          </div>
          <div className="lg:col-span-2 space-y-8">
            <ArchitectureGraph logs={logs} />
            <LiveTerminal logs={logs} />
          </div>
        </div>
      </div>
    </main>
  );
}
