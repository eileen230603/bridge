# Integraciones pendientes

## Servidor médico/PACS/RIS

AP1 usa la API REST real de Symphony PACS para buscar estudios y descargar el ZIP completo. La dirección y rutas se configuran en `studyApi`.

ANTES DE INSTALAR EN LA CLÍNICA:

Confirmar `studyApi.baseUrl`, `/getestudios`, `/DescargaEstudio` y el acceso de red desde la estación AP1.

## Epson

TD Bridge observa un Monitoring Folder. La aplicación cooperante deposita allí un Job Description File (JDF); TD Bridge valida el trabajo, prepara la imagen y la impresión mediante Total Disc Maker y lo entrega al Discproducer. La compatibilidad del dispositivo depende de los modelos Epson Discproducer admitidos por la versión instalada de TD Bridge, no de lógica específica en AP1.

La instalación local de TD Bridge 10.0.1.0 contiene `TD Bridge TRG_E_F_22.pdf`. `BuildJDF` implementa el subconjunto oficial para un CD de datos: `JOB_ID`, `COPIES`, `DISC_TYPE=CD`, `FORMAT=JOLIET`, entradas `DATA` explícitas y `LABEL` PNG. No fija `PUBLISHER` ni stackers a un modelo; TD Bridge aplica sus valores de entorno.

El envío usa staging y publicación segura: copia a `<trabajo>.jdf.tmp`, hace `fsync`, cierra el archivo y lo renombra a `<trabajo>.jdf`. AP1 queda en `QueuedForEpson`; el depósito no implica que el disco se haya completado.

Pendiente para una instalación real:

- Confirmar permisos de la cuenta del servicio TD Bridge sobre las rutas de contenido y etiqueta.
- Elegir ajustes locales opcionales (publisher, stackers, área/calidad de impresión) sin fijarlos a un modelo en AP1.
- El monitoreo durante la ejecución correlaciona cada trabajo por `JOB_ID` y observa JDF/RJD/INP/STP/DON/ERR. El historial ya se conserva en SQLite entre reinicios. Queda ampliar la interpretación de STF si se requiere mayor detalle de progreso, dispositivo, discos y tinta. Véase [rutas e historial](ap1-paths-history.md).
- La prueba de grabación en la instalación actual fue confirmada como satisfactoria por el usuario.

## Viewer

AP2 ya integra Cornerstone y su cargador DICOM. Se verificaron renderizado, miniaturas, navegación con rueda y cambio de serie con un paquete local. El ejecutable Windows actualizado se integra en AP1. Véase [verificación del visor](ap2-viewer.md).

Quedan fuera de esta verificación: importar discos DICOMDIR ajenos al formato de AP1, comprobar cada sintaxis de transferencia y DICOM multiframe, pruebas de rendimiento extremo y validación de los builds macOS/Linux.
