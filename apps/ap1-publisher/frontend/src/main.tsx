import React from "react";
import "./theme";
import { createRoot } from "react-dom/client";
import "./style.css";
import "./jobs.css";
import "./settings.css";
import "./components/LicenciaModal.css";
import { SplashScreen, StartupError } from "./components/SplashScreen";
import { LicenseGuard } from "./components/LicenseGuard";
import { LicenseProvider, useLicense } from "./components/LicenseContext";
import medicareLogo from "./assets/MEDICARESOFTPNG.png";
import LicenciaModal from "./components/LicenciaModal";
import { StudiesView } from "./components/StudiesView";
import { HistoryView } from "./components/HistoryView";
import { SettingsView } from "./components/SettingsView";
import { Sidebar } from "./components/Sidebar";
import { TitleBarDrag } from "./components/TitleBarDrag";
import { ConfirmDialog } from "./components/ConfirmDialog";
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
type EpsonConfig = {
  discType: string;
  format: string;
  printMode: number;
  labelType: number;
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
type DiscLabelConfig = { hospitalName: string; logoPath: string };

const api = () => window.go?.main?.App;

const MINIMUM_SPLASH_MS = 6_000;
const SPLASH_FADE_MS = 300;
const delay = (milliseconds: number) =>
  new Promise<void>((resolve) => window.setTimeout(resolve, milliseconds));

function App({initialStatus, initialJobs,}: {
  initialStatus: SystemStatus;
  initialJobs: DiscJob[];
}) {
  const {isValid} = useLicense();
  const [isModalOpen, setIsModalOpen] = React.useState(false);  // modal de licencia
  const today = new Date().toISOString().slice(0, 10);
  const prior = new Date(Date.now() - 7 * 864e5).toISOString().slice(0, 10);
  const [from, setFrom] = React.useState(prior);
  const [to, setTo] = React.useState(today);
  const [studies, setStudies] = React.useState<Study[]>([]);
  const [studyQuery, setStudyQuery] = React.useState("");
  const [jobs, setJobs] = React.useState<DiscJob[]>(initialJobs);
  const [historyError, setHistoryError] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  type View = "studies" | "history" | "settings";
  const [view, setViewRaw] = React.useState<View>("studies");
  // Configuración avisa si hay cambios sin guardar antes de salir.
  const [settingsDirty, setSettingsDirty] = React.useState(false);
  const [pendingView, setPendingView] = React.useState<View | null>(null);
  const setView = (next: View) => {
    if (view === "settings" && next !== "settings" && settingsDirty) { setPendingView(next); return; }
    setViewRaw(next);
  };
  const [searched, setSearched] = React.useState(false);
  const [searchError, setSearchError] = React.useState("");

  const [submittingStudies, setSubmittingStudies] = React.useState<Set<string>>(
    () => new Set(),
  );
  const [pressedStudies, setPressedStudies] = React.useState<Set<string>>(
    () => new Set(),
  );
  const [message, setMessage] = React.useState("Listo para buscar estudios");
  const [status, setStatus] = React.useState<SystemStatus>(initialStatus);
  async function refreshStatus() {
    const value = await api()?.GetSystemStatus();
    if (value) setStatus(value);
  }
  async function search() {
    setLoading(true);
    setSearchError("");
    try {
      const x = await api()?.SearchStudies(from, to);
      setStudies(x ?? []);
      setSearched(true);
      setMessage(
        (x?.length ?? 0) === 0
          ? "No se encontraron estudios en ese rango."
          : `${x?.length ?? 0} estudios encontrados`,
      );
    } catch (e) {
      setStudies([]);
      setSearched(false);
      setMessage(String(e));
      setSearchError(String(e).replace(/^Error:\s*/, ""));
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
          String(value ?? "")
            .toLocaleLowerCase("es")
            .includes(normalizedStudyQuery),
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
      setSearchError(`No se pudo grabar: ${String(e).replace(/^Error:\s*/, "")}`);
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
        const error = await api()?.GetHistoryError();
        if (active) setHistoryError(error ?? "");
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
    <LicenseGuard>
    <div className="shell appEnter withSide">
      <TitleBarDrag />
      <Sidebar
        view={view}
        onNavigate={setView}
        activeJobs={jobs.filter((j) => j.status !== "Completed" && j.status !== "Failed").length}
        licenseValid={isValid}
        onLicense={() => setIsModalOpen(true)}
      />
      <LicenciaModal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)}></LicenciaModal>
      {pendingView && (
        <ConfirmDialog
          title="Cambios sin guardar"
          message="Hay cambios sin guardar en Configuración. Si sale ahora, se descartan."
          confirm="Salir sin guardar"
          cancel="Seguir editando"
          onCancel={() => setPendingView(null)}
          onConfirm={() => { setSettingsDirty(false); setViewRaw(pendingView); setPendingView(null); }}
        />
      )}
      <main>
        {view === "settings" ? (
          <SettingsView onBack={() => setView("studies")} onSaved={refreshStatus} onDirtyChange={setSettingsDirty} />
        ) : view === "studies" ? (
          <StudiesView
            from={from}
            to={to}
            setFrom={setFrom}
            setTo={setTo}
            studies={searched ? studies : null}
            loading={loading}
            error={searchError}
            jobs={jobs}
            submitting={submittingStudies}
            canRecord={isValid}
            serverOk={status.studyServer === "Conectado"}
            tdBridgeOk={status.tdBridgeConfigured}
            onSearch={search}
            onCancel={cancelSearch}
            onRecord={publish}
            onOpenHistory={() => setView("history")}
          />
        ) : (
          <HistoryView
            jobs={jobs}
            historyError={historyError}
            onBack={() => setView("studies")}
          />
        )}
      </main>
    </div>
    </LicenseGuard>
  );
} // aqui termina App
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
  const [phase, setPhase] = React.useState<
    "loading" | "exiting" | "ready" | "error"
  >("loading");
  const [initialData, setInitialData] = React.useState<{
    status: SystemStatus;
    jobs: DiscJob[];
  } | null>(null);
  const [error, setError] = React.useState("");

  React.useEffect(() => {
    let active = true;
    const initialize = async () => {
      const initApp = async () => {
        const backend = api();
        if (!backend) throw new Error("No hay conexión con el runtime nativo.");
        const [status, jobs] = await Promise.all([
          backend.GetSystemStatus(),
          backend.ListJobs(),
        ]);
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
      if (outcome.failure || !outcome.data) {
        setError(
          outcome.failure || "La inicialización no devolvió datos válidos.",
        );
        setPhase("error");
        return;
      }
      setInitialData(outcome.data);
      setPhase("exiting");
      await delay(SPLASH_FADE_MS);
      if (active) setPhase("ready");
    };
    initialize();
    return () => {
      active = false;
    };
  }, []);

  if (phase === "loading" || phase === "exiting")
    return <SplashScreen exiting={phase === "exiting"} />;
  if (phase === "error") return <StartupError detail={error} />;
  return (
    <App initialStatus={initialData!.status} initialJobs={initialData!.jobs} />
  );
}

function friendlyStartupError(error: unknown) {
  const text = error instanceof Error ? error.message : String(error);
  return text.split("\n")[0] || "Error desconocido durante la inicialización.";
}

createRoot(document.getElementById("root")!).render(<LicenseProvider><Root /></LicenseProvider>);
