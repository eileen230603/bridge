import React from "react";
import "./views.css";
import "./settingsView.css";
import { ConfirmDialog } from "./ConfirmDialog";

type Server = { protocol: "http" | "https"; host: string; port: number; timeoutSeconds: number; apiKey?: string };
type SCP = { aeTitle: string; port: number; autoStart: boolean };
type ModeId = "v1" | "v2" | "scp";
type ModeSettings = { active: ModeId; v1: Server; v2: Server; scp: SCP };
type Robot = { enabled: boolean; supported: boolean; discType: string; format: string; printMode: number; labelType: number; defaultCopies: number };
type Drives = { enabled: boolean; selected: string[]; verify: boolean; eject: boolean; defaultCopies: number };
type DestSettings = { robot: Robot; drives: Drives };
type DriveStatus = { id: string; name: string; label: string; canWriteCd: boolean; canWriteDvd: boolean; canWriteBd: boolean; selected: boolean; state: string; stateLabel: string };
type DestStatus = { robot: { enabled: boolean; available: boolean; message: string }; drives: DriveStatus[]; drivesError?: string };
type DiscLabel = { hospitalName: string; logoPath: string };
type Paper = { enabled: boolean; printer: string; size: string; auto: boolean };
type LabelSettings = { label: DiscLabel; paper: Paper; sizes?: { id: string; label: string }[] };
type Conn = { status: string; message: string };
type DiscProfile = { type: string; label: string; default: string; formats: { id: string; label: string }[]; note?: string };
type SectionId = "mode" | "destination" | "label" | "about";
type Editable = Exclude<SectionId, "about">;
type AppInfo = { name: string; company: string; version: string; commit: string; buildDate: string; devMode: boolean };

const api = () => (window as any).go?.main?.App;
const errText = (e: unknown) => String((e as any)?.message ?? e).replace(/^Error:\s*/, "");
const same = (a: unknown, b: unknown) => JSON.stringify(a) === JSON.stringify(b);

const I = (d: React.ReactNode, size = 20) => (
  <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">{d}</svg>
);
const ICON = {
  mode: I(<><path d="M4 6h10M18 6h2M4 12h4M12 12h8M4 18h12M20 18h0" /><circle cx="16" cy="6" r="2" /><circle cx="10" cy="12" r="2" /><circle cx="18" cy="18" r="2" /></>),
  destination: I(<><circle cx="12" cy="12" r="9" /><circle cx="12" cy="12" r="2.5" /></>),
  label: I(<><rect x="4" y="4" width="16" height="16" rx="3" /><path d="M8 14l3-3 3 3 2-2 2 2" /><circle cx="9" cy="9" r="1" /></>),
  about: I(<><circle cx="12" cy="12" r="9" /><path d="M12 11v5M12 7.5v.01" /></>),
  robot: I(<><path d="M6 9V3h12v6" /><rect x="3" y="9" width="18" height="9" rx="2" /><path d="M7 14h10v7H7z" /></>),
  disc: I(<><circle cx="12" cy="12" r="9" /><circle cx="12" cy="12" r="2.5" /></>),
  paper: I(<><rect x="4" y="3" width="16" height="18" rx="2" /><path d="M8 8h8M8 12h8M8 16h5" /></>),
  refresh: I(<><path d="M20 11a8 8 0 1 0-2.3 5.7" /><path d="M20 4v7h-7" /></>, 18),
  play: I(<path d="M8 5l11 7-11 7z" fill="currentColor" />, 16),
  eye: I(<><path d="M2 12s3.6-7 10-7 10 7 10 7-3.6 7-10 7S2 12 2 12z" /><circle cx="12" cy="12" r="3" /></>, 18),
  eyeOff: I(<><path d="M3 3l18 18M10.6 5.1A10 10 0 0 1 12 5c6.4 0 10 7 10 7a17 17 0 0 1-3.2 4.2M6.6 6.6A17 17 0 0 0 2 12s3.6 7 10 7a9.7 9.7 0 0 0 4.4-1" /></>, 18),
};
const SECTIONS: { id: SectionId; title: string }[] = [
  { id: "mode", title: "Modo" },
  { id: "destination", title: "Destino" },
  { id: "label", title: "Etiqueta" },
  { id: "about", title: "Acerca de" },
];
const MODES: { id: ModeId; title: string; desc: string }[] = [
  { id: "v1", title: "Servidor V1", desc: "Consulta los estudios en el servidor Symphony V1." },
  { id: "v2", title: "Servidor V2", desc: "Consulta el servidor Symphony V2 autenticándose con una API Key." },
  { id: "scp", title: "DICOM SCP", desc: "Recibe estudios por DICOM (StoreSCP) y los lee de la base local del sistema." },
];
const SECTION_NAME: Record<Editable, string> = { mode: "Modo", destination: "Destino", label: "Etiqueta" };

function Toggle({ on, onChange, disabled, label }: { on: boolean; onChange: (v: boolean) => void; disabled?: boolean; label: string }) {
  return <button type="button" role="switch" aria-checked={on} aria-label={label} className={`s-toggle ${on ? "on" : ""}`} disabled={disabled} onClick={() => onChange(!on)}><i /></button>;
}
function Check({ on, onChange, children, disabled }: { on: boolean; onChange: (v: boolean) => void; children?: React.ReactNode; disabled?: boolean }) {
  return (
    <label className={`s-check ${disabled ? "off" : ""}`}>
      <input type="checkbox" checked={on} disabled={disabled} onChange={(e) => onChange(e.target.checked)} />
      <span className="s-check-box" aria-hidden="true">{on && I(<path d="M5 12l5 5 9-10" />, 14)}</span>
      {children && <span>{children}</span>}
    </label>
  );
}
function Stepper({ value, onChange, label }: { value: number; onChange: (v: number) => void; label: string }) {
  const set = (v: number) => onChange(Math.min(99, Math.max(1, v || 1)));
  return (
    <div className="s-stepper" role="group" aria-label={label}>
      <button type="button" onClick={() => set(value - 1)} disabled={value <= 1} aria-label="Menos">−</button>
      <input value={value} inputMode="numeric" onChange={(e) => set(Number(e.target.value.replace(/\D/g, "")))} aria-label={label} />
      <button type="button" onClick={() => set(value + 1)} disabled={value >= 99} aria-label="Más">+</button>
    </div>
  );
}
function Pill({ tone, children, dot = true }: { tone: string; children: React.ReactNode; dot?: boolean }) {
  return <span className={`v-pill tone-${tone} s-pill`}>{dot && <i className="s-pill-dot" />}{children}</span>;
}

export function SettingsView({ onSaved, onDirtyChange }: { onBack?: () => void; onSaved: () => void; onDirtyChange?: (dirty: boolean) => void }) {
  const [section, setSection] = React.useState<SectionId>("mode");
  const [loaded, setLoaded] = React.useState(false);

  // Cada sección guarda su copia «guardada» para saber si hay cambios y poder descartarlos.
  const [mode, setMode] = React.useState<ModeSettings | null>(null);
  const [modeSaved, setModeSaved] = React.useState<ModeSettings | null>(null);
  const [dest, setDest] = React.useState<DestSettings | null>(null);
  const [destSaved, setDestSaved] = React.useState<DestSettings | null>(null);
  const [lbl, setLbl] = React.useState<LabelSettings | null>(null);
  const [lblSaved, setLblSaved] = React.useState<LabelSettings | null>(null);

  const [status, setStatus] = React.useState<DestStatus | null>(null);
  const [detecting, setDetecting] = React.useState(false);
  const [profiles, setProfiles] = React.useState<DiscProfile[]>([]);
  const [printers, setPrinters] = React.useState<string[] | null>(null);
  const [printerError, setPrinterError] = React.useState("");
  const [preview, setPreview] = React.useState("");
  const [labelError, setLabelError] = React.useState("");
  const [conn, setConn] = React.useState<Conn>({ status: "No probado", message: "" });
  const [showKey, setShowKey] = React.useState(false);
  const [info, setInfo] = React.useState<AppInfo | null>(null);
  const [busy, setBusy] = React.useState<Editable | "test" | null>(null);
  const [fb, setFb] = React.useState<Record<Editable, { ok: boolean; text: string } | null>>({ mode: null, destination: null, label: null });
  const [pending, setPending] = React.useState<SectionId | null>(null);
  const [sizes, setSizes] = React.useState<{ id: string; label: string }[]>([]);

  const dirty: Record<Editable, boolean> = {
    mode: !!mode && !same(mode, modeSaved),
    destination: !!dest && !same(dest, destSaved),
    label: !!lbl && !same(lbl, lblSaved),
  };
  const anyDirty = dirty.mode || dirty.destination || dirty.label;
  React.useEffect(() => { onDirtyChange?.(anyDirty); }, [anyDirty]);
  React.useEffect(() => () => onDirtyChange?.(false), []);

  React.useEffect(() => {
    (async () => {
      const b = api();
      const [m, d, l, p, i] = await Promise.all([b?.GetModeConfig(), b?.GetDestinationConfig(), b?.GetLabelSettings(), b?.GetDiscProfiles?.(), b?.GetAppInfo?.()]);
      if (m) { setMode(m); setModeSaved(m); }
      if (d) { setDest(d); setDestSaved(d); }
      if (l) { const v = { label: l.label, paper: l.paper }; setLbl(v); setLblSaved(v); if (l.sizes) setSizes(l.sizes); }
      if (p) setProfiles(p);
      if (i) setInfo(i);
      setLoaded(true);
    })().catch(() => setLoaded(true));
  }, []);

  const detect = React.useCallback(async () => {
    setDetecting(true);
    try { setStatus(await api()?.GetDestinationStatus()); } catch { /* se mantiene el último estado */ }
    finally { setDetecting(false); }
  }, []);
  // En Destino se vuelve a detectar cada pocos segundos para reflejar discos insertados.
  React.useEffect(() => {
    if (section !== "destination") return;
    detect();
    const t = window.setInterval(detect, 5000);
    return () => window.clearInterval(t);
  }, [section, detect]);

  const loadPrinters = React.useCallback(async () => {
    setPrinterError("");
    try { setPrinters((await api()?.ListPrinters()) ?? []); } catch (e) { setPrinters([]); setPrinterError(errText(e)); }
  }, []);
  React.useEffect(() => { if (section === "label" && printers === null) loadPrinters(); }, [section, printers, loadPrinters]);

  React.useEffect(() => {
    if (section !== "label" || !lbl) return;
    let active = true;
    const t = window.setTimeout(async () => {
      try {
        const p = await api()?.GetDiscLabelPreview(lbl.label);
        if (active) { setPreview(p ?? ""); setLabelError(""); }
      } catch {
        if (active) { setPreview(""); setLabelError("No se pudo cargar el logo seleccionado."); }
      }
    }, 250);
    return () => { active = false; window.clearTimeout(t); };
  }, [section, lbl?.label.hospitalName, lbl?.label.logoPath]);

  const setFeedback = (id: Editable, f: { ok: boolean; text: string } | null) => setFb((v) => ({ ...v, [id]: f }));

  async function save(id: Editable, override?: DestSettings) {
    setBusy(id);
    setFeedback(id, null);
    try {
      if (id === "mode" && mode) { await api()?.SaveModeConfig(mode); setModeSaved(mode); }
      if (id === "destination" && dest) { const v = override ?? dest; await api()?.SaveDestinationConfig(v); setDest(v); setDestSaved(v); detect(); }
      if (id === "label" && lbl) {
        await api()?.SaveLabelSettings(lbl);
        const fresh = await api()?.GetLabelSettings();
        const v = fresh ? { label: fresh.label, paper: fresh.paper } : lbl;
        setLbl(v); setLblSaved(v);
      }
      setFeedback(id, { ok: true, text: "Guardado" });
      onSaved();
    } catch (e) {
      setFeedback(id, { ok: false, text: errText(e) || "No se pudo guardar." });
    } finally {
      setBusy(null);
    }
  }
  function discard(id: Editable) {
    if (id === "mode") setMode(modeSaved);
    if (id === "destination") setDest(destSaved);
    if (id === "label") setLbl(lblSaved);
    setFeedback(id, null);
  }
  function go(next: SectionId) {
    if (next === section) return;
    if (section !== "about" && dirty[section as Editable]) { setPending(next); return; }
    setSection(next);
  }

  async function test() {
    if (!mode || mode.active === "scp") return;
    setBusy("test");
    try {
      const server = mode.active === "v1" ? mode.v1 : mode.v2;
      setConn((await api()?.TestServerConnection(server)) ?? { status: "Error", message: "No se pudo conectar al servidor." });
      onSaved();
    } finally { setBusy(null); }
  }
  async function pickLogo() {
    const path = await api()?.SelectDiscLabelLogo();
    if (path) setLbl((c) => c && { ...c, label: { ...c.label, logoPath: path } });
  }

  // ---------- Modo ----------
  const server = mode ? (mode.active === "v2" ? mode.v2 : mode.v1) : null;
  const setServer = (patch: Partial<Server>) =>
    setMode((m) => m && (m.active === "v2" ? { ...m, v2: { ...m.v2, ...patch } } : { ...m, v1: { ...m.v1, ...patch } }));
  const modeTitle = (id?: ModeId) => MODES.find((x) => x.id === id)?.title ?? "";

  // ---------- Destino ----------
  const robot = dest?.robot;
  const drives = dest?.drives;
  const setRobot = (patch: Partial<Robot>) => setDest((d) => d && { ...d, robot: { ...d.robot, ...patch } });
  const setDrives = (patch: Partial<Drives>) => setDest((d) => d && { ...d, drives: { ...d.drives, ...patch } });
  const profile = profiles.find((p) => p.type === robot?.discType);
  const pickDisc = (type: string) => { const p = profiles.find((x) => x.type === type); if (p) setRobot({ discType: type, format: p.default }); };
  const pickSurface = (labelType: number) =>
    setDest((d) => d && { ...d, robot: { ...d.robot, labelType, printMode: labelType === 3 ? 1 : d.robot.printMode === 2 && labelType !== 1 && labelType !== 2 ? 0 : d.robot.printMode } });
  const speedOptions = robot ? [
    ...(robot.labelType === 3 ? [] : [{ value: 0, label: "Según el equipo" }]),
    { value: 1, label: "Calidad" },
    ...(robot.labelType === 1 || robot.labelType === 2 ? [{ value: 2, label: "Rápida" }] : []),
  ] : [];
  const toggleDrive = (id: string, on: boolean) =>
    setDrives({ selected: on ? [...(drives?.selected ?? []), id] : (drives?.selected ?? []).filter((x) => x !== id) });
  const detected = status?.drives ?? [];
  const selectedReady = detected.filter((d) => drives?.selected.includes(d.id) && d.state === "listo").length;
  const selectedCount = detected.filter((d) => drives?.selected.includes(d.id)).length;
  const kinds = (d: DriveStatus) => [d.canWriteCd && "CD", d.canWriteDvd && "DVD", d.canWriteBd && "Blu-ray"].filter(Boolean).join(" · ") || "Solo lectura";
  const driveTone = (s: string) => (s === "listo" ? "ok" : s === "sin-disco" ? "slate" : "amber");
  const enabledNames = dest ? [dest.robot.enabled && "Robot Epson", dest.drives.enabled && `Unidades${selectedCount ? ` (${detected.filter((d) => drives?.selected.includes(d.id)).map((d) => d.label).join(", ")})` : ""}`].filter(Boolean).join(" · ") : "";
  const destError = fb.destination && !fb.destination.ok ? fb.destination.text : "";
  const fallback: DestSettings | null =
    dest && destError.startsWith("Robot Epson no disponible") && dest.drives.enabled ? { ...dest, robot: { ...dest.robot, enabled: false } }
    : dest && destError.startsWith("Unidades CD/DVD no disponibles") && dest.robot.enabled ? { ...dest, drives: { ...dest.drives, enabled: false } }
    : null;

  // ---------- Pie de guardado común ----------
  const Footer = ({ id, detail, extra }: { id: Editable; detail: React.ReactNode; extra?: React.ReactNode }) => {
    const f = fb[id];
    const isDirty = dirty[id];
    const tone = f && !f.ok ? "bad" : isDirty ? "warn" : "ok";
    return (
      <div className="s-footbar">
        <i className={`s-foot-dot ${tone}`} />
        <div className="s-foot-text">
          <strong className={tone}>{isDirty ? "Cambios sin guardar" : f && !f.ok ? "No se guardó" : "Guardado"}</strong>
          <span className={f && !f.ok ? "bad" : ""}>{f && !f.ok ? f.text : detail}</span>
        </div>
        {extra}
        {isDirty && <button className="v-btn v-btn-ghost" onClick={() => discard(id)} disabled={busy === id}>Descartar</button>}
        <button className="v-btn v-btn-primary" disabled={!loaded || busy === id || (!isDirty && !(f && !f.ok))} onClick={() => save(id)}>
          {busy === id ? <span className="v-spin" /> : null} Guardar
        </button>
      </div>
    );
  };

  return (
    <div className="v-wrap">
      <div className="v-toolbar">
        <h2 className="v-title">Configuración</h2>
      </div>

      <div className="v-panel s-layout">
        <nav className="s-side" aria-label="Secciones de configuración">
          {SECTIONS.map((s) => (
            <button key={s.id} className={`s-nav ${section === s.id ? "on" : ""}`} onClick={() => go(s.id)}>
              {ICON[s.id]}
              <span>{s.title}</span>
              {s.id !== "about" && dirty[s.id as Editable] && <i className="s-nav-dirty" title="Cambios sin guardar" />}
            </button>
          ))}
        </nav>

        <div className="s-body">
          {!loaded && <div className="s-content"><p className="s-hint">Cargando configuración…</p></div>}

          {loaded && section === "mode" && mode && server && (
            <>
              <div className="s-content">
                <section className="s-section">
                  <h3>Modo de trabajo</h3>
                  <div className="s-seg" role="radiogroup" aria-label="Modo de trabajo">
                    {MODES.map((m) => (
                      <button key={m.id} role="radio" aria-checked={mode.active === m.id} className={mode.active === m.id ? "on" : ""} onClick={() => { setMode({ ...mode, active: m.id }); setConn({ status: "No probado", message: "" }); }}>
                        {m.title}
                      </button>
                    ))}
                  </div>
                  <p className="s-hint">{MODES.find((m) => m.id === mode.active)?.desc}</p>

                  {mode.active !== "scp" && (
                    <>
                      <div className="s-grid">
                        <label className="s-wide">Servidor / IP
                          <input value={server.host} onChange={(e) => setServer({ host: e.target.value })} placeholder="192.168.0.10" />
                        </label>
                        <label>Puerto
                          <input type="number" min="1" max="65535" value={server.port} onChange={(e) => setServer({ port: Number(e.target.value) })} />
                        </label>
                        <label>Protocolo
                          <select value={server.protocol} onChange={(e) => setServer({ protocol: e.target.value as "http" | "https" })}>
                            <option value="http">http</option>
                            <option value="https">https</option>
                          </select>
                        </label>
                        <label>Tiempo de espera (s)
                          <input type="number" min="1" value={server.timeoutSeconds} onChange={(e) => setServer({ timeoutSeconds: Number(e.target.value) })} />
                        </label>
                        {mode.active === "v2" && (
                          <label className="s-wide">API Key
                            <div className="s-secret">
                              <input type={showKey ? "text" : "password"} value={server.apiKey ?? ""} autoComplete="off" spellCheck={false} onChange={(e) => setServer({ apiKey: e.target.value })} placeholder="Clave entregada por el servidor V2" />
                              <button type="button" className="s-secret-btn" onClick={() => setShowKey((v) => !v)} aria-label={showKey ? "Ocultar clave" : "Mostrar clave"} title={showKey ? "Ocultar clave" : "Mostrar clave"}>{showKey ? ICON.eyeOff : ICON.eye}</button>
                            </div>
                            <span className="s-field-hint">Se envía en cada petición para autenticar la comunicación con el servidor.</span>
                          </label>
                        )}
                      </div>
                      <div className="s-test">
                        <button className="v-btn v-btn-ghost" disabled={busy === "test"} onClick={test}>
                          {busy === "test" ? <span className="v-spin" /> : null} Probar conexión
                        </button>
                        <span className={`s-conn ${conn.status === "Conectado" ? "ok" : conn.status === "Error" ? "bad" : ""}`}>
                          <i className="v-dot" /> {conn.status === "Conectado" && mode.active === "v2" ? "Conectado · clave válida" : conn.status}
                          {conn.message && <small> · {conn.message}</small>}
                        </span>
                      </div>
                    </>
                  )}

                  {mode.active === "scp" && (
                    <>
                      <div className="s-group">
                        <h4>Receptor DICOM (StoreSCP)</h4>
                        <div className="s-grid">
                          <label>AE Title
                            <input value={mode.scp.aeTitle} maxLength={16} onChange={(e) => setMode({ ...mode, scp: { ...mode.scp, aeTitle: e.target.value.toUpperCase() } })} />
                          </label>
                          <label>Puerto
                            <input type="number" min="1" max="65535" value={mode.scp.port} onChange={(e) => setMode({ ...mode, scp: { ...mode.scp, port: Number(e.target.value) } })} />
                          </label>
                        </div>
                        <div className="s-scp-state">
                          <span className="s-conn"><i className="v-dot" /> Servidor apagado</span>
                          <button className="v-btn v-btn-primary s-btn-sm" disabled title="El receptor DICOM todavía no está disponible en esta versión.">{ICON.play} Iniciar</button>
                        </div>
                        <Check on={mode.scp.autoStart} onChange={(v) => setMode({ ...mode, scp: { ...mode.scp, autoStart: v } })}>Iniciar al abrir la aplicación</Check>
                      </div>
                      <div className="s-group">
                        <h4>Direcciones en escucha</h4>
                        <p className="s-hint">Sin direcciones en escucha. El servidor está detenido.</p>
                      </div>
                      <p className="s-hint warn">El receptor DICOM y la búsqueda en la base local todavía no están disponibles en esta versión. La configuración se guarda.</p>
                    </>
                  )}
                </section>
              </div>
              <Footer id="mode" detail={`Modo de trabajo: ${modeTitle(mode.active)}`} />
            </>
          )}

          {loaded && section === "destination" && dest && robot && drives && (
            <>
              <div className="s-content">
                <section className="s-section">
                  <h3>Destinos de grabación</h3>
                  <p className="s-hint">Habilita los destinos disponibles. El destino se elige al pulsar Grabar en cada estudio.</p>

                  <div className={`s-card ${!robot.supported ? "dim" : ""} ${status && robot.enabled && !status.robot.available ? "bad" : ""}`}>
                    <div className="s-card-head">
                      {ICON.robot}
                      <h4>Robot Epson</h4>
                      {!robot.supported ? <Pill tone="slate">Solo Windows</Pill>
                        : !robot.enabled ? <Pill tone="slate">Deshabilitado</Pill>
                        : !status ? <Pill tone="slate">Comprobando…</Pill>
                        : status.robot.available ? <Pill tone="ok">Listo</Pill> : <Pill tone="bad">No disponible</Pill>}
                      <span className="v-spacer" />
                      <Toggle on={robot.enabled} disabled={!robot.supported} onChange={(v) => setRobot({ enabled: v })} label="Habilitar robot Epson" />
                    </div>
                    <p className={`s-hint ${status && robot.enabled && !status.robot.available ? "bad" : ""}`}>
                      {!robot.supported ? "Disponible solo en Windows." : status && robot.enabled && !status.robot.available ? status.robot.message : "Graba e imprime sobre el disco mediante TD Bridge."}
                    </p>
                    {robot.supported && (
                      <fieldset className="s-card-body" disabled={!robot.enabled}>
                        <div className="s-row">
                          <div className="s-field">
                            <span className="s-lbl">Tipo de disco</span>
                            <div className="s-seg" role="radiogroup" aria-label="Tipo de disco">
                              {profiles.map((p) => (
                                <button key={p.type} type="button" role="radio" aria-checked={robot.discType === p.type} className={robot.discType === p.type ? "on" : ""} onClick={() => pickDisc(p.type)}>{p.label}</button>
                              ))}
                            </div>
                          </div>
                          <label className="s-field">Formato
                            <select value={robot.format} onChange={(e) => setRobot({ format: e.target.value })}>
                              {profile?.formats.map((f) => (<option key={f.id} value={f.id}>{f.label}</option>))}
                            </select>
                          </label>
                          <div className="s-field">
                            <span className="s-lbl">Copias</span>
                            <Stepper value={robot.defaultCopies} onChange={(v) => setRobot({ defaultCopies: v })} label="Copias por defecto del robot" />
                          </div>
                        </div>
                        {profile && robot.format !== profile.default && <p className="s-hint">Formato recomendado para {profile.label}: {profile.formats.find((f) => f.id === profile.default)?.label}.</p>}
                        {profile?.note && <p className="s-hint warn">{profile.note}</p>}
                        <div className="s-grid">
                          <label>Impresión de etiqueta
                            <select value={robot.labelType} onChange={(e) => pickSurface(Number(e.target.value))}>
                              <option value={0}>Según el equipo</option>
                              <option value={1}>Estándar CD/DVD</option>
                              <option value={2}>Alta calidad CD/DVD</option>
                              <option value={3}>Brillante (Epson)</option>
                            </select>
                          </label>
                          <label>Calidad de impresión
                            <select value={robot.printMode} onChange={(e) => setRobot({ printMode: Number(e.target.value) })}>
                              {speedOptions.map((o) => (<option key={o.value} value={o.value}>{o.label}</option>))}
                            </select>
                          </label>
                        </div>
                      </fieldset>
                    )}
                  </div>

                  <div className="s-card">
                    <div className="s-card-head">
                      {ICON.disc}
                      <h4>Unidades CD/DVD</h4>
                      {!status ? <Pill tone="slate">Comprobando…</Pill>
                        : detected.length === 0 ? <Pill tone="slate">Sin unidades</Pill>
                        : selectedCount === 0 ? <Pill tone="slate">{detected.length} detectada{detected.length > 1 ? "s" : ""}</Pill>
                        : <Pill tone={selectedReady ? "ok" : "slate"}>{selectedReady} de {selectedCount} lista{selectedCount > 1 ? "s" : ""}</Pill>}
                      <span className="v-spacer" />
                      <button className="v-btn v-btn-ghost s-btn-sm" onClick={detect} disabled={detecting}>{detecting ? <span className="v-spin" /> : ICON.refresh} Actualizar</button>
                      <Toggle on={drives.enabled} onChange={(v) => setDrives({ enabled: v })} label="Habilitar unidades CD/DVD" />
                    </div>
                    <p className="s-hint">El tipo de disco y el formato se detectan del disco insertado. Windows y macOS.</p>
                    <fieldset className="s-card-body" disabled={!drives.enabled}>
                      {status?.drivesError ? <p className="s-hint bad">{status.drivesError}</p>
                        : detected.length === 0 ? <p className="s-hint">{status ? "No se detectaron grabadoras conectadas a este equipo." : "Buscando unidades…"}</p>
                        : (
                          <ul className="s-drives">
                            {detected.map((d) => (
                              <li key={d.id}>
                                <Check on={drives.selected.includes(d.id)} onChange={(v) => toggleDrive(d.id, v)} />
                                <span className="s-drive-name"><b>{d.label}</b> {d.name}</span>
                                <span className="s-drive-kinds">{kinds(d)}</span>
                                <Pill tone={driveTone(d.state)}>{d.stateLabel}</Pill>
                              </li>
                            ))}
                          </ul>
                        )}
                      <div className="s-row s-row-center">
                        <Check on={drives.verify} onChange={(v) => setDrives({ verify: v })}>Verificar al terminar</Check>
                        <Check on={drives.eject} onChange={(v) => setDrives({ eject: v })}>Expulsar al terminar</Check>
                        <span className="s-inline-field">
                          <span className="s-lbl">Copias</span>
                          <Stepper value={drives.defaultCopies} onChange={(v) => setDrives({ defaultCopies: v })} label="Copias por defecto en unidades" />
                        </span>
                      </div>
                    </fieldset>
                  </div>
                </section>
              </div>
              <Footer
                id="destination"
                detail={enabledNames ? `Habilitados: ${enabledNames}` : "Sin destinos habilitados"}
                extra={fallback && (
                  <button className="v-btn v-btn-ghost" disabled={busy === "destination"} onClick={() => save("destination", fallback)}>
                    {fallback.robot.enabled ? "Guardar sin las unidades" : "Guardar sin el robot"}
                  </button>
                )}
              />
            </>
          )}

          {loaded && section === "label" && lbl && (
            <>
              <div className="s-content">
                <section className="s-section">
                  <h3>Etiqueta</h3>
                  <p className="s-hint">Se imprime sobre el disco con el robot, o en papel adhesivo al grabar en una unidad.</p>
                  <div className="s-label">
                    <div className="s-label-form">
                      <label>Nombre del hospital
                        <input value={lbl.label.hospitalName} placeholder="Nombre del hospital" onChange={(e) => setLbl({ ...lbl, label: { ...lbl.label, hospitalName: e.target.value } })} />
                      </label>
                      <label>Logo
                        <div className="s-logo">
                          <input value={lbl.label.logoPath} placeholder="Logo predeterminado" readOnly />
                          <button className="v-btn v-btn-ghost" onClick={pickLogo}>Elegir</button>
                          {lbl.label.logoPath && <button className="v-btn v-btn-ghost" onClick={() => setLbl({ ...lbl, label: { ...lbl.label, logoPath: "" } })}>Quitar</button>}
                        </div>
                      </label>
                      {labelError && <p className="s-hint bad">{labelError}</p>}
                    </div>
                    <div className="s-preview">
                      {preview ? <img src={preview} alt="Vista previa de la etiqueta" /> : <span>Generando vista previa…</span>}
                    </div>
                  </div>

                  <div className="s-card">
                    <div className="s-card-head">
                      {ICON.paper}
                      <h4>Etiqueta en papel</h4>
                      <Pill tone="blue" dot={false}>Solo unidades CD/DVD</Pill>
                      <span className="v-spacer" />
                      <Toggle on={lbl.paper.enabled} onChange={(v) => setLbl({ ...lbl, paper: { ...lbl.paper, enabled: v } })} label="Habilitar etiqueta en papel" />
                    </div>
                    <p className="s-hint">Para pegar en el disco o en la caja. El robot Epson imprime directamente sobre el disco.</p>
                    <fieldset className="s-card-body" disabled={!lbl.paper.enabled}>
                      <div className="s-grid">
                        <label>Impresora
                          <div className="s-logo">
                            <select value={lbl.paper.printer} onChange={(e) => setLbl({ ...lbl, paper: { ...lbl.paper, printer: e.target.value } })}>
                              <option value="">{printers === null ? "Buscando impresoras…" : "Elija una impresora"}</option>
                              {lbl.paper.printer && printers && !printers.includes(lbl.paper.printer) && <option value={lbl.paper.printer}>{lbl.paper.printer} (no encontrada)</option>}
                              {(printers ?? []).map((p) => <option key={p} value={p}>{p.replace(/_/g, " ")}</option>)}
                            </select>
                            <button type="button" className="v-btn v-btn-ghost s-icon-btn" onClick={loadPrinters} title="Actualizar impresoras" aria-label="Actualizar impresoras">{ICON.refresh}</button>
                          </div>
                        </label>
                        <label>Tamaño de etiqueta
                          <select value={lbl.paper.size} onChange={(e) => setLbl({ ...lbl, paper: { ...lbl.paper, size: e.target.value } })}>
                            {sizes.map((s) => <option key={s.id} value={s.id}>{s.label}</option>)}
                          </select>
                        </label>
                      </div>
                      {printerError && <p className="s-hint bad">{printerError}</p>}
                      <div className="s-field">
                        <span className="s-lbl">Impresión</span>
                        <div className="s-seg" role="radiogroup" aria-label="Impresión">
                          {[{ v: true, t: "Automática" }, { v: false, t: "Manual" }].map((o) => (
                            <button key={o.t} type="button" role="radio" aria-checked={lbl.paper.auto === o.v} className={lbl.paper.auto === o.v ? "on" : ""} onClick={() => setLbl({ ...lbl, paper: { ...lbl.paper, auto: o.v } })}>{o.t}</button>
                          ))}
                        </div>
                      </div>
                      <p className="s-hint">{lbl.paper.auto ? "Automática: se imprime solo cuando la grabación se verifica con éxito." : "Manual: la etiqueta se imprime desde el Historial cuando la necesite."}</p>
                    </fieldset>
                  </div>
                </section>
              </div>
              <Footer id="label" detail="Se aplica al robot Epson y a la etiqueta en papel" />
            </>
          )}

          {section === "about" && info && (
            <div className="s-content">
              <section className="s-section">
                <h3>Acerca de</h3>
                <dl className="s-about">
                  <dt>Aplicación</dt><dd>{info.name}</dd>
                  <dt>Versión</dt><dd>{info.version}{info.devMode ? " (modo desarrollo)" : ""}</dd>
                  <dt>Compilación</dt><dd>{info.commit}{info.buildDate ? ` · ${info.buildDate}` : ""}</dd>
                  <dt>Desarrollado por</dt><dd>{info.company}</dd>
                </dl>
              </section>
            </div>
          )}
        </div>
      </div>

      {pending && section !== "about" && (
        <ConfirmDialog
          title="Cambios sin guardar"
          message={`Hay cambios sin guardar en ${SECTION_NAME[section as Editable]}. Si continúa, se descartan.`}
          confirm="Descartar cambios"
          cancel="Seguir editando"
          onCancel={() => setPending(null)}
          onConfirm={() => { discard(section as Editable); setSection(pending); setPending(null); }}
        />
      )}
    </div>
  );
}
