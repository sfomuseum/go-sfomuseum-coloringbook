package colouringbook

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
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

	letter_w := 8.5
	letter_h := 11.0

	dpi := 150.0

	margin_x := 0.5
	margin_y := 0.5

	max_w := letter_h - (margin_x * 2)
	max_h := letter_w - (margin_y * 2.6)

	footer_y := 8.5 - (margin_y * 2.5) // max_h + 0.1

	log.Println("WUT", max_h)

	qr_w := 0.4
	qr_h := 0.4
	qr_margin := 0.5

	line_h := 0.15

	logo_w := 1.29
	logo_h := 0.4

	if Orientation(opts.Image) == "P" {
		max_h = letter_h - (margin_y * 2.6)
		max_w = letter_w - (margin_x * 2)
		footer_y = 11.0 - (margin_y * 2.5) // max_h + 0.9
	}

	dims := opts.Image.Bounds()
	im_w := float64(dims.Max.X) / dpi
	im_h := float64(dims.Max.Y) / dpi

	im_x := margin_x
	im_y := margin_y

	// Scale image if necessary

	resize_image := false
	max_dim := 0.0

	// To do: Actually determine scale factor/size here...

	if im_h > max_h && im_w > max_w {
		resize_image = true
		max_dim = min(max_w, max_h)

		log.Println("RESIZE A", max_dim)
	} else if im_h > max_h {
		resize_image = true
		max_dim = max_h

		log.Println("RESIZE B", max_dim)
	} else if im_w > max_w {
		resize_image = true
		max_dim = max_w

		log.Println("RESIZE C", max_dim)
	} else {
	}

	if resize_image {

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

		log.Println("FFFFUUUUU", im_w, im_h)
	}

	// if im_h > im_w && im_w < max_w {
	im_x = margin_x + ((max_w - im_w) / 2.0)
	log.Printf("OMG %02f = %02f + ((%02f - %02f) / 2.0)\n", im_x, margin_x, max_w, im_w)
	// } else {
	im_y = margin_y + ((max_h - im_h) / 4.0)
	log.Printf("OMG %02f = %02f + ((%02f - %02f) / 2.0)\n", im_y, margin_y, max_h, im_h)
	// }

	log.Printf("BBQ %02f, %02f @ %02f, %02f\n", im_x, im_y, im_w, im_h)

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
