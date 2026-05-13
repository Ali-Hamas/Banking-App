export function CheckMark() {
  return (
    <svg className="checkmark" viewBox="0 0 80 80" width="96" height="96">
      <circle className="checkmark-circle" cx="40" cy="40" r="36" />
      <path className="checkmark-tick" d="M24 42 L36 54 L58 30" fill="none" />
    </svg>
  );
}
