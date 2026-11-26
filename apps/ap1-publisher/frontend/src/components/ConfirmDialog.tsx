import React from "react";
import "./LicenciaModal.css";

type Props = { title: string; message: string; confirm: string; cancel: string; onConfirm: () => void; onCancel: () => void };

// Diálogo de confirmación con el mismo vidrio que la modal de licencia.
// No usa window.confirm: el WebView de macOS no lo muestra.
export function ConfirmDialog({ title, message, confirm, cancel, onConfirm, onCancel }: Props) {
  const cancelRef = React.useRef<HTMLButtonElement>(null);
  React.useEffect(() => {
    cancelRef.current?.focus();
    const onKey = (e: KeyboardEvent) => { if (e.key === "Escape") onCancel(); };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onCancel]);
  return (
    <div className="license-modal-overlay" onMouseDown={(e) => { if (e.target === e.currentTarget) onCancel(); }}>
      <div className="license-modal-card" role="alertdialog" aria-modal="true" aria-labelledby="confirm-title" aria-describedby="confirm-text">
        <div className="license-modal-head">
          <h3 className="license-modal-title" id="confirm-title">{title}</h3>
        </div>
        <p id="confirm-text" style={{ margin: 0, fontSize: 15, color: "var(--muted)", lineHeight: 1.45 }}>{message}</p>
        <div className="license-actions-row">
          <button type="button" ref={cancelRef} className="btn-secondary" onClick={onCancel}>{cancel}</button>
          <button type="button" className="btn-primary" onClick={onConfirm}>{confirm}</button>
        </div>
      </div>
    </div>
  );
}
