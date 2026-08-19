using DicomDiscPublisher.Dicom;
using FellowOakDicom.Network;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;

namespace DicomDiscPublisher.Server;

public sealed class DicomServerWorker : BackgroundService
{
    private readonly IDicomServerFactory _serverFactory;
    private readonly DicomReceiverOptions _options;
    private readonly ILogger<DicomServerWorker> _logger;
    private IDicomServer? _server;

    public DicomServerWorker(
        IDicomServerFactory serverFactory,
        DicomReceiverOptions options,
        ILogger<DicomServerWorker> logger)
    {
        _serverFactory = serverFactory;
        _options = options;
        _logger = logger;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        Directory.CreateDirectory(_options.IncomingPath);
        _server = _serverFactory.Create<DicomReceiverService>(_options.Port);

        _logger.LogInformation("DICOM Server: ONLINE");
        _logger.LogInformation("AE Title: {AeTitle}", _options.AeTitle);
        _logger.LogInformation("Port: {Port}", _options.Port);
        _logger.LogInformation("Waiting for DICOM studies...");

        try
        {
            await Task.Delay(Timeout.Infinite, stoppingToken);
        }
        catch (OperationCanceledException) when (stoppingToken.IsCancellationRequested)
        {
            _logger.LogInformation("Deteniendo servidor DICOM");
        }
    }

    public override Task StopAsync(CancellationToken cancellationToken)
    {
        _server?.Stop();
        _server?.Dispose();
        _server = null;
        return base.StopAsync(cancellationToken);
    }
}
