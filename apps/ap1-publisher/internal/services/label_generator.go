package services

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"

	"github.com/local/dicom-disc-suite/shared/models"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const labelSize = 1700

// GenerateDiscLabel creates a simple 1700x1700 PNG, the native image size
// recommended by the TD Bridge Technical Reference Guide.
func GenerateDiscLabel(path string, study models.Study) (err error) {
	img := image.NewRGBA(image.Rect(0, 0, labelSize, labelSize))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	teal := color.RGBA{R: 10, G: 137, B: 126, A: 255}
	dark := color.RGBA{R: 20, G: 45, B: 62, A: 255}
	for y := 0; y < 120; y++ {
		draw.Draw(img, image.Rect(0, y, labelSize, y+1), &image.Uniform{C: teal}, image.Point{}, draw.Src)
	}
	lines := []string{
		"DICOM DISC",
		"Paciente: " + study.PatientName,
		"ID: " + study.PatientID,
		"Estudio: " + study.StudyDescription,
		"Modalidad: " + study.Modality,
		"Fecha: " + study.StudyDate,
	}
	// Draw to a smaller canvas and scale 3x so the dependency-free bitmap font
	// remains legible on the printer's recommended 1700px canvas.
	small := image.NewRGBA(image.Rect(0, 0, 567, 567))
	draw.Draw(small, small.Bounds(), image.Transparent, image.Point{}, draw.Src)
	d := &font.Drawer{Dst: small, Src: &image.Uniform{C: dark}, Face: basicfont.Face7x13}
	for i, line := range lines {
		d.Dot = fixed.P(55, 75+i*42)
		d.DrawString(cleanLabelText(line))
	}
	for y := 0; y < small.Bounds().Dy(); y++ {
		for x := 0; x < small.Bounds().Dx(); x++ {
			c := small.RGBAAt(x, y)
			if c.A == 0 {
				continue
			}
			for sy := 0; sy < 3; sy++ {
				for sx := 0; sx < 3; sx++ {
					img.SetRGBA(x*3+sx, y*3+sy, c)
				}
			}
		}
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := f.Close()
		if err == nil {
			err = closeErr
		}
	}()
	return png.Encode(f, img)
}

func cleanLabelText(s string) string {
	r := strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "Á", "A", "É", "E", "Í", "I", "Ó", "O", "Ú", "U", "ñ", "n", "Ñ", "N")
	return r.Replace(s)
}
