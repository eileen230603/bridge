# DicomDiscSuite

Base ejecutable de dos aplicaciones de escritorio Wails + Go + React + TypeScript para preparar y visualizar discos DICOM. El prototipo .NET preexistente permanece en `src/` y en `DicomDiscPublisher.sln`; no fue reemplazado ni integrado con el código nuevo.

## Estructura

```text
apps/
  ap1-publisher/       AP1: consulta Symphony, paquete y publicación TD Bridge
    frontend/          React + TypeScript + Vite
    internal/          adapters, configuración y servicios
  ap2-viewer/          AP2: lector portátil independiente
    frontend/          React + TypeScript + Vite
    internal/          validación del entorno
shared/                modelos y filesystem seguro
runtime/               paquetes temporales, staging Epson, completados y logs
docs/                  arquitectura, flujo e integraciones pendientes
src/                   prototipo .NET conservado
```

## AP1 — Symphony Media Export

```text
Buscar en Symphony → descargar ZIP DICOM → preparar paquete → agregar AP2
→ crear manifest → preparar trabajo Epson → entregar al Hot Folder
```

Desde `apps/ap1-publisher`:

```powershell
npm --prefix frontend install
wails dev
```

La configuración está en `config.json`; todas las rutas y datos de conexión son configurables. **Grabar CD** descarga el ZIP mediante `/DescargaEstudio/:uuid`, crea `runtime/temp/<EST_UID>/` con `data/`, `AP2/`, `study.json` y `label/`, genera el JDF real en staging y lo entrega atómicamente a TD Bridge.

### Antes de instalar en la clínica

Confirmar `studyApi.baseUrl`, las rutas REST y permisos de lectura del Monitoring Folder configurado para TD Bridge.

## AP2 — Portable DICOM Viewer

```text
Ejecutarse desde CD → encontrar /data → cargar study.json → visualizar DICOM
```

Desde `apps/ap2-viewer`:

```powershell
npm --prefix frontend install
wails dev
```

En producción busca `data/` y `study.dat` junto al ejecutable. Para desarrollo puede establecerse `DICOM_VIEWER_CONTENT_DIR` a una carpeta de paquete. El visor integrado usa Cornerstone y carga los DICOM locales sin un servidor PACS. Permite seleccionar series y navegar imágenes con la rueda del ratón. Lee el manifiesto cifrado `study.dat` y la carpeta `data/`. Véase [verificación del visor](docs/ap2-viewer.md).

## Implementado

- Dos aplicaciones Wails independientes y dos interfaces React responsivas.
- Modelos compartidos y estados de trabajo tipados.
- `HttpStudyRepository` con búsqueda y descarga ZIP desde Symphony PACS.
- `StudyPackageBuilder` y generación de `study.json`.
- `TdBridgePublisher` con staging y publicación segura al Monitoring Folder.
- Configuración JSON, logging estructurado y limpieza configurable en modo dry-run.
- Descubrimiento de contenido y manifiesto en AP2.
- `ExecutionEnvironmentValidator` con implementación permisiva de desarrollo.

## Pendiente de instalación

- Grabación con TD Bridge/Epson: validada por el usuario en la instalación actual.
- Ampliar la matriz de verificación de formatos DICOM y plataformas; el renderizado y la navegación local de AP2 están implementados.
- Etiqueta PNG real, base de datos y autenticación.
- Validación de ejecución desde CD y builds arm64/Linux/macOS.

Véanse [arquitectura](docs/architecture.md), [flujo](docs/workflow.md) e [integraciones pendientes](docs/pending-integrations.md).

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

## Rutas e historial de AP1

AP1 incluye configuración portátil sin datos de activación del desarrollador y conserva el historial en SQLite entre reinicios. Véase [rutas, instalación y respaldo](docs/ap1-paths-history.md).
