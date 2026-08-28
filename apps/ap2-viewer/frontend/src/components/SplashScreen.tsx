import { useEffect, useState } from "react";
import symphonyLogo from "../assets/logo.png";
import "./SplashScreen.css";



const SPLASH_DURATION_MS = 10_000;

function progressMessage(elapsed: number) {
  if (elapsed >= 4_000) return "Visor listo";
  if (elapsed >= 3_000) return "Cargando estudio";
  if (elapsed >= 2_000) return "Inicializando visor DICOM";
  if (elapsed >= 1_000) return "Preparando imágenes";
  return "Leyendo estudio";
}

export function SplashScreen({ exiting = false }: { exiting?: boolean }) {
  const [elapsed, setElapsed] = useState(0);
  useEffect(() => {
    const startedAt = performance.now();
    const timer = window.setInterval(() => setElapsed(Math.min(performance.now() - startedAt, SPLASH_DURATION_MS)), 100);
    return () => window.clearInterval(timer);
  }, []);
  const progress = exiting ? 100 : Math.min(100, Math.max(3, elapsed / SPLASH_DURATION_MS * 100));
  const detail = exiting ? "Visor listo" : progressMessage(elapsed);
  return (
    <main className={`startupScreen${exiting ? " startupScreenExit" : ""}`} role="status" aria-live="polite">
      <div className="startupAmbient" aria-hidden="true" />
      <div className="startupParticles" aria-hidden="true">{Array.from({ length: 11 }, (_, index) => <i key={index} />)}</div>
      <div className="startupWaves" aria-hidden="true">
        <svg viewBox="0 0 1600 240" preserveAspectRatio="none">
          <path className="wave waveOne" d="M-80 125 C170 20 330 225 590 120 S1010 20 1240 128 S1540 205 1720 92" />
          <path className="wave waveTwo" d="M-100 160 C120 70 370 215 620 145 S1010 65 1280 145 S1510 205 1710 122" />
          <path className="wave waveThree" d="M-60 92 C240 190 390 30 700 105 S1110 190 1390 90 S1590 48 1710 75" />
        </svg>
      </div>
      <section className="startupContent">
        <div className="startupLogoGlow" aria-hidden="true" />
        <img className="startupLogo" src={symphonyLogo} alt="Symphony" />
        <h1>Symphony</h1>
        <p className="startupProduct">DICOM VIEWER</p>
        <div className="startupSpinner" aria-hidden="true" />
        <p className="startupTitle">Iniciando visor...</p>
        <p className="startupDetail">{detail}</p>
        <div className="startupProgress" aria-label={`Progreso ${Math.round(progress)}%`}>
          <div className="startupProgressFill" style={{ width: `${progress}%` }}><i /></div>
        </div>
        <p className="startupWait">Por favor espere</p>
      </section>
    </main>
  );
}

export function StartupError({ detail }: { detail: string }) {
  return <main className="startupScreen" role="alert"><section className="startupContent startupError">
    <img className="startupLogo startupErrorLogo" src={symphonyLogo} alt="Symphony" />
    <div className="startupErrorIcon" aria-hidden="true">!</div>
    <h1>No se pudo abrir el estudio.</h1><p className="startupDetail">{detail}</p>
  </section></main>;
}
