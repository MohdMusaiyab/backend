import Link from 'next/link';
import { ServerCrash, Home } from 'lucide-react';

export default function NotFound() {
  return (
    <div className="min-h-screen bg-black text-white flex flex-col items-center justify-center p-4 selection:bg-cyan-500/30">
      
      {/* Background glow effects */}
      <div className="fixed top-0 left-1/2 -translate-x-1/2 w-[800px] h-[300px] bg-cyan-500/10 blur-[120px] rounded-full pointer-events-none" />
      <div className="fixed bottom-0 left-1/2 -translate-x-1/2 w-[600px] h-[300px] bg-purple-500/10 blur-[120px] rounded-full pointer-events-none" />

      <div className="relative bg-neutral-900/50 border border-white/10 rounded-3xl p-10 md:p-14 flex flex-col items-center text-center max-w-lg shadow-[0_0_50px_rgba(0,0,0,0.5)] backdrop-blur-xl z-10">
        
        <div className="w-24 h-24 bg-neutral-950 rounded-2xl border-2 border-neutral-800 flex items-center justify-center mb-8 shadow-[0_0_30px_rgba(34,211,238,0.15)] relative">
          <ServerCrash className="w-12 h-12 text-cyan-400 animate-pulse" />
          <div className="absolute -top-3 -right-3 bg-red-500 text-white text-[10px] font-bold px-2 py-1 rounded-full uppercase tracking-widest border border-red-500/50">
            404
          </div>
        </div>
        
        <h1 className="text-3xl font-bold mb-3 tracking-tight text-white">Lost in the Queue</h1>
        
        <p className="text-neutral-400 mb-8 max-w-sm text-sm leading-relaxed">
          The route you are looking for has been dropped. It might not exist, or it was routed directly to the Dead Letter Queue.
        </p>
        
        <Link 
          href="/" 
          className="flex items-center px-6 py-3 bg-white text-black font-semibold rounded-xl hover:bg-neutral-200 transition-all active:scale-95 shadow-[0_0_20px_rgba(255,255,255,0.15)] hover:shadow-[0_0_30px_rgba(255,255,255,0.3)]"
        >
          <Home className="w-4 h-4 mr-2" />
          Return to Dashboard
        </Link>
      </div>
    </div>
  );
}
