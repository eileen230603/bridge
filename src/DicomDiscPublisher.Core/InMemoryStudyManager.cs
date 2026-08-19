using System.Collections.Concurrent;

namespace DicomDiscPublisher.Core;

public sealed class InMemoryStudyManager : IStudyManager
{
    private readonly ConcurrentDictionary<string, DicomStudy> _studies = new();
    private readonly ConcurrentDictionary<string, object> _studyLocks = new();
    private readonly ConcurrentDictionary<string, ConcurrentDictionary<string, byte>> _instanceIds = new();

    public Task RegisterInstanceAsync(
        StudyInstanceRegistration instance,
        CancellationToken cancellationToken = default)
    {
        cancellationToken.ThrowIfCancellationRequested();

        if (string.IsNullOrWhiteSpace(instance.StudyInstanceUid))
            throw new ArgumentException("StudyInstanceUID es obligatorio.", nameof(instance));
        if (string.IsNullOrWhiteSpace(instance.SopInstanceUid))
            throw new ArgumentException("SOPInstanceUID es obligatorio.", nameof(instance));

        var studyLock = _studyLocks.GetOrAdd(instance.StudyInstanceUid, _ => new object());
        var instanceIds = _instanceIds.GetOrAdd(
            instance.StudyInstanceUid,
            _ => new ConcurrentDictionary<string, byte>(StringComparer.Ordinal));

        lock (studyLock)
        {
            if (!instanceIds.TryAdd(instance.SopInstanceUid, 0))
                return Task.CompletedTask;

            var study = _studies.GetOrAdd(instance.StudyInstanceUid, _ => new DicomStudy
            {
                PatientId = instance.PatientId,
                PatientName = instance.PatientName,
                StudyInstanceUid = instance.StudyInstanceUid,
                StudyDescription = instance.StudyDescription,
                Modality = instance.Modality,
                StudyDate = instance.StudyDate,
                ReceivedAt = instance.ReceivedAt,
                LastImageReceivedAt = instance.ReceivedAt,
                ImageCount = 0,
                Status = StudyStatus.Receiving
            });

            study.ImageCount++;
            study.LastImageReceivedAt = instance.ReceivedAt;
            study.Status = StudyStatus.Receiving;
        }

        return Task.CompletedTask;
    }

    public Task<DicomStudy?> GetStudyAsync(
        string studyInstanceUid,
        CancellationToken cancellationToken = default)
    {
        cancellationToken.ThrowIfCancellationRequested();
        _studies.TryGetValue(studyInstanceUid, out var study);
        return Task.FromResult(study);
    }

    public Task<IReadOnlyCollection<DicomStudy>> GetStudiesAsync(
        CancellationToken cancellationToken = default)
    {
        cancellationToken.ThrowIfCancellationRequested();
        IReadOnlyCollection<DicomStudy> studies = _studies.Values
            .OrderBy(study => study.ReceivedAt)
            .ToArray();
        return Task.FromResult(studies);
    }

    public Task<bool> TryMarkAsReceivedAsync(
        string studyInstanceUid,
        DateTimeOffset expectedLastImageReceivedAt,
        CancellationToken cancellationToken = default)
    {
        cancellationToken.ThrowIfCancellationRequested();

        if (!_studies.TryGetValue(studyInstanceUid, out var study))
            return Task.FromResult(false);

        var studyLock = _studyLocks.GetOrAdd(studyInstanceUid, _ => new object());
        lock (studyLock)
        {
            if (study.Status != StudyStatus.Receiving ||
                study.LastImageReceivedAt != expectedLastImageReceivedAt)
                return Task.FromResult(false);

            study.Status = StudyStatus.Received;
            return Task.FromResult(true);
        }
    }
}