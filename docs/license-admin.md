# Administrador de licencias Symphony

Abrir `license-admin-private/Symphony License Admin.exe`. Ahora es una aplicación gráfica, sin consola.

## Emitir y consultar

1. Pulsar **Nueva licencia**.
2. Introducir cliente o institución y el Machine ID de AP1.
3. Elegir fecha de vencimiento o licencia permanente.
4. Generar y copiar el token, o guardarlo como archivo .txt.
5. Entregar el token al cliente para pegarlo en AP1 > Activar Licencia.

El panel muestra todas las licencias, activas, próximas a vencer (30 días) y vencidas. Incluye búsqueda por cliente/identificador y el tiempo restante, actualizado cada 30 segundos. Las licencias permanentes cuentan como activas. El vencimiento ocurre al terminar la fecha seleccionada, en UTC.

“Activa” significa vigente según el token, no confirma que se haya instalado en el equipo del cliente: AP1 valida sin enviar confirmaciones a este administrador.

## SQLite y respaldo

La base SQLite `license-admin-private/licenses.sqlite` se crea al abrir la aplicación. Al iniciar se importan automáticamente los archivos `licencia-*.txt` que estén junto al ejecutable, siempre que su firma corresponda a esta clave. No se duplican tokens ya importados y se conservan también los vencidos. Archivos inválidos o de otro emisor no se importan.

Para un respaldo completo, cerrar el administrador y copiar en privado la carpeta `license-admin-private`. Contiene la base y `issuer-private.key`, necesaria para emitir nuevas licencias. No distribuir esa carpeta a clientes. Está excluida de Git; la clave nunca se incrusta en AP1 ni en el ejecutable administrador.

La aplicación pública solo contiene la clave de verificación en `third_party/symphonylicensemanager-go/license/license.go`.

## Compilar

Desde `tools/license-admin`:

```
wails build
```

Copiar `build/bin/Symphony License Admin.exe` a la carpeta privada, junto a la clave existente. No reemplazar ni regenerar esa clave.

La interfaz está en `tools/license-admin/ui`, sin dependencias npm. SQLite usa el controlador Go `modernc.org/sqlite` (https://modernc.org/sqlite).

La compatibilidad por consola se conserva con `-machine <ID> -expires AAAA-MM-DD` y `-keys <carpeta>`. También registra esas emisiones en SQLite. `-init` es únicamente para configurar un emisor nuevo y nunca sobrescribe una clave existente; cambiar la clave requiere recompilar AP1 con la pública correspondiente.

## Verificar

```
go test ./tools/license-admin
```

Incluye firma, persistencia, importación sin duplicados, entradas inválidas y límites de vencimiento. La prueba de compatibilidad con el emisor local se omite donde la clave privada no está instalada.