import React from "react";
import logo from "../assets/logo.png";
import "./sidebar.css";
import { useTheme } from "../theme";

export type View = "studies" | "history" | "settings";

const I = (d: React.ReactNode) => (
  <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round">{d}</svg>
);
const ICON = {
  studies: I(<><rect x="3" y="4" width="18" height="16" rx="3" /><path d="M3 9h18M8 14h4" /></>),
  history: I(<><path d="M3 12a9 9 0 1 0 3-6.7L3 8" /><path d="M3 3v5h5" /><path d="M12 7v5l3 2" /></>),
  settings: I(<><circle cx="12" cy="12" r="3" /><path d="M19.4 15a1.7 1.7 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-1.8-.3 1.7 1.7 0 0 0-1 1.5V21a2 2 0 1 1-4 0v-.1a1.7 1.7 0 0 0-1.1-1.5 1.7 1.7 0 0 0-1.8.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.7 1.7 0 0 0 .3-1.8 1.7 1.7 0 0 0-1.5-1H3a2 2 0 1 1 0-4h.1a1.7 1.7 0 0 0 1.5-1.1 1.7 1.7 0 0 0-.3-1.8l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.7 1.7 0 0 0 1.8.3H9a1.7 1.7 0 0 0 1-1.5V3a2 2 0 1 1 4 0v.1a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.8-.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0-.3 1.8V9a1.7 1.7 0 0 0 1.5 1H21a2 2 0 1 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1z" /></>),
  key: I(<><circle cx="8" cy="15" r="4" /><path d="M10.8 12.2 20 3M16 7l3 3" /></>),
  sun: I(<><circle cx="12" cy="12" r="4" /><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" /></>),
  moon: I(<path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z" />),
};

type Props = {
  view: View;
  onNavigate: (v: View) => void;
  activeJobs: number;
  licenseValid: boolean;
  onLicense: () => void;
};

export function Sidebar({ view, onNavigate, activeJobs, licenseValid, onLicense }: Props) {
  const { theme, toggle } = useTheme();
  const item = (id: View, text: string, badge = 0) => (
    <button key={id} className={`sb-item ${view === id ? "on" : ""}`} onClick={() => onNavigate(id)} aria-current={view === id ? "page" : undefined}>
      <span className="sb-ico">{ICON[id]}</span>
      <span className="sb-text">{text}</span>
      {badge > 0 && <span className="sb-badge">{badge}</span>}
    </button>
  );
  const next = theme === "dark" ? "claro" : "oscuro";

  return (
    <aside className="sidebar">
      <div className="sb-brand">
        <img className="brandLogo" src={logo} alt="Symphony Media Export" />
        <div className="sb-name"><b>SYMPHONY</b><span>MEDIA EXPORT</span></div>
      </div>

      <div className="sb-label">Menú</div>
      <nav className="sb-nav">
        {item("studies", "Estudios")}
        {item("history", "Historial", activeJobs)}
        {item("settings", "Configuración")}
      </nav>

      <div className="sb-spacer" />
      <div className="sb-rule" />
      <button className={`sb-item sb-license ${licenseValid ? "ok" : "bad"}`} onClick={onLicense}>
        <span className="sb-ico">{ICON.key}</span>
        <span className="sb-text">{licenseValid ? "Licencia activa" : "Licencia inactiva"}</span>
        <i className="sb-status" />
      </button>
      <button className="sb-item sb-theme" onClick={toggle} aria-label={`Cambiar a tema ${next}`} title={`Cambiar a tema ${next}`}>
        <span className="sb-ico">{theme === "dark" ? ICON.sun : ICON.moon}</span>
        <span className="sb-text">Tema {theme === "dark" ? "oscuro" : "claro"}</span>
      </button>
    </aside>
  );
}
