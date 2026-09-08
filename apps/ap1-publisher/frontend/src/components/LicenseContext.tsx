import React, { createContext, useContext, useState, useEffect, useCallback } from "react";
import { GetLicenseToken, ValidateLicense } from "../../wailsjs/go/main/App";
import { main } from "../../wailsjs/go/models";

interface LicenseContextType {
  isValid: boolean;
  isChecking: boolean;
  status: main.LicenseResponse | null;
  checkLicense: () => Promise<void>;
}

const LicenseContext = createContext<LicenseContextType>({
  isValid: false,
  isChecking: true,
  status: null,
  checkLicense: async () => {},
});

export const LicenseProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [isValid, setIsValid] = useState<boolean>(false);
  const [isChecking, setIsChecking] = useState<boolean>(true);
  const [status, setStatus] = useState<main.LicenseResponse | null>(null);

  const checkLicense = useCallback(async () => {
    setIsChecking(true);

    try {
      // 1. Obtener el token de Go con timeout de seguridad
      const tokenPromise = GetLicenseToken();
      const timeoutPromise = new Promise<string>((_, reject) =>
        setTimeout(() => reject(new Error("Timeout obteniendo token")), 3000)
      );

      const token = await Promise.race([tokenPromise, timeoutPromise]);

      if (!token || token.trim() === "") {
        setIsValid(false);
        setStatus(null);
      } else {
        // 2. Validar el token existente
        const res = await ValidateLicense(token.trim());
        setIsValid(res?.isValid ?? false);
        setStatus(res ?? null);
      }
    } catch (err) {
      console.error("[LicenseContext] Error al verificar la licencia:", err);
      setIsValid(false);
      setStatus(null);
    } finally {
      // 3. Forzar el fin de la carga de forma transparente
      setIsChecking(false);
    }
  }, []);

  useEffect(() => {
    checkLicense();
  }, [checkLicense]);

  return (
    <LicenseContext.Provider value={{ isValid, isChecking, status, checkLicense }}>
      {children}
    </LicenseContext.Provider>
  );
};

export const useLicense = () => useContext(LicenseContext);