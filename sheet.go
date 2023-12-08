package colouringbook

import (
	"context"
	"fmt"
	"os"

	"github.com/boombuler/barcode/qr"
	"github.com/go-pdf/fpdf"
	"github.com/go-pdf/fpdf/contrib/barcode"
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

	pdf.ImageOptions(opts.Image, 1.375, 0.75, 8.25, 6.375, false, im_opts, 0, "")

	// QR code

	pdf.SetY(7.25 + 0.5)
	pdf.SetX(1.375 + 0.5)

	key := barcode.RegisterQR(pdf, opts.URL, qr.H, qr.Unicode)
	barcode.Barcode(pdf, key, 1.375, 7.25, .4, .4, false)

	// Metadata

	pdf.SetY(7.21)
	pdf.SetX(1.375 + 0.5)

	blurb := fmt.Sprintf("%s (%s)\n%s\n%s", opts.Title, opts.Date, opts.CreditLine, opts.AccessionNumber)

	pdf.MultiCell(6.375-0.5, .15, blurb, "", "", false)
	return nil
}
