# Política de versionado

Symphony Media Export usa **Versionado Semántico 2.0** (`MAYOR.MENOR.PARCHE`, p. ej. `2.0.0`).
La serie 1.x corresponde al prototipo inicial; **2.0.0** es la primera versión de la suite integrada
(AP1 + AP2 + licencias) con la nueva interfaz.

## Cuándo sube cada número

| Cambio | Sube | Ejemplos |
|---|---|---|
| Rompe compatibilidad | **MAYOR** | Cambia el formato de `study.dat`/manifest o el cifrado y AP2 de versiones anteriores ya no abre los discos; cambia el formato de licencias o de `config.json` sin migración automática; cambia el contrato con Symphony o TD Bridge. |
| Funcionalidad nueva compatible | **MENOR** | Nueva vista o sección de configuración, nuevo filtro, nuevo tipo de disco, nuevo campo opcional en la configuración. |
| Corrección sin cambio funcional | **PARCHE** | Bug en la búsqueda, texto, ajuste visual, rendimiento, dependencia con parche de seguridad. |

Al subir MAYOR o MENOR, los números menores se reinician a 0 (`2.3.4` → `2.4.0` → `3.0.0`).

## Versiones preliminares

Para probar antes de liberar se usa el sufijo `-rc.N` (`2.1.0-rc.1`) o `-beta.N`. Una preliminar nunca se entrega
a una clínica. Al liberar, se publica el número sin sufijo.

## Dónde vive la versión

Una sola fuente de verdad: el archivo **`VERSION`** (raíz del repositorio). Debe coincidir con:

1. `apps/ap1-publisher/version.go` → `Version`
2. `apps/ap1-publisher/wails.json` → `info.productVersion`

`make version-check` (y `make build-ap1`) falla si no coinciden; también lo comprueba `go test`.
El commit y la fecha de compilación se inyectan al compilar con `make build-ap1` y se ven en
**Configuración → Acerca de**; la versión también aparece al pie del menú lateral y en el log de inicio.

## Cómo liberar una versión

1. Parta de una rama actualizada y con pruebas en verde (`go test ./...`).
2. Decida el número según la tabla de arriba.
3. Actualice los tres lugares (`VERSION`, `version.go`, `wails.json`) y `CHANGELOG.md`
   (mueva lo de *Sin liberar* a la nueva versión, con fecha `AAAA-MM-DD`).
4. `make version-check`, luego `make build-ap1` (y `make build-ap2` si cambió AP2).
5. Commit `chore(release): vX.Y.Z` y etiqueta anotada: `git tag -a vX.Y.Z -m "Symphony Media Export X.Y.Z"`.
6. Suba la rama y la etiqueta: `git push && git push origin vX.Y.Z`.

Las etiquetas no se mueven ni se reescriben. Si una versión liberada tiene un error, se corrige con una versión de PARCHE nueva.

## Compatibilidad AP1 ↔ AP2

AP1 empaqueta el visor (AP2) dentro de cada disco junto con `study.dat`, de modo que **cada disco es autocontenido**:
el visor del disco siempre entiende su propio `study.dat`. Cambiar el formato de `study.dat` es un cambio MAYOR
si afecta a discos ya grabados que se abran con un visor distinto al que traen.

## Mensajes de commit

Se recomienda el estilo *Conventional Commits* (`feat:`, `fix:`, `chore:`, `docs:`), que facilita decidir el siguiente
número de versión y redactar el changelog.
