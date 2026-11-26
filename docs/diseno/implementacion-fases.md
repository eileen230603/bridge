# Implementación de v2.1 (Modos) y v2.2 (Destinos) por fases

Rama: `feature/modos-y-destinos`. Diseño de referencia: Penpot (`AP1 v2.1 · Modos de trabajo`, `AP1 v2.2 · Destinos`) y `propuesta-flujo-destinos.md`.

## Fase 1 · Configuración (hecha)

**Configuración y migración** (`internal/config/destinations.go`)
- Claves nuevas en `config.json`: `mode` (`active`, `v2`, `scp`), `drives` y `paperLabel`.
- Servidor V1 sigue en `studyApi` y el robot sigue en `epson` (`epson.enabled` = robot habilitado). Una configuración v2.0 carga sin cambios: modo V1, robot como estaba, unidades deshabilitadas, verificar activado y copias 1.
- Un `false` guardado no se pisa con el valor por defecto (se distingue «clave ausente»).

**Detección** (`internal/discdrive`)
- macOS: `drutil list/info/status`. Windows: IMAPI2 por PowerShell (sin ventana). Otros sistemas: no disponible.
- Una unidad está **lista** solo con un disco virgen, no regrabable y de un tipo que la unidad graba.
- Lista de impresoras: `lpstat -e` en macOS y `Win32_Printer` en Windows.
- Modo desarrollo (`make dev-ap1`): simula D: con DVD-R virgen y E: sin disco.

**API para la interfaz** (`app_settings.go`)
- `GetModeConfig` / `SaveModeConfig`: valida solo el modo seleccionado, guarda los tres espacios; la API Key nunca pasa a V1.
- `GetDestinationConfig` / `SaveDestinationConfig` / `GetDestinationStatus`: guarda solo si cada destino habilitado está operativo y hay al menos uno habilitado. Un guardado rechazado no toca el archivo.
- `GetLabelSettings` / `SaveLabelSettings` / `ListPrinters`.
- La app arranca aunque el robot esté deshabilitado. Grabar con el robot deshabilitado da un aviso claro (grabar en unidades llega en la fase 2).

**Interfaz** (`SettingsView.tsx`)
- Menú: Modo · Destino · Etiqueta · Acerca de («Grabado» pasa a la tarjeta del robot).
- Pie fijo en cada sección: «Guardado / Cambios sin guardar / No se guardó», con Descartar y Guardar. Guardar no se activa hasta cargar los datos (corrige el riesgo de pisar la configuración con valores por defecto).
- Aviso al cambiar de sección o salir de Configuración con cambios sin guardar.
- Destino: tarjetas con interruptor, estado real detectado cada 5 s, «Guardar sin el robot» / «Guardar sin las unidades» cuando uno no está disponible.

## Fase 2 · Grabación (pendiente)
- Cola por recurso, candado por unidad y preparación en paralelo con tope.
- Grabación en unidades: Windows (IMAPI2) y macOS (`hdiutil makehybrid` + `drutil burn`), verificar y expulsar, nombre del disco = paciente.
- Diálogo «Grabar disco», pie de Estudios por recurso, estados neutros e Historial con «En curso» y columna Destino.

## Fase 3 · DICOM SCP y etiqueta en papel (pendiente)
- Receptor StoreSCP (iniciar/detener, inicio automático, direcciones en escucha) y búsqueda en la base SQLite. Requiere el esquema de la base.
- Impresión de la etiqueta en papel (automática tras verificar, o manual desde el Historial) e «Imprimir prueba».

## Supuestos por confirmar
- **API Key de V2:** se envía en la cabecera `X-API-Key`. Confirmar con el servidor V2.
- **Modo DICOM SCP:** en la fase 1 se puede guardar, pero Buscar avisa que aún no está disponible.
- **Grabación en unidades:** sin probar con hardware real (este equipo no tiene grabadora). La detección se probó con salidas de ejemplo de `drutil` e IMAPI2.
