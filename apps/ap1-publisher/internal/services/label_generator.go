package services

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"os"
	"strings"

	"github.com/local/dicom-disc-suite/shared/models"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const labelSize = 1700

type DiscLabelBranding struct {
	HospitalName  string
	LogoPath      string
	SymphonyLogo  []byte
	MediGlobeLogo []byte
}

func GenerateDiscLabel(path string, study models.Study) (err error) {
	return GenerateDiscLabelWithBranding(path, study, DiscLabelBranding{})
}

func GenerateDiscLabelWithBranding(path string, study models.Study, branding DiscLabelBranding) error {
	raw, err := GenerateDiscLabelPNG(study, branding)
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func GenerateDiscLabelPNG(study models.Study, branding DiscLabelBranding) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, labelSize, labelSize))
	teal := color.RGBA{R: 10, G: 137, B: 126, A: 255}
	dark := color.RGBA{R: 20, G: 45, B: 62, A: 255}
	muted := color.RGBA{R: 92, G: 112, B: 124, A: 255}
	drawDiscSurface(img, image.Pt(labelSize/2, labelSize/2), 810, 170)

	parsedFont, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}
	titleFace, err := labelFace(parsedFont, 44)
	if err != nil {
		return nil, err
	}
	valueFace, err := labelFace(parsedFont, 40)
	if err != nil {
		return nil, err
	}
	captionFace, err := labelFace(parsedFont, 27)
	if err != nil {
		return nil, err
	}

	hospitalName := strings.TrimSpace(branding.HospitalName)
	if hospitalName == "" {
		hospitalName = "NOMBRE DEL HOSPITAL"
	}
	var symphonyImage image.Image
	if len(branding.SymphonyLogo) > 0 {
		symphonyImage, _, _ = image.Decode(bytes.NewReader(branding.SymphonyLogo))
	}
	var mediGlobeImage image.Image
	if len(branding.MediGlobeLogo) > 0 {
		mediGlobeImage, _, _ = image.Decode(bytes.NewReader(branding.MediGlobeLogo))
	}
	var primaryLogo image.Image
	if strings.TrimSpace(branding.LogoPath) != "" {
		logoFile, err := os.Open(branding.LogoPath)
		if err != nil {
			return nil, err
		}
		logo, _, decodeErr := image.Decode(logoFile)
		_ = logoFile.Close()
		if decodeErr != nil {
			return nil, decodeErr
		}
		primaryLogo = logo
	}
	if primaryLogo != nil {
		drawLogo(img, primaryLogo, image.Rect(650, 105, 1050, 315))
	} else {
		drawLogoPlaceholder(img, captionFace, teal, image.Rect(650, 105, 1050, 315))
	}
	drawCenteredText(img, titleFace, dark, hospitalName, 850, 390, 850)

	drawRule(img, 180, 560, 585, teal)
	drawRule(img, 1115, 560, 1520, teal)

	drawTextBlock(img, captionFace, valueFace, muted, dark, "PACIENTE", study.PatientName, 205, 665, 365)
	drawTextBlock(img, captionFace, valueFace, muted, dark, "ID", study.PatientID, 205, 855, 350)
	drawTextBlock(img, captionFace, valueFace, muted, dark, "ESTUDIO", study.StudyDescription, 1135, 665, 360)
	drawTextBlock(img, captionFace, valueFace, muted, dark, "MODALIDAD", study.Modality, 1165, 875, 315)
	drawTextBlock(img, captionFace, valueFace, muted, dark, "FECHA", study.StudyDate, 1060, 1110, 430)

	if symphonyImage != nil {
		drawLogo(img, symphonyImage, image.Rect(300, 1180, 820, 1450))
	}
	draw.Draw(img, image.Rect(848, 1210, 852, 1470), &image.Uniform{C: muted}, image.Point{}, draw.Over)
	if mediGlobeImage != nil {
		drawLogo(img, mediGlobeImage, image.Rect(880, 1180, 1400, 1450))
	}

	var output bytes.Buffer
	if err := png.Encode(&output, img); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func labelFace(parsed *opentype.Font, size float64) (font.Face, error) {
	return opentype.NewFace(parsed, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
}

func drawDiscSurface(img *image.RGBA, center image.Point, outerRadius, holeRadius int) {
	disc := color.RGBA{R: 249, G: 251, B: 251, A: 255}
	edge := color.RGBA{R: 10, G: 137, B: 126, A: 255}
	hubEdge := color.RGBA{R: 188, G: 205, B: 211, A: 255}
	for y := center.Y - outerRadius; y <= center.Y+outerRadius; y++ {
		for x := center.X - outerRadius; x <= center.X+outerRadius; x++ {
			dx, dy := x-center.X, y-center.Y
			distance := dx*dx + dy*dy
			switch {
			case distance > outerRadius*outerRadius:
				continue
			case distance >= (outerRadius-7)*(outerRadius-7):
				img.SetRGBA(x, y, edge)
			case distance <= holeRadius*holeRadius:
				if distance >= (holeRadius-7)*(holeRadius-7) {
					img.SetRGBA(x, y, hubEdge)
				}
			default:
				img.SetRGBA(x, y, disc)
			}
		}
	}
}

func drawRule(img draw.Image, startX, y, endX int, lineColor color.Color) {
	draw.Draw(img, image.Rect(startX, y, endX, y+4), &image.Uniform{C: lineColor}, image.Point{}, draw.Over)
}

func drawTextBlock(img draw.Image, captionFace, valueFace font.Face, captionColor, valueColor color.Color, caption, value string, x, y, maxWidth int) {
	drawLabelText(img, captionFace, captionColor, caption, x, y, maxWidth)
	drawLabelText(img, valueFace, valueColor, value, x, y+55, maxWidth)
}

func drawCenteredText(img draw.Image, face font.Face, textColor color.Color, text string, centerX, baselineY, maxWidth int) {
	text = truncateLabelText(face, cleanLabelText(text), maxWidth)
	width := font.MeasureString(face, text).Ceil()
	drawLabelText(img, face, textColor, text, centerX-width/2, baselineY, maxWidth)
}

func drawLabelText(img draw.Image, face font.Face, textColor color.Color, text string, x, baselineY, maxWidth int) {
	text = truncateLabelText(face, cleanLabelText(strings.TrimSpace(text)), maxWidth)
	if text == "" {
		text = "—"
	}
	drawer := &font.Drawer{Dst: img, Src: &image.Uniform{C: textColor}, Face: face, Dot: fixed.P(x, baselineY)}
	drawer.DrawString(text)
}

func truncateLabelText(face font.Face, text string, maxWidth int) string {
	if font.MeasureString(face, text).Ceil() <= maxWidth {
		return text
	}
	runes := []rune(text)
	for len(runes) > 1 && font.MeasureString(face, string(runes)+"…").Ceil() > maxWidth {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

func drawLogoPlaceholder(destination draw.Image, face font.Face, placeholderColor color.Color, bounds image.Rectangle) {
	const border = 2
	draw.Draw(destination, image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Min.Y+border), &image.Uniform{C: placeholderColor}, image.Point{}, draw.Over)
	draw.Draw(destination, image.Rect(bounds.Min.X, bounds.Max.Y-border, bounds.Max.X, bounds.Max.Y), &image.Uniform{C: placeholderColor}, image.Point{}, draw.Over)
	draw.Draw(destination, image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Min.X+border, bounds.Max.Y), &image.Uniform{C: placeholderColor}, image.Point{}, draw.Over)
	draw.Draw(destination, image.Rect(bounds.Max.X-border, bounds.Min.Y, bounds.Max.X, bounds.Max.Y), &image.Uniform{C: placeholderColor}, image.Point{}, draw.Over)
	drawCenteredText(destination, face, placeholderColor, "INSERTE IMAGEN", (bounds.Min.X+bounds.Max.X)/2, (bounds.Min.Y+bounds.Max.Y)/2+8, bounds.Dx()-30)
}

func drawLogo(destination draw.Image, source image.Image, bounds image.Rectangle) {
	sourceBounds := source.Bounds()
	scale := min(float64(bounds.Dx())/float64(sourceBounds.Dx()), float64(bounds.Dy())/float64(sourceBounds.Dy()))
	width := int(float64(sourceBounds.Dx()) * scale)
	height := int(float64(sourceBounds.Dy()) * scale)
	target := image.Rect(
		bounds.Min.X+(bounds.Dx()-width)/2,
		bounds.Min.Y+(bounds.Dy()-height)/2,
		bounds.Min.X+(bounds.Dx()+width)/2,
		bounds.Min.Y+(bounds.Dy()+height)/2,
	)
	xdraw.CatmullRom.Scale(destination, target, source, sourceBounds, draw.Over, nil)
}

func cleanLabelText(s string) string {
	r := strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "Á", "A", "É", "E", "Í", "I", "Ó", "O", "Ú", "U", "ñ", "n", "Ñ", "N")
	return r.Replace(s)
}
