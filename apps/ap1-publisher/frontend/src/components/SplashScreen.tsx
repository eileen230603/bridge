import symphonyLogo from "../assets/logo.png";
import "./SplashScreen.css";

export function SplashScreen({ exiting = false }: { exiting?: boolean }) {
  return (
    <main className={`startupScreen${exiting ? " startupScreenExit" : ""}`} role="status" aria-live="polite">
      <div className="startupAmbient" aria-hidden="true" />
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
        <p className="startupProduct">DICOM DISC PUBLISHER</p>
        <div className="startupSpinner" aria-hidden="true" />
        <p className="startupTitle">Iniciando sistema...</p>
        <p className="startupDetail">Preparando DICOM Disc Publisher</p>
      </section>
      <p className="startupBrand">Symphony Medical</p>
    </main>
  );
}

export function StartupError({ detail }: { detail: string }) {
  return <main className="startupScreen" role="alert"><section className="startupContent startupError">
    <img className="startupLogo startupErrorLogo" src={symphonyLogo} alt="Symphony" />
    <div className="startupErrorIcon" aria-hidden="true">!</div>
    <h1>No se pudo iniciar la aplicación.</h1><p className="startupDetail">{detail}</p>
  </section></main>;
}