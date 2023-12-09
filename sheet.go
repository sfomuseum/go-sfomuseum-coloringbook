package colouringbook

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/boombuler/barcode/qr"
	"github.com/go-pdf/fpdf"
	"github.com/go-pdf/fpdf/contrib/barcode"
	"github.com/sfomuseum/go-sfomuseum-colouringbook/static"
)

type AddSheetOptions struct {
	Image           string
	Title           string
	Date            string
	CreditLine      string
	AccessionNumber string
	URL             string
}

func AddSheet(ctx context.Context, pdf *fpdf.Fpdf, opts *AddSheetOptions) error {

	margin_x := 1.375
	margin_y := 0.75

	max_w := 8.25
	max_h := 6.375

	qr_w := 0.4
	qr_h := 0.4
	qr_margin := 0.5

	footer_y := 7.25 // derive from max_h + something
	line_h := 0.15

	logo_w := 1.29
	logo_h := 0.4

	// To do: things in portrait mode...

	pdf.SetFont("Helvetica", "", 8)

	pdf.AddPage()

	r, err := os.Open(opts.Image)

	if err != nil {
		return fmt.Errorf("Failed to open image for reading, %w", err)
	}

	defer r.Close()

	im_opts := fpdf.ImageOptions{
		ImageType: "png",
		ReadDpi:   false,
	}

	info := pdf.RegisterImageOptionsReader(opts.Image, im_opts, r)
	info.SetDpi(150)

	pdf.ImageOptions(opts.Image, margin_x, margin_y, max_w, max_h, false, im_opts, 0, "")

	// QR code

	pdf.SetY(footer_y)
	pdf.SetX(margin_x)

	key := barcode.RegisterQR(pdf, opts.URL, qr.H, qr.Unicode)
	barcode.Barcode(pdf, key, margin_x, footer_y, qr_h, qr_w, false)

	// Metadata

	pdf.SetY(footer_y - 0.04)
	pdf.SetX(margin_x + qr_margin)

	blurb := fmt.Sprintf("%s (%s)\n%s\n%s", opts.Title, opts.Date, opts.CreditLine, opts.AccessionNumber)

	pdf.MultiCell(max_w-qr_margin-logo_w, line_h, blurb, "", "", false)

	// Add SFO Museum logo

	logo_r, err := static.FS.Open("logo.png")

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
