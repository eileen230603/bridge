import React from "react";
import { createRoot } from "react-dom/client";
import "./style.css";
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
  status: string;
  errorMessage?: string;
};
type SystemStatus = {
  studyServer: string;
  studyApiConfigured: boolean;
  pacs: string;
  pacsConfigured: boolean;
  tdBridge: string;
  tdBridgeConfigured: boolean;
};

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
    studyApiConfigured: false,
    pacs: "No probado",
    pacsConfigured: false,
    tdBridge: "Error",
    tdBridgeConfigured: false,
  });
  async function refreshStatus() {
    const value = await api()?.GetSystemStatus();
    if (value) setStatus(value);
  }
  async function testPacs() {
    setLoading(true);
    try {
      await api()?.TestPacsConnection();
      setMessage("Conexión PACS verificada mediante C-ECHO");
    } catch (e) {
      setMessage(String(e));
    } finally {
      await refreshStatus();
      setLoading(false);
    }
  }
  async function search() {
    setLoading(true);
    try {
      const x = await api()?.SearchStudies(from, to);
      setStudies(x ?? []);
      setMessage((x?.length ?? 0) === 0 ? "No se encontraron estudios en ese rango." : `${x?.length ?? 0} estudios encontrados`);
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
        <button className="settings" onClick={testPacs} disabled={loading}>
          Probar conexión PACS
        </button>
      </header>
      <main>
        <section className="hero">
          <div>
            <p className="eyebrow">BÚSQUEDA HTTP · RECUPERACIÓN DICOM</p>
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
            Servidor de estudios: {status.studyServer}
          </span>
          <span className={status.pacs === "Conectado" ? "ok" : "bad"}>
            DICOM: {status.pacs}
          </span>
          <span className={status.tdBridgeConfigured ? "ok" : "bad"}>
            TD Bridge: {status.tdBridge}
          </span>
          {!status.pacsConfigured && (
            <strong>DICOM no configurado. Revise config.json.</strong>
          )}
		  {!status.studyApiConfigured && <strong>Servidor de estudios no configurado. Revise config.json.</strong>}
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
              <div className="tr" key={`${s.studyInstanceUID}-${s.patientName}-${s.studyDate}-${s.studyDescription}`}>
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
                  <button className="action" onClick={() => publish(s)} disabled={!s.studyInstanceUID} title={!s.studyInstanceUID ? "El servidor no proporcionó un StudyInstanceUID DICOM" : undefined}>
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
            jobs.map((j) => (
              <div className="job" key={j.id}>
                <span>{j.patientName}</span>
                <b>{label(j.status)}</b>
                <div className="progress">
                  <i />
                </div>
              </div>
            ))
          )}
        </section>
      </main>
    </div>
  );
}
function label(s: string) {
  return (
    (
      {
        Preparing: "Preparando",
        Ready: "Listo",
        QueuedForEpson: "Enviado a Epson",
        Publishing: "Enviado a Epson",
        Completed: "Completado",
        Failed: "Error",
      } as Record<string, string>
    )[s] ?? s
  );
}
createRoot(document.getElementById("root")!).render(<App />);
