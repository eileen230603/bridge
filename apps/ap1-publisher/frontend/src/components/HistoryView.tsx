import React from "react";
import { DiscJob, cleanName } from "./types";
import "./views.css";

type Props = { jobs: DiscJob[]; historyError: string; onBack: () => void };

const STATUS: Record<string, { text: string; tone: "ok" | "bad" | "wait" }> = {
  Completed: { text: "Completado", tone: "ok" },
  Failed: { text: "Error", tone: "bad" },
  Pending: { text: "En espera", tone: "wait" },
  Downloading: { text: "Descargando", tone: "wait" },
  Preparing: { text: "Preparando", tone: "wait" },
  Ready: { text: "Listo", tone: "wait" },
  QueuedForEpson: { text: "En cola", tone: "wait" },
  Processing: { text: "Procesando", tone: "wait" },
  Publishing: { text: "Grabando", tone: "wait" },
};

function when(j: DiscJob) {
  const d = new Date(j.updatedAt || j.createdAt || "");
  if (isNaN(+d)) return "";
  return d.toLocaleString("es", { day: "2-digit", month: "short", hour: "2-digit", minute: "2-digit" });
}

export function HistoryView({ jobs, historyError, onBack }: Props) {
  const [onlyErrors, setOnlyErrors] = React.useState(false);
  const [open, setOpen] = React.useState<string | null>(null);
  const errors = jobs.filter((j) => j.status === "Failed").length;
  const list = (onlyErrors ? jobs.filter((j) => j.status === "Failed") : jobs)
    .slice()
    .sort((a, b) => +new Date(b.createdAt || 0) - +new Date(a.createdAt || 0));

  return (
    <div className="v-wrap">
      <div className="v-toolbar">
        <h2 className="v-title">Historial</h2>
        <span className="v-spacer" />
        <button className={`v-chip ${onlyErrors ? "on" : ""}`} onClick={() => setOnlyErrors((v) => !v)} disabled={errors === 0 && !onlyErrors}>
          Con error{errors > 0 ? ` · ${errors}` : ""}
        </button>
      </div>

      {historyError && <div className="v-alert" role="alert">{historyError}</div>}

      <div className="v-panel">
        {list.length === 0 ? (
          <div className="v-empty">
            <p>{onlyErrors ? "Sin errores" : "Aún no hay trabajos"}</p>
          </div>
        ) : (
          <div className="v-scroll">
            <table className="v-table">
              <thead>
                <tr>
                  <th>Estado</th>
                  <th>Paciente</th>
                  <th>Estudio</th>
                  <th>Fecha</th>
                  <th className="c">Detalle</th>
                </tr>
              </thead>
              <tbody>
                {list.map((j) => {
                  const st = STATUS[j.status] ?? { text: j.status, tone: "wait" as const };
                  const failed = j.status === "Failed";
                  const isOpen = open === j.id;
                  return (
                    <React.Fragment key={j.id}>
                      <tr className={failed ? "v-failed" : ""}>
                        <td><span className={`v-pill tone-${st.tone}`}>{st.tone === "wait" && <i className="v-pulse" />}{st.text}</span></td>
                        <td className="strong">{cleanName(j.patientName)}</td>
                        <td>{j.studyDescription}</td>
                        <td className="muted nowrap">{when(j)}</td>
                        <td className="c">
                          {failed && (
                            <button className={`v-icon v-rec-failed ${isOpen ? "on" : ""}`} onClick={() => setOpen(isOpen ? null : j.id)} title="Ver detalle del error" aria-label="Ver detalle del error">
                              <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="9" /><path d="M12 8v5M12 16.5v.01" /></svg>
                            </button>
                          )}
                        </td>
                      </tr>
                      {failed && isOpen && (
                        <tr className="v-detail-row">
                          <td colSpan={5}>
                            <div className="v-err">{j.errorMessage || "El trabajo falló."}</div>
                            <dl className="v-detail">
                              {j.errorCode && (<><dt>Código</dt><dd>{j.errorCode}</dd></>)}
                              {j.technicalStatus && (<><dt>Estado TD Bridge</dt><dd>{j.technicalStatus}</dd></>)}
                              {j.detailStatus && (<><dt>Detalle</dt><dd>{j.detailStatus}</dd></>)}
                              <dt>Trabajo</dt><dd>{j.id}</dd>
                            </dl>
                          </td>
                        </tr>
                      )}
                    </React.Fragment>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
