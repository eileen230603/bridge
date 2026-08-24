import { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
// Importar el visor de la librería y sus estilos
import { DicomViewer,DicomStudy } from "@medicaresoft/dicom-viewer";
import "@medicaresoft/dicom-viewer/style.css";

import { LoadStudy, GetDicomFile } from "../wailsjs/go/main/App";

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

function Root() {
  const [loading, setLoading] = useState(true);
  const [viewerState, setViewerState] = useState<State | null>(null);

  useEffect(() => {
    LoadStudy()
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
            series: []
          }
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
  const handleGetDicomFile = async (uidOrFilename: string): Promise<Blob> => {
    const base64Data = await GetDicomFile(uidOrFilename);

    if (!base64Data) {
      throw new Error(`El archivo ${uidOrFilename} no existe.`);
    }

    const binaryString = window.atob(base64Data);
    const len = binaryString.length;
    const bytes = new Uint8Array(len);
    for (let i = 0; i < len; i++) {
      bytes[i] = binaryString.charCodeAt(i);
    }

    return new Blob([bytes.buffer], { type: "application/dicom" });
  };

  if (!state || !state.manifest) {
    return <div>Cargando datos del estudio...</div>;
  }

  const m = state.manifest;

  // Formateador ultra-robusto que mapea tanto campos PACS como camelCase
 const formattedStudy: DicomStudy = {
  id: state.manifest.id,
  studyInstanceUid: state.manifest.studyInstanceUid,
  patientName: state.manifest.patientName || "Paciente Sin Nombre",
  patientId: state.manifest.patientId || "N/A",
  series: (state.manifest.series || []).map((s: any) => ({
    id: s.id || s.seriesInstanceUid,
    seriesInstanceUid: s.seriesInstanceUid,
    modality: s.modality || "CT",
    name: s.name || "Serie",
    files: (s.files || []).map((f: any) => ({
      position: f.position,
      instanceUid: f.instanceUid,
    })),
  })),
};

  return (
    <div style={{ width: "100vw", height: "100vh", backgroundColor: "#000" }}>
      <DicomViewer
        study={formattedStudy as DicomStudy}
        getDicomFile={handleGetDicomFile}
        theme={{
          primary:"#097d22",
          primaryForeground:"#fff",
          secondary:"#a910ab",
          secondaryForeground:"#fff"
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
      <Root />
    //</React.StrictMode>
  );
}