using FellowOakDicom;
using FellowOakDicom.Network;
using FellowOakDicom.Network.Client;
using Microsoft.Extensions.Configuration;

const string defaultOutputFileName = "sample.dcm";

try
{
    var options = LoadOptions();
    var command = args.FirstOrDefault()?.ToLowerInvariant();

    switch (command)
    {
        case "echo":
            await SendEchoAsync(options);
            return 0;
        case "store":
            if (args.Length < 2 || string.IsNullOrWhiteSpace(args[1]))
            {
                Console.Error.WriteLine("Uso: store <ruta\archivo.dcm>");
                return 1;
            }

            await SendStoreAsync(options, args[1]);
            return 0;
        case "generate":
            GenerateSampleFile(defaultOutputFileName);
            return 0;
        case "generate-study":
            if (!TryGetCount(args, out var count))
                return 1;

            GenerateStudy("test-study", count);
            return 0;
        case "store-study":
            if (args.Length < 2 || string.IsNullOrWhiteSpace(args[1]))
            {
                Console.Error.WriteLine("Uso: store-study <carpeta>");
                return 1;
            }

            await SendStudyAsync(options, args[1]);
            return 0;
        case "append-image":
            if (args.Length < 2 || string.IsNullOrWhiteSpace(args[1]))
            {
                Console.Error.WriteLine("Uso: append-image <carpeta>");
                return 1;
            }

            await AppendImageAsync(options, args[1]);
            return 0;
        default:
            PrintUsage();
            return 1;
    }
}
catch (Exception exception)
{
    Console.Error.WriteLine($"Error: {exception}");
    return 1;
}

static DicomServerOptions LoadOptions()
{
    var configuration = new ConfigurationBuilder()
        .SetBasePath(AppContext.BaseDirectory)
        .AddJsonFile("appsettings.json", optional: false, reloadOnChange: false)
        .Build();

    return configuration.GetSection(DicomServerOptions.SectionName).Get<DicomServerOptions>()
        ?? throw new InvalidOperationException("No se encontro la seccion DicomServer en appsettings.json.");
}

static async Task SendEchoAsync(DicomServerOptions options, CancellationToken cancellationToken = default)
{
    Console.WriteLine($"Connecting to {options.CalledAeTitle} at {options.Host}:{options.Port}...");

    var request = new DicomCEchoRequest();
    var responseStatus = DicomStatus.ProcessingFailure;
    request.OnResponseReceived += (_, response) => responseStatus = response.Status;

    var client = CreateClient(options);
    await client.AddRequestAsync(request);
    await client.SendAsync(cancellationToken: cancellationToken);

    Console.WriteLine($"C-ECHO response: {responseStatus.State}");
    if (responseStatus.State != DicomState.Success)
        throw new InvalidOperationException($"El servidor devolvio el estado DICOM {responseStatus}.");
}

static async Task SendStoreAsync(
    DicomServerOptions options,
    string filePath,
    CancellationToken cancellationToken = default)
{
    if (!File.Exists(filePath))
        throw new FileNotFoundException("No existe el archivo DICOM indicado.", filePath);

    var file = await Task.Run(() => DicomFile.Open(filePath, FileReadOption.ReadAll), cancellationToken);
    Console.WriteLine($"Sending {Path.GetFileName(filePath)}...");

    var request = new DicomCStoreRequest(file);
    var responseStatus = DicomStatus.ProcessingFailure;
    request.OnResponseReceived += (_, response) => responseStatus = response.Status;

    var client = CreateClient(options);
    await client.AddRequestAsync(request);
    await client.SendAsync(cancellationToken: cancellationToken);

    Console.WriteLine($"C-STORE response: {responseStatus.State}");
    if (responseStatus.State != DicomState.Success)
        throw new InvalidOperationException($"El servidor devolvio el estado DICOM {responseStatus}.");
}

static void GenerateSampleFile(string filePath)
{
    var file = new DicomFile(CreateDataset(
        DicomUID.Generate().UID,
        DicomUID.Generate().UID,
        DicomUID.Generate().UID,
        1));
    file.Save(filePath);
    Console.WriteLine($"Generated {Path.GetFullPath(filePath)}");
}

static void GenerateStudy(string directoryPath, int count)
{
    Directory.CreateDirectory(directoryPath);
    foreach (var oldFile in Directory.EnumerateFiles(directoryPath, "*.dcm"))
        File.Delete(oldFile);

    var studyUid = DicomUID.Generate().UID;
    var seriesUid = DicomUID.Generate().UID;
    for (var index = 1; index <= count; index++)
    {
        var filePath = Path.Combine(directoryPath, $"image{index:000}.dcm");
        var dataset = CreateDataset(
            studyUid,
            seriesUid,
            DicomUID.Generate().UID,
            index);
        new DicomFile(dataset).Save(filePath);
    }

    Console.WriteLine($"Generated {count} DICOM instances in {Path.GetFullPath(directoryPath)}");
}

static async Task AppendImageAsync(
    DicomServerOptions options,
    string directoryPath,
    CancellationToken cancellationToken = default)
{
    if (!Directory.Exists(directoryPath))
        throw new DirectoryNotFoundException($"No existe la carpeta: {directoryPath}");

    var sourcePath = Directory.GetFiles(directoryPath, "*.dcm")
        .OrderBy(path => path, StringComparer.OrdinalIgnoreCase)
        .FirstOrDefault()
        ?? throw new InvalidOperationException("La carpeta no contiene archivos .dcm.");
    var sourceFile = await Task.Run(
        () => DicomFile.Open(sourcePath, FileReadOption.ReadAll),
        cancellationToken);
    var sourceDataset = sourceFile.Dataset;
    var instanceNumber = 1;
    string outputPath;
    do
    {
        outputPath = Path.Combine(directoryPath, $"image{instanceNumber++:000}.dcm");
    } while (File.Exists(outputPath));

    var dataset = CreateDataset(
        GetString(sourceDataset, DicomTag.StudyInstanceUID),
        GetString(sourceDataset, DicomTag.SeriesInstanceUID),
        DicomUID.Generate().UID,
        instanceNumber - 1,
        GetString(sourceDataset, DicomTag.PatientID),
        GetString(sourceDataset, DicomTag.PatientName),
        GetString(sourceDataset, DicomTag.StudyDescription),
        GetString(sourceDataset, DicomTag.Modality),
        GetString(sourceDataset, DicomTag.StudyDate));
    new DicomFile(dataset).Save(outputPath);

    await SendStoreAsync(options, outputPath, cancellationToken);
    Console.WriteLine($"Appended {Path.GetFileName(outputPath)} to {Path.GetFullPath(directoryPath)}");
}

static DicomDataset CreateDataset(
    string studyUid,
    string seriesUid,
    string sopUid,
    int instanceNumber,
    string patientId = "TEST001",
    string patientName = "PACIENTE^PRUEBA",
    string studyDescription = "ESTUDIO DE PRUEBA",
    string modality = "CT",
    string? studyDate = null)
{
    return new DicomDataset
    {
        { DicomTag.SOPClassUID, DicomUID.SecondaryCaptureImageStorage.UID },
        { DicomTag.SOPInstanceUID, sopUid },
        { DicomTag.PatientID, patientId },
        { DicomTag.PatientName, patientName },
        { DicomTag.StudyInstanceUID, studyUid },
        { DicomTag.SeriesInstanceUID, seriesUid },
        { DicomTag.Modality, modality },
        { DicomTag.StudyDescription, studyDescription },
        { DicomTag.StudyDate, studyDate ?? DateTime.Now.ToString("yyyyMMdd") },
        { DicomTag.SeriesNumber, "1" },
        { DicomTag.InstanceNumber, instanceNumber.ToString() },
        { DicomTag.StudyID, "TEST001" }
    };
}

static async Task SendStudyAsync(
    DicomServerOptions options,
    string directoryPath,
    CancellationToken cancellationToken = default)
{
    if (!Directory.Exists(directoryPath))
        throw new DirectoryNotFoundException($"No existe la carpeta: {directoryPath}");

    var filePaths = Directory.GetFiles(directoryPath, "*.dcm")
        .OrderBy(path => path, StringComparer.OrdinalIgnoreCase)
        .ToArray();
    if (filePaths.Length == 0)
        throw new InvalidOperationException("La carpeta no contiene archivos .dcm.");

    var statuses = new DicomStatus[filePaths.Length];
    var client = CreateClient(options);
    for (var index = 0; index < filePaths.Length; index++)
    {
        var request = new DicomCStoreRequest(
            await Task.Run(() => DicomFile.Open(filePaths[index], FileReadOption.ReadAll), cancellationToken));
        var responseIndex = index;
        request.OnResponseReceived += (_, response) => statuses[responseIndex] = response.Status;
        await client.AddRequestAsync(request);
    }

    for (var index = 0; index < filePaths.Length; index++)
        Console.WriteLine($"Sending {index + 1}/{filePaths.Length}...");

    await client.SendAsync(cancellationToken: cancellationToken);

    var acceptedCount = statuses.Count(status => status.State == DicomState.Success);
    if (acceptedCount != filePaths.Length)
        throw new InvalidOperationException($"Solo se aceptaron {acceptedCount}/{filePaths.Length} instancias.");

    Console.WriteLine();
    Console.WriteLine("Study sent successfully.");
    Console.WriteLine($"{acceptedCount}/{filePaths.Length} instances accepted.");
}

static bool TryGetCount(string[] arguments, out int count)
{
    if (arguments.Length < 2 || !int.TryParse(arguments[1], out count) || count <= 0)
    {
        Console.Error.WriteLine("Uso: generate-study <cantidad positiva>");
        count = 0;
        return false;
    }

    return true;
}

static string GetString(DicomDataset dataset, DicomTag tag)
{
    return dataset.TryGetString(tag, out var value) ? value : string.Empty;
}

static IDicomClient CreateClient(DicomServerOptions options)
{
    return DicomClientFactory.Create(
        options.Host,
        options.Port,
        useTls: false,
        options.CallingAeTitle,
        options.CalledAeTitle);
}

static void PrintUsage()
{
    Console.WriteLine("Uso:");
    Console.WriteLine("  echo");
    Console.WriteLine("  store <ruta\\archivo.dcm>");
    Console.WriteLine("  generate");
    Console.WriteLine("  generate-study <cantidad>");
    Console.WriteLine("  store-study <carpeta>");
    Console.WriteLine("  append-image <carpeta>");
}

sealed class DicomServerOptions
{
    public const string SectionName = "DicomServer";
    public string Host { get; set; } = "127.0.0.1";
    public int Port { get; set; } = 11112;
    public string CallingAeTitle { get; set; } = "TEST_SCU";
    public string CalledAeTitle { get; set; } = "DICOM_BURNER";
}
