import React, { useRef, useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import cornerstone from "cornerstone-core";
import cornerstoneWADOImageLoader from "cornerstone-wado-image-loader";
import dicomParser from "dicom-parser";
import "./style.css";

// Configurar dependencias
cornerstoneWADOImageLoader.external.cornerstone = cornerstone;
cornerstoneWADOImageLoader.external.dicomParser = dicomParser;

// Desactivar Workers para decodificar directo en el hilo principal
try {
  cornerstoneWADOImageLoader.configure({
    beforeSend: function (xhr: any) {},
    useWebWorkers: false,
  });
} catch (e) {
  console.warn("Configuración de WADO Image Loader omite workers");
}
// Función para cargar los bytes de Go y renderizar la imagen en el canvas
async function displayImage(
  filename: string,
  element: HTMLDivElement,
  setError: (msg: string | null) => void,
) {
  try {
    setError(null);

    // 1. Habilitar el visor SOLAMENTE si aún no está activo
    try {
      cornerstone.getEnabledElement(element);
    } catch (e) {
      cornerstone.enable(element);
    }

    // 2. Obtener la cadena Base64 desde Go
    // @ts-ignore
    const base64Data = await window.go.main.App.GetDicomFile(filename);
    if (!base64Data) {
      throw new Error(`El archivo ${filename} no existe o está vacío.`);
    }

    // 3. Decodificar Base64 a Uint8Array
    const binaryString = window.atob(base64Data);
    const len = binaryString.length;
    const bytes = new Uint8Array(len);
    for (let i = 0; i < len; i++) {
      bytes[i] = binaryString.charCodeAt(i);
    }

    // 4. Registrar en Cornerstone
    const file = new File([bytes.buffer], filename, {
      type: "application/dicom",
    });
    const imageId = cornerstoneWADOImageLoader.wadouri.fileManager.add(file);

    // 5. Cargar y renderizar la imagen sin reconstruir el lienzo
    const image = await cornerstone.loadImage(imageId);

    cornerstone.displayImage(element, image);
    cornerstone.fitToWindow(element);
    cornerstone.resize(element, true);
  } catch (err: any) {
    console.error("Error al renderizar DICOM:", err);
    const message =
      err?.error ||
      err?.message ||
      (typeof err === "string" ? err : JSON.stringify(err));
    setError(String(message));
  }
}
type Series = {
  seriesInstanceUID: string;
  seriesNumber: number;
  description: string;
  modality: string;
  instanceCount: number;
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
declare global {
  interface Window {
    go?: { main?: { App?: { LoadStudy(): Promise<State> } } };
  }
}
const AVAILABLE_IMAGES = [
  "image1.dcm",
  "image2.dcm",
  "image3.dcm",
  "image4.dcm",
];

function App() {
  const [state, setState] = React.useState<State>();
  //estado para seleccionar la serie
  const [selectedSeriesIndex, setSelectedSeriesIndex] = useState<number>(0);
  const [currentImageIndex, setCurrentImageIndex] = useState<number>(0);
  const [imageError, setImageError] = useState<string | null>(null);
  const viewportRef = useRef<HTMLDivElement>(null);
  // 1. Cargar estudio al montar la aplicación
  useEffect(() => {
    // @ts-ignore
    window.go?.main?.App?.LoadStudy().then((res) => {
      console.log("Estudio cargado:", res);
      setState(res);
    });
  }, []);
  const currentSeries = state?.manifest?.series?.[selectedSeriesIndex];

  // Determina el archivo dinámico que se debe cargar
  // Cada serie o corte selecciona un archivo distinto de AVAILABLE_IMAGES
  const activeFilename =
    AVAILABLE_IMAGES[
      (selectedSeriesIndex + currentImageIndex) % AVAILABLE_IMAGES.length
    ];
  // Inicializar el elemento Cornerstone cuando el viewport esté listo
  useEffect(() => {
    const element = viewportRef.current;
    if (!element || !state) return;

    console.log("Cargando imagen:", activeFilename);
    displayImage(activeFilename, element, setImageError);
  }, [activeFilename, state]);

  useEffect(() => {
    const element = viewportRef.current;
    if (!element) return;

    const maxCount = currentSeries?.instanceCount || 4;

    const handleWheel = (e: WheelEvent) => {
      e.preventDefault();
      if (e.deltaY > 0) {
        setCurrentImageIndex((prev) => Math.min(prev + 1, maxCount - 1));
      } else if (e.deltaY < 0) {
        setCurrentImageIndex((prev) => Math.max(prev - 1, 0));
      }
    };

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "ArrowDown" || e.key === "ArrowRight") {
        setCurrentImageIndex((prev) => Math.min(prev + 1, maxCount - 1));
      } else if (e.key === "ArrowUp" || e.key === "ArrowLeft") {
        setCurrentImageIndex((prev) => Math.max(prev - 1, 0));
      }
    };

    element.addEventListener("wheel", handleWheel, { passive: false });
    window.addEventListener("keydown", handleKeyDown);

    return () => {
      element.removeEventListener("wheel", handleWheel);
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [currentSeries]);

  // Punto 6: Cambiar entre Series
  const handleSelectSeries = (index: number) => {
    console.log("Serie seleccionada:", index);
    setSelectedSeriesIndex(index);
    setCurrentImageIndex(0); // Reiniciar al primer corte al cambiar de serie
  };

  if (!state) return <div className="center">Cargando estudio…</div>;
  if (!state.dataFound)
    return (
      <div className="missing">
        <div className="disc">◉</div>
        <h1>No se encontró el estudio DICOM.</h1>
        <p>
          La carpeta <b>“data”</b> debe encontrarse junto al visualizador.
        </p>
        <small>{state.dataPath}</small>
      </div>
    );
  const m = state.manifest;
  // const currentSeries = m?.series?.[selectedIndex];
  return (
    <div className="viewer">
      <header>
        <div className="logo">D</div>
        <b>DICOM VIEWER</b>
        <span>VISUALIZADOR PORTÁTIL</span>
      </header>
      <div className="workspace">
        <aside>
          <p>SERIES</p>
          {m?.series?.length ? (
            m.series.map((s, i) => (
              <button
                //Activamos la clase si coincide con el índice seleccionado
                className={selectedSeriesIndex === i ? "active" : ""}
                key={s.seriesInstanceUID || i}
                // Cambiamos la serie al hacer clic
                onClick={() => handleSelectSeries(i)}
              >
                <span>
                  <b>{s.description || `SERIE ${s.seriesNumber}`}</b>
                  <small>{s.modality}</small>
                </span>
                <em>{s.instanceCount}</em>
              </button>
            ))
          ) : (
            <div className="noSeries">Sin detalle de series</div>
          )}
        </aside>
        <main>
          <div className="viewport" ref={viewportRef}>
            <div className="cross">+</div>

            {/* Etiqueta Overlay con estado actual y archivo cargado */}
            <div className="overlay-info">
              <span>
                SERIE: {selectedSeriesIndex + 1}/{m?.series?.length || 1}
              </span>
              <span>
                CORTE: {currentImageIndex + 1}/
                {currentSeries?.instanceCount || 1}
              </span>
              <small style={{ opacity: 0.7 }}>{activeFilename}</small>
            </div>

            {imageError && (
              <div className="placeholder" style={{ color: "#ff6b6b" }}>
                <div>⚠️</div>
                <h2>ERROR AL CARGAR IMAGEN</h2>
                <p>{imageError}</p>
              </div>
            )}
          </div>
        </main>
      </div>
      <footer>
        <div>
          <small>PACIENTE</small>
          <b>{m?.patient.name || "Sin información"}</b>
          <span>ID {m?.patient.id || "—"}</span>
        </div>
        <div>
          <small>ESTUDIO</small>
          <b>
            {m?.modality || "DICOM"} ·{" "}
            {m?.studyDescription || "Sin descripción"}
          </b>
        </div>
        <div>
          <small>FECHA</small>
          <b>
            {m?.studyDate
              ? new Date(m.studyDate + "T00:00").toLocaleDateString("es")
              : "—"}
          </b>
          <span>{state.imageCount} archivos detectados</span>
        </div>
      </footer>
    </div>
  );
}
createRoot(document.getElementById("root")!).render(<App />);
