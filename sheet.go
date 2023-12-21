package coloringbook

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"os"

	"github.com/boombuler/barcode/qr"
	"github.com/go-pdf/fpdf"
	"github.com/go-pdf/fpdf/contrib/barcode"
	"github.com/nfnt/resize"
	"github.com/sfomuseum/go-sfomuseum-coloringbook/outline"
	"github.com/sfomuseum/go-sfomuseum-coloringbook/static"
)

type AddSheetOptions struct {
	Image           image.Image
	ImageReader     io.Reader
	ImagePath       string
	Title           string
	Date            string
	CreditLine      string
	AccessionNumber string
	URL             string
	Outline         *outline.OutlineOptions
}

func AddSheet(ctx context.Context, pdf *fpdf.Fpdf, opts *AddSheetOptions) error {

	letter_w := 8.5
	letter_h := 11.0

	logo_w := 1.0
	logo_h := 0.3

	dpi := 150.0

	margin_x := 0.5
	margin_y := 0.5

	max_w := letter_h - (margin_x * 2)
	max_h := letter_w - (margin_y * 3.75)

	footer_y := margin_y + max_h + 0.15

	qr_w := 0.4
	qr_h := 0.4
	qr_margin := 0.5

	line_h := 0.15

	if Orientation(opts.Image) == "P" {
		max_h = letter_h - (margin_y * 3.75)
		max_w = letter_w - (margin_x * 2)
		footer_y = margin_y + max_h + 0.1
	}

	dims := opts.Image.Bounds()
	im_w := float64(dims.Max.X) / dpi
	im_h := float64(dims.Max.Y) / dpi

	im_x := margin_x
	im_y := margin_y

	log.Printf("MAX w %02f h %02f\n", max_w, max_h)
	log.Printf("IMAGE w %02f h %02f\n", im_w, im_h)

	if im_h > max_h || im_w > max_w {

		new_w := uint(max_w * dpi)
		new_h := uint(max_h * dpi)

		new_im := resize.Thumbnail(new_w, new_h, opts.Image, resize.Lanczos3)
		bounds := new_im.Bounds()

		log.Printf("NEW DIMS w %d (%02f) %d (%02f)\n", bounds.Max.X, float64(bounds.Max.X)/dpi, bounds.Max.Y, float64(bounds.Max.Y)/dpi)

		resized_fh, err := os.CreateTemp("", "*.png")

		if err != nil {
			return fmt.Errorf("Failed to create resized temp file, %w", err)
		}

		resized_path := resized_fh.Name()
		defer os.Remove(resized_path)

		err = png.Encode(resized_fh, new_im)

		if err != nil {
			return fmt.Errorf("Failed to write resized image, %w", err)
		}

		err = resized_fh.Close()

		if err != nil {
			return fmt.Errorf("Failed to close resized image after writing, %w", err)
		}

		resized_r, err := os.Open(resized_path)

		if err != nil {
			return fmt.Errorf("Failed to open %s for reading, %w", resized_path, err)
		}

		defer resized_r.Close()

		opts.Image = new_im
		opts.ImagePath = resized_path
		opts.ImageReader = resized_r

		new_dims := new_im.Bounds()
		im_w = float64(new_dims.Max.X) / dpi
		im_h = float64(new_dims.Max.Y) / dpi
	}

	log.Printf("IMAGE w %02f h %02f\n", im_w, im_h)

	if im_w > max_w {
		return fmt.Errorf("Image width (%02f) is still greater than max width (%02f)", im_w, max_w)
	}

	if im_h > max_h {
		return fmt.Errorf("Image height (%02f) is still greater than max height (%02f)", im_h, max_h)
	}

	im_x = margin_x + ((max_w - im_w) / 2.0)
	im_y = margin_y + ((max_h - im_h) / 4.0)

	log.Printf("OFFSET w %02f h %02f\n", im_x, im_y)

	pdf.SetFont("Helvetica", "", 8)

	pdf.AddPage()

	im_opts := fpdf.ImageOptions{
		ImageType: "png",
		ReadDpi:   false,
	}

	info := pdf.RegisterImageOptionsReader(opts.ImagePath, im_opts, opts.ImageReader)
	info.SetDpi(dpi)

	pdf.ImageOptions(opts.ImagePath, im_x, im_y, im_w, im_h, false, im_opts, 0, "")

	// QR code

	pdf.SetY(footer_y)
	pdf.SetX(margin_x)

	key := barcode.RegisterQR(pdf, opts.URL, qr.H, qr.Unicode)
	barcode.Barcode(pdf, key, margin_x, footer_y, qr_h, qr_w, false)

	// Metadata

	pdf.SetY(footer_y - 0.04)
	pdf.SetX(margin_x + qr_margin)

	blurb := fmt.Sprintf("%s (%s)\n%s\n%s", opts.Title, opts.Date, opts.CreditLine, opts.AccessionNumber)

	cell_x := max_w - qr_margin - logo_w
	cell_y := line_h

	pdf.MultiCell(cell_x, cell_y, blurb, "", "", false)

	// Add SFO Museum logo

	logo_r, err := static.FS.Open("logo-150.png")

	if err != nil {
		return fmt.Errorf("Failed to open SFOM logo, %w", err)
	}

	defer logo_r.Close()

	logo_tmpfile, err := os.CreateTemp("", "*.png")

	if err != nil {
		return fmt.Errorf("Failed to create logo temp file, %w", err)
	}

	logo_path := logo_tmpfile.Name()
	defer os.Remove(logo_path)

	_, err = io.Copy(logo_tmpfile, logo_r)

	if err != nil {
		return fmt.Errorf("Failed to copy logo, %w", err)
	}

	err = logo_tmpfile.Close()

	if err != nil {
		return fmt.Errorf("Failed to close logo temp file, %w", err)
	}

	logo_opts := fpdf.ImageOptions{
		ImageType: "png",
		ReadDpi:   false,
	}

	logo_r, err = os.Open(logo_path)

	if err != nil {
		return fmt.Errorf("Failed to open logo PNG file %s, %w", logo_path, err)
	}

	logo_info := pdf.RegisterImageOptionsReader(logo_path, logo_opts, logo_r)
	logo_info.SetDpi(150)

	pdf.ImageOptions(logo_path, (margin_x+max_w)-logo_w, footer_y, logo_w, logo_h, false, logo_opts, 0, "")

	return nil
}
