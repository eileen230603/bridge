# Flujo de publicación

1. AP1 consulta estudios a través de `StudyRepository`.
2. El usuario selecciona **Grabar CD**.
3. El repositorio recupera las instancias en `runtime/temp/<StudyInstanceUID>/data`.
4. `StudyPackageBuilder` prepara el visualizador, manifiesto y directorio de etiqueta.
5. `TdBridgePublisher.CreateJob` valida el paquete y crea en staging un JDF para CD de datos conforme al Technical Reference Guide E/F revisión 22.
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

AP1 no tiene fallback de simulación. La búsqueda ejecuta C-ECHO y C-FIND reales; la recuperación usa C-MOVE hacia el Storage SCP configurado. Los tests de transporte TD Bridge usan exclusivamente carpetas temporales.
