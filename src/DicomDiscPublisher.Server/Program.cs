using DicomDiscPublisher.Dicom;
using DicomDiscPublisher.Server;
using DicomDiscPublisher.Core;
using FellowOakDicom;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;

var host = Host.CreateDefaultBuilder(args)
	.ConfigureAppConfiguration((_, configuration) =>
	{
		configuration.AddJsonFile(
			Path.Combine(AppContext.BaseDirectory, "appsettings.json"),
			optional: false,
			reloadOnChange: false);
	})
	.ConfigureServices((context, services) =>
	{
		var dicomOptions = context.Configuration
			.GetSection(DicomReceiverOptions.SectionName)
			.Get<DicomReceiverOptions>() ?? new DicomReceiverOptions();
		var studyProcessingOptions = context.Configuration
			.GetSection(StudyProcessingOptions.SectionName)
			.Get<StudyProcessingOptions>() ?? new StudyProcessingOptions();

		services.AddFellowOakDicom();
		services.AddSingleton(dicomOptions);
		services.AddSingleton(studyProcessingOptions);
		services.AddSingleton<StudyStorage>();
		services.AddSingleton<IStudyManager, InMemoryStudyManager>();
		services.AddHostedService<DicomServerWorker>();
		services.AddHostedService<StudiesCommandWorker>();
		services.AddHostedService<StudyCompletionWorker>();
	})
	.Build();

DicomSetupBuilder.UseServiceProvider(host.Services);

Console.CancelKeyPress += (_, eventArgs) =>
{
	eventArgs.Cancel = true;
	host.Services.GetRequiredService<IHostApplicationLifetime>().StopApplication();
};

Console.WriteLine("========================================");
Console.WriteLine(" DICOM DISC PUBLISHER");
Console.WriteLine("========================================");
Console.WriteLine();

await host.RunAsync();
