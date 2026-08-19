using DicomDiscPublisher.Core;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;

namespace DicomDiscPublisher.Server;

public sealed class StudyCompletionWorker : BackgroundService
{
    private readonly IStudyManager _studyManager;
    private readonly StudyProcessingOptions _options;
    private readonly ILogger<StudyCompletionWorker> _logger;

    public StudyCompletionWorker(
        IStudyManager studyManager,
        StudyProcessingOptions options,
        ILogger<StudyCompletionWorker> logger)
    {
        _studyManager = studyManager;
        _options = options;
        _logger = logger;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        using var timer = new PeriodicTimer(TimeSpan.FromSeconds(_options.ScanIntervalSeconds));
        while (await timer.WaitForNextTickAsync(stoppingToken))
        {
            await ScanAsync(stoppingToken);
        }
    }

    private async Task ScanAsync(CancellationToken cancellationToken)
    {
        var now = DateTimeOffset.UtcNow;
        var timeout = TimeSpan.FromSeconds(_options.CompletionTimeoutSeconds);
        var studies = await _studyManager.GetStudiesAsync(cancellationToken);

        foreach (var study in studies.Where(study => study.Status == StudyStatus.Receiving))
        {
            var expectedLastImageReceivedAt = study.LastImageReceivedAt;
            if (now - expectedLastImageReceivedAt < timeout)
                continue;

            if (await _studyManager.TryMarkAsReceivedAsync(
                    study.StudyInstanceUid,
                    expectedLastImageReceivedAt,
                    cancellationToken))
            {
                _logger.LogInformation("[STUDY] Study completed receiving\n[STUDY] UID: {StudyUid}\n[STUDY] Images: {ImageCount}\n[STUDY] Status: {Status}",
                    study.StudyInstanceUid, study.ImageCount, StudyStatus.Received);
            }
        }
    }
}