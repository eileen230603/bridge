import React, { useState, useEffect } from 'react';
import { GetMachineID, ValidateLicense } from '../../wailsjs/go/main/App';
import { main } from '../../wailsjs/go/models';

interface LicenciaModalProps {
  isOpen: boolean;
  onClose: () => void;
  onLicenseValidated?: (status: main.LicenseResponse) => void;
}

export default function LicenciaModal({ isOpen, onClose, onLicenseValidated }: LicenciaModalProps) {
  const [machineID, setMachineID] = useState<string>('');
  const [licenseKey, setLicenseKey] = useState<string>('');
  const [status, setStatus] = useState<main.LicenseResponse | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [copied, setCopied] = useState<boolean>(false);

  useEffect(() => {
    if (isOpen) {
      GetMachineID()
        .then((id: string) => setMachineID(id))
        .catch((err: unknown) => console.error("Error al obtener Machine ID:", err));
    }
  }, [isOpen]);

  const handleCopyID = (): void => {
    navigator.clipboard.writeText(machineID);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleValidate = async (e: React.FormEvent<HTMLFormElement>): Promise<void> => {
    e.preventDefault();
    if (!licenseKey.trim()) return;

    setLoading(true);
    setStatus(null);

    try {
      const res: main.LicenseResponse = await ValidateLicense(licenseKey.trim());
      setStatus(res);
      if (res.isValid && onLicenseValidated) {
        onLicenseValidated(res);
      }
    } catch {
      setStatus(
        new main.LicenseResponse({
          isValid: false,
          machineId: machineID,
          isVm: false,
          error: 'Error inesperado al validar la licencia.',
        })
      );
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div style={{
      position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
      backgroundColor: 'rgba(0,0,0,0.6)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000
    }}>
      <div style={{ background: '#fff', padding: '24px', borderRadius: '8px', width: '480px', boxShadow: '0 4px 20px rgba(0,0,0,0.2)' }}>
        <h3 style={{ marginTop: 0, marginBottom: '16px' }}>Gestión de Licencia del Sistema</h3>
        
        <div style={{ marginBottom: '16px' }}>
          <label style={{ display: 'block', fontSize: '12px', fontWeight: 'bold', marginBottom: '6px', color: '#555' }}>
            IDENTIFICADOR DEL EQUIPO (MACHINE ID)
          </label>
          <div style={{ display: 'flex', gap: '8px' }}>
            <input 
              type="text" 
              value={machineID || 'Cargando...'} 
              readOnly 
              style={{ flex: 1, padding: '8px 12px', background: '#f4f4f5', border: '1px solid #ccc', borderRadius: '4px', fontFamily: 'monospace', fontSize: '13px' }}
            />
            <button 
              type="button"
              onClick={handleCopyID} 
              style={{ padding: '8px 12px', background: '#e4e4e7', border: '1px solid #ccc', borderRadius: '4px', cursor: 'pointer', fontWeight: 'bold' }}>
              {copied ? '¡Copiado!' : 'Copiar'}
            </button>
          </div>
        </div>

        <form onSubmit={handleValidate}>
          <div style={{ marginBottom: '16px' }}>
            <label style={{ display: 'block', fontSize: '12px', fontWeight: 'bold', marginBottom: '6px', color: '#555' }}>
              CLAVE DE LICENCIA
            </label>
            <textarea 
              rows={4} 
              value={licenseKey} 
              onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setLicenseKey(e.target.value)}
              placeholder="Ingrese o pegue aquí el token de licencia..."
              style={{ width: '100%', padding: '8px', border: '1px solid #ccc', borderRadius: '4px', resize: 'vertical', fontFamily: 'monospace' }}
              required
            />
          </div>

          {status && (
            <div style={{ 
              padding: '10px 12px', 
              borderRadius: '4px', 
              marginBottom: '16px',
              fontSize: '14px',
              backgroundColor: status.isValid ? '#d1e7dd' : '#f8d7da',
              color: status.isValid ? '#0f5132' : '#842029',
              border: `1px solid ${status.isValid ? '#badbcc' : '#f5c2c7'}`
            }}>
              {status.isValid ? '✓ Licencia válida y activada correctamente.' : `✕ Error: ${status.error || 'Licencia inválida'}`}
            </div>
          )}

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
            <button 
              type="button" 
              onClick={onClose} 
              style={{ padding: '8px 16px', background: '#fff', border: '1px solid #ccc', borderRadius: '4px', cursor: 'pointer' }}>
              Cerrar
            </button>
            <button 
              type="submit" 
              disabled={loading} 
              style={{ padding: '8px 16px', background: '#0d6efd', color: '#fff', border: 'none', borderRadius: '4px', cursor: 'pointer', fontWeight: 'bold' }}>
              {loading ? 'Validando...' : 'Activar Licencia'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}