export type Study = {
  studyInstanceUID: string;
  patientId?: string;
  patientName: string;
  studyDescription: string;
  studyDate: string;
  modality: string;
  seriesCount: number;
  instanceCount: number;
};

export type DiscJob = {
  id: string;
  studyInstanceUID: string;
  patientId?: string;
  patientName: string;
  studyDescription: string;
  status: string;
  createdAt?: string;
  updatedAt?: string;
  epsonState?: string;
  errorCode?: string;
  technicalStatus?: string;
  detailStatus?: string;
  errorMessage?: string;
};

export const isActive = (j: DiscJob) => j.status !== "Completed" && j.status !== "Failed";

// "DOE^JOHN" -> "DOE JOHN"
export const cleanName = (n: string) => (n || "").replace(/\^+/g, " ").trim();

const MODALITIES: Record<string, string> = {
  CT: "Tomografía", MR: "Resonancia", CR: "Rayos X", DX: "Rayos X", DR: "Rayos X",
  US: "Ecografía", MG: "Mamografía", NM: "Medicina nuclear", PT: "PET", XA: "Angiografía",
  RF: "Fluoroscopia", OT: "Otro", ES: "Endoscopia", PX: "Panorámica", RG: "Rayos X",
};
// "CT\\MR" -> ["Tomografía", "Resonancia"]
export const modalityNames = (m: string) =>
  (m || "").split(/[\\/,; ]+/).filter(Boolean).map((c) => MODALITIES[c.toUpperCase()] ?? c);
export const modalityTone = (m: string) => {
  const c = (m || "").split(/[\\/,; ]+/)[0]?.toUpperCase();
  return ["CT"].includes(c) ? "blue" : ["MR"].includes(c) ? "violet" : ["CR", "DX", "DR", "RG"].includes(c) ? "green" : ["US", "MG"].includes(c) ? "amber" : "slate";
};
