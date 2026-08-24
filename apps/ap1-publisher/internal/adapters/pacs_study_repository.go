package adapters

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/amrshadid/go-dicom/dataelem"
	"github.com/amrshadid/go-dicom/dataset"
	"github.com/amrshadid/go-dicom/filebase"
	"github.com/amrshadid/go-dicom/filewriter"
	"github.com/amrshadid/go-dicom/network"
	"github.com/amrshadid/go-dicom/tag"
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/config"
	"github.com/local/dicom-disc-suite/shared/models"
)

const (
	associationTimeout = 30 * time.Second
	retrieveTimeout    = 15 * time.Minute
	implementationUID  = "1.2.826.0.1.3680043.10.1000.1"
)

type receiveSession struct {
	studyUID   string
	directory  string
	received   map[string]struct{}
	writeError error
}

// PacsStudyRepository implements classic DICOM Query/Retrieve as a Study Root
// SCU and runs the Storage SCP used as the destination of C-MOVE operations.
type PacsStudyRepository struct {
	Config config.PACSConfig
	Logger *slog.Logger

	mu        sync.Mutex
	sessions  map[string]*receiveSession
	scp       *network.SCP
	cancelSCP context.CancelFunc
}

func NewPacsStudyRepository(cfg config.PACSConfig, logger *slog.Logger) *PacsStudyRepository {
	return &PacsStudyRepository{Config: cfg, Logger: logger, sessions: make(map[string]*receiveSession)}
}

func (p *PacsStudyRepository) Start(ctx context.Context) error {
	if !p.Config.IsConfigured() {
		return errors.New("PACS no configurado")
	}
	scpContext, cancel := context.WithCancel(ctx)
	p.cancelSCP = cancel
	p.logger().Info("[SCP] Starting", "address", fmt.Sprintf("0.0.0.0:%d", p.Config.ReceivePort))
	scp := network.NewSCP(network.SCPConfig{AETitle: p.Config.MoveDestinationAETitle, BindAddress: "0.0.0.0", Port: p.Config.ReceivePort})
	abstractSyntaxes := append([]string{network.VerificationSOPClassUID}, network.AllStorageSOPClassUIDs()...)
	scp.SetSupportedAbstractSyntaxes(abstractSyntaxes)
	scp.SetSupportedTransferSyntaxes(network.DefaultTransferSyntaxes())
	scp.SetHandler(&diagnosticStorageHandler{
		StorageHandler: network.StorageHandler{OnStore: p.storeInstance},
		logger:         p.logger(),
	})
	p.scp = scp
	started := make(chan error, 1)
	go func() { started <- scp.ListenAndServe(scpContext) }()
	deadline := time.NewTimer(3 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case err := <-started:
			return fmt.Errorf("iniciar Storage SCP: %w", err)
		case <-ticker.C:
			if scp.Addr() != "" {
				p.logger().Info("[SCP] Waiting for association", "ae_title", p.Config.MoveDestinationAETitle, "address", scp.Addr())
				return nil
			}
		case <-deadline.C:
			return errors.New("timeout al iniciar Storage SCP")
		}
	}
}

type diagnosticStorageHandler struct {
	network.StorageHandler
	logger *slog.Logger
}

func (h *diagnosticStorageHandler) HandleCEcho(ctx context.Context, req *network.CEchoRequest) (*network.CEchoResponse, error) {
	h.logAssociation(ctx)
	h.logger.Info("[SCP] C-ECHO received")
	return h.StorageHandler.HandleCEcho(ctx, req)
}

func (h *diagnosticStorageHandler) HandleCStore(ctx context.Context, req *network.CStoreRequest) (*network.CStoreResponse, error) {
	h.logAssociation(ctx)
	h.logger.Info("[SCP] C-STORE received", "sop_class_uid", req.AffectedSOPClass, "sop_instance_uid", req.AffectedSOPInstance)
	return h.StorageHandler.HandleCStore(ctx, req)
}

func (h *diagnosticStorageHandler) logAssociation(ctx context.Context) {
	info := network.AssociationInfoFromContext(ctx)
	if info == nil {
		return
	}
	h.logger.Info("[SCP] Association accepted", "remote", info.RemoteAddr, "calling_ae", info.CallingAE, "called_ae", info.CalledAE)
}

func (p *PacsStudyRepository) Close() error {
	if p.cancelSCP != nil {
		p.cancelSCP()
	}
	if p.scp == nil {
		return nil
	}
	return p.scp.Close()
}

func (p *PacsStudyRepository) Echo(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, associationTimeout)
	defer cancel()
	scu := p.newSCU()
	if err := scu.Associate(ctx, network.DefaultVerificationContexts()); err != nil {
		return err
	}
	defer func() { _ = scu.Release(context.Background()) }()
	return scu.Echo(ctx)
}

func (p *PacsStudyRepository) SearchStudies(ctx context.Context, from, to time.Time) ([]models.Study, error) {
	query := dataset.NewDataset()
	for _, elem := range []*dataelem.DataElement{
		dataelem.NewDataElement(tag.New(0x0008, 0x0052), dataelem.CS, []byte("STUDY")),
		dataelem.NewDataElement(tag.New(0x0008, 0x0020), dataelem.DA, []byte(from.Format("20060102")+"-"+to.Format("20060102"))),
		dataelem.NewDataElement(tag.New(0x0010, 0x0010), dataelem.PN, []byte{}),
		dataelem.NewDataElement(tag.New(0x0010, 0x0020), dataelem.LO, []byte{}),
		dataelem.NewDataElement(tag.New(0x0020, 0x000D), dataelem.UI, []byte{}),
		dataelem.NewDataElement(tag.New(0x0008, 0x1030), dataelem.LO, []byte{}),
		dataelem.NewDataElement(tag.New(0x0008, 0x0061), dataelem.CS, []byte{}),
		dataelem.NewDataElement(tag.New(0x0020, 0x1206), dataelem.IS, []byte{}),
		dataelem.NewDataElement(tag.New(0x0020, 0x1208), dataelem.IS, []byte{}),
	} {
		if err := query.Add(elem); err != nil {
			return nil, err
		}
	}
	ctx, cancel := context.WithTimeout(ctx, associationTimeout)
	defer cancel()
	scu := p.newSCU()
	contexts := []network.PresentationContextItem{{ID: 1, AbstractSyntax: network.StudyRootQueryRetrieveFind, TransferSyntaxes: network.DefaultTransferSyntaxes()}}
	if err := scu.Associate(ctx, contexts); err != nil {
		return nil, err
	}
	defer func() { _ = scu.Release(context.Background()) }()
	results, err := scu.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	studies := make([]models.Study, 0)
	for result := range results {
		if result.Err != nil {
			return nil, result.Err
		}
		if result.DataSet == nil {
			continue
		}
		uid := cleanDICOMString(result.DataSet.GetStringByKeyword("StudyInstanceUID"))
		if uid == "" {
			continue
		}
		studies = append(studies, models.Study{
			StudyInstanceUID: uid,
			PatientID:        cleanDICOMString(result.DataSet.GetStringByKeyword("PatientID")),
			PatientName:      formatPersonName(result.DataSet.GetStringByKeyword("PatientName")),
			StudyDescription: cleanDICOMString(result.DataSet.GetStringByKeyword("StudyDescription")),
			StudyDate:        formatDICOMDate(result.DataSet.GetStringByKeyword("StudyDate")),
			Modality:         cleanDICOMString(result.DataSet.GetStringByKeyword("ModalitiesInStudy")),
			SeriesCount:      dicomInt(result.DataSet.GetStringByKeyword("NumberOfStudyRelatedSeries")),
			InstanceCount:    dicomInt(result.DataSet.GetStringByKeyword("NumberOfStudyRelatedInstances")),
		})
	}
	return studies, nil
}

func (p *PacsStudyRepository) RetrieveStudy(ctx context.Context, studyUID, destination string) error {
	if strings.TrimSpace(studyUID) == "" {
		return errors.New("StudyInstanceUID vacío")
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return err
	}
	session := &receiveSession{studyUID: studyUID, directory: destination, received: make(map[string]struct{})}
	p.mu.Lock()
	if len(p.sessions) != 0 {
		p.mu.Unlock()
		return errors.New("ya existe una recuperación DICOM en curso")
	}
	p.sessions[studyUID] = session
	p.mu.Unlock()
	defer func() {
		p.mu.Lock()
		delete(p.sessions, studyUID)
		p.mu.Unlock()
	}()
	query := dataset.NewDataset()
	_ = query.Add(dataelem.NewDataElement(tag.New(0x0008, 0x0052), dataelem.CS, []byte("STUDY")))
	_ = query.Add(dataelem.NewDataElement(tag.New(0x0020, 0x000D), dataelem.UI, []byte(studyUID)))
	ctx, cancel := context.WithTimeout(ctx, retrieveTimeout)
	defer cancel()
	scu := p.newSCU()
	contexts := []network.PresentationContextItem{{ID: 1, AbstractSyntax: network.StudyRootQueryRetrieveMove, TransferSyntaxes: network.DefaultTransferSyntaxes()}}
	if err := scu.Associate(ctx, contexts); err != nil {
		return err
	}
	defer func() { _ = scu.Release(context.Background()) }()
	if err := scu.Move(ctx, query, p.Config.MoveDestinationAETitle); err != nil {
		return err
	}
	p.mu.Lock()
	received := len(session.received)
	writeErr := session.writeError
	p.mu.Unlock()
	if writeErr != nil {
		return writeErr
	}
	if received == 0 {
		return errors.New("C-MOVE finalizó sin recibir instancias DICOM")
	}
	p.logger().Info("DICOM study received", "study_uid", studyUID, "instances", received)
	return nil
}

func (p *PacsStudyRepository) storeInstance(ctx context.Context, sopClassUID, sopInstanceUID string, ds *dataset.Dataset) uint16 {
	if ds == nil {
		return network.StatusUnableToProcess
	}
	studyUID := cleanDICOMString(ds.GetStringByKeyword("StudyInstanceUID"))
	sopInstanceUID = cleanDICOMString(sopInstanceUID)
	if sopInstanceUID == "" {
		sopInstanceUID = cleanDICOMString(ds.GetStringByKeyword("SOPInstanceUID"))
	}
	p.mu.Lock()
	session := p.sessions[studyUID]
	if session == nil || session.studyUID != studyUID || sopInstanceUID == "" {
		p.mu.Unlock()
		return network.StatusDataSetNotMatch
	}
	if _, duplicate := session.received[sopInstanceUID]; duplicate {
		p.mu.Unlock()
		return network.StatusSuccess
	}
	p.mu.Unlock()
	transferSyntax := network.ExplicitVRLittleEndianUID
	if info := network.AssociationInfoFromContext(ctx); info != nil {
		for _, accepted := range info.AcceptedContexts {
			if accepted.AbstractSyntax == sopClassUID {
				transferSyntax = accepted.TransferSyntax
				break
			}
		}
	}
	path := filepath.Join(session.directory, sopInstanceUID+".dcm")
	if err := writePart10(path, ds, sopClassUID, sopInstanceUID, transferSyntax, p.Config.CallingAETitle); err != nil {
		p.mu.Lock()
		session.writeError = fmt.Errorf("guardar instancia %s: %w", sopInstanceUID, err)
		p.mu.Unlock()
		p.logger().Error("[SCP] C-STORE failed", "sop_instance_uid", sopInstanceUID, "error", err)
		return network.StatusOutOfResources
	}
	p.mu.Lock()
	session.received[sopInstanceUID] = struct{}{}
	p.mu.Unlock()
	return network.StatusSuccess
}

func (p *PacsStudyRepository) newSCU() *network.SCU {
	return network.NewSCU(network.SCUConfig{CallingAE: p.Config.CallingAETitle, CalledAE: p.Config.CalledAETitle, Address: fmt.Sprintf("%s:%d", p.Config.Host, p.Config.Port)})
}

func (p *PacsStudyRepository) logger() *slog.Logger {
	if p.Logger != nil {
		return p.Logger
	}
	return slog.Default()
}

func cleanDICOMString(value string) string { return strings.Trim(string([]byte(value)), " \x00") }
func formatPersonName(value string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(cleanDICOMString(value), "^", " ")), " ")
}
func dicomInt(value string) int { n, _ := strconv.Atoi(cleanDICOMString(value)); return n }
func formatDICOMDate(value string) string {
	v := cleanDICOMString(value)
	if len(v) == 8 {
		return v[:4] + "-" + v[4:6] + "-" + v[6:]
	}
	return v
}

func writePart10(path string, ds *dataset.Dataset, sopClassUID, sopInstanceUID, transferSyntax, sourceAE string) (err error) {
	if transferSyntax != network.ImplicitVRLittleEndianUID && transferSyntax != network.ExplicitVRLittleEndianUID {
		return fmt.Errorf("transfer syntax no admitida para almacenamiento: %s", transferSyntax)
	}
	elements, err := writerElements(ds)
	if err != nil {
		return err
	}
	temporary := path + ".tmp"
	f, err := os.OpenFile(temporary, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		if err != nil {
			_ = os.Remove(temporary)
		}
	}()
	w := filewriter.NewDCMFileWriter(filebase.NewFileWriter(f))
	w.SetExplicitVR(true)
	w.SetLittleEndian(true)
	if err = w.WritePreamble(); err != nil {
		return err
	}
	if err = w.WriteDICMPrefix(); err != nil {
		return err
	}
	meta := part10Meta(sopClassUID, sopInstanceUID, transferSyntax, sourceAE)
	if err = w.WriteDataElements(meta); err != nil {
		return err
	}
	if transferSyntax == network.ImplicitVRLittleEndianUID {
		w.SetExplicitVR(false)
	}
	if err = w.WriteDataElements(elements); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	if err = os.Rename(temporary, path); err != nil {
		return err
	}
	return nil
}

func part10Meta(sopClassUID, sopInstanceUID, transferSyntax, sourceAE string) []*filewriter.DataElement {
	items := []*filewriter.DataElement{
		writerElement(tag.New(0x0002, 0x0001), "OB", []byte{0, 1}),
		writerElement(tag.New(0x0002, 0x0002), "UI", paddedValue("UI", []byte(sopClassUID))),
		writerElement(tag.New(0x0002, 0x0003), "UI", paddedValue("UI", []byte(sopInstanceUID))),
		writerElement(tag.New(0x0002, 0x0010), "UI", paddedValue("UI", []byte(transferSyntax))),
		writerElement(tag.New(0x0002, 0x0012), "UI", paddedValue("UI", []byte(implementationUID))),
		writerElement(tag.New(0x0002, 0x0013), "SH", paddedValue("SH", []byte("DICOMDISC_AP1"))),
		writerElement(tag.New(0x0002, 0x0016), "AE", paddedValue("AE", []byte(sourceAE))),
	}
	var length uint32
	for _, item := range items {
		length += encodedElementLength(item)
	}
	value := make([]byte, 4)
	binary.LittleEndian.PutUint32(value, length)
	return append([]*filewriter.DataElement{writerElement(tag.New(0x0002, 0x0000), "UL", value)}, items...)
}

func writerElements(ds *dataset.Dataset) ([]*filewriter.DataElement, error) {
	converted := filewriter.ElementsFromDataset(ds)
	out := make([]*filewriter.DataElement, 0, len(converted))
	for _, elem := range converted {
		if elem.Tag.Group() != 0x0002 {
			out = append(out, elem)
		}
	}
	return out, nil
}

func paddedValue(vr string, value []byte) []byte {
	if len(value)%2 == 0 {
		return value
	}
	pad := byte(' ')
	if vr == "UI" || vr == "OB" || vr == "OW" || vr == "UN" {
		pad = 0
	}
	return append(value, pad)
}

func writerElement(t tag.Tag, vr string, value []byte) *filewriter.DataElement {
	return &filewriter.DataElement{Tag: t, VR: vr, Value: value, Length: uint32(len(value))}
}

func encodedElementLength(elem *filewriter.DataElement) uint32 {
	header := uint32(8)
	switch elem.VR {
	case "OB", "OD", "OF", "OL", "OW", "SQ", "UC", "UN", "UR", "UT":
		header = 12
	}
	return header + elem.Length
}
