# Flujo de publicación

1. AP1 consulta Symphony mediante `GET /getestudios?inicio=YYYYMMDD&final=YYYYMMDD`.
2. El usuario selecciona **Grabar CD**.
3. `HttpStudyRepository` descarga `GET /DescargaEstudio/:uuid`, valida el ZIP y extrae los DICOM en `runtime/temp/<EST_UID>/data`.
4. `StudyPackageBuilder` prepara el visualizador, manifiesto y directorio de etiqueta.
5. `TdBridgePublisher.CreateJob` valida el paquete y crea en staging un JDF para CD de datos conforme al Technical Reference Guide E/F revisión 22.
6. `TdBridgePublisher.SubmitJob` copia el archivo como `.tmp`, lo sincroniza y lo renombra a `.jdf` en el Monitoring Folder.
7. TD Bridge recoge el JDF y solicita la grabación/impresión a un Epson Discproducer compatible.
8. AP1 consulta cada dos segundos los archivos de estado del mismo `JOB_ID` y actualiza la sección Trabajos. Esta lista se mantiene sólo durante la ejecución actual.

```text
AP1
 -> Symphony REST API
 -> StudyPackageBuilder
 -> TdBridgePublisher.CreateJob
 -> JDF
 -> TdBridgePublisher.SubmitJob
 -> Monitoring Folder
 -> TD Bridge
 -> Epson Discproducer
```

AP1 no tiene fallback de simulación ni dependencia de DICOM Query/Retrieve. Los tests HTTP y de transporte TD Bridge usan exclusivamente servidores y carpetas temporales.
