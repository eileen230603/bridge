namespace DicomDiscPublisher.Core;

public interface IStudyManager
{
    Task RegisterInstanceAsync(StudyInstanceRegistration instance, CancellationToken cancellationToken = default);

    Task<DicomStudy?> GetStudyAsync(string studyInstanceUid, CancellationToken cancellationToken = default);

    Task<IReadOnlyCollection<DicomStudy>> GetStudiesAsync(CancellationToken cancellationToken = default);

    Task<bool> TryMarkAsReceivedAsync(
        string studyInstanceUid,
        DateTimeOffset expectedLastImageReceivedAt,
        CancellationToken cancellationToken = default);
}

public sealed record StudyInstanceRegistration(
    string PatientId,
    string PatientName,
    string StudyInstanceUid,
    string StudyDescription,
    string Modality,
    DateTime? StudyDate,
    string SopInstanceUid,
    DateTimeOffset ReceivedAt);