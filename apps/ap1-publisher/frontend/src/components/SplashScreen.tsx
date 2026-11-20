import logo from "../assets/logo.png";
import companyLogo from "../assets/MEDICARESOFTPNG.png";
import "./SplashScreen.css";
import { TitleBarDrag } from "./TitleBarDrag";

export function SplashScreen({ exiting = false }: { exiting?: boolean }) {
  return (
    <main className={`startupScreen${exiting ? " startupScreenExit" : ""}`} role="status" aria-live="polite">
      <TitleBarDrag />
      <section className="startupCard">
        <img className="startupLogo" src={logo} alt="Symphony Media Export" />
        <h1>SYMPHONY</h1>
        <p className="startupProduct">MEDIA EXPORT</p>
        <div className="startupSpinner" aria-hidden="true" />
        <p className="startupTitle">Iniciando sistema…</p>
      </section>
      <p className="startupBrand">
        <img src={companyLogo} alt="" aria-hidden="true" />
        <span>© {new Date().getFullYear()} MedicareSoft Ingeniería SRL</span>
      </p>
    </main>
  );
}

export function StartupError({ detail }: { detail: string }) {
  return (
    <main className="startupScreen" role="alert">
      <TitleBarDrag />
      <section className="startupCard startupError">
        <img className="startupLogo" src={logo} alt="Symphony Media Export" />
        <div className="startupErrorIcon" aria-hidden="true">!</div>
        <h2>No se pudo iniciar la aplicación</h2>
        <p className="startupDetail">{detail}</p>
      </section>
    </main>
  );
}
