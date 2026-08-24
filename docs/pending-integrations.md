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
- El monitoreo durante la ejecución correlaciona cada trabajo por `JOB_ID` y observa JDF/RJD/INP/STP/DON/ERR. Falta persistir trabajos entre reinicios y, si se requiere mayor detalle, interpretar STF para progreso, códigos de dispositivo, discos y tinta.
- Validar físicamente el soporte, capacidad, impresión y resultado final con Total Disc Maker/TD Bridge instalados.

## Viewer

Quedan pendientes Cornerstone, codecs, indexación DICOM/DICOMDIR, validación desde medio óptico, firma/integridad, manejo de estudios grandes y matriz de builds multiplataforma.
