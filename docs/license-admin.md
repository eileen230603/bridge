# Emisor de licencias Symphony AP1

El administrador abre `license-admin-private/Symphony License Admin.exe`.

1. Pegar el Machine ID que muestra AP1.
2. Escribir la fecha de vencimiento AAAA-MM-DD o pulsar Enter para una licencia permanente.
3. El programa muestra el token y guarda un archivo `licencia-<Machine ID>-<fecha>.txt` junto al ejecutable.
4. Entregar solamente ese token al usuario, quien lo pega en AP1 > licencia inactiva > Activar Licencia.

Las fechas vencen al finalizar el día indicado en UTC. Se necesita el AP1 actualizado con la nueva clave pública.

## Archivos del administrador

`license-admin-private/issuer-private.key` es la clave de emisión. Mantener una copia de seguridad privada; quien la tenga puede emitir licencias. No distribuir la carpeta del administrador a clientes. La carpeta completa está excluida de Git. La clave nunca se incrusta en AP1 ni en el ejecutable del emisor: el emisor la lee del archivo que lo acompaña.

`issuer-public.key` puede compartirse. Su valor está incorporado en `third_party/symphonylicensemanager-go/license/license.go`. La clave anterior de la biblioteca ya no se acepta.

Para compilar el emisor desde la raíz:

```
go build -o "license-admin-private/Symphony License Admin.exe" ./tools/license-admin
```

Para emitir por consola:

```
"Symphony License Admin.exe" -machine <Machine-ID> -expires 2027-12-31
```

La opción `-keys` permite especificar la carpeta privada. `-init` es solo para establecer un emisor NUEVO; no sobrescribe claves existentes. Un nuevo par de claves requiere actualizar la clave pública y recompilar AP1. No ejecutar para emitir cada licencia.

Pruebas:

```
go test ./tools/license-admin ./apps/ap1-publisher/...
```

La prueba local de compatibilidad comprueba que la clave privada instalada emite tokens aceptados por la aplicación; se omite donde no está instalada esa clave.