# Integraciones pendientes

## Servidor médico/PACS/RIS

AP1 implementa DICOM clásico: C-ECHO, C-FIND Study Root, C-MOVE y Storage SCP/C-STORE. Falta sustituir los placeholders de `config.json` por los datos reales, registrar el AE de destino en el PACS y validar la declaración de conformidad del PACS, los SOP Classes y transfer syntaxes utilizados por la clínica.

ANTES DE INSTALAR EN LA CLÍNICA:

Cambiar `host`, `port`, `calledAETitle`, `callingAETitle`, `moveDestinationAETitle` y `receivePort` por los datos reales entregados por el administrador PACS.

La configuración actual de `apps/ap1-publisher/config.json` es una **CONFIGURACIÓN DE DESARROLLO CON ORTHANC LOCAL. CAMBIAR ESTOS VALORES POR LOS DEL PACS REAL DE LA CLÍNICA.**

## Epson

TD Bridge observa un Monitoring Folder. La aplicación cooperante deposita allí un Job Description File (JDF); TD Bridge valida el trabajo, prepara la imagen y la impresión mediante Total Disc Maker y lo entrega al Discproducer. La compatibilidad del dispositivo depende de los modelos Epson Discproducer admitidos por la versión instalada de TD Bridge, no de lógica específica en AP1.

La instalación local de TD Bridge 10.0.1.0 contiene `TD Bridge TRG_E_F_22.pdf`. `BuildJDF` implementa el subconjunto oficial para un CD de datos: `JOB_ID`, `COPIES`, `DISC_TYPE=CD`, `FORMAT=JOLIET`, entradas `DATA` explícitas y `LABEL` PNG. No fija `PUBLISHER` ni stackers a un modelo; TD Bridge aplica sus valores de entorno.

El envío usa staging y publicación segura: copia a `<trabajo>.jdf.tmp`, hace `fsync`, cierra el archivo y lo renombra a `<trabajo>.jdf`. AP1 queda en `QueuedForEpson`; el depósito no implica que el disco se haya completado.

Pendiente para una instalación real:

- Confirmar permisos de la cuenta del servicio TD Bridge sobre las rutas de contenido y etiqueta.
- Elegir ajustes locales opcionales (publisher, stackers, área/calidad de impresión) sin fijarlos a un modelo en AP1.
- Integrar `TdBridgeJobMonitor` en el ciclo de actualización de la UI. El monitor mínimo ya mapea las extensiones oficiales JDF/RJD/INP/STP/DON/ERR; falta exponer errores detallados, discos y tinta mediante STF.
- Validar físicamente el soporte, capacidad, impresión y resultado final con Total Disc Maker/TD Bridge instalados.

## Viewer

Quedan pendientes Cornerstone, codecs, indexación DICOM/DICOMDIR, validación desde medio óptico, firma/integridad, manejo de estudios grandes y matriz de builds multiplataforma.
