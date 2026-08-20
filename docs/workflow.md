# Flujo de publicación

1. AP1 consulta estudios a través de `StudyRepository`.
2. El usuario selecciona **Grabar CD**.
3. El repositorio recupera las instancias en `runtime/temp/<StudyInstanceUID>/data`.
4. `StudyPackageBuilder` prepara el visualizador, manifiesto y directorio de etiqueta.
5. `EpsonPublisher.CreateJob` crea y cierra el trabajo en staging.
6. `EpsonPublisher.SubmitJob` lo expone de forma segura en el Hot Folder.
7. En producción, TD Bridge procesará el trabajo y la PP-100III grabará e imprimirá el disco.

Los pasos 1, 3, 5, 6 y 7 están simulados en esta etapa. Los `.dcm` generados contienen texto y no son objetos DICOM.
