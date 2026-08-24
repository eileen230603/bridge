# DicomDiscSuite

Base ejecutable de dos aplicaciones de escritorio Wails + Go + React + TypeScript para preparar y visualizar discos DICOM. El prototipo .NET preexistente permanece en `src/` y en `DicomDiscPublisher.sln`; no fue reemplazado ni integrado con el código nuevo.

## Estructura

```text
apps/
  ap1-publisher/       AP1: consulta PACS, paquete y publicación TD Bridge
    frontend/          React + TypeScript + Vite
    internal/          adapters, configuración y servicios
  ap2-viewer/          AP2: lector portátil independiente
    frontend/          React + TypeScript + Vite
    internal/          validación del entorno
shared/                modelos, contrato DICOM y filesystem seguro
runtime/               paquetes temporales, staging Epson, completados y logs
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

La configuración está en `config.json`; todas las rutas y datos de conexión son configurables. **Grabar CD** recupera el estudio por C-MOVE, crea `runtime/temp/<StudyInstanceUID>/` con `data/`, `AP2/`, `study.json` y `label/`, genera el JDF real en staging y lo entrega atómicamente a TD Bridge.

### Antes de instalar en la clínica

Cambiar en `apps/ap1-publisher/config.json` los valores `pacs.host`, `pacs.port`, `pacs.calledAETitle`, `pacs.callingAETitle`, `pacs.moveDestinationAETitle` y `pacs.receivePort` por los datos reales entregados por el administrador PACS. El administrador debe registrar el AE de destino y permitir C-ECHO, C-FIND Study Root, C-MOVE y C-STORE hacia AP1.

La configuración incluida actualmente es una **CONFIGURACIÓN DE DESARROLLO CON ORTHANC LOCAL. CAMBIAR ESTOS VALORES POR LOS DEL PACS REAL DE LA CLÍNICA.**

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
- `PacsStudyRepository` con C-ECHO, C-FIND Study Root, C-MOVE y Storage SCP/C-STORE.
- `StudyPackageBuilder` y generación de `study.json`.
- `TdBridgePublisher` con staging y publicación segura al Monitoring Folder.
- Configuración JSON, logging estructurado y limpieza configurable en modo dry-run.
- Descubrimiento de contenido y manifiesto en AP2.
- `ExecutionEnvironmentValidator` con implementación permisiva de desarrollo.

## Pendiente de instalación

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
# Configuración del servidor de estudios AP1

ESTA ES LA DIRECCIÓN ACTUAL DEL SERVIDOR DE ESTUDIOS.
SI CAMBIA EL SERVIDOR, MODIFICAR `studyApi.baseUrl` EN config.json.
