# Rutas e historial de AP1

## Configuración de instalación

El ejecutable incluye `config.defaults.json`, sin licencia, servidor privado ni rutas del equipo de desarrollo. Si no existe configuración externa, los datos se crean en `runtime/` junto al ejecutable. Utilizar una carpeta con permisos de escritura.

Orden de selección: `AP1_CONFIG` explícito, `config.json` junto al ejecutable, configuración del proyecto para builds en `build/bin`, y configuración del directorio de trabajo para desarrollo. Las rutas relativas se resuelven desde la carpeta del archivo de configuración y se conservan relativas al guardar. Las rutas absolutas configuradas para Epson se respetan.

En este proyecto, `config.json` conserva la licencia y el servidor existentes, pero usa `../../runtime/` para mantener las carpetas actuales. El logo usa `assets/hospital-logo.png`. La carpeta de órdenes Epson continúa siendo la configurada en esta estación.

Al llevar solo el ejecutable a una instalación nueva, configurar el servidor y activar la licencia del equipo desde la aplicación. Para trasladar una instalación completa, cerrar AP1 y copiar también su configuración, `runtime/` y `assets/`, conservando la estructura relativa. Revisar la carpeta de órdenes de TD Bridge en el equipo de destino.

## Historial

`historyFile` configura la base SQLite; por defecto es `runtime/history.sqlite`. En el proyecto se ubica en `../../runtime/history.sqlite`, fuera de la carpeta temporal de paquetes.

Se registra cada trabajo antes de comenzar la descarga, antes de enviarlo a Epson y al cambiar su estado. Cada identificador tiene una sola fila que se actualiza. La pantalla «Historial de trabajos» recupera los registros al abrir AP1.

Al reiniciar, los trabajos cuya preparación se interrumpió pasan a fallidos. Los trabajos enviados o que estaban en fase de envío conservan su registro y se consultan en TD Bridge; no se vuelven a enviar automáticamente. Los estados completado y fallido se conservan aunque Epson retire sus archivos de seguimiento.

Un error al abrir la base impide iniciar con un historial vacío ficticio. Un error al escribir se comunica en la pantalla; si ocurre antes del envío, no se entrega el trabajo a Epson. Si ocurre después del envío, el mensaje indica que no debe repetirse la grabación.

El historial empieza a persistirse con esta versión. Los trabajos que versiones anteriores guardaban solamente en memoria no se recuperan automáticamente. Para respaldarlo, cerrar AP1 y copiar la base SQLite junto con la configuración.

## Verificación

- `go test ./apps/ap1-publisher/...`: rutas tras guardar y trasladar, carpeta Epson explícita, defaults sin activación, reapertura SQLite, actualización sin duplicados, base corrupta, recuperación de estados desde Epson y fallo de persistencia antes de descargar.
- `npm --prefix apps/ap1-publisher/frontend run build`.
- `wails build -s` desde `apps/ap1-publisher`.