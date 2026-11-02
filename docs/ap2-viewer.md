# Visor portátil AP2

AP2 abre los paquetes producidos por AP1: `study.dat` cifrado y `data/` junto al ejecutable. En desarrollo puede indicarse la carpeta mediante `DICOM_VIEWER_CONTENT_DIR`.

## Uso

Abrir `Symphony Viewer.exe` en el disco. Seleccionar una serie en la columna izquierda y usar la rueda del ratón sobre la imagen para recorrer los cortes. La barra del visor incluye ventana/nivel, zoom, desplazamiento y reproducción.

La lectura permanece vinculada al estudio abierto. Las imágenes se ordenan por `position`; las series SR y las vacías no aparecen. Las miniaturas se renderizan con el mismo decodificador Cornerstone que las imágenes, sin interpretar los bytes manualmente. Si falla una lectura, la interfaz informa del problema y permite volver a abrir el estudio.

## Comprobaciones realizadas

- `go test ./apps/ap2-viewer`: manifiesto cifrado, rutas portátiles, orden, fijación de la carpeta del estudio, archivos ausentes e identificadores inválidos.
- `npm --prefix apps/ap2-viewer/frontend run build`: TypeScript y empaquetado de la interfaz.
- `node tools/ap2-smoke-test.mjs`: Chrome local con un paquete existente; imagen de 320×320, navegación en una serie de 33 imágenes, cambio de serie y 17 miniaturas. No hubo excepciones de JavaScript. Esta prueba suministra los bytes por un adaptador local; las pruebas Go cubren la lectura nativa.
- Compilación Wails para Windows e integración del ejecutable en AP1.

Las pruebas usan una muestra local; no equivalen a verificar todas las modalidades, sintaxis de transferencia, imágenes multiframe o plataformas. La lectura de discos DICOMDIR de terceros no forma parte del contrato del paquete AP1. La grabación física ya fue validada por el usuario en su instalación.

## Actualizar el visor incluido en los discos

1. Compilar la interfaz de AP2 y ejecutar `wails build -s` desde `apps/ap2-viewer`.
2. Copiar `build/bin/Portable DICOM Viewer.exe` a `apps/ap1-publisher/viewer-builds/windows/Symphony Viewer.exe`.
3. Compilar AP1 para incorporar ese ejecutable. Los discos ya grabados conservan su versión anterior.