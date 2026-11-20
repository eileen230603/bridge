import "./titlebar.css";

/** Franja superior que permite mover la ventana (solo visible/activa en macOS). */
export function TitleBarDrag() {
  return <div className="titlebar-drag" aria-hidden="true" />;
}
