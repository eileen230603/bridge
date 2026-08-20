# Integraciones pendientes

## Servidor médico/PACS/RIS

Antes de implementar se necesita confirmar protocolo (DIMSE C-FIND/C-MOVE/C-GET, DICOMweb, REST u otro), endpoints, AE Titles si aplican, puertos, TLS, credenciales, modelo de filtros, transferencia y comportamiento ante estudios incompletos.

## Epson

Se requiere documentación oficial de TD Bridge/SDK para la PP-100III, formato real del Job, reglas y ubicación del Hot Folder, estados/confirmaciones, tratamiento de errores, plantillas de impresión y requisitos de instalación/licencia. El archivo `.mock-job` actual no representa ningún formato Epson.

## Viewer

Quedan pendientes Cornerstone, codecs, indexación DICOM/DICOMDIR, validación desde medio óptico, firma/integridad, manejo de estudios grandes y matriz de builds multiplataforma.
