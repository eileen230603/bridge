import React from "react";
import { createRoot } from "react-dom/client";
import "./style.css";
import "./jobs.css";
import "./settings.css";
type Study = {
  studyInstanceUID: string;
  patientName: string;
  studyDescription: string;
  studyDate: string;
  modality: string;
  seriesCount: number;
  instanceCount: number;
};

type DiscJob = {
  id: string;
  patientName: string;
  studyDescription: string;
  status: string;
  epsonState?: string;
  errorCode?: string;
  technicalStatus?: string;
  detailStatus?: string;
  errorMessage?: string;
};
type SystemStatus = {
  studyServer: string;
  studyServerAddress: string;
  studyApiConfigured: boolean;
  tdBridge: string;
  tdBridgeConfigured: boolean;
};
type ServerConfig = {
  protocol: "http" | "https";
  host: string;
  port: number;
  timeoutSeconds: number;
};
type ConnectionTestResult = { status: string; message: string };

const api = () => window.go?.main?.App;

function App() {
  const today = new Date().toISOString().slice(0, 10);
  const prior = new Date(Date.now() - 7 * 864e5).toISOString().slice(0, 10);
  const [from, setFrom] = React.useState(prior);
  const [to, setTo] = React.useState(today);
  const [studies, setStudies] = React.useState<Study[]>([]);
  const [jobs, setJobs] = React.useState<DiscJob[]>([]);
  const [loading, setLoading] = React.useState(false);
  const [message, setMessage] = React.useState("Listo para buscar estudios");
  const [status, setStatus] = React.useState<SystemStatus>({
    studyServer: "No probado",
    studyServerAddress: "",
    studyApiConfigured: false,
    tdBridge: "Error",
    tdBridgeConfigured: false,
  });
  const [settingsOpen, setSettingsOpen] = React.useState(false);
  const [serverConfig, setServerConfig] = React.useState<ServerConfig>({
    protocol: "http",
    host: "192.168.0.102",
    port: 4000,
    timeoutSeconds: 60,
  });
  const [connection, setConnection] = React.useState<ConnectionTestResult>({
    status: "No probado",
    message: "",
  });
  const [settingsBusy, setSettingsBusy] = React.useState(false);
  async function openSettings() {
    const value = await api()?.GetServerConfig();
    if (value) setServerConfig(value);
    setConnection({ status: "No probado", message: "" });
    setSettingsOpen(true);
  }
  async function testConnection() {
    setSettingsBusy(true);
    try {
      setConnection(
        (await api()?.TestServerConnection(serverConfig)) ?? {
          status: "Error",
          message: "No se pudo conectar al servidor.",
        },
      );
      await refreshStatus();
    } finally {
      setSettingsBusy(false);
    }
  }
  async function saveSettings() {
    setSettingsBusy(true);
    try {
      await api()?.SaveServerConfig(serverConfig);
      setSettingsOpen(false);
      await refreshStatus();
      setMessage("Configuración del servidor guardada.");
    } catch (e) {
      setConnection({ status: "Error", message: friendlyError(e) });
    } finally {
      setSettingsBusy(false);
    }
  }
  async function refreshStatus() {
    const value = await api()?.GetSystemStatus();
    if (value) setStatus(value);
  }
  async function search() {
    setLoading(true);
    try {
      const x = await api()?.SearchStudies(from, to);
      setStudies(x ?? []);
      setMessage(
        (x?.length ?? 0) === 0
          ? "No se encontraron estudios en ese rango."
          : `${x?.length ?? 0} estudios encontrados`,
      );
    } catch (e) {
      setStudies([]);
      setMessage(String(e));
    } finally {
      await refreshStatus();
      setLoading(false);
    }
  }
  async function publish(s: Study) {
    setMessage(`Preparando ${s.studyDescription}…`);
    try {
      const j = await api()?.CreateDiscJob(s.studyInstanceUID);
      if (j) setJobs((v) => [j, ...v]);
      setMessage("Trabajo entregado a TD Bridge");
    } catch (e) {
      setMessage(`Error: ${String(e)}`);
      const x = await api()?.ListJobs();
      setJobs(x ?? []);
    }
  }
  React.useEffect(() => {
    refreshStatus();
    let active = true;
    const refreshJobs = async () => {
      try {
        const value = await api()?.ListJobs();
        if (active && value) setJobs(value);
      } catch {
        // A transient polling error must not hide the last known job state.
      }
    };
    refreshJobs();
    const timer = window.setInterval(refreshJobs, 2000);
    return () => {
      active = false;
      window.clearInterval(timer);
    };
  }, []);
  return (
    <div className="shell">
      <header>
        <div className="brand">
          <span className="mark">D</span>
          <div>
            <b>DICOM DISC PUBLISHER</b>
            <small>Gestión de medios médicos</small>
          </div>
        </div>
        <button className="settings" onClick={openSettings}>
          ⚙ Configuración
        </button>
      </header>
      <main>
        <section className="hero">
          <div>
            <p className="eyebrow">BÚSQUEDA DE ESTUDIOS · SYMPHONY PACS</p>
            <h1>Estudios disponibles</h1>
            <p>
              Seleccione un rango para consultar estudios y preparar su
              publicación.
            </p>
          </div>
          <div className="filters">
            <label>
              Desde
              <input
                type="date"
                value={from}
                max={to}
                onChange={(e) => setFrom(e.target.value)}
              />
            </label>
            <label>
              Hasta
              <input
                type="date"
                value={to}
                min={from}
                onChange={(e) => setTo(e.target.value)}
              />
            </label>
            <button onClick={search} disabled={loading}>
              {loading ? "Buscando…" : "Buscar estudios"}
            </button>
          </div>
        </section>
        <section className="systemStatus">
          <span className={status.studyServer === "Conectado" ? "ok" : "bad"}>
            Servidor: {status.studyServerAddress || "Sin configurar"} · ●{" "}
            {status.studyServer}
          </span>
          <span className={status.tdBridgeConfigured ? "ok" : "bad"}>
            TD Bridge: {status.tdBridge}
          </span>
          {!status.studyApiConfigured && (
            <strong>
              Servidor de estudios no configurado. Revise config.json.
            </strong>
          )}
        </section>
        <section className="panel">
          <div className="panelTitle">
            <h2>Estudios</h2>
            <span>{studies.length} resultados</span>
          </div>
          <div className="table">
            <div className="tr th">
              <span>Paciente</span>
              <span>Fecha</span>
              <span>Modalidad</span>
              <span>Estudio</span>
              <span>Imágenes</span>
              <span></span>
            </div>
            {studies.map((s) => (
              <div
                className="tr"
                key={`${s.studyInstanceUID}-${s.patientName}-${s.studyDate}-${s.studyDescription}`}
              >
                <span className="patient">{s.patientName}</span>
                <span>
                  {new Date(s.studyDate + "T00:00").toLocaleDateString("es")}
                </span>
                <span>
                  <i>{s.modality}</i>
                </span>
                <span>{s.studyDescription}</span>
                <span>{s.instanceCount}</span>
                <span>
                  <button
                    className="action"
                    onClick={() => publish(s)}
                    disabled={!s.studyInstanceUID}
                    title={
                      !s.studyInstanceUID
                        ? "El servidor no proporcionó EST_UID"
                        : undefined
                    }
                  >
                    Grabar CD
                  </button>
                </span>
              </div>
            ))}
          </div>
        </section>
        <section className="panel jobs">
          <div className="panelTitle">
            <h2>Trabajos</h2>
            <span className="statusline">● {message}</span>
          </div>
          {jobs.length === 0 ? (
            <div className="empty">
              Los trabajos preparados aparecerán aquí.
            </div>
          ) : (
            <div className="jobList">
              <div className="job jobHeader">
                <span>Paciente</span>
                <span>Estudio</span>
                <span>Job</span>
                <span>Estado</span>
              </div>
              {jobs.map((j) => (
                <div className="job" key={j.id}>
                  <span>{j.patientName}</span>
                  <span>{j.studyDescription}</span>
                  <code>{j.id}</code>
                  <span>
                    <b className={`jobBadge status-${j.status}`}>
                      {label(j.status)}
                    </b>
                    {j.status === "Failed" && j.errorMessage && (
                      <>
                        <small className="friendlyError">
                          {j.errorMessage}
                        </small>
                        <details>
                          <summary>Ver detalle</summary>
                          <small>
                            {j.errorCode && (
                              <>
                                Código: {j.errorCode}
                                <br />
                              </>
                            )}
                            {j.technicalStatus && (
                              <>
                                Estado TD Bridge: {j.technicalStatus}
                                <br />
                              </>
                            )}
                            {j.detailStatus && (
                              <>
                                Detalle: {j.detailStatus}
                                <br />
                              </>
                            )}
                            Mensaje: {j.errorMessage}
                          </small>
                        </details>
                      </>
                    )}
                  </span>
                </div>
              ))}
            </div>
          )}
        </section>
      </main>
      {settingsOpen && (
        <div
          className="modalBackdrop"
          role="presentation"
          onMouseDown={(e) => {
            if (e.target === e.currentTarget) setSettingsOpen(false);
          }}
        >
          <section
            className="settingsModal"
            role="dialog"
            aria-modal="true"
            aria-labelledby="settings-title"
          >
            <h2 id="settings-title">CONFIGURACIÓN DEL SERVIDOR</h2>
            <label>
              Servidor / IP
              <input
                autoFocus
                value={serverConfig.host}
                onChange={(e) =>
                  setServerConfig((v) => ({ ...v, host: e.target.value }))
                }
              />
            </label>
            <label>
              Puerto
              <input
                type="number"
                min="1"
                max="65535"
                value={serverConfig.port}
                onChange={(e) =>
                  setServerConfig((v) => ({
                    ...v,
                    port: Number(e.target.value),
                  }))
                }
              />
            </label>
            <label>
              Protocolo
              <select
                value={serverConfig.protocol}
                onChange={(e) =>
                  setServerConfig((v) => ({
                    ...v,
                    protocol: e.target.value as "http" | "https",
                  }))
                }
              >
                <option value="http">http</option>
                <option value="https">https</option>
              </select>
            </label>
            <label>
              Timeout
              <div className="inputSuffix">
                <input
                  type="number"
                  min="1"
                  value={serverConfig.timeoutSeconds}
                  onChange={(e) =>
                    setServerConfig((v) => ({
                      ...v,
                      timeoutSeconds: Number(e.target.value),
                    }))
                  }
                />
                <span>segundos</span>
              </div>
            </label>
            <button
              className="testButton"
              disabled={settingsBusy}
              onClick={testConnection}
            >
              {settingsBusy ? "Probando…" : "Probar conexión"}
            </button>
            <div
              className={`connectionState ${connection.status === "Conectado" ? "connected" : connection.status === "Error" ? "failed" : ""}`}
            >
              Estado: ● {connection.status}
              {connection.message && <small>{connection.message}</small>}
            </div>
            <div className="modalActions">
              <button
                className="cancelButton"
                disabled={settingsBusy}
                onClick={() => setSettingsOpen(false)}
              >
                Cancelar
              </button>
              <button
                className="saveButton"
                disabled={settingsBusy}
                onClick={saveSettings}
              >
                Guardar
              </button>
            </div>
          </section>
        </div>
      )}
    </div>
  );
}
function friendlyError(error: unknown) {
  const text = String(error);
  for (const message of [
    "Ingrese un servidor válido.",
    "Puerto inválido.",
    "Protocolo inválido.",
    "El timeout debe ser mayor a 0.",
    "No se pudo guardar la configuración.",
  ])
    if (text.includes(message)) return message;
  return "No se pudo guardar la configuración.";
}
function label(s: string) {
  return (
    (
      {
        Preparing: "Preparando",
        Ready: "Listo",
        QueuedForEpson: "Enviado a Epson",
        Processing: "Procesando",
        Publishing: "Enviado a Epson",
        Completed: "Completado",
        Failed: "Error",
      } as Record<string, string>
    )[s] ?? s
  );
}
createRoot(document.getElementById("root")!).render(<App />);
