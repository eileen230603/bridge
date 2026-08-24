# Arquitectura

DicomDiscSuite añade dos ejecutables Wails independientes al prototipo .NET existente. Ambos comparten únicamente modelos y utilidades pequeñas mediante el módulo Go raíz.

## AP1 — Publisher

La UI invoca métodos del `App` enlazado por Wails. `PacsStudyRepository` ejecuta C-ECHO, C-FIND Study Root y C-MOVE mediante `github.com/amrshadid/go-dicom`; su Storage SCP recibe C-STORE en el puerto configurado. `TdBridgePublisher` crea y entrega el JDF real. `StudyPackageBuilder` crea `data/`, `AP2/`, `label/` y `study.json` bajo una carpeta temporal por Study Instance UID.

La entrega Epson se prepara fuera del Hot Folder. El archivo se escribe y se cierra por completo antes de `SubmitJob`. Se intenta un `rename` atómico; si origen y destino están en volúmenes distintos, se copia a un nombre `.part`, se sincroniza, se cierra y sólo entonces se renombra al nombre visible final.

`TempCleanupService` examina antigüedad de directorios únicamente cuando `cleanupEnabled` está activo y, aun así, sólo registra qué eliminaría. No borra datos.

## AP2 — Viewer

AP2 obtiene el directorio de su ejecutable, valida el entorno mediante `ExecutionEnvironmentValidator`, busca `data/` y carga opcionalmente `study.json`. `DevelopmentEnvironmentValidator` permite toda ejecución. TODO: detectar ejecución desde medio óptico.

## Portabilidad

- Wails depende del WebView nativo: WebView2 en Windows y motores distintos en Linux/macOS.
- La futura integración Cornerstone deberá probar codecs, memoria y aceleración en cada plataforma.
- Epson TD Bridge/SDK probablemente será Windows-only y debe permanecer detrás de `EpsonPublisher`.
- La resolución de rutas usa `path/filepath`; AP2 no presupone letras de unidad.
- Se prevén builds Windows amd64 y arm64, pero aún no se automatizan ni prueban.
