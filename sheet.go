package colouringbook

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"

	"github.com/aaronland/go-image/resize"
	"github.com/boombuler/barcode/qr"
	"github.com/go-pdf/fpdf"
	"github.com/go-pdf/fpdf/contrib/barcode"
	"github.com/sfomuseum/go-sfomuseum-colouringbook/static"
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
}

func AddSheet(ctx context.Context, pdf *fpdf.Fpdf, opts *AddSheetOptions) error {

	dpi := 150.0

	margin_x := 1.375
	margin_y := 0.75

	max_w := 11.0 - (margin_x * 2) // 8.25
	max_h := 8.5 - (margin_y * 2)  // 6.375

	qr_w := 0.4
	qr_h := 0.4
	qr_margin := 0.5

	footer_y := max_h + 0.25 // 7.25 // derive from max_h + something
	line_h := 0.15

	logo_w := 1.29
	logo_h := 0.4

	if Orientation(opts.Image) == "P" {

		max_h = 10.25 - (margin_y * 2) // 8.25
		max_w = 8.5 - (margin_x * 2)   // 6.375

		footer_y = max_h + 0.9
	}

	dims := opts.Image.Bounds()
	im_w := float64(dims.Max.X) / dpi
	im_h := float64(dims.Max.Y) / dpi

	im_x := margin_x
	im_y := margin_y * 0.75

	// Scale image if necessary

	if im_h > max_h || im_w > max_w {

		max_dim := max(max_w, max_h)

		new_im, err := resize.ResizeImage(ctx, opts.Image, int(max_dim*dpi))

		if err != nil {
			return fmt.Errorf("Failed to resize image, %w", err)
		}

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

	if im_h > im_w {
		im_x = margin_x + ((max_w - im_w) / 2.0)
	}

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
