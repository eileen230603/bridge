import React, { useState, useEffect } from 'react';
import { GetMachineID, SaveLicenseToken } from '../../wailsjs/go/main/App';
import { useLicense } from './LicenseContext';
import './LicenciaModal.css';

interface LicenciaModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const LicenciaModal: React.FC<LicenciaModalProps> = ({ isOpen, onClose }) => {
  const { checkLicense, isValid } = useLicense();
  const [machineId, setMachineId] = useState<string>('');
  const [tokenInput, setTokenInput] = useState<string>('');
  const [errorMsg, setErrorMsg] = useState<string>('');
  const [isSaving, setIsSaving] = useState<boolean>(false);
  const [copied, setCopied] = useState<boolean>(false);

  useEffect(() => {
    if (isOpen) {
      GetMachineID()
        .then((id) => setMachineId(id))
        .catch((err) => setErrorMsg('No se pudo obtener el Machine ID: ' + err));
    }
  }, [isOpen]);

  useEffect(() => {
    if (!isOpen) return;
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  const handleCopy = () => {
    navigator.clipboard.writeText(machineId);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleSave = async () => {
    if (!tokenInput.trim()) {
      setErrorMsg('Ingrese un token de licencia válido.');
      return;
    }

    setIsSaving(true);
    setErrorMsg('');

    try {
      const res = await SaveLicenseToken(tokenInput.trim());
      if (res.isValid) {
        await checkLicense();
        onClose();
      } else {
        setErrorMsg(res.error || 'La licencia no es válida.');
      }
    } catch (err: any) {
      setErrorMsg(err?.message || 'Error al guardar la licencia.');
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="license-modal-overlay" onMouseDown={(e) => { if (e.target === e.currentTarget) onClose(); }}>
      <div className="license-modal-card" role="dialog" aria-modal="true" aria-labelledby="license-title">
        <div className="license-modal-head">
          <h3 className="license-modal-title" id="license-title">Licencia</h3>
          <span className={`license-status ${isValid ? 'ok' : 'bad'}`}>{isValid ? 'Activa' : 'Inactiva'}</span>
        </div>

        <div className="license-field-group">
          <label className="license-label">Identificador de este equipo (Machine ID):</label>
          <div className="license-input-row">
            <input
              type="text"
              readOnly
              value={machineId}
              className="license-input-text"
            />
            <button type="button" onClick={handleCopy} className="btn-secondary">
              {copied ? '¡Copiado!' : 'Copiar'}
            </button>
          </div>
        </div>

        <div className="license-field-group">
          <label className="license-label">Token de Licencia:</label>
          <textarea
            rows={4}
            value={tokenInput}
            onChange={(e) => setTokenInput(e.target.value)}
            placeholder="Pegue aquí la clave de licencia generada..."
            className="license-textarea"
          />
        </div>

        {errorMsg && <div className="license-error-banner">{errorMsg}</div>}

        <div className="license-actions-row">
          <button type="button" onClick={onClose} className="btn-secondary">
            Cancelar
          </button>
          <button
            type="button"
            onClick={handleSave}
            disabled={isSaving}
            className="btn-primary"
          >
            {isSaving ? 'Guardando...' : 'Activar Licencia'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default LicenciaModal;