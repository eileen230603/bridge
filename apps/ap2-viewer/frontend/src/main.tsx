import { useRef, useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import cornerstone from "cornerstone-core";
import cornerstoneWADOImageLoader from "cornerstone-wado-image-loader";
import dicomParser from "dicom-parser";
import "./style.css";
import {LoadStudy} from "../wailsjs/go/main/App";
import React from "react";

function Root() {
  const [loading, setLoading] = useState(true);
  const [viewerState, setViewerState] = useState<any>(null);

  useEffect(() => {
    // Al iniciar, invocamos la comprobación del backend
    LoadStudy()
      .then((state) => {
        setViewerState(state);
      })
      .catch((err) => {
        setViewerState({ error: `Error inesperado del sistema: ${err}` });
      })
      .finally(() => {
        setLoading(false);
      });
  }, []);

  // 1. PANTALLA DE CARGA
  if (loading) {
    return (
      <div style={{ display: 'flex', height: '100vh', alignItems: 'center', justifyContent: 'center', backgroundColor: '#0f172a', color: '#fff' }}>
        <p>Cargando visor DICOM...</p>
      </div>
    );
  }

  // 2. BLOQUEO: EJECUCIÓN FUERA DEL CD
  if (viewerState?.error) {
    return (
      <div style={{ display: 'flex', height: '100vh', alignItems: 'center', justifyContent: 'center', backgroundColor: '#020617', color: '#fff', padding: '20px' }}>
        <div style={{ maxWidth: '420px', border: '1px solid rgba(239, 68, 68, 0.3)', backgroundColor: 'rgba(15, 23, 42, 0.9)', padding: '30px', borderRadius: '12px', textAlign: 'center' }}>
          <div style={{ fontSize: '40px', marginBottom: '10px' }}>⚠️</div>
          <h2 style={{ color: '#ef4444', marginBottom: '10px', fontSize: '20px' }}>Acceso Denegado</h2>
          <p style={{ fontSize: '14px', color: '#cbd5e1', marginBottom: '20px', lineHeight: '1.5' }}>
            {viewerState.error}
          </p>
          <div style={{ fontSize: '12px', color: '#94a3b8', backgroundColor: '#1e293b', padding: '12px', borderRadius: '8px' }}>
            Este software no puede ejecutarse desde el disco local o pendrives USB. Inserte el CD/DVD original enviado por el centro médico para continuar.
          </div>
        </div>
      </div>
    );
  }

  // 3. ERROR: NO SE ENCONTRÓ LA CARPETA DATA
  if (!viewerState?.dataFound) {
    return (
      <div style={{ display: 'flex', height: '100vh', alignItems: 'center', justifyContent: 'center', backgroundColor: '#0f172a', color: '#f59e0b' }}>
        <p>No se encontró la carpeta <code>data/</code> contigua al ejecutable.</p>
      </div>
    );
  }

  // 4. ÉXITO: RENDERIZAR TU VISOR DICOM
 // 4. ÉXITO: RENDERIZAR TU VISOR DICOM
  return <App />;
}



// Configurar dependencias
cornerstoneWADOImageLoader.external.cornerstone = cornerstone;
cornerstoneWADOImageLoader.external.dicomParser = dicomParser;

cornerstoneWADOImageLoader.configure({
  beforeSend: function () {},
  webWorkerManager: {
    maxWebWorkers: 0,
    autoRegister: false,
  },
});

// Cache para almacenar las imágenes ya decodificadas y evitar lecturas repetidas
const imageCache = new Map<string, any>();
// Función para cargar los bytes de Go y renderizar la imagen en el canvas
async function displayImage(
  filename: string,
  element: HTMLDivElement,
  setError: (msg: string | null) => void
) {
  try {
    setError(null);

    try {
      cornerstone.getEnabledElement(element);
    } catch (e) {
      cornerstone.enable(element);
    }

    let imageId = imageCache.get(filename);

    if (!imageId) {
      // @ts-ignore
      const base64Data = await window.go.main.App.GetDicomFile(filename);
      if (!base64Data) {
        throw new Error(`El archivo ${filename} no existe o está vacío.`);
      }

      const binaryString = window.atob(base64Data);
      const len = binaryString.length;
      const bytes = new Uint8Array(len);
      for (let i = 0; i < len; i++) {
        bytes[i] = binaryString.charCodeAt(i);
      }

      const file = new File([bytes.buffer], filename, { type: "application/dicom" });
      imageId = cornerstoneWADOImageLoader.wadouri.fileManager.add(file);
      imageCache.set(filename, imageId);
    }

    const image = await cornerstone.loadImage(imageId);
    cornerstone.displayImage(element, image);
    cornerstone.fitToWindow(element);
    cornerstone.resize(element, true);

  } catch (err: any) {
    console.error("Error al renderizar DICOM:", err);
    const message = err?.error || err?.message || (typeof err === "string" ? err : JSON.stringify(err));
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
const SERIES_MAP: Record<number, string[]> = {
  0: ["image1.dcm"],
  1: ["image1.dcm"],
  2: ["image1.dcm"],
  3: ["image1.dcm"],
};
type SeriesViewportState = {
  zoom: number;
  pan: { x: number; y: number };
  voi: { windowWidth: number; windowCenter: number };
};

function App() {
  const [state, setState] = useState<State>();
  const [selectedSeriesIndex, setSelectedSeriesIndex] = useState<number>(0);
  const [seriesPositions, setSeriesPositions] = useState<Record<number, number>>({});
  const [viewportStates, setViewportStates] = useState<Record<number, SeriesViewportState>>({});

  // Estados de arrastre (Pan con Clic Izquierdo, Window/Level con Clic Derecho)
  const [isPanning, setIsPanning] = useState<boolean>(false);
  const [isWlDragging, setIsWlDragging] = useState<boolean>(false);
  const dragStartRef = useRef<{ x: number; y: number }>({ x: 0, y: 0 });

  const [imageError, setImageError] = useState<string | null>(null);
  const viewportRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // @ts-ignore
    window.go?.main?.App?.LoadStudy().then(setState);
  }, []);

  const currentSeries = state?.manifest?.series?.[selectedSeriesIndex];
  const currentImageIndex = seriesPositions[selectedSeriesIndex] || 0;
  const seriesFiles = SERIES_MAP[selectedSeriesIndex] || ["image1.dcm"];
  const activeFilename = seriesFiles[currentImageIndex % seriesFiles.length];

  // Obtener estado activo guardado o inicializar valores por defecto
  const currentViewportState = viewportStates[selectedSeriesIndex] || {
    zoom: 1.0,
    pan: { x: 0, y: 0 },
    voi: { windowWidth: 400, windowCenter: 40 },
  };

  // Aplicar estado general al Viewport de Cornerstone
  const applyViewportState = (
    zoom: number,
    pan: { x: number; y: number },
    voi?: { windowWidth: number; windowCenter: number }
  ) => {
    const element = viewportRef.current;
    if (!element) return;

    try {
      const viewport = cornerstone.getViewport(element);
      if (viewport) {
        viewport.scale = zoom;
        viewport.translation.x = pan.x;
        viewport.translation.y = pan.y;

        if (voi && voi.windowWidth > 0) {
          viewport.voi.windowWidth = voi.windowWidth;
          viewport.voi.windowCenter = voi.windowCenter;
        }

        cornerstone.setViewport(element, viewport);
      }
    } catch (e) {}
  };

  useEffect(() => {
    const element = viewportRef.current;
    if (!element || !state) return;

    displayImage(activeFilename, element, setImageError).then(() => {
      // Al cargar, leer los valores nativos de Window/Level si no se han sobreescrito
      try {
        const viewport = cornerstone.getViewport(element);
        if (viewport && (!currentViewportState.voi || currentViewportState.voi.windowWidth === 400)) {
          currentViewportState.voi = {
            windowWidth: viewport.voi.windowWidth,
            windowCenter: viewport.voi.windowCenter,
          };
        }
      } catch (e) {}

      applyViewportState(
        currentViewportState.zoom,
        currentViewportState.pan,
        currentViewportState.voi
      );
    });
  }, [activeFilename, state, selectedSeriesIndex]);

  // Controles de Zoom
  const updateZoom = (newScale: number) => {
    const scale = Math.min(Math.max(newScale, 0.2), 5.0);
    setViewportStates((prev) => ({
      ...prev,
      [selectedSeriesIndex]: {
        ...(prev[selectedSeriesIndex] || { pan: { x: 0, y: 0 }, voi: currentViewportState.voi }),
        zoom: scale,
      },
    }));
    applyViewportState(scale, currentViewportState.pan, currentViewportState.voi);
  };

  const handleZoomIn = () => updateZoom(currentViewportState.zoom + 0.2);
  const handleZoomOut = () => updateZoom(currentViewportState.zoom - 0.2);

  // Restablecer Zoom, Pan y Window/Level
  const handleResetAll = () => {
    const element = viewportRef.current;
    if (!element) return;
    try {
      cornerstone.fitToWindow(element);
      
      // Volver a la imagen original para obtener los W/L nativos
      const enabledElement = cornerstone.getEnabledElement(element);
      const defaultWw = enabledElement.image.windowWidth || 400;
      const defaultWc = enabledElement.image.windowCenter || 40;

      const viewport = cornerstone.getViewport(element);
      if (viewport) {
        viewport.translation = { x: 0, y: 0 };
        viewport.voi.windowWidth = defaultWw;
        viewport.voi.windowCenter = defaultWc;
        cornerstone.setViewport(element, viewport);

        setViewportStates((prev) => ({
          ...prev,
          [selectedSeriesIndex]: {
            zoom: viewport.scale,
            pan: { x: 0, y: 0 },
            voi: { windowWidth: defaultWw, windowCenter: defaultWc },
          },
        }));
      }
    } catch (e) {}
  };

  // Presets Rápidos de Window/Level (Punto 9)
  const applyPreset = (ww: number, wc: number) => {
    const newVoi = { windowWidth: ww, windowCenter: wc };
    setViewportStates((prev) => ({
      ...prev,
      [selectedSeriesIndex]: {
        ...currentViewportState,
        voi: newVoi,
      },
    }));
    applyViewportState(currentViewportState.zoom, currentViewportState.pan, newVoi);
  };

  // Manejadores de Eventos del Ratón (Clic Izquierdo = Pan, Clic Derecho = Window/Level)
  const handleMouseDown = (e: React.MouseEvent<HTMLDivElement>) => {
    dragStartRef.current = { x: e.clientX, y: e.clientY };

    if (e.button === 0) {
      // Clic Izquierdo -> Pan
      setIsPanning(true);
    } else if (e.button === 2) {
      // Clic Derecho -> Window/Level
      setIsWlDragging(true);
    }
  };

  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    const element = viewportRef.current;
    if (!element) return;

    // Arrastre para PAN
    if (isPanning) {
      try {
        const viewport = cornerstone.getViewport(element);
        if (viewport) {
          const deltaX = (e.clientX - dragStartRef.current.x) / viewport.scale;
          const deltaY = (e.clientY - dragStartRef.current.y) / viewport.scale;

          const newPan = {
            x: currentViewportState.pan.x + deltaX,
            y: currentViewportState.pan.y + deltaY,
          };

          viewport.translation.x = newPan.x;
          viewport.translation.y = newPan.y;
          cornerstone.setViewport(element, viewport);

          setViewportStates((prev) => ({
            ...prev,
            [selectedSeriesIndex]: {
              ...currentViewportState,
              pan: newPan,
            },
          }));

          dragStartRef.current = { x: e.clientX, y: e.clientY };
        }
      } catch (err) {}
    }

    // Arrastre para WINDOW / LEVEL
    if (isWlDragging) {
      try {
        const viewport = cornerstone.getViewport(element);
        if (viewport) {
          const deltaX = e.clientX - dragStartRef.current.x;
          const deltaY = e.clientY - dragStartRef.current.y;

          const newWw = Math.max(1, (currentViewportState.voi?.windowWidth || 400) + deltaX * 2);
          const newWc = (currentViewportState.voi?.windowCenter || 40) + deltaY * 2;

          const newVoi = { windowWidth: newWw, windowCenter: newWc };

          viewport.voi.windowWidth = newWw;
          viewport.voi.windowCenter = newWc;
          cornerstone.setViewport(element, viewport);

          setViewportStates((prev) => ({
            ...prev,
            [selectedSeriesIndex]: {
              ...currentViewportState,
              voi: newVoi,
            },
          }));

          dragStartRef.current = { x: e.clientX, y: e.clientY };
        }
      } catch (err) {}
    }
  };

  const handleMouseUp = () => {
    setIsPanning(false);
    setIsWlDragging(false);
  };

  // Manejar Scroll (Cortes/Zoom) y Teclado
  useEffect(() => {
    const element = viewportRef.current;
    if (!element) return;

    const maxCount = currentSeries?.instanceCount || seriesFiles.length;

    const updateIndex = (delta: number) => {
      setSeriesPositions((prev) => {
        const currentIndex = prev[selectedSeriesIndex] || 0;
        const newIndex = (currentIndex + delta + maxCount) % maxCount;
        return {
          ...prev,
          [selectedSeriesIndex]: newIndex,
        };
      });
    };

    const handleWheel = (e: WheelEvent) => {
      e.preventDefault();
      if (e.ctrlKey || e.metaKey || e.shiftKey) {
        if (e.deltaY < 0) {
          updateZoom(currentViewportState.zoom + 0.1);
        } else {
          updateZoom(currentViewportState.zoom - 0.1);
        }
      } else {
        if (e.deltaY > 0) updateIndex(1);
        else if (e.deltaY < 0) updateIndex(-1);
      }
    };

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "ArrowDown" || e.key === "ArrowRight") updateIndex(1);
      else if (e.key === "ArrowUp" || e.key === "ArrowLeft") updateIndex(-1);
      else if (e.key === "+" || e.key === "=") handleZoomIn();
      else if (e.key === "-") handleZoomOut();
      else if (e.key.toLowerCase() === "r") handleResetAll();
    };

    element.addEventListener("wheel", handleWheel, { passive: false });
    window.addEventListener("keydown", handleKeyDown);

    return () => {
      element.removeEventListener("wheel", handleWheel);
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [
    selectedSeriesIndex,
    currentSeries,
    seriesFiles.length,
    currentViewportState.zoom,
  ]);

  const handleSelectSeries = (index: number) => {
    setSelectedSeriesIndex(index);
  };

  if (!state) return <div className="center">Cargando estudio…</div>;
  if (!state.dataFound)
    return (
      <div className="missing">
        <div className="disc">◉</div>
        <h1>No se encontró el estudio DICOM.</h1>
        <small>{state.dataPath}</small>
      </div>
    );

  const m = state.manifest;

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
                className={selectedSeriesIndex === i ? "active" : ""}
                key={s.seriesInstanceUID || i}
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
          <div
            className={`viewport ${isPanning ? "panning" : ""} ${isWlDragging ? "wl-dragging" : ""}`}
            ref={viewportRef}
            onMouseDown={handleMouseDown}
            onMouseMove={handleMouseMove}
            onMouseUp={handleMouseUp}
            onMouseLeave={handleMouseUp}
            onContextMenu={(e) => e.preventDefault()} // Desactivar menú contextual del clic derecho
          >
            <div className="cross">+</div>

            {/* Barra de Herramientas de Zoom y Presets W/L */}
            <div className="zoom-toolbar">
              <button onClick={handleZoomOut} title="Alejar (-)">−</button>
              <span>{Math.round((currentViewportState.zoom || 1) * 100)}%</span>
              <button onClick={handleZoomIn} title="Acercar (+)">+</button>
              
              <div className="divider" />
              
              <span className="wl-label">W/L:</span>
              <button onClick={() => applyPreset(2000, 500)} className="preset-btn" title="Hueso">HUESO</button>
              <button onClick={() => applyPreset(400, 40)} className="preset-btn" title="Tejido Blando">TEJIDO</button>
              <button onClick={() => applyPreset(1500, -600)} className="preset-btn" title="Pulmón">PULMÓN</button>

              <div className="divider" />

              <button onClick={handleResetAll} className="reset-btn" title="Restablecer todo (R)">RESET</button>
            </div>

            {/* Overlay informativo de Window / Level */}
            <div className="overlay-info">
              <span>SERIE: {selectedSeriesIndex + 1}/{m?.series?.length || 1}</span>
              <span>CORTE: {currentImageIndex + 1}/{currentSeries?.instanceCount || 1}</span>
              <span>WW: {Math.round(currentViewportState.voi?.windowWidth || 0)} / WL: {Math.round(currentViewportState.voi?.windowCenter || 0)}</span>
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
            {m?.modality || "DICOM"} · {m?.studyDescription || "Sin descripción"}
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

const container = document.getElementById("root");
if (container) {
  createRoot(container).render(
    <React.StrictMode>
      <Root />
    </React.StrictMode>
  );
}