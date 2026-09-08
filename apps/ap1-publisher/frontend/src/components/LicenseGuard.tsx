import React, { useState, useEffect } from "react";
import { useLicense } from "./LicenseContext";

export const LicenseGuard: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isChecking } = useLicense();
  const [forceShow, setForceShow] = useState(true);

  useEffect(() => {
    // Temporizador de rescate: a los 2s quita el spinner pase lo que pase
    const timer = setTimeout(() => {
      setForceShow(false);
    }, 2000);

    return () => clearTimeout(timer);
  }, []);

  // Si Wails responde rápido usa isChecking, de lo contrario lo fuerza a false con forceShow
  if (isChecking && forceShow) {
    return (
      <div style={{
        display: "flex",
        height: "100vh",
        width: "100vw",
        alignItems: "center",
        justifyContent: "center",
        backgroundColor: "#0f172a",
        color: "#f8fafc",
        fontFamily: "sans-serif"
      }}>
        <div style={{ textAlign: "center" }}>
          <div className="spinner" style={{ margin: "0 auto 12px" }}></div>
          <p style={{ margin: 0, fontSize: "14px", color: "#94a3b8" }}>
            Verificando licencia del sistema...
          </p>
        </div>
      </div>
    );
  }

  return <>{children}</>;
};