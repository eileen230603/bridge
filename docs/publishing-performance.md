# Rendimiento de publicación Epson

AP1 usa TD Bridge sin fijar un modelo ni una velocidad nominal. La PP-100III y otros Discproducer deben tener una versión de TD Bridge/Total Disc Maker compatible con el equipo. Esto no agrega soporte de grabación a modelos que solo imprimen, ni Blu-ray a unidades que no lo admiten.

En Configuración → Parámetros del Grabador, elegir la superficie real del disco y el modo Rápida. Se ofrecen Estándar y Alta calidad para impresión rápida. Brillante/certificada Epson requiere Calidad. Por defecto se conservan los ajustes de impresión del equipo, para no asumir qué superficie se utiliza.

Los campos JSON epson.printMode y epson.labelType tienen valor 0 por defecto (heredar).
printMode: 1 = calidad, 2 = rápida.
labelType: 1 = estándar, 2 = alta calidad, 3 = brillante/certificada.
El modo 3, exclusivo de PP-100AP en la referencia de Epson, no se expone.
Las combinaciones incompatibles se rechazan al cargar/guardar y antes de emitir un JDF.

La velocidad de grabación sigue automática: se omite WRITING_SPEED, que según TD Bridge solicita la mayor disponible. No se cambia COMPARE ni la finalización. No se modifican tiempos de secado ni ajustes del controlador.

La descarga y la extracción de visores se solapan. Los ejecutables se copian mediante streaming. La etiqueta se genera una vez y study.dat se escribe una vez. Cada trabajo tiene una carpeta exclusiva en temporaryDirectory; repetir un estudio no reemplaza los archivos de otro trabajo en cola. La limpieza existente sigue siendo dry-run.

Los logs incluyen duration_ms para descarga, extracción y preparación total. Los tiempos de descarga y extracción se solapan: no deben sumarse. Para comparar, publicar el mismo estudio con el mismo soporte, registrar preparación y tiempo hasta completado, y comprobar legibilidad de la etiqueta. La mejora física depende del equipo, soporte y ajustes previos; no hay un porcentaje garantizado. Los cambios de impresión se aplican a nuevos trabajos; los ya iniciados conservan su configuración.

Referencia: Epson, TD Bridge Technical Reference Guide, pp. 45 y 48 (copia del manual del fabricante):
https://www.imagingpacs.com/sites/imagingpacs.com/files/TD-Bridge-TRG-E-13.pdf