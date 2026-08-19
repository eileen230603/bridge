using System.Text;
using DicomDiscPublisher.Core;
using FellowOakDicom;
using FellowOakDicom.Network;
using Microsoft.Extensions.Logging;

namespace DicomDiscPublisher.Dicom;

public sealed class DicomReceiverService : DicomService, IDicomServiceProvider, IDicomCEchoProvider, IDicomCStoreProvider
{
    private readonly StudyStorage _storage;
    private readonly IStudyManager _studyManager;
    private readonly ILogger _logger;

    public DicomReceiverService(
        INetworkStream stream,
        Encoding fallbackEncoding,
        ILogger logger,
        DicomServiceDependencies dependencies,
        StudyStorage storage,
        IStudyManager studyManager)
        : base(stream, fallbackEncoding, logger, dependencies)
    {
        _storage = storage;
        _studyManager = studyManager;
        _logger = logger;
    }

    public Task OnReceiveAssociationRequestAsync(DicomAssociation association)
    {
        var remote = association.RemoteHost ?? "desconocido";
        _logger.LogInformation("Nueva asociacion desde {RemoteHost}, AE remoto {CallingAe}", remote, association.CallingAE);

        foreach (var presentationContext in association.PresentationContexts)
        {
            presentationContext.SetResult(DicomPresentationContextResult.Accept);
        }

        return SendAssociationAcceptAsync(association);
    }

    public Task OnReceiveAssociationReleaseRequestAsync()
    {
        _logger.LogInformation("Solicitud de cierre de asociacion recibida");
        return Task.CompletedTask;
    }

    public void OnReceiveAbort(DicomAbortSource source, DicomAbortReason reason)
    {
        _logger.LogWarning("Asociacion abortada. Origen: {Source}, razon: {Reason}", source, reason);
    }

    public void OnConnectionClosed(Exception? exception)
    {
        if (exception is null)
            _logger.LogInformation("Asociacion cerrada");
        else
            _logger.LogWarning(exception, "Asociacion cerrada con error");
    }

    public Task<DicomCEchoResponse> OnCEchoRequestAsync(DicomCEchoRequest request)
    {
        _logger.LogInformation("C-ECHO recibido");
        return Task.FromResult(new DicomCEchoResponse(request, DicomStatus.Success));
    }

    public async Task<DicomCStoreResponse> OnCStoreRequestAsync(DicomCStoreRequest request)
    {
        try
        {
            var dataset = request.Dataset;
            var patientId = GetString(dataset, DicomTag.PatientID);
            var patientName = GetString(dataset, DicomTag.PatientName);
            var studyUid = GetString(dataset, DicomTag.StudyInstanceUID);
            var sopUid = GetString(dataset, DicomTag.SOPInstanceUID);
            var receivedAt = DateTimeOffset.UtcNow;

            await _storage.StoreAsync(request.File, CancellationToken.None);
            await _studyManager.RegisterInstanceAsync(new StudyInstanceRegistration(
                patientId,
                patientName,
                studyUid,
                GetString(dataset, DicomTag.StudyDescription),
                GetString(dataset, DicomTag.Modality),
                ParseStudyDate(GetString(dataset, DicomTag.StudyDate)),
                sopUid,
                receivedAt));

            var study = await _studyManager.GetStudyAsync(studyUid);
            if (study is not null)
            {
                _logger.LogInformation("[DICOM] Study: {StudyUid}\n[DICOM] Patient: {PatientName}\n[DICOM] Modality: {Modality}\n[DICOM] Images: {ImageCount}\n[DICOM] Status: {Status}",
                    study.StudyInstanceUid, study.PatientName, study.Modality, study.ImageCount, study.Status);
            }
            return new DicomCStoreResponse(request, DicomStatus.Success);
        }
        catch (Exception exception)
        {
            _logger.LogError(exception, "Error procesando solicitud C-STORE");
            return new DicomCStoreResponse(request, DicomStatus.ProcessingFailure);
        }
    }

    public Task OnCStoreRequestExceptionAsync(string tempFileName, Exception e)
    {
        _logger.LogError(e, "Excepcion C-STORE para archivo temporal {TempFileName}", tempFileName);
        return Task.CompletedTask;
    }

    private static string GetString(DicomDataset dataset, DicomTag tag)
    {
        return dataset.TryGetString(tag, out var value) ? value : string.Empty;
    }

    private static DateTime? ParseStudyDate(string value)
    {
        return DateTime.TryParseExact(value, "yyyyMMdd", null,
            System.Globalization.DateTimeStyles.None, out var date) ? date : null;
    }
}
