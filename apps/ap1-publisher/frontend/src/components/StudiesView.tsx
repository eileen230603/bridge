import React from "react";
import { DiscJob, Study, cleanName, isActive, modalityNames, modalityTone } from "./types";
import "./views.css";

type Props = {
  from: string;
  to: string;
  setFrom: (v: string) => void;
  setTo: (v: string) => void;
  studies: Study[] | null; // null = aún no se ha buscado
  loading: boolean;
  error: string;
  jobs: DiscJob[];
  submitting: Set<string>;
  canRecord: boolean;
  serverOk: boolean;
  tdBridgeOk: boolean;
  onSearch: () => void;
  onCancel: () => void;
  onRecord: (s: Study) => void;
  onOpenHistory: () => void;
};

const Icon = {
  search: (
    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><circle cx="11" cy="11" r="7" /><path d="m20 20-3.5-3.5" /></svg>
  ),
  disc: (
    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="9" /><circle cx="12" cy="12" r="2.5" /></svg>
  ),
  history: (
    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M3 12a9 9 0 1 0 3-6.7L3 8" /><path d="M3 3v5h5" /><path d="M12 7v5l3 2" /></svg>
  ),
  check: (
    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"><path d="m5 12 5 5 9-10" /></svg>
  ),
  burning: (
    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><circle cx="12" cy="12" r="9" /><circle cx="12" cy="12" r="2.5" fill="currentColor" /><path d="M12 3a9 9 0 0 1 9 9" strokeWidth="3" /></svg>
  ),
  queue: (
    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="9" /><path d="M12 7v5l3 2" /></svg>
  ),
  retry: (
    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"><path d="M3 12a9 9 0 0 1 15.5-6.2L21 8" /><path d="M21 3v5h-5" /><path d="M21 12a9 9 0 0 1-15.5 6.2L3 16" /><path d="M3 21v-5h5" /></svg>
  ),
  cross: (
    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round"><path d="M6 6l12 12M18 6 6 18" /></svg>
  ),
};

function formatDate(d: string) {
  const date = new Date(d + "T00:00");
  return isNaN(+date) ? d : date.toLocaleDateString("es", { day: "2-digit", month: "short" });
}

export function StudiesView(p: Props) {
  const [query, setQuery] = React.useState("");
  const q = query.trim().toLocaleLowerCase("es");
  const visible = (p.studies ?? []).filter(
    (s) =>
      !q ||
      cleanName(s.patientName).toLocaleLowerCase("es").includes(q) ||
      (s.patientId ?? "").toLocaleLowerCase("es").includes(q),
  );
  const activeJobs = p.jobs.filter(isActive).length;

  return (
    <div className="v-wrap">
      <div className="v-toolbar">
        <div className="v-search">
          {Icon.search}
          <input
            type="search"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Paciente o ID"
            aria-label="Filtrar por nombre o ID del paciente"
          />
        </div>
        <div className="v-dates">
          <input type="date" value={p.from} max={p.to} onChange={(e) => p.setFrom(e.target.value)} aria-label="Desde" />
          <span aria-hidden>→</span>
          <input type="date" value={p.to} min={p.from} onChange={(e) => p.setTo(e.target.value)} aria-label="Hasta" />
        </div>
        {p.loading ? (
          <button className="v-btn v-btn-ghost" onClick={p.onCancel} title="Cancelar búsqueda">
            <span className="v-spin" /> Cancelar
          </button>
        ) : (
          <button className="v-btn v-btn-primary" onClick={p.onSearch} disabled={!p.canRecord} title={!p.canRecord ? "Licencia requerida" : undefined}>
            {Icon.search} Buscar
          </button>
        )}
      </div>

      {p.error && <div className="v-alert" role="alert">{p.error}</div>}

      <div className="v-panel">
        {p.studies === null || visible.length === 0 ? (
          <div className="v-empty">
            <div className="v-empty-icon">{Icon.search}</div>
            <p>{p.studies === null ? "Elija las fechas y pulse Buscar" : "Sin resultados"}</p>
          </div>
        ) : (
          <div className="v-scroll">
            <table className="v-table">
              <thead>
                <tr>
                  <th>Modalidad</th>
                  <th>Paciente</th>
                  <th>ID</th>
                  <th>Estudio</th>
                  <th>Fecha</th>
                  <th className="c">Grabar</th>
                </tr>
              </thead>
              <tbody>
                {visible.map((s) => {
                  const sj = p.jobs.filter((j) => j.studyInstanceUID === s.studyInstanceUID);
                  const busy = p.submitting.has(s.studyInstanceUID);
                  const done = sj.some((j) => j.status === "Completed");
                  const processing = sj.some(isActive);
                  const failed = !done && !processing && sj.some((j) => j.status === "Failed");
                  const burning = sj.some((j) => j.status === "Processing" || j.status === "Publishing");
                  const state = busy ? "busy" : done ? "done" : burning ? "burn" : processing ? "queued" : failed ? "failed" : "idle";
                  const tip = { busy: "Preparando…", done: "Grabado", queued: "En cola de Epson", burn: "Grabando…", failed: "Falló · reintentar", idle: "Grabar disco" }[state];
                  return (
                    <tr key={s.studyInstanceUID}>
                      <td>
                        {modalityNames(s.modality).map((n) => (
                          <span key={n} className={`v-pill tone-${modalityTone(s.modality)}`}>{n}</span>
                        ))}
                      </td>
                      <td className="strong">{cleanName(s.patientName)}</td>
                      <td className="muted">{s.patientId ?? ""}</td>
                      <td>{s.studyDescription}</td>
                      <td className="muted nowrap">{formatDate(s.studyDate)}</td>
                      <td className="c">
                        <button
                          className={`v-icon v-rec-${state}`}
                          disabled={!p.canRecord || !s.studyInstanceUID || busy || processing || done}
                          onClick={() => p.onRecord(s)}
                          title={tip}
                          aria-label={tip}
                        >
                          {state === "busy" ? <span className="v-spin" /> : state === "done" ? Icon.check : state === "burn" ? Icon.burning : state === "queued" ? Icon.queue : state === "failed" ? Icon.retry : Icon.disc}
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <div className="v-foot">
        <span className={`v-dot ${p.serverOk ? "ok" : "bad"}`} title="Servidor de estudios" /> Servidor
        <span className={`v-dot ${p.tdBridgeOk ? "ok" : "bad"}`} title="TD Bridge" /> Epson
        {p.studies && <span className="v-count">{visible.length} estudios</span>}
      </div>
    </div>
  );
}
