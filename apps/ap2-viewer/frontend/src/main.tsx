import { useEffect, useState } from "react";
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
<<<<<<< HEAD
const MINIMUM_SPLASH_MS = 1_500;
=======
const MINIMUM_SPLASH_MS = 10_000;
>>>>>>> origin/eileen
const SPLASH_FADE_MS = 500;
const delay = (milliseconds: number) => new Promise<void>((resolve) => window.setTimeout(resolve, milliseconds));

function startupError(state: State): string | null {
  if (state.error) {
    const message = state.error.toLowerCase();
    if (message.includes("directorio") || message.includes("data")) return "No se encontró la carpeta data.";
    if (message.includes("manifiesto") && message.includes("válido")) return "El archivo del estudio no es válido.";
    if (message.includes("manifiesto") || message.includes("study.json")) return "No se encontró study.json.";
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

  const handleGetDicomFile = async (param: any): Promise<Blob> => {
    let actualUid =
      typeof param === "string"
        ? param
        : param?.instanceUid || param?.file || param?.id || "";
    const base64Data = await safeGetDicomFile(actualUid);

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
  };

  // Cargar y convertir los archivos DICOM a formato PNG ligero para las miniaturas
  useEffect(() => {
    if (!state?.manifest?.series) return;

    state.manifest.series
      .filter((s: any) => s.modality?.trim().toUpperCase() !== "SR")
      .forEach(async (s: any) => {
        const firstFile = s.files?.[0];
        const uid =
          typeof firstFile === "string" ? firstFile : firstFile?.instanceUid;
        if (!uid || thumbnailUrls[uid]) return;

        // Los documentos no pixelados que sí sean visibles usan un icono.
        if (s.modality === "DOC") {
          const docCanvas = document.createElement("canvas");
          docCanvas.width = 128;
          docCanvas.height = 128;
          const ctx = docCanvas.getContext("2d");
          if (ctx) {
            ctx.fillStyle = "#1e1e2d";
            ctx.fillRect(0, 0, 128, 128);
            ctx.fillStyle = "#8888a0";
            ctx.font = "bold 14px sans-serif";
            ctx.textAlign = "center";
            ctx.fillText("DOCUMENTO", 64, 60);
            ctx.fillText("DICOM", 64, 78);
            setThumbnailUrls((prev) => ({
              ...prev,
              [uid]: docCanvas.toDataURL("image/png"),
            }));
          }
          return;
        }

        try {
        const blob = await handleGetDicomFile(uid);
        const arrayBuffer = await blob.arrayBuffer();
        const byteArray = new Uint8Array(arrayBuffer);
        const dataView = new DataView(arrayBuffer);

        let rows = 256;
        let cols = 256;
        let pixelOffset = -1;
        let isBigEndian = false;

        for (let i = 0; i < byteArray.length - 8; i++) {
          // Detectar Transfer Syntax para verificar si es Big-Endian (1.2.840.10008.1.2.2)
          if (
            byteArray[i] === 0x02 &&
            byteArray[i + 1] === 0x00 &&
            byteArray[i + 2] === 0x10 &&
            byteArray[i + 3] === 0x00
          ) {
            if (
              byteArray[i + 8] === 0x31 &&
              byteArray[i + 9] === 0x2e &&
              byteArray[i + 10] === 0x32 &&
              byteArray[i + 11] === 0x2e &&
              byteArray[i + 12] === 0x38
            ) {
              // Verifica si contiene la secuencia Big Endian Explicit
              const str = String.fromCharCode(
                ...byteArray.slice(i + 8, i + 25),
              );
              if (str.includes("1.2.840.10008.1.2.2")) isBigEndian = true;
            }
          }
          // Rows (0x0028, 0x0010)
          if (
            byteArray[i] === 0x28 &&
            byteArray[i + 1] === 0x00 &&
            byteArray[i + 2] === 0x10 &&
            byteArray[i + 3] === 0x00
          ) {
            rows = dataView.getUint16(i + 8, !isBigEndian);
          }
          // Cols (0x0028, 0x0011)
          if (
            byteArray[i] === 0x28 &&
            byteArray[i + 1] === 0x00 &&
            byteArray[i + 2] === 0x11 &&
            byteArray[i + 3] === 0x00
          ) {
            cols = dataView.getUint16(i + 8, !isBigEndian);
          }
          // PixelData (0x7FE0, 0x0010)
          if (
            byteArray[i] === 0xe0 &&
            byteArray[i + 1] === 0x7f &&
            byteArray[i + 2] === 0x10 &&
            byteArray[i + 3] === 0x00
          ) {
            pixelOffset = i + 12;
            break;
          }
        }

        if (pixelOffset !== -1 && rows > 0 && cols > 0) {
          const totalPixels = rows * cols;
          const rawValues = new Uint16Array(totalPixels);
          let minVal = 65535;
          let maxVal = 0;

          for (let i = 0; i < totalPixels; i++) {
            const rawIdx = pixelOffset + i * 2;
            if (rawIdx + 1 < byteArray.length) {
              const val = dataView.getUint16(rawIdx, !isBigEndian);
              rawValues[i] = val;
              if (val < minVal) minVal = val;
              if (val > maxVal) maxVal = val;
            }
          }

          const range = maxVal - minVal || 1;
          const origCanvas = document.createElement("canvas");
          origCanvas.width = cols;
          origCanvas.height = rows;
          const origCtx = origCanvas.getContext("2d");

          if (origCtx) {
            const imgData = origCtx.createImageData(cols, rows);
            const data = imgData.data;

            for (let i = 0; i < totalPixels; i++) {
              const normalized = Math.floor(
                ((rawValues[i] - minVal) / range) * 255,
              );
              const outIdx = i * 4;
              data[outIdx] = normalized;
              data[outIdx + 1] = normalized;
              data[outIdx + 2] = normalized;
              data[outIdx + 3] = 255;
            }

            origCtx.putImageData(imgData, 0, 0);

            const thumbCanvas = document.createElement("canvas");
            thumbCanvas.width = 128;
            thumbCanvas.height = 128;
            const thumbCtx = thumbCanvas.getContext("2d");

            if (thumbCtx) {
              thumbCtx.drawImage(origCanvas, 0, 0, 128, 128);
              setThumbnailUrls((prev) => ({
                ...prev,
                [uid]: thumbCanvas.toDataURL("image/png"),
              }));
            }
          }
        }
        } catch (e) {
          console.error("Error procesando miniatura:", e);
        }
      });
  }, [state]);

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

  if (!state || !state.manifest) {
    return (
      <div style={{ padding: "24px", color: "#fff", background: "#040d17" }}>
        <h2>No se pudo cargar el estudio</h2>
        <p>{state?.error || "No se encontró study.json ni la carpeta data."}</p>
        {state?.dataPath && <small>Ruta buscada: {state.dataPath}</small>}
      </div>
    );
  }

  // 6. Formateo del objeto DicomStudy
  const formattedStudy: DicomStudy = {
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
  };

  return (
    <div className="viewer-enter" style={{ backgroundColor: "#000" }}>
      <Viewer
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
