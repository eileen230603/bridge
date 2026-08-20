# Integraciones pendientes

## Servidor médico/PACS/RIS

Antes de implementar se necesita confirmar protocolo (DIMSE C-FIND/C-MOVE/C-GET, DICOMweb, REST u otro), endpoints, AE Titles si aplican, puertos, TLS, credenciales, modelo de filtros, transferencia y comportamiento ante estudios incompletos.

## Epson

TD Bridge observa un Monitoring Folder. La aplicación cooperante deposita allí un Job Description File (JDF); TD Bridge valida el trabajo, prepara la imagen y la impresión mediante Total Disc Maker y lo entrega al Discproducer. La compatibilidad del dispositivo depende de los modelos Epson Discproducer admitidos por la versión instalada de TD Bridge, no de lógica específica en AP1.

El repositorio no contiene el Technical Reference Guide correspondiente a TD Bridge 10.0.1.0 ni un JDF real validado. Por seguridad, `BuildJDF` no inventa campos y devuelve un error explícito. `SubmitJob` sí está listo para recibir un JDF ya terminado y entregarlo de forma atómica.

El envío usa staging y publicación segura: copia a `<trabajo>.jdf.tmp`, hace `fsync`, cierra el archivo y lo renombra a `<trabajo>.jdf`. AP1 queda en `QueuedForEpson`; el depósito no implica que el disco se haya completado.

Pendiente para una instalación real:

- Obtener del capítulo de Job Description File: sintaxis y codificación; extensión y nombre; campos obligatorios; representación de un CD de datos y copias; referencia/rules de rutas del contenido; configuración de filesystem; referencia de etiqueta y formatos admitidos; escaping/quoting; límites; y un ejemplo completo validado.
- Confirmar permisos de la cuenta del servicio TD Bridge sobre las rutas de contenido y etiqueta.
- Elegir ajustes locales opcionales (publisher, stackers, área/calidad de impresión) sin fijarlos a un modelo en AP1.
- Implementar `EpsonJobMonitor` leyendo STF y/o las transiciones JDF → RJD → INP → DON/ERR/STP.
- Mapear estados, errores de grabación, falta de discos/tinta y equipo offline. `NoopEpsonJobMonitor` conserva por ahora el último estado conocido.
- Validar físicamente el soporte, capacidad, impresión y resultado final con Total Disc Maker/TD Bridge instalados.

## Viewer

Quedan pendientes Cornerstone, codecs, indexación DICOM/DICOMDIR, validación desde medio óptico, firma/integridad, manejo de estudios grandes y matriz de builds multiplataforma.
