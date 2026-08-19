using System.Collections.Concurrent;
using FellowOakDicom;
using Microsoft.Extensions.Logging;

namespace DicomDiscPublisher.Dicom;

public sealed class StudyStorage
{
    private readonly DicomReceiverOptions _options;
    private readonly ILogger<StudyStorage> _logger;
    private readonly ConcurrentDictionary<string, object> _studyLocks = new();

    public StudyStorage(DicomReceiverOptions options, ILogger<StudyStorage> logger)
    {
        _options = options;
        _logger = logger;
    }

    public async Task StoreAsync(DicomFile file, CancellationToken cancellationToken)
    {
        var dataset = file.Dataset;
        var studyUid = GetRequiredString(dataset, DicomTag.StudyInstanceUID, "unknown-study");
        var sopUid = GetRequiredString(dataset, DicomTag.SOPInstanceUID, Guid.NewGuid().ToString("N"));
        var studyDirectory = Path.Combine(_options.IncomingPath, studyUid);
        Directory.CreateDirectory(studyDirectory);

        var studyLock = _studyLocks.GetOrAdd(studyUid, _ => new object());
        string destinationPath;

        lock (studyLock)
        {
            destinationPath = GetAvailablePath(studyDirectory, sopUid);
            file.Save(destinationPath);
        }

        await Task.CompletedTask;
        cancellationToken.ThrowIfCancellationRequested();
        _logger.LogInformation("Archivo DICOM guardado: {FileName} para estudio {StudyInstanceUid}",
            Path.GetFileName(destinationPath), studyUid);
    }

    private static string GetAvailablePath(string directory, string sopUid)
    {
        var basePath = Path.Combine(directory, $"{sopUid}.dcm");
        if (!File.Exists(basePath))
            return basePath;

        var suffix = 1;
        string candidate;
        do
        {
            candidate = Path.Combine(directory, $"{sopUid}_{suffix++}.dcm");
        } while (File.Exists(candidate));

        return candidate;
    }

    private static string GetRequiredString(DicomDataset dataset, DicomTag tag, string fallback)
    {
        var value = GetString(dataset, tag);
        return string.IsNullOrWhiteSpace(value) ? fallback : value;
    }

    private static string GetString(DicomDataset dataset, DicomTag tag)
    {
        return dataset.TryGetString(tag, out var value) ? value : string.Empty;
    }

}
