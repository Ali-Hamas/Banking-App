import { useEffect, useState, ReactNode } from "react";

interface Props {
  atmId: string;
  children: ReactNode;
}

export function AtmFrame({ atmId, children }: Props) {
  const [now, setNow] = useState(() => new Date());
  useEffect(() => {
    const t = setInterval(() => setNow(new Date()), 30_000);
    return () => clearInterval(t);
  }, []);

  const clock = now.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });

  return (
    <div className="kiosk">
      <div className="kiosk-side kiosk-side-left" aria-hidden>
        {[0, 1, 2, 3, 4].map((i) => (
          <div key={i} className="kiosk-key" />
        ))}
      </div>

      <div className="atm">
        <div className="atm-statusbar">
          <span className="atm-status-brand">DEMOBANK</span>
          <span className="atm-status-id">{atmId}</span>
          <span className="atm-status-icons">
            <IconLock />
            <IconSignal />
            <span className="atm-status-clock">{clock}</span>
          </span>
        </div>

        <div className="atm-screen">
          <div className="scanlines" aria-hidden />
          <div className="screen-content">{children}</div>
        </div>

        <div className="atm-footer">
          <span>Secured by Cardless Authentication Layer</span>
          <span className="atm-footer-dot" />
          <span>TLS 1.3 · HS256</span>
        </div>
      </div>

      <div className="kiosk-side kiosk-side-right" aria-hidden>
        {[0, 1, 2, 3, 4].map((i) => (
          <div key={i} className="kiosk-key" />
        ))}
      </div>
    </div>
  );
}

function IconLock() {
  return (
    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2">
      <rect x="4" y="11" width="16" height="10" rx="2" />
      <path d="M8 11V7a4 4 0 0 1 8 0v4" />
    </svg>
  );
}

function IconSignal() {
  return (
    <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor">
      <rect x="2" y="14" width="3" height="6" rx="1" />
      <rect x="8" y="10" width="3" height="10" rx="1" />
      <rect x="14" y="6" width="3" height="14" rx="1" />
      <rect x="20" y="2" width="3" height="18" rx="1" opacity="0.4" />
    </svg>
  );
}
