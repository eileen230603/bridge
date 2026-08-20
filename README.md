# DicomDiscSuite

Base ejecutable de dos aplicaciones de escritorio Wails + Go + React + TypeScript para preparar y visualizar discos DICOM. El prototipo .NET preexistente permanece en `src/` y en `DicomDiscPublisher.sln`; no fue reemplazado ni integrado con el código nuevo.

## Estructura

```text
apps/
  ap1-publisher/       AP1: búsqueda, paquete y publicación simulada
    frontend/          React + TypeScript + Vite
    internal/          adapters, configuración y servicios
  ap2-viewer/          AP2: lector portátil independiente
    frontend/          React + TypeScript + Vite
    internal/          validación del entorno
shared/                modelos, contrato DICOM y filesystem seguro
runtime/               temp, epson-hotfolder, completed y logs
docs/                  arquitectura, flujo e integraciones pendientes
src/                   prototipo .NET conservado
```

## AP1 — DICOM Disc Publisher

```text
Buscar estudio → obtener DICOM → preparar paquete → agregar AP2
→ crear manifest → preparar trabajo Epson → entregar al Hot Folder
```

Desde `apps/ap1-publisher`:

```powershell
npm --prefix frontend install
wails dev
```

La configuración está en `config.json`; todas las rutas son configurables. **Grabar CD** crea `runtime/temp/<StudyInstanceUID>/` con `data/`, `AP2/`, `study.json` y `label/`. El trabajo mock se prepara fuera del Hot Folder y se entrega sólo después de cerrarlo.

## AP2 — Portable DICOM Viewer

```text
Ejecutarse desde CD → encontrar /data → cargar study.json → visualizar DICOM
```

Desde `apps/ap2-viewer`:

```powershell
npm --prefix frontend install
wails dev
```

En producción busca `data/` y `study.json` junto al ejecutable. Para desarrollo puede establecerse `DICOM_VIEWER_CONTENT_DIR` a una carpeta de paquete. El área de imagen es intencionalmente un placeholder.

## Implementado

- Dos aplicaciones Wails independientes y dos interfaces React responsivas.
- Modelos compartidos y estados de trabajo tipados.
- `StudyRepository` con cinco estudios mock y recuperación de tres placeholders marcados como no-DICOM.
- `StudyPackageBuilder` y generación de `study.json`.
- `EpsonPublisher` mock con staging y publicación segura al Hot Folder.
- Configuración JSON, logging estructurado y limpieza configurable en modo dry-run.
- Descubrimiento de contenido y manifiesto en AP2.
- `ExecutionEnvironmentValidator` con implementación permisiva de desarrollo.

## Simulado o pendiente

- PACS/RIS, DICOM Query/Retrieve o DICOMweb.
- Formato Job/SDK/TD Bridge Epson, grabación e impresión física.
- Cornerstone, codecs y renderizado DICOM.
- Etiqueta PNG real, base de datos y autenticación.
- Validación de ejecución desde CD y builds arm64/Linux/macOS.

No se elegirá protocolo médico ni formato Epson hasta disponer de requisitos y documentación reales. Véanse [arquitectura](docs/architecture.md), [flujo](docs/workflow.md) e [integraciones pendientes](docs/pending-integrations.md).

## Toolchain y compilación

Requiere Go, Node.js/npm, WebView2 y Wails v2. Si Wails no está instalado, tras instalar Go use:

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails doctor
```

Builds independientes:

```powershell
cd apps/ap1-publisher
wails build

cd ../ap2-viewer
wails build
```

## Prototipo .NET

El receptor de pruebas C-ECHO/C-STORE sigue disponible sin cambios bajo `src/`. Se compila como antes con `dotnet build DicomDiscPublisher.sln`.
