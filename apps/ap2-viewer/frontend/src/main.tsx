import { useCallback, useEffect, useMemo, useState } from "react";
import { createRoot } from "react-dom/client";
// Importar el visor de la librería y sus estilos
import type { DicomStudy } from "@medicaresoft/dicom-viewer";
import "@medicaresoft/dicom-viewer/style.css";
import { LoadStudy, GetDicomFile } from "../wailsjs/go/main/App";
import "./style123.css";
import { SplashScreen, StartupError } from "./components/SplashScreen";


// 1. Estructura exacta que genera el backend Go / Eileen (ViewerStudy)
// export interface DicomStudy2 {
//   id: string;
//   studyInstanceUid: string;
//   patientName: string;
//   patientId: string;
//   patientBirthDate?: string;
//   patientAge?: string;
//   patientSex?: string;
//   studyDate?: string;
//   studyTime?: string;
//   studyDescription?: string;
//   series: Array<{
//     id: string;
//     seriesInstanceUid: string;
//     modality: string;
//     name: string;
//     files: Array<{ position: number; instanceUid: string }>;
//   }>;
// }

// 2. Estado retornado por LoadStudy() en Go
type State = {
  dataFound: boolean;
  dataPath: string;
  manifest?: any;
  imageCount: number;
  error?: string;
};
const safeLoadStudy = async (): Promise<State> => {
  if (
    typeof window !== "undefined" &&
    (window as any)["go"]?.["main"]?.["App"]?.["LoadStudy"]
  ) {
    return (await LoadStudy()) as unknown as State;
  }

  // Si estamos en localhost (web puro sin Wails runtime), lanzamos error capturable
  throw new Error("Sin conexión al runtime nativo Wails (Modo Web)");
};

const safeGetDicomFile = async (uidOrFilename: string): Promise<string> => {
  if (
    typeof window !== "undefined" &&
    (window as any)["go"]?.["main"]?.["App"]?.["GetDicomFile"]
  ) {
    return await GetDicomFile(uidOrFilename);
  }
  return "";
};
type StartupState = "loading" | "ready" | "error";
const MINIMUM_SPLASH_MS = 1_500;
const SPLASH_FADE_MS = 500;
const delay = (milliseconds: number) => new Promise<void>((resolve) => window.setTimeout(resolve, milliseconds));

function startupError(state: State): string | null {
  if (state.error) {
    const message = state.error.toLowerCase();
    if (message.includes("directorio")) return "No se encontró la carpeta data.";
    if (message.includes("manifiesto") && message.includes("válido")) return "El archivo del estudio no es válido.";
    if (message.includes("descifrar")) return "El archivo study.dat está dañado o no es compatible.";
    if (message.includes("study.dat")) return "No se pudo leer study.dat junto a la carpeta data.";
    return state.error.split("\n")[0];
  }
  if (!state.dataFound) return "No se encontró la carpeta data.";
  if (!state.manifest || !Array.isArray(state.manifest.series)) return "El archivo del estudio no es válido.";
  return null;
}

function Root() {
  const [startup, setStartup] = useState<StartupState | "exiting">("loading");
  const [error, setError] = useState("");
  const [viewerState, setViewerState] = useState<State | null>(null);
  const [Viewer, setViewer] = useState<any>(null);

  useEffect(() => {
    let active = true;
    const initialize = async () => {
      const initViewer = async () => {
        const state = await safeLoadStudy();
        const detail = startupError(state);
        if (detail) throw new Error(detail);
        const module = await import("@medicaresoft/dicom-viewer");
        return { state, Viewer: module.DicomViewer };
      };
      const [outcome] = await Promise.all([
        initViewer().then(
          (data) => ({ data, failure: "" }),
          (reason) => ({ data: null, failure: reason instanceof Error ? reason.message : "Ocurrió un error durante el inicio." }),
        ),
        delay(MINIMUM_SPLASH_MS),
      ]);
      if (!active) return;
      if (!outcome.data) {
        setError(outcome.failure);
        setStartup("error");
        return;
      }
      try {
        setViewerState(outcome.data.state);
        setViewer(() => outcome.data!.Viewer);
        setStartup("exiting");
        await delay(SPLASH_FADE_MS);
        if (!active) return;
        setStartup("ready");
      } catch (err) {
        if (!active) return;
        setError(err instanceof Error ? err.message : "Ocurrió un error durante el inicio.");
        setStartup("error");
      }
    };
    initialize();
    return () => { active = false; };
  }, []);

  if (startup === "loading" || startup === "exiting") return <SplashScreen exiting={startup === "exiting"} />;
  if (startup === "error") return <StartupError detail={error} />;

  return <App state={viewerState!} Viewer={Viewer} />;
}

function App({ state, Viewer }: { state: State; Viewer: any }) {
  const [thumbnailUrls, setThumbnailUrls] = useState<Record<string, string>>(
    {},
  );

  const [fileError, setFileError] = useState("");
  const handleGetDicomFile = useCallback(async (param: any): Promise<Blob> => {
    let actualUid =
      typeof param === "string"
        ? param
        : param?.instanceUid || param?.file || param?.id || "";
    let base64Data: string;
    try {
      base64Data = await safeGetDicomFile(actualUid);
    } catch {
      setFileError("No se pudo leer una imagen del disco. Comprueba que el disco o la carpeta sigan disponibles y vuelve a abrir el estudio.");
      throw new Error("No se pudo leer la imagen DICOM.");
    }

    if (!base64Data) {
      throw new Error(`El archivo ${actualUid} no existe.`);
    }

    const cleanBase64 = base64Data.replace(/[\r\n\s]/g, "");
    const binaryString = window.atob(cleanBase64);
    const len = binaryString.length;
    const bytes = new Uint8Array(len);
    for (let i = 0; i < len; i++) {
      bytes[i] = binaryString.charCodeAt(i);
    }

    return new Blob([bytes.buffer], { type: "application/dicom" });
  }, []);

  // Use the same DICOM decoder as the viewport, including compressed and signed pixels.
  useEffect(() => {
    let cancelled = false;
    const series = state.manifest?.series || [];
    const buildThumbnails = async () => {
      for (const s of series) {
        if (cancelled) return;
        const uid = s.files?.[0]?.instanceUid;
        if (!uid || ["SR", "DOC"].includes(s.modality?.trim().toUpperCase())) continue;
        let url = "";
        try {
          const blob = await handleGetDicomFile(uid);
          if (cancelled) return;
          url = URL.createObjectURL(blob);
          const cornerstone = (window as any).cornerstone;
          if (!cornerstone) continue;
          const image = await cornerstone.loadImage("wadouri:" + url);
          if (cancelled) return;
          const canvas = document.createElement("canvas");
          canvas.width = 128;
          canvas.height = 128;
          cornerstone.renderToCanvas(canvas, image);
          const thumbnail = canvas.toDataURL("image/png");
          setThumbnailUrls(prev => ({ ...prev, [uid]: thumbnail }));
        } catch {
          // A missing thumbnail must not prevent opening other series.
        } finally {
          if (url) URL.revokeObjectURL(url);
        }
      }
    };
    void buildThumbnails();
    return () => { cancelled = true; };
  }, [state, handleGetDicomFile]);

  const handleGetThumbnailUrl = (param: any): string => {
    const uid =
      typeof param === "string"
        ? param
        : param?.instanceUid || param?.file || param?.id || "";
    return thumbnailUrls[uid] || "";
  };

  // Desbloquear clic derecho (menú contextual)
  useEffect(() => {
    const enableContextMenu = (e: MouseEvent) => {
      e.stopPropagation();
    };

    window.addEventListener("contextmenu", enableContextMenu, true);
    return () =>
      window.removeEventListener("contextmenu", enableContextMenu, true);
  }, []);

  // 6. Formateo del objeto DicomStudy
  const formattedStudy: DicomStudy = useMemo(() => ({
    id: state.manifest?.id || "N/A",
    studyInstanceUid: state.manifest?.studyInstanceUid || "",
    patientName: state.manifest?.patientName || "Paciente Sin Nombre",
    patientId: state.manifest?.patientId || "N/A",
    patientBirthDate: state.manifest?.patientBirthDate,
    patientSex: state.manifest?.patientSex,
    studyDate: state.manifest?.studyDate,
    studyTime: state.manifest?.studyTime,
    studyDescription: state.manifest?.studyDescription,
    series: (state.manifest?.series || [])
      .filter((s: any) => s.modality?.trim().toUpperCase() !== "SR")
      .map((s: any) => ({
        id: s.id || s.seriesInstanceUid,
        seriesInstanceUid: s.seriesInstanceUid,
        modality: s.modality || "MR",
        name: s.name || s.seriesDescription || "Serie",
        files: (s.files || []).map((f: any, idx: number) => ({
          position: f.position ?? idx + 1,
          instanceUid: typeof f === "string" ? f : f.instanceUid,
        })),
      })),
  }), [state.manifest]);

  return (
    <div className="viewer-enter" style={{ backgroundColor: "#000" }}>
      {fileError && <div role="alert" className="viewer-file-error">{fileError}<button onClick={() => window.location.reload()}>Volver a abrir</button></div>}
      <Viewer
        maxConcurrentDownloads={3}
        study={formattedStudy}
        getDicomFile={handleGetDicomFile}
        getThumbnailUrl={handleGetThumbnailUrl}
        theme={{
          primary: "#0a5f8a",
          primaryForeground: "#ffffff",
          secondary: "#e63946",
          secondaryForeground: "#ffffff",
          // superficies opcionales (tema oscuro completo):
          background: "#0e1017",
          chrome: "#060608",
          panel: "#1a1d27",
          panelHover: "#202432",
          border: "#242735",
          // ...
        }}
      />
    </div>
  );
}
// Renderizado principal en el DOM
const container = document.getElementById("root");
if (container) {
  createRoot(container).render(
    //<React.StrictMode>
    <Root />,
    //</React.StrictMode>
  );
}
