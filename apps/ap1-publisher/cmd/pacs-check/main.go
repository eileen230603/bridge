package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/amrshadid/go-dicom/filebase"
	"github.com/amrshadid/go-dicom/filereader"
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/adapters"
	"github.com/local/dicom-disc-suite/apps/ap1-publisher/internal/config"
)

func main() {
	configPath := flag.String("config", "apps/ap1-publisher/config.json", "ruta de config.json")
	fromValue := flag.String("from", "2004-01-19", "fecha inicial YYYY-MM-DD")
	toValue := flag.String("to", "2004-01-19", "fecha final YYYY-MM-DD")
	echoOnly := flag.Bool("echo-only", false, "ejecutar únicamente C-ECHO")
	retrieve := flag.Bool("retrieve", false, "recuperar mediante C-MOVE el único estudio encontrado")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fatal(err)
	}
	from, err := time.Parse("2006-01-02", *fromValue)
	if err != nil {
		fatal(err)
	}
	to, err := time.Parse("2006-01-02", *toValue)
	if err != nil {
		fatal(err)
	}
	repo := adapters.NewPacsStudyRepository(cfg.PACS, slog.Default())
	ctx := context.Background()
	if err := repo.Echo(ctx); err != nil {
		fatal(fmt.Errorf("C-ECHO: %w", err))
	}
	fmt.Println("C-ECHO: Success")
	if *echoOnly {
		return
	}
	studies, err := repo.SearchStudies(ctx, from, to)
	if err != nil {
		fatal(fmt.Errorf("C-FIND: %w", err))
	}
	fmt.Printf("C-FIND: Success (%d estudios)\n", len(studies))
	for _, study := range studies {
		fmt.Printf("PatientName=%s\nPatientID=%s\nStudyInstanceUID=%s\nStudyDate=%s\nModality=%s\nStudyDescription=%s\nInstances=%d\n\n", study.PatientName, study.PatientID, study.StudyInstanceUID, study.StudyDate, study.Modality, study.StudyDescription, study.InstanceCount)
	}
	if !*retrieve {
		return
	}
	if len(studies) != 1 {
		fatal(fmt.Errorf("C-MOVE requiere exactamente un estudio; C-FIND devolvió %d", len(studies)))
	}
	studyUID := studies[0].StudyInstanceUID
	destination := filepath.Join(cfg.TemporaryDirectory, studyUID, "data")
	if err := repo.Start(ctx); err != nil {
		fatal(fmt.Errorf("Storage SCP: %w", err))
	}
	defer func() { _ = repo.Close() }()
	if err := repo.RetrieveStudy(ctx, studyUID, destination); err != nil {
		fatal(fmt.Errorf("C-MOVE: %w", err))
	}
	files, err := filepath.Glob(filepath.Join(destination, "*.dcm"))
	if err != nil {
		fatal(err)
	}
	for _, path := range files {
		f, err := os.Open(path)
		if err != nil {
			fatal(err)
		}
		dicomFile, readErr := filereader.ReadDICOMFile(filebase.NewFileReader(f))
		_ = f.Close()
		if readErr != nil {
			fatal(fmt.Errorf("C-STORE produjo un archivo Part-10 inválido %s: %w", path, readErr))
		}
		ds := dicomFile.GetDataset()
		if ds.GetStringByKeyword("StudyInstanceUID") != studyUID {
			fatal(fmt.Errorf("StudyInstanceUID recibido no coincide en %s", path))
		}
		if ds.GetStringByKeyword("SOPInstanceUID") == "" {
			fatal(fmt.Errorf("SOPInstanceUID vacío en %s", path))
		}
	}
	fmt.Printf("C-MOVE: Success\nC-STORE: Success (%d DICOM)\nDestination=%s\n", len(files), destination)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
