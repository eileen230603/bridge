# Registro de cambios de diseño — AP1 v2.1 · Modos de trabajo

- **Base:** v2.0.0 (estética glass sobre grafito). En Penpot está guardada como versión `base 2.0.0 · estética glass`.
- **Archivo Penpot:** servidor local `192.168.0.157`, archivo «New File 2».
  - Página `AP1 · Maqueta`: estado actual de la app (v2.0.0). No se modifica.
  - Página `AP1 v2.1 · Modos de trabajo`: pantallas propuestas (24 tableros, oscuro y claro).
  - Página `AP1 v2.1 · Registro y flujo`: fichas y diagrama de flujo.
- **Alcance:** solo diseño. Nada de esto está implementado en el código.
- **Estado de todas las fichas:** Propuesto.

| ID | Pantalla | Cambio | Impacto |
|----|----------|--------|---------|
| C-001 | Configuración | La pestaña «Servidor» pasa a llamarse «Modo» | Estilo |
| C-002 | Configuración › Modo | Selector segmentado Servidor V1 · Servidor V2 · DICOM SCP; el modo activo es el seleccionado | Estilo + lógica |
| C-003 | Modo › Servidor V1 | Mismos campos que hoy | Solo traslado |
| C-004 | Modo › Servidor V2 | Campo API Key | Lógica |
| C-005 | Modo › DICOM SCP | AE Title, puerto, estado de StoreSCP, inicio automático, IPs en escucha | Lógica nueva |
| C-006 | Modo › DICOM SCP | ~~Ruta de la base SQLite~~ **Descartado**: la base es interna del sistema | Sin cambio |
| C-007 | Configuración › Modo | Botón Guardar inferior global | Lógica |
| C-008 | Estudios | Buscar depende del modo; chip de modo y estado «Sin configurar» | Lógica |

## C-001 · Pestaña «Modo»
- **Qué:** la navegación lateral de Configuración muestra «Modo» (ícono de controles) en lugar de «Servidor».
- **Por qué:** el sistema ya no trabaja solo con Symphony V1.
- **Aceptación:** el resto de secciones (Grabado, Etiqueta, Acerca de) no cambia.

## C-002 · Selector de modo
- **Qué:** control segmentado en la parte superior de «Modo de trabajo», con el mismo estilo del selector CD · DVD · Blu-ray.
- **Regla:** solo hay un modo trabajando a la vez y **el modo activo es el de la pestaña seleccionada**. No hay indicador aparte de «modo guardado».
- **Aceptación:** el modo se aplica al pulsar Guardar (C-007).
- **Pantallas:** `Modo · Servidor V1`, `Modo · Servidor V2`, `Modo · DICOM SCP activo`, `Modo · DICOM SCP apagado`.

## C-003 · Servidor V1
- **Campos (sin cambios):** Servidor / IP, Puerto, Protocolo (http/https), Tiempo de espera (s), botón Probar conexión con estado.

## C-004 · Servidor V2
- **Campos:** los de V1 más **API Key**.
- **API Key:** campo enmascarado con botón para mostrar u ocultar; texto de ayuda «Se envía en cada petición para autenticar la comunicación con el servidor».
- **Aceptación:** Probar conexión valida host y clave; el resultado dice «Conectado · clave válida» o el motivo del error.

## C-005 · DICOM SCP
- **Campos:** AE Title, Puerto.
- **Estado de StoreSCP:** un indicador («Servidor activo» en verde o «Servidor apagado» en gris) y, al lado, un botón de acción: **Detener** con ícono stop (cuadrado) cuando está activo e **Iniciar** con ícono play (triángulo) cuando está apagado.
- **Inicio automático:** casilla «Iniciar al abrir la aplicación».
- **Direcciones en escucha:** lista de `IP : puerto` (con interfaz, por ejemplo Ethernet o Local) que se actualiza al iniciar o detener. Apagado: «Sin direcciones en escucha. El servidor está detenido.»
- **Aceptación:** el estado del botón refleja el estado real del proceso; las IPs salen de las interfaces de red del equipo.

## C-006 · Base SQLite (descartado)
- **Decisión:** la base SQLite es interna del sistema y no se configura desde la interfaz. No hay campo de ruta ni botón Elegir en Configuración.
- **Efecto en el diseño:** el modo DICOM SCP solo muestra el receptor (AE Title, puerto, estado, inicio automático) y las direcciones en escucha.

## C-007 · Guardar global
- **Qué:** pie fijo en la parte inferior de «Modo», con el botón **Guardar**.
- **Estados del pie:** «Cambios sin guardar» (punto ámbar) o «Guardado» (punto verde), más la línea «Modo de trabajo: …» con el modo de la pestaña seleccionada.
- **Aceptación:** al guardar, el sistema trabaja bajo el modo de la pestaña seleccionada. Si faltan campos o la conexión no se valida, no se guarda y se muestra el motivo.

## C-008 · Estudios según el modo
- **Pie de Estudios:** chip «Modo: Servidor V1 / Servidor V2 / DICOM SCP» y, en SCP, «SCP activo · puerto» y «Base local».
- **Sin modo guardado:** alerta «No se puede buscar…» y estado vacío con botón «Ir a Configuración».
- **Pantallas:** `Estudios · Servidor V2 (con resultados)`, `Estudios · DICOM SCP (con resultados)`, `Estudios · Sin configurar`.

## Destinos de grabación (v2.2 · dibujado en Penpot)
Detalle en `propuesta-flujo-destinos.md`. Página Penpot `AP1 v2.2 · Destinos` (14 tableros, oscuro y claro); versión guardada `v2.2 · Destinos (propuesta)`.

| Tablero | Fichas |
|---------|--------|
| `Destino` | C-009, C-010, C-011 (robot listo, unidades D: lista y E: sin disco, pie «Guardado») |
| `Destino · Robot no disponible` | C-009, C-010: tarjeta en rojo, Guardar deshabilitado y botón «Guardar sin el robot» |
| `Etiqueta · papel` | C-012: hospital, logo, vista previa disco/papel, tarjeta «Etiqueta en papel» con impresora, tamaño, Automática/Manual e «Imprimir prueba» |
| `Estudios · Grabar disco` | C-013, C-014: diálogo con la unidad D: elegida, barra de capacidad, copias y etiqueta en papel; pie con un chip por recurso |
| `Estudios · Grabar disco · no cabe` | C-013: 5,9 GB en DVD, barra roja, aviso y Grabar deshabilitado |
| `Estudios · Sin destino listo` | C-014: botones Grabar atenuados, tooltip con enlace a Destino, chips «No disponible / Sin disco» |
| `Historial · En curso` | C-015: carriles por recurso (En cola, Esperando disco 2 de 2, Grabando 43 %), columna Destino y filtro «Destino: Todos» |

Notas de dibujo: en la barra lateral de Configuración, «Grabado» pasa a ser «Destino». La sección «Grabaciones simultáneas» (avanzado) y la variante macOS (tarjeta del robot atenuada «Disponible solo en Windows») no están dibujadas.

| ID | Pantalla | Cambio | Impacto |
|----|----------|--------|---------|
| C-009 | Configuración | Nueva sección «Destino» bajo Modo, con tarjetas (no pestañas) y pie Guardar | Estilo + lógica |
| C-010 | Destino › Robot Epson | Interruptor Habilitado, estado, tipo de disco, formato, impresión, copias (absorbe «Grabado») | Traslado + lógica |
| C-011 | Destino › Unidades CD/DVD | Lista de unidades con estado, verificar (activo), expulsar, copias | Lógica nueva |
| C-012 | Configuración › Etiqueta | Común a ambos destinos; bloque «Etiqueta en papel» | Lógica nueva |
| C-013 | Estudios | Diálogo «Grabar disco»: destino, barra de capacidad, copias, etiqueta en papel | Lógica nueva |
| C-014 | Estudios | Pie con estado por recurso; Grabar deshabilitado sin destino listo; estados neutros | Estilo + lógica |
| C-015 | Historial | Columnas Destino y Copias, filtro por destino, sección «En curso» | Lógica |

## Pantallas sin cambio de contenido
Historial, Configuración › Grabado, Etiqueta, Acerca de y la modal de Licencia se copiaron a v2.1. En Configuración solo cambia el nombre de la pestaña lateral (C-001).

## Pendiente
- Logo real en la barra lateral (en Penpot hay un círculo vacío).
- Estados de error del modo SCP (base no accesible, puerto ocupado) como pantallas propias; hoy solo están descritos en el flujo.
- Revisión visual con exportación desde Penpot (falla desde la IP local por la URL pública configurada).
- Flujo de trabajo en Penpot: la API de plugins no copia tableros entre páginas. Las páginas nuevas se arman serializando los tableros base (con capas ocultas incluidas) y reconstruyéndolos; luego se edita encima.
