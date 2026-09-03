import React from "react";
import { createRoot } from "react-dom/client";
import "./style.css";
import "./jobs.css";
import "./settings.css";
import { SplashScreen, StartupError } from "./components/SplashScreen";
<<<<<<< HEAD
=======
import medicareLogo from "./assets/MEDICARESOFTPNG.png";
>>>>>>> origin/eileen
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
  studyInstanceUID: string;
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
<<<<<<< HEAD

const api = () => window.go?.main?.App;

const MINIMUM_SPLASH_MS = 4_000;
const SPLASH_FADE_MS = 200;
=======
type DiscLabelConfig = { hospitalName: string; logoPath: string };

const api = () => window.go?.main?.App;

const MINIMUM_SPLASH_MS = 6_000;
const SPLASH_FADE_MS = 300;
>>>>>>> origin/eileen
const delay = (milliseconds: number) => new Promise<void>((resolve) => window.setTimeout(resolve, milliseconds));

function App({ initialStatus, initialJobs }: { initialStatus: SystemStatus; initialJobs: DiscJob[] }) {
  const today = new Date().toISOString().slice(0, 10);
  const prior = new Date(Date.now() - 7 * 864e5).toISOString().slice(0, 10);
  const [from, setFrom] = React.useState(prior);
  const [to, setTo] = React.useState(today);
  const [studies, setStudies] = React.useState<Study[]>([]);
<<<<<<< HEAD
=======
  const [studyQuery, setStudyQuery] = React.useState("");
>>>>>>> origin/eileen
  const [jobs, setJobs] = React.useState<DiscJob[]>(initialJobs);
  const [loading, setLoading] = React.useState(false);
  const [submittingStudies, setSubmittingStudies] = React.useState<Set<string>>(
    () => new Set(),
  );
  const [pressedStudies, setPressedStudies] = React.useState<Set<string>>(
    () => new Set(),
  );
  const [message, setMessage] = React.useState("Listo para buscar estudios");
  const [status, setStatus] = React.useState<SystemStatus>(initialStatus);
  const [settingsOpen, setSettingsOpen] = React.useState(false);
  const [serverConfig, setServerConfig] = React.useState<ServerConfig>({
    protocol: "http",
    host: "192.168.0.102",
    port: 4000,
    timeoutSeconds: 60,
<<<<<<< HEAD
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
=======
  });
  const [discLabelConfig, setDiscLabelConfig] = React.useState<DiscLabelConfig>({ hospitalName: "", logoPath: "" });
  const [labelPreview, setLabelPreview] = React.useState("");
  const [labelError, setLabelError] = React.useState("");
  const [connection, setConnection] = React.useState<ConnectionTestResult>({
    status: "No probado",
    message: "",
  });
  const [settingsBusy, setSettingsBusy] = React.useState(false);
  async function openSettings() {
    const backend = api();
    const [server, labelConfig] = await Promise.all([
      backend?.GetServerConfig(),
      backend?.GetDiscLabelConfig(),
    ]);
    if (server) setServerConfig(server);
    if (labelConfig) setDiscLabelConfig(labelConfig);
    setConnection({ status: "No probado", message: "" });
    setSettingsOpen(true);
  }
  async function selectDiscLabelLogo() {
    const path = await api()?.SelectDiscLabelLogo();
    if (path) setDiscLabelConfig((current) => ({ ...current, logoPath: path }));
  }
  React.useEffect(() => {
    if (!settingsOpen) return;
    let active = true;
    const timer = window.setTimeout(async () => {
      try {
        const preview = await api()?.GetDiscLabelPreview(discLabelConfig);
        if (active) {
          setLabelPreview(preview ?? "");
          setLabelError("");
        }
      } catch {
        if (active) {
          setLabelPreview("");
          setLabelError("No se pudo cargar el logo seleccionado.");
        }
      }
    }, 250);
    return () => {
      active = false;
      window.clearTimeout(timer);
    };
  }, [settingsOpen, discLabelConfig]);
>>>>>>> origin/eileen
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
<<<<<<< HEAD
=======
      await api()?.SaveDiscLabelConfig(discLabelConfig);
>>>>>>> origin/eileen
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
  async function cancelSearch() {
    await api()?.CancelSearch();
    setMessage("Búsqueda cancelada.");
  }
  const normalizedStudyQuery = studyQuery.trim().toLocaleLowerCase("es");
  const filteredStudies = normalizedStudyQuery
    ? studies.filter((study) =>
        [
          study.patientName,
          study.studyDescription,
          study.modality,
          study.studyDate,
          study.studyInstanceUID,
        ].some((value) =>
          String(value ?? "").toLocaleLowerCase("es").includes(normalizedStudyQuery),
        ),
      )
    : studies;
  async function publish(s: Study) {
    setPressedStudies((current) => {
      const next = new Set(current);
      next.add(s.studyInstanceUID);
      return next;
    });
    setSubmittingStudies((current) => {
      const next = new Set(current);
      next.add(s.studyInstanceUID);
      return next;
    });
    setMessage(`Preparando ${s.studyDescription}…`);
    try {
      const j = await api()?.CreateDiscJob(s.studyInstanceUID);
      if (j) setJobs((v) => [j, ...v]);
      setMessage("Trabajo entregado a TD Bridge");
    } catch (e) {
      setMessage(`Error: ${String(e)}`);
      const x = await api()?.ListJobs();
      setJobs(x ?? []);
    } finally {
      setSubmittingStudies((current) => {
        const next = new Set(current);
        next.delete(s.studyInstanceUID);
        return next;
      });
    }
  }
  React.useEffect(() => {
    let active = true;
    const refreshJobs = async () => {
      try {
        const value = await api()?.ListJobs();
        if (active && value) setJobs(value);
      } catch {
        // A transient polling error must not hide the last known job state.
      }
    };
    const timer = window.setInterval(refreshJobs, 2000);
    return () => {
      active = false;
      window.clearInterval(timer);
    };
  }, []);
  return (
    <div className="shell appEnter">
      <header>
        <div className="brand">
          <img className="brandLogo" src={medicareLogo} alt="Medicare Soft" />
          <div>
            <b>SYMPHONY MEDIA EXPORT</b>
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
            {loading ? (
              <button className="cancelSearch" onClick={cancelSearch}>
                Cancelar búsqueda
              </button>
            ) : (
              <button onClick={search}>Buscar estudios</button>
            )}
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
        <section className="panel studiesPanel">
          <div className="panelTitle">
            <h2>Estudios</h2>
            <div className="studySearch">
              <input
                type="search"
                value={studyQuery}
                onChange={(event) => setStudyQuery(event.target.value)}
                placeholder="Buscar paciente o estudio…"
                aria-label="Buscar dentro de los estudios"
              />
              {studyQuery && (
                <button onClick={() => setStudyQuery("")} aria-label="Limpiar búsqueda">
                  Limpiar
                </button>
              )}
              <span>
                {filteredStudies.length}
                {studyQuery ? ` de ${studies.length}` : ""} resultados
              </span>
            </div>
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
<<<<<<< HEAD
            {studies.map((s) => {
=======
            {filteredStudies.map((s) => {
>>>>>>> origin/eileen
              const studyJobs = jobs.filter(
                (job) => job.studyInstanceUID === s.studyInstanceUID,
              );
              const isSubmitting = submittingStudies.has(s.studyInstanceUID);
              const isRecorded = studyJobs.some(
                (job) => job.status === "Completed",
              );
              const isProcessing = studyJobs.some(
                (job) => job.status !== "Completed" && job.status !== "Failed",
              );
              const hasFailed =
                !isRecorded &&
                !isProcessing &&
                studyJobs.some((job) => job.status === "Failed");
              const wasPressed =
                pressedStudies.has(s.studyInstanceUID) ||
                isRecorded ||
                isProcessing;
              const buttonState = hasFailed
                ? "recordingFailed"
                : wasPressed
                  ? "recordingSelected"
                  : "";

              return (
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
                    className={`action ${buttonState}`}
                    onClick={() => publish(s)}
                    disabled={
                      !s.studyInstanceUID ||
                      isSubmitting ||
                      isProcessing ||
                      isRecorded
                    }
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
              );
            })}
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
<<<<<<< HEAD
            <h2 id="settings-title">CONFIGURACIÓN DEL SERVIDOR</h2>
=======
            <h2 id="settings-title">CONFIGURACIÓN</h2>
            <div className="settingsGrid">
            <section className="settingsSection">
            <h3>Servidor de estudios</h3>
>>>>>>> origin/eileen
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
<<<<<<< HEAD
=======
            </section>
            <section className="settingsSection discLabelSettings">
              <h3>Etiqueta del disco</h3>
              <label>
                Nombre del hospital
                <input
                  value={discLabelConfig.hospitalName}
                  placeholder="Nombre del hospital"
                  onChange={(event) => setDiscLabelConfig((current) => ({ ...current, hospitalName: event.target.value }))}
                />
              </label>
              <label>
                Logo del hospital
                <div className="logoPicker">
                  <input value={discLabelConfig.logoPath} placeholder="Identidad Symphony predeterminada" readOnly />
                  <button className="testButton" onClick={selectDiscLabelLogo}>Seleccionar</button>
                  {discLabelConfig.logoPath && (
                    <button className="cancelButton" onClick={() => setDiscLabelConfig((current) => ({ ...current, logoPath: "" }))}>
                      Quitar
                    </button>
                  )}
                </div>
              </label>
              <div className="labelPreview">
                {labelPreview ? (
                  <img src={labelPreview} alt="Vista previa de la etiqueta del disco" />
                ) : (
                  <span>Generando vista previa…</span>
                )}
              </div>
              {labelError && <p className="labelError">{labelError}</p>}
              <small className="labelHint">
                Sin nombre ni logo personalizado se utilizará la identidad de Symphony.
              </small>
            </section>
            </div>
>>>>>>> origin/eileen
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
function Root() {
  const [phase, setPhase] = React.useState<"loading" | "exiting" | "ready" | "error">("loading");
  const [initialData, setInitialData] = React.useState<{ status: SystemStatus; jobs: DiscJob[] } | null>(null);
  const [error, setError] = React.useState("");

  React.useEffect(() => {
    let active = true;
    const initialize = async () => {
      const initApp = async () => {
        const backend = api();
        if (!backend) throw new Error("No hay conexión con el runtime nativo.");
        const [status, jobs] = await Promise.all([backend.GetSystemStatus(), backend.ListJobs()]);
        return { status, jobs: jobs ?? [] };
      };
      const [outcome] = await Promise.all([
        initApp().then(
          (data) => ({ data, failure: "" }),
          (reason) => ({ data: null, failure: friendlyStartupError(reason) }),
        ),
        delay(MINIMUM_SPLASH_MS),
      ]);
      if (!active) return;
      if (outcome.failure || !outcome.data) { setError(outcome.failure || "La inicialización no devolvió datos válidos."); setPhase("error"); return; }
      setInitialData(outcome.data);
      setPhase("exiting");
      await delay(SPLASH_FADE_MS);
      if (active) setPhase("ready");
    };
    initialize();
    return () => { active = false; };
  }, []);

  if (phase === "loading" || phase === "exiting") return <SplashScreen exiting={phase === "exiting"} />;
  if (phase === "error") return <StartupError detail={error} />;
  return <App initialStatus={initialData!.status} initialJobs={initialData!.jobs} />;
}

function friendlyStartupError(error: unknown) {
  const text = error instanceof Error ? error.message : String(error);
  return text.split("\n")[0] || "Error desconocido durante la inicialización.";
}

createRoot(document.getElementById("root")!).render(<Root />);
<<<<<<< HEAD
=======


>>>>>>> origin/eileen
