import { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
// Importar el visor de la librería y sus estilos
import { DicomViewer, DicomStudy } from "@medicaresoft/dicom-viewer";
import "@medicaresoft/dicom-viewer/style.css";
import { LoadStudy, GetDicomFile } from "../wailsjs/go/main/App";
import "./style123.css";



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
function Root() {
  const [loading, setLoading] = useState(true);
  const [viewerState, setViewerState] = useState<State | null>(null);

  useEffect(() => {
    safeLoadStudy()
      .then((state) => {
        setViewerState(state);
      })
      .catch((err) => {
        // En lugar de guardar null o romper, guardamos un estado vacío funcional
        setViewerState({
          dataFound: false,
          dataPath: "",
          imageCount: 0,
          error: "Sin conexión al runtime nativo Wails",
          manifest: {
            id: "N/A",
            studyInstanceUid: "",
            patientName: "DESCONECTADO (MODO WEB)",
            patientId: "N/A",
            series: [],
          },
        });
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  if (loading) return <div>Cargando...</div>;

  return <App state={viewerState!} />;
}

function App({ state }: { state: State }) {
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

    state.manifest.series.forEach(async (s: any) => {
      const firstFile = s.files?.[0];
      const uid =
        typeof firstFile === "string" ? firstFile : firstFile?.instanceUid;
      if (!uid || thumbnailUrls[uid]) return;

      // Si es un reporte o documento sin imagen (SR, DOC), renderizar icono de documento
      if (s.modality === "SR" || s.modality === "DOC") {
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
          ctx.fillText("DICOM SR", 64, 78);
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
    return <div>Cargando datos del estudio...</div>;
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
    series: (state.manifest?.series || []).map((s: any) => ({
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
    <div style={{ width: "100vw", height: "100vh", backgroundColor: "#000" }}>
      <DicomViewer
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
