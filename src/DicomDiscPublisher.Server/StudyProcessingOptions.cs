namespace DicomDiscPublisher.Server;

public sealed class StudyProcessingOptions
{
    public const string SectionName = "StudyProcessing";
    public int CompletionTimeoutSeconds { get; set; } = 15;
    public int ScanIntervalSeconds { get; set; } = 2;
}