namespace DicomDiscPublisher.Dicom;

public sealed class DicomReceiverOptions
{
    public const string SectionName = "Dicom";

    public string AeTitle { get; init; } = "DICOM_BURNER";
    public int Port { get; init; } = 11112;
    public string IncomingPath { get; init; } = "data/incoming";
}
