namespace DicomDiscPublisher.Core;

public enum StudyStatus
{
    Receiving,
    Received,
    Preparing,
    ReadyToPublish,
    Publishing,
    Completed,
    Failed
}
