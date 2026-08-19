namespace DicomDiscPublisher.Core;

public sealed class DicomStudy
{
    public Guid Id { get; init; } = Guid.NewGuid();
    public string PatientId { get; init; } = string.Empty;
    public string PatientName { get; init; } = string.Empty;
    public string StudyInstanceUid { get; init; } = string.Empty;
    public string StudyDescription { get; init; } = string.Empty;
    public string Modality { get; init; } = string.Empty;
    public DateTime? StudyDate { get; init; }
    public DateTimeOffset ReceivedAt { get; init; } = DateTimeOffset.UtcNow;
    public DateTimeOffset LastImageReceivedAt { get; set; } = DateTimeOffset.UtcNow;
    public int ImageCount { get; set; }
    public StudyStatus Status { get; set; } = StudyStatus.Receiving;
}
