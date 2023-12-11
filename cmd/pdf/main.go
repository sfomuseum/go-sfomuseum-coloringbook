package main

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"

	aa_bucket "github.com/aaronland/gocloud-blob/bucket"
	"github.com/go-pdf/fpdf"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-sfomuseum-colouringbook"
	sfom_writer "github.com/sfomuseum/go-sfomuseum-writer/v3"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-reader"
	_ "github.com/whosonfirst/go-reader-http"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader"
	"github.com/whosonfirst/go-writer/v3"
	_ "gocloud.dev/blob/fileblob"
)

func main() {

	var object_image string
	var object_id int64
	var reader_uri string
	var writer_uri string
	var bucket_uri string
	var filename string
	var update_object bool

	var mode string

	fs := flagset.NewFlagSet("colouringbook")

	fs.StringVar(&object_image, "object-image", "", "...")
	fs.Int64Var(&object_id, "object-id", 0, "...")
	fs.StringVar(&reader_uri, "reader-uri", "https://static.sfomuseum.org/data/", "...")
	fs.StringVar(&bucket_uri, "bucket-uri", "cwd://", "...")
	fs.StringVar(&filename, "filename", "", "...")
	fs.StringVar(&writer_uri, "writer-uri", "stdout://", "...")
	fs.BoolVar(&update_object, "update-object", false, "...")
	fs.StringVar(&mode, "mode", "cli", "...")

	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVars(fs, "SFOMUSEUM")

	if err != nil {
		log.Fatalf("Failed to set flags from environment variables, %v", err)
	}

	ctx := context.Background()

	r, err := reader.NewReader(ctx, reader_uri)

	if err != nil {
		log.Fatalf("Failed to create reader, %v", err)
	}

	var wr writer.Writer

	if update_object {

		wr, err = writer.NewWriter(ctx, writer_uri)

		if err != nil {
			log.Fatalf("Failed to create new writer, %v", err)
		}
	}

	// Set up bucket

	if bucket_uri == "cwd://" {

		cwd, err := os.Getwd()

		if err != nil {
			log.Fatalf("Failed to derive current working directory, %v", err)
		}

		bucket_uri = fmt.Sprintf("file://%s", cwd)
	}

	bucket, err := aa_bucket.OpenBucket(ctx, bucket_uri)

	if err != nil {
		log.Fatalf("Failed to open bucket, %v", err)
	}

	defer bucket.Close()

	run := func(ctx context.Context, object_id int64) error {

		// Get object metadata

		body, err := wof_reader.LoadBytes(ctx, r, object_id)

		if err != nil {
			return fmt.Errorf("Failed to load feature for object, %v", err)
		}

		title_rsp := gjson.GetBytes(body, "properties.wof:name")
		date_rsp := gjson.GetBytes(body, "properties.sfomuseum:date")
		creditline_rsp := gjson.GetBytes(body, "properties.sfomuseum:creditline")
		accession_number_rsp := gjson.GetBytes(body, "properties.sfomuseum:accession_number")

		url := fmt.Sprintf("https://collection.sfomuseum.org/objects/%d/", object_id)

		// Derive contoured image if necessary

		if object_image == "" {

			primary_rsp := gjson.GetBytes(body, "properties.millsfield:primary_image")

			if !primary_rsp.Exists() {
				return fmt.Errorf("Object is missing primary image property")
			}

			image_id := primary_rsp.Int()

			derived_image, err := colouringbook.DeriveObjectImage(ctx, r, image_id)

			if err != nil {
				return fmt.Errorf("Failed to derive object image, %v", err)
			}

			defer os.Remove(derived_image)

			object_image = derived_image
		}

		im_r, err := os.Open(object_image)

		if err != nil {
			return fmt.Errorf("Failed to open %s for reading, %w", object_image, err)
		}

		defer im_r.Close()

		im, _, err := image.Decode(im_r)

		if err != nil {
			return fmt.Errorf("Failed to decode image %s, %w", object_image, err)
		}

		orientation := colouringbook.Orientation(im)

		log.Println("O", orientation)

		im_r.Seek(0, 0)

		// Create PDF

		pdf := fpdf.New(orientation, "in", "Letter", "")

		// Add sheet to colouring book

		sheet_opts := &colouringbook.AddSheetOptions{
			Image:           im,
			ImagePath:       object_image,
			ImageReader:     im_r,
			URL:             url,
			Title:           title_rsp.String(),
			Date:            date_rsp.String(),
			CreditLine:      creditline_rsp.String(),
			AccessionNumber: accession_number_rsp.String(),
		}

		err = colouringbook.AddSheet(ctx, pdf, sheet_opts)

		if err != nil {
			return fmt.Errorf("Failed to add sheet, %v", err)
		}

		// Publish PDF file

		if filename == "" {
			filename = fmt.Sprintf("%d-coloringbook.pdf", object_id)
		}

		pdf_wr, err := bucket.NewWriter(ctx, filename, nil)

		if err != nil {
			return fmt.Errorf("Failed to create new writer for %s, %v", filename, err)
		}

		err = pdf.OutputAndClose(pdf_wr)

		if err != nil {
			return fmt.Errorf("Failed to write %s, %v", filename, err)
		}

		log.Printf("Wrote %s\n", filename)

		// Update object record

		if update_object {

			updates := map[string]interface{}{
				"properties.millsfield:has_coloring_book": 1,
			}

			has_updates, new_body, err := export.AssignPropertiesIfChanged(ctx, body, updates)

			if err != nil {
				return fmt.Errorf("Failed to assign updates to object record, %v", err)
			}

			if has_updates {

				_, err := sfom_writer.WriteBytes(ctx, wr, new_body)

				if err != nil {
					return fmt.Errorf("Failed to update object record, %v", err)
				}

				err = wr.Close(ctx)

				if err != nil {
					return fmt.Errorf("Failed to close object update writer, %v", err)
				}
			}
		}

		return nil
	}

	switch mode {
	case "cli":
		run(ctx, object_id)
	case "lambda":
		log.Fatalf("Not implemented")
	default:
		log.Fatalf("Invalid mode")
	}
}
