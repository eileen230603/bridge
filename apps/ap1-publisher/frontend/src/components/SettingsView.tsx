import React from "react";
import "./views.css";
import "./settingsView.css";

type ServerConfig = { protocol: "http" | "https"; host: string; port: number; timeoutSeconds: number };
type EpsonConfig = { discType: string; format: string; printMode: number; labelType: number };
type DiscLabelConfig = { hospitalName: string; logoPath: string };
type Conn = { status: string; message: string };
type DiscProfile = { type: string; label: string; default: string; formats: { id: string; label: string }[]; note?: string };
type SectionId = "server" | "burn" | "label" | "about";
type AppInfo = { name: string; company: string; version: string; commit: string; buildDate: string; devMode: boolean };

const api = () => (window as any).go?.main?.App;
const errText = (e: unknown) => String((e as any)?.message ?? e).replace(/^Error:\s*/, "");

const I = (d: React.ReactNode) => (
  <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">{d}</svg>
);
const ICONS: Record<SectionId, React.ReactNode> = {
  server: I(<><rect x="3" y="4" width="18" height="6" rx="2" /><rect x="3" y="14" width="18" height="6" rx="2" /><path d="M7 7h.01M7 17h.01" /></>),
  burn: I(<><circle cx="12" cy="12" r="9" /><circle cx="12" cy="12" r="2.5" /></>),
  about: I(<><circle cx="12" cy="12" r="9" /><path d="M12 11v5M12 7.5v.01" /></>),
  label: I(<><rect x="4" y="4" width="16" height="16" rx="3" /><path d="M8 14l3-3 3 3 2-2 2 2" /><circle cx="9" cy="9" r="1" /></>),
};
const SECTIONS: { id: SectionId; title: string }[] = [
  { id: "server", title: "Servidor" },
  { id: "burn", title: "Grabado" },
  { id: "label", title: "Etiqueta" },
  { id: "about", title: "Acerca de" },
];

function Feedback({ f }: { f: { ok: boolean; text: string } | null }) {
  if (!f) return null;
  return <span className={`s-feedback ${f.ok ? "ok" : "bad"}`} role="status">{f.ok ? "✓ " : ""}{f.text}</span>;
}

export function SettingsView({ onBack, onSaved }: { onBack: () => void; onSaved: () => void }) {
  const [section, setSection] = React.useState<SectionId>("server");
  const [server, setServer] = React.useState<ServerConfig>({ protocol: "http", host: "", port: 4000, timeoutSeconds: 60 });
  const [epson, setEpson] = React.useState<EpsonConfig>({ discType: "DVD", format: "UDF102", printMode: 0, labelType: 0 });
  const [label, setLabel] = React.useState<DiscLabelConfig>({ hospitalName: "", logoPath: "" });
  const [preview, setPreview] = React.useState("");
  const [labelError, setLabelError] = React.useState("");
  const [conn, setConn] = React.useState<Conn>({ status: "No probado", message: "" });
  const [profiles, setProfiles] = React.useState<DiscProfile[]>([]);
  const [advanced, setAdvanced] = React.useState(false);
  React.useEffect(() => { api()?.GetDiscProfiles?.().then(setProfiles).catch(() => {}); }, []);
  const profile = profiles.find((p) => p.type === epson.discType);
  const formatLabel = (id: string) => profile?.formats.find((f) => f.id === id)?.label ?? id;
  // Al cambiar de disco, el formato pasa al recomendado de ese disco: nunca queda una combinación inválida.
  const pickDisc = (type: string) => {
    const p = profiles.find((x) => x.type === type);
    if (!p) return;
    setEpson((v) => ({ ...v, discType: type, format: p.default }));
    setAdvanced(false);
  };
  const pickSurface = (labelType: number) =>
    setEpson((v) => ({ ...v, labelType, printMode: labelType === 3 ? 1 : v.printMode === 2 && labelType !== 1 && labelType !== 2 ? 0 : v.printMode }));
  const speedOptions = [
    ...(epson.labelType === 3 ? [] : [{ value: 0, label: "Según el equipo" }]),
    { value: 1, label: "Calidad" },
    ...(epson.labelType === 1 || epson.labelType === 2 ? [{ value: 2, label: "Rápida" }] : []),
  ];
  const [info, setInfo] = React.useState<AppInfo | null>(null);
  React.useEffect(() => { api()?.GetAppInfo?.().then(setInfo).catch(() => {}); }, []);
  const [busy, setBusy] = React.useState<SectionId | "test" | null>(null);
  const [fb, setFb] = React.useState<Record<SectionId, { ok: boolean; text: string } | null>>({ server: null, burn: null, label: null, about: null });

  React.useEffect(() => {
    (async () => {
      const b = api();
      const [s, l, e] = await Promise.all([b?.GetServerConfig(), b?.GetDiscLabelConfig(), b?.GetEpsonConfig?.()]);
      if (s) setServer(s);
      if (l) setLabel(l);
      if (e) setEpson({ discType: e.discType || "DVD", format: e.format || "UDF102", printMode: e.printMode ?? 0, labelType: e.labelType ?? 0 });
    })();
  }, []);

  React.useEffect(() => {
    if (section !== "label") return;
    let active = true;
    const t = window.setTimeout(async () => {
      try {
        const p = await api()?.GetDiscLabelPreview(label);
        if (active) { setPreview(p ?? ""); setLabelError(""); }
      } catch {
        if (active) { setPreview(""); setLabelError("No se pudo cargar el logo seleccionado."); }
      }
    }, 250);
    return () => { active = false; window.clearTimeout(t); };
  }, [section, label]);

  const setFeedback = (id: SectionId, f: { ok: boolean; text: string } | null) => setFb((v) => ({ ...v, [id]: f }));

  async function save(id: SectionId) {
    setBusy(id);
    setFeedback(id, null);
    try {
      if (id === "server") await api()?.SaveServerConfig(server);
      if (id === "burn") await api()?.SaveEpsonConfig?.(epson);
      if (id === "label") await api()?.SaveDiscLabelConfig(label);
      setFeedback(id, { ok: true, text: "Guardado" });
      onSaved();
    } catch (e) {
      setFeedback(id, { ok: false, text: errText(e) || "No se pudo guardar." });
    } finally {
      setBusy(null);
    }
  }
  async function test() {
    setBusy("test");
    try {
      setConn((await api()?.TestServerConnection(server)) ?? { status: "Error", message: "No se pudo conectar al servidor." });
      onSaved();
    } finally {
      setBusy(null);
    }
  }
  async function pickLogo() {
    const path = await api()?.SelectDiscLabelLogo();
    if (path) setLabel((c) => ({ ...c, logoPath: path }));
  }

  const SaveRow = ({ id }: { id: SectionId }) => (
    <div className="s-actions">
      <Feedback f={fb[id]} />
      <button className="v-btn v-btn-primary" disabled={busy === id} onClick={() => save(id)}>
        {busy === id ? <span className="v-spin" /> : null} Guardar
      </button>
    </div>
  );

  return (
    <div className="v-wrap">
      <div className="v-toolbar">
        <h2 className="v-title">Configuración</h2>
      </div>

      <div className="v-panel s-layout">
        <nav className="s-side" aria-label="Secciones de configuración">
          {SECTIONS.map((s) => (
            <button key={s.id} className={`s-nav ${section === s.id ? "on" : ""}`} onClick={() => setSection(s.id)}>
              {ICONS[s.id]}
              <span>{s.title}</span>
            </button>
          ))}
        </nav>

        <div className="s-body">
          {section === "server" && (
            <section className="s-section">
              <h3>Servidor de estudios</h3>
              <div className="s-grid">
                <label className="s-wide">Servidor / IP
                  <input value={server.host} onChange={(e) => setServer((v) => ({ ...v, host: e.target.value }))} placeholder="192.168.0.10" />
                </label>
                <label>Puerto
                  <input type="number" min="1" max="65535" value={server.port} onChange={(e) => setServer((v) => ({ ...v, port: Number(e.target.value) }))} />
                </label>
                <label>Protocolo
                  <select value={server.protocol} onChange={(e) => setServer((v) => ({ ...v, protocol: e.target.value as "http" | "https" }))}>
                    <option value="http">http</option>
                    <option value="https">https</option>
                  </select>
                </label>
                <label>Tiempo de espera (s)
                  <input type="number" min="1" value={server.timeoutSeconds} onChange={(e) => setServer((v) => ({ ...v, timeoutSeconds: Number(e.target.value) }))} />
                </label>
              </div>
              <div className="s-test">
                <button className="v-btn v-btn-ghost" disabled={busy === "test"} onClick={test}>
                  {busy === "test" ? <span className="v-spin" /> : null} Probar conexión
                </button>
                <span className={`s-conn ${conn.status === "Conectado" ? "ok" : conn.status === "Error" ? "bad" : ""}`}>
                  <i className="v-dot" /> {conn.status}
                  {conn.message && <small> · {conn.message}</small>}
                </span>
              </div>
              <SaveRow id="server" />
            </section>
          )}

          {section === "burn" && (
            <section className="s-section">
              <h3>Grabado</h3>

              <div className="s-group">
                <h4>Disco</h4>
                <div className="s-seg" role="radiogroup" aria-label="Tipo de disco">
                  {profiles.map((p) => (
                    <button key={p.type} role="radio" aria-checked={epson.discType === p.type} className={epson.discType === p.type ? "on" : ""} onClick={() => pickDisc(p.type)}>
                      {p.label}
                    </button>
                  ))}
                </div>
                {profile && (
                  <div className="s-format">
                    <span className="s-format-label">Formato</span>
                    <span className="s-format-value">
                      {formatLabel(epson.format)}
                      {epson.format === profile.default && <em> · recomendado</em>}
                    </span>
                    {profile.formats.length > 1 && (
                      <button className="v-link" onClick={() => setAdvanced((v) => !v)}>{advanced ? "Ocultar" : "Cambiar"}</button>
                    )}
                    {advanced && profile.formats.length > 1 && (
                      <select className="s-format-select" value={epson.format} onChange={(e) => setEpson((v) => ({ ...v, format: e.target.value }))} aria-label="Formato de archivos">
                        {profile.formats.map((f) => (
                          <option key={f.id} value={f.id}>{f.label}{f.id === profile.default ? " (recomendado)" : ""}</option>
                        ))}
                      </select>
                    )}
                  </div>
                )}
                {profile?.note && <p className="s-hint warn">{profile.note}</p>}
              </div>

              <div className="s-group">
                <h4>Impresión de la etiqueta</h4>
                <div className="s-grid">
                  <label>Superficie imprimible
                    <select value={epson.labelType} onChange={(e) => pickSurface(Number(e.target.value))}>
                      <option value={0}>Según el equipo</option>
                      <option value={1}>Estándar CD/DVD</option>
                      <option value={2}>Alta calidad CD/DVD</option>
                      <option value={3}>Brillante (Epson)</option>
                    </select>
                  </label>
                  <label>Velocidad de impresión
                    <select value={epson.printMode} onChange={(e) => setEpson((v) => ({ ...v, printMode: Number(e.target.value) }))}>
                      {speedOptions.map((o) => (<option key={o.value} value={o.value}>{o.label}</option>))}
                    </select>
                  </label>
                </div>
                <p className="s-hint">{epson.labelType === 3 ? "La superficie brillante imprime siempre en calidad." : epson.labelType === 0 ? "Con «Según el equipo» no se puede forzar la impresión rápida." : "La impresión rápida está disponible con esta superficie."}</p>
              </div>
              <SaveRow id="burn" />
            </section>
          )}

          {section === "label" && (
            <section className="s-section">
              <h3>Etiqueta del disco</h3>
              <div className="s-label">
                <div className="s-label-form">
                  <label>Nombre del hospital
                    <input value={label.hospitalName} placeholder="Nombre del hospital" onChange={(e) => setLabel((c) => ({ ...c, hospitalName: e.target.value }))} />
                  </label>
                  <label>Logo
                    <div className="s-logo">
                      <input value={label.logoPath} placeholder="Logo predeterminado" readOnly />
                      <button className="v-btn v-btn-ghost" onClick={pickLogo}>Elegir</button>
                      {label.logoPath && <button className="v-btn v-btn-ghost" onClick={() => setLabel((c) => ({ ...c, logoPath: "" }))}>Quitar</button>}
                    </div>
                  </label>
                  {labelError && <p className="s-hint bad">{labelError}</p>}
                </div>
                <div className="s-preview">
                  {preview ? <img src={preview} alt="Vista previa de la etiqueta" /> : <span>Generando vista previa…</span>}
                </div>
              </div>
              <SaveRow id="label" />
            </section>
          )}
          {section === "about" && info && (
            <section className="s-section">
              <h3>Acerca de</h3>
              <dl className="s-about">
                <dt>Aplicación</dt><dd>{info.name}</dd>
                <dt>Versión</dt><dd>{info.version}{info.devMode ? " (modo desarrollo)" : ""}</dd>
                <dt>Compilación</dt><dd>{info.commit}{info.buildDate ? ` · ${info.buildDate}` : ""}</dd>
                <dt>Desarrollado por</dt><dd>{info.company}</dd>
              </dl>
            </section>
          )}
        </div>
      </div>
    </div>
  );
}
