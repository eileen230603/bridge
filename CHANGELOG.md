# Changelog

Formato basado en [Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/) y
[Versionado Semántico](https://semver.org/lang/es/). Ver [docs/versioning.md](docs/versioning.md).

## [Sin liberar]

### Cambiado
- Configuración → Grabado: el tipo de disco (CD, DVD, Blu-ray) define los formatos válidos; el formato recomendado se
  selecciona solo y "Cambiar" muestra únicamente los permitidos. Impresión de la etiqueta separada del disco.
- Se validan las combinaciones disco/formato al guardar y al leer `config.json`; las configuraciones antiguas inválidas
  (p. ej. Blu-ray con UDF 1.02) se migran al formato recomendado.

### Pendiente
- Confirmar con la documentación de TD Bridge el valor `FORMAT` para Blu-ray (hoy `UDF250`).

## [2.0.0] - 2026-11-20

Primera versión de la suite integrada con la nueva interfaz.

### Agregado
- Menú lateral con vistas separadas: **Estudios**, **Historial** y **Configuración**.
- Lista de estudios en tabla con filtro por paciente/ID al teclear y búsqueda por rango de fechas.
- Historial de trabajos con estado final y detalle de errores.
- Configuración por secciones (servidor, grabado, etiqueta del disco) con guardado independiente y sección **Acerca de**.
- Tema claro y oscuro que sigue al sistema, con botón para alternar.
- Modo desarrollo con datos de prueba (`make dev-ap1`).
- Versionado: archivo `VERSION`, versión visible en la app y política en `docs/versioning.md`.

### Cambiado
- La aplicación pasa a llamarse **Symphony Media Export** (antes "DICOM Disc Publisher"), con nuevo ícono.
- Nueva identidad visual (vidrio, paleta teal y rojo del logo), pantalla de carga y modal de licencia.
- Tamaño mínimo de ventana: 1100×700.

[Sin liberar]: https://github.com/eileen230603/bridge/compare/v2.0.0...HEAD
