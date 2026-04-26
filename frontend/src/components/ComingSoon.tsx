interface ComingSoonProps {
  pageName: string
  tag: string
}

export default function ComingSoon({ pageName, tag }: ComingSoonProps) {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen gap-6 px-8">
      <div className="flex flex-col items-center gap-4 max-w-sm text-center">
        <div className="w-16 h-16 rounded-2xl flex items-center justify-center bg-primary/8 border border-primary/20">
          <svg viewBox="0 0 32 32" fill="none" className="w-8 h-8">
            <circle cx="16" cy="16" r="14" stroke="currentColor" strokeWidth="1.5" strokeOpacity="0.3" className="text-primary" />
            <circle cx="16" cy="16" r="8"  stroke="currentColor" strokeWidth="1.5" strokeOpacity="0.5" className="text-primary" />
            <circle cx="16" cy="16" r="3"  fill="currentColor" className="text-primary" />
            <path d="M16 4V16L24 24" stroke="currentColor" strokeWidth="2" strokeLinecap="round" className="text-primary" />
          </svg>
        </div>

        <div className="space-y-1">
          <p className="font-mono text-[9px] tracking-widest text-primary">{tag}</p>
          <h1 className="font-mono font-black text-2xl tracking-wide text-foreground uppercase">{pageName}</h1>
        </div>

        <div className="w-full h-px bg-border" />

        <div className="space-y-2">
          <p className="font-mono text-[10px] tracking-widest text-muted-foreground uppercase">
            Runway Under Construction
          </p>
          <p className="font-mono text-xs text-muted-foreground/60 leading-relaxed">
            This module is on approach. Stand by for clearance.
          </p>
        </div>

        <div className="flex items-center gap-2 px-4 py-2 rounded-lg bg-primary/4 border border-primary/12">
          <div className="w-1.5 h-1.5 rounded-full bg-primary animate-pulse" />
          <span className="font-mono text-[10px] tracking-widest text-primary font-bold">COMING SOON</span>
          <div className="w-1.5 h-1.5 rounded-full bg-primary animate-pulse" />
        </div>
      </div>
    </div>
  )
}
