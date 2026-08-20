# Flujo de publicación

1. AP1 consulta estudios a través de `StudyRepository`.
2. El usuario selecciona **Grabar CD**.
3. El repositorio recupera las instancias en `runtime/temp/<StudyInstanceUID>/data`.
4. `StudyPackageBuilder` prepara el visualizador, manifiesto y directorio de etiqueta.
5. `TdBridgePublisher.CreateJob` validará el paquete y creará el JDF en staging cuando se incorpore la especificación oficial.
6. `TdBridgePublisher.SubmitJob` copia el archivo como `.tmp`, lo sincroniza y lo renombra a `.jdf` en el Monitoring Folder.
7. TD Bridge recoge el JDF y solicita la grabación/impresión a un Epson Discproducer compatible.

```text
AP1
 -> StudyPackageBuilder
 -> TdBridgePublisher.CreateJob
 -> JDF
 -> TdBridgePublisher.SubmitJob
 -> Monitoring Folder
 -> TD Bridge
 -> Epson Discproducer
```

Actualmente `BuildJDF` se detiene con un error claro porque el formato exacto no está documentado en el repositorio. No se deposita ningún archivo en TD Bridge. Con `publisher: "mock"`, los pasos de Epson siguen simulados. Los `.dcm` generados por el repositorio mock contienen texto y no son objetos DICOM.
