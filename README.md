# DicomDiscPublisher

Base de una aplicacion de consola para recibir estudios DICOM desde un tomografo o PACS. La integracion con Epson Discproducer PP-100III, base de datos, interfaz grafica y DICOMDIR quedan fuera de esta primera etapa.

## Requisitos

- Windows
- .NET SDK 8
- Acceso de red desde el SCU DICOM al puerto configurado

## Restaurar, compilar y ejecutar

```powershell
dotnet restore DicomDiscPublisher.sln
dotnet build DicomDiscPublisher.sln
dotnet run --project src/DicomDiscPublisher.Server
```

Por defecto el receptor usa AE Title `DICOM_BURNER` y puerto `11112`. La configuracion esta en `src/DicomDiscPublisher.Server/appsettings.json`; `IncomingPath` es relativo al directorio de ejecucion y por defecto guarda los estudios en `data/incoming`.

## Funcionalidad implementada

- SCP DICOM basado en fo-dicom 5.2.6.
- C-ECHO.
- C-STORE para objetos DICOM.
- Lectura de PatientID, PatientName, StudyInstanceUID, SeriesInstanceUID, SOPInstanceUID, StudyDescription, Modality y StudyDate.
- Agrupacion por StudyInstanceUID.
- Gestion en memoria de estudios mediante `IStudyManager` e `InMemoryStudyManager`.
- Deduplicacion de instancias por SOPInstanceUID, incluso si se reenvia el mismo estudio.
- Comando `studies` para listar los estudios conocidos durante la ejecucion del servidor.
- Deteccion automatica de fin de recepcion configurable mediante `StudyCompletionWorker`.
- Guardado como `<SOPInstanceUID>.dcm`, usando sufijo cuando el SOP llega duplicado para evitar sobrescritura.
- Conteo de imagenes por estudio y estado inicial `Receiving` en memoria.
- Cancelacion mediante Ctrl+C y logging con Microsoft.Extensions.Logging.

El log actual muestra PatientID y PatientName para facilitar pruebas. Debe anonimizarse o eliminarse antes de usar el sistema en produccion.

## Pruebas DICOM

El cliente de prueba incluido permite probar el receptor sin DCMTK:

Terminal 1:

```powershell
dotnet run --project src/DicomDiscPublisher.Server
```

Terminal 2:

```powershell
dotnet run --project src/DicomDiscPublisher.TestClient -- echo
dotnet run --project src/DicomDiscPublisher.TestClient -- generate
dotnet run --project src/DicomDiscPublisher.TestClient -- store sample.dcm
dotnet run --project src/DicomDiscPublisher.TestClient -- generate-study 20
dotnet run --project src/DicomDiscPublisher.TestClient -- store-study test-study
dotnet run --project src/DicomDiscPublisher.TestClient -- append-image test-study
```

El archivo recibido se guarda en `data/incoming/<StudyInstanceUID>/<SOPInstanceUID>.dcm`.
El host, puerto y AE Titles del cliente se configuran en
`src/DicomDiscPublisher.TestClient/appsettings.json`.
Mientras el servidor esta ejecutandose, escribir `studies` en su consola muestra
los estudios agrupados y su cantidad actual de instancias. Reenviar `test-study`
no incrementa la cantidad porque las instancias se deduplican por SOPInstanceUID.
Un estudio pasa de `Receiving` a `Received` despues de 15 segundos sin nuevas
instancias; el intervalo de escaneo y el timeout se configuran en
`src/DicomDiscPublisher.Server/appsettings.json`. `append-image` vuelve a ponerlo
en `Receiving` y reinicia el timeout.

Para C-ECHO se puede usar DCMTK:

```powershell
echo DICOM_BURNER | Out-Null
echoscu -v -aet TEST_SCU -aec DICOM_BURNER 127.0.0.1 11112
```

Para C-STORE, enviar un archivo de prueba:

```powershell
storescu -v -aet TEST_SCU -aec DICOM_BURNER 127.0.0.1 11112 .\sample.dcm
```

Tambien se puede configurar cualquier PACS o SCU con:

- Host: la IP de la maquina que ejecuta el receptor
- Puerto: `11112`
- Called AE: `DICOM_BURNER`
- Calling AE: cualquier valor

No hay restriccion por IP ni por Calling AE en esta etapa.

## Estructura

```text
DicomDiscPublisher/
|-- DicomDiscPublisher.sln
|-- README.md
`-- src/
    |-- DicomDiscPublisher.Core/
    |   |-- DicomStudy.cs
    |   `-- StudyStatus.cs
    |-- DicomDiscPublisher.Dicom/
    |   |-- DicomReceiverOptions.cs
    |   |-- DicomReceiverService.cs
    |   `-- StudyStorage.cs
    `-- DicomDiscPublisher.Server/
        |-- DicomServerWorker.cs
        |-- Program.cs
        `-- appsettings.json
    `-- DicomDiscPublisher.TestClient/
        |-- Program.cs
        |-- appsettings.json
        `-- DicomDiscPublisher.TestClient.csproj
```
