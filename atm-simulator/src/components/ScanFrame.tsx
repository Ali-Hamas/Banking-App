import { ReactNode } from "react";

interface Props {
  remaining: number;
  ttl: number;
  children: ReactNode;
}

// Circular SVG timer ring that wraps a QR code. Stroke unwinds as time passes.
export function ScanFrame({ remaining, ttl, children }: Props) {
  const size = 296;
  const r = (size - 12) / 2; // stroke 8 + padding
  const c = 2 * Math.PI * r;
  const pct = Math.max(0, Math.min(1, remaining / Math.max(1, ttl)));
  const dash = c * pct;

  return (
    <div className="scan-frame">
      <svg className="scan-ring" width={size} height={size}>
        <circle cx={size / 2} cy={size / 2} r={r} className="scan-ring-track" />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={r}
          className="scan-ring-progress"
          strokeDasharray={`${dash} ${c}`}
        />
      </svg>

      <div className="scan-corners" aria-hidden>
        <span className="scan-corner scan-corner-tl" />
        <span className="scan-corner scan-corner-tr" />
        <span className="scan-corner scan-corner-bl" />
        <span className="scan-corner scan-corner-br" />
        <span className="scan-line" />
      </div>

      <div className="scan-qr">{children}</div>
    </div>
  );
}
