import React, { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";

// 1. Importar el visor de la nueva librería y sus estilos
import { DicomViewer } from "@medicaresoft/dicom-viewer";
import "@medicaresoft/dicom-viewer/style.css";

import "./style.css";
// Funciones expuestas por el backend de Wails
import { LoadStudy, GetDicomFile } from "../wailsjs/go/main/App";

interface DicomStudy {
  id: string;
  studyInstanceUid: string;
  patientName: string;
  patientId: string;
  patientBirthDate?: string;
  patientAge?: string;
  patientSex?: string;
  studyDate?: string;
  studyTime?: string;
  studyDescription?: string;
  series: Array<{
    id: string;
    seriesInstanceUid: string;
    modality: string;
    name: string;
    files: Array<{ position: number; instanceUid: string }>;
  }>;
}

type Series = {
  seriesInstanceUID: string;
  seriesNumber: number;
  description: string;
  modality: string;
  instanceCount: number;
  instanceUids?: string[]; // Si el backend envía las listas de UIDs o nombres de archivo por serie
};

type Manifest = {
  patient: { id: string; name: string };
  studyDescription: string;
  modality: string;
  studyDate: string;
  series: Series[];
};

type State = {
  dataFound: boolean;
  dataPath: string;
  manifest?: Manifest;
  imageCount: number;
  error?: string;
};

// Mapa de respaldo por defecto si la serie no tiene lista explícita de archivos
const SERIES_MAP: Record<number, string[]> = {
  0: ["image1.dcm"],
  1: ["image1.dcm"],
  2: ["image1.dcm"],
  3: ["image1.dcm"],
};

function Root() {
  const [loading, setLoading] = useState(true);
  const [viewerState, setViewerState] = useState<State | null>(null);

  useEffect(() => {
    // Al iniciar, invocamos la comprobación del backend Wails
    LoadStudy()
      .then((state) => {
        setViewerState(state);
      })
      .catch((err) => {
        setViewerState({
          error: `Error inesperado del sistema: ${err}`,
          dataFound: false,
          dataPath: "",
          imageCount: 0,
        });
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  // 1. PANTALLA DE CARGA
  if (loading) {
    return (
      <div
        style={{
          display: "flex",
          height: "100vh",
          alignItems: "center",
          justifyContent: "center",
          backgroundColor: "#0f172a",
          color: "#fff",
        }}
      >
        <p>Cargando visor DICOM...</p>
      </div>
    );
  }

  // 2. BLOQUEO: EJECUCIÓN FUERA DEL CD
  if (viewerState?.error) {
    return (
      <div
        style={{
          display: "flex",
          height: "100vh",
          alignItems: "center",
          justifyContent: "center",
          backgroundColor: "#020617",
          color: "#fff",
          padding: "20px",
        }}
      >
        <div
          style={{
            maxWidth: "420px",
            border: "1px solid rgba(239, 68, 68, 0.3)",
            backgroundColor: "rgba(15, 23, 42, 0.9)",
            padding: "30px",
            borderRadius: "12px",
            textAlign: "center",
          }}
        >
          <div style={{ fontSize: "40px", marginBottom: "10px" }}>⚠️</div>
          <h2 style={{ color: "#ef4444", marginBottom: "10px", fontSize: "20px" }}>
            Acceso Denegado
          </h2>
          <p style={{ fontSize: "14px", color: "#cbd5e1", marginBottom: "20px", lineHeight: "1.5" }}>
            {viewerState.error}
          </p>
          <div
            style={{
              fontSize: "12px",
              color: "#94a3b8",
              backgroundColor: "#1e293b",
              padding: "12px",
              borderRadius: "8px",
            }}
          >
            Este software no puede ejecutarse desde el disco local o pendrives USB. Inserte el CD/DVD original enviado por el centro médico para continuar.
          </div>
        </div>
      </div>
    );
  }

  // 3. ERROR: NO SE ENCONTRÓ LA CARPETA DATA
  if (!viewerState?.dataFound) {
    return (
      <div
        style={{
          display: "flex",
          height: "100vh",
          alignItems: "center",
          justifyContent: "center",
          backgroundColor: "#0f172a",
          color: "#f59e0b",
        }}
      >
        <p>
          No se encontró la carpeta <code>data/</code> contigua al ejecutable.
        </p>
      </div>
    );
  }

  // 4. ÉXITO: RENDERIZAR EL NUEVO VISOR DE MEDICARESOFT
  return <App state={viewerState} />;
}

function App({ state }: { state: State }) {
  // Función puente entre el DicomViewer de la librería y tu backend de Wails
  const handleGetDicomFile = async (uidOrFilename: string): Promise<Blob> => {
    // Invocamos la función de Go expuesta por Wails que devuelve el Base64
    const base64Data = await GetDicomFile(uidOrFilename);

    if (!base64Data) {
      throw new Error(`El archivo ${uidOrFilename} no existe o está vacío.`);
    }

    // Convertimos la cadena Base64 a Uint8Array y luego a Blob
    const binaryString = window.atob(base64Data);
    const len = binaryString.length;
    const bytes = new Uint8Array(len);
    for (let i = 0; i < len; i++) {
      bytes[i] = binaryString.charCodeAt(i);
    }

    return new Blob([bytes.buffer], { type: "application/dicom" });
  };

  // Construcción del objeto de datos en base al Manifest devuelto por Go
  const studyData: DicomStudy = {
    id: state.manifest?.patient?.id || "study-1",
    studyInstanceUid: state.manifest?.patient?.id || "1.2.840.10008.1",
    patientName: state.manifest?.patient?.name || "Desconocido",
    patientId: state.manifest?.patient?.id || "N/A",
    studyDescription: state.manifest?.studyDescription || "Estudio Radiológico",
    studyDate: state.manifest?.studyDate || "",
    series: (state.manifest?.series || []).map((s, seriesIndex) => {
      const filenames = s.instanceUids || SERIES_MAP[seriesIndex] || ["image1.dcm"];

      return {
        id: s.seriesInstanceUID || `series-${seriesIndex}`,
        seriesInstanceUid: s.seriesInstanceUID || `series-uid-${seriesIndex}`,
        modality: s.modality || "CR",
        name: s.description || `Serie ${s.seriesNumber || seriesIndex + 1}`,
        files: filenames.map((filename, fileIndex) => ({
          position: fileIndex + 1,
          instanceUid: filename, // Este ID o nombre de archivo se pasa a handleGetDicomFile
        })),
      };
    }),
  };

  return (
    <div style={{ width: "100vw", height: "100vh", backgroundColor: "#000" }}>
      <DicomViewer study={studyData} getDicomFile={handleGetDicomFile} />
    </div>
  );
}

// Renderizado principal en el DOM
const container = document.getElementById("root");
if (container) {
  createRoot(container).render(
    <React.StrictMode>
      <Root />
    </React.StrictMode>
  );
}