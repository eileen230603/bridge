using DicomDiscPublisher.Core;
using Microsoft.Extensions.Hosting;

namespace DicomDiscPublisher.Server;

public sealed class StudiesCommandWorker : BackgroundService
{
    private readonly IStudyManager _studyManager;

    public StudiesCommandWorker(IStudyManager studyManager)
    {
        _studyManager = studyManager;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        while (!stoppingToken.IsCancellationRequested)
        {
            var command = await Task.Run(Console.ReadLine, stoppingToken);
            if (command is null)
                break;

            if (string.Equals(command.Trim(), "studies", StringComparison.OrdinalIgnoreCase))
                await PrintStudiesAsync(stoppingToken);
        }
    }

    private async Task PrintStudiesAsync(CancellationToken cancellationToken)
    {
        var studies = await _studyManager.GetStudiesAsync(cancellationToken);
        Console.WriteLine("-----------------------------------------------------------------------");
        Console.WriteLine("PATIENT             MODALITY    IMAGES    STATUS       STUDY UID");
        Console.WriteLine("-----------------------------------------------------------------------");

        foreach (var study in studies)
        {
            Console.WriteLine("{0,-19} {1,-10} {2,-9} {3,-12} {4}",
                Truncate(study.PatientName, 19),
                Truncate(study.Modality, 10),
                study.ImageCount,
                study.Status,
                study.StudyInstanceUid);
        }

        Console.WriteLine("-----------------------------------------------------------------------");
    }

    private static string Truncate(string value, int length)
    {
        return value.Length <= length ? value : value[..(length - 1)] + "...";
    }
}
