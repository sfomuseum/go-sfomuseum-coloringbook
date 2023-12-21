package main

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aaronland/gocloud-blob-s3"
	aa_bucket "github.com/aaronland/gocloud-blob/bucket"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-pdf/fpdf"
	"github.com/nfnt/resize"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-sfomuseum-coloringbook"
	"github.com/sfomuseum/go-coloringbook/outline"
	sfom_writer "github.com/sfomuseum/go-sfomuseum-writer/v3"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-reader"
	_ "github.com/whosonfirst/go-reader-http"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader"
	"github.com/whosonfirst/go-whosonfirst-uri"
	gh_writer "github.com/whosonfirst/go-writer-github/v3"
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
	var append_tree bool
	var access_token_uri string

	var contour_iterations int
	var contour_scale float64
	var contour_format string

	var vtracer_precision int
	var vtracer_speckle int

	var use_batik bool
	var path_batik string

	var mode string

	fs := flagset.NewFlagSet("coloringbook")

	fs.StringVar(&object_image, "object-image", "", "...")
	fs.Int64Var(&object_id, "object-id", 0, "...")
	fs.StringVar(&reader_uri, "reader-uri", "https://static.sfomuseum.org/data/", "...")
	fs.StringVar(&bucket_uri, "bucket-uri", "cwd://", "...")
	fs.StringVar(&filename, "filename", "", "...")
	fs.StringVar(&writer_uri, "writer-uri", "stdout://", "...")
	fs.BoolVar(&update_object, "update-object", false, "...")
	fs.StringVar(&mode, "mode", "cli", "...")
	fs.BoolVar(&append_tree, "append-tree", false, "...")
	fs.StringVar(&access_token_uri, "access-token-uri", "", "...")

	fs.IntVar(&contour_iterations, "contour-iteration", 8, "...")
	fs.Float64Var(&contour_scale, "contour-scale", 1.0, "...")
	fs.StringVar(&contour_format, "contour-format", "png", "...")

	fs.IntVar(&vtracer_precision, "vtracer-precision", 6, "...")
	fs.IntVar(&vtracer_speckle, "vtracer-speckle", 8, "...")

	fs.BoolVar(&use_batik, "use-batik", true, "...")
	fs.StringVar(&path_batik, "path-batik", "/usr/local/src/batik-1.17/batik-rasterizer-1.17.jar", "...")

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

		primary_rsp := gjson.GetBytes(body, "properties.millsfield:primary_image")

		if !primary_rsp.Exists() {
			return fmt.Errorf("Object is missing primary image property")
		}

		image_id := primary_rsp.Int()

		// Derive contoured image if necessary

		if object_image == "" {

			contour_opts := &outline.ContourOptions{
				Iterations: contour_iterations,
				Scale:      contour_scale,
				Format:     contour_format,
			}

			trace_opts := &outline.TraceOptions{
				Precision: vtracer_precision,
				Speckle:   vtracer_speckle,
			}

			raster_opts := &outline.RasterizeOptions{
				UseBatik:  use_batik,
				PathBatik: path_batik,
			}

			outline_opts := &outline.OutlineOptions{
				Contour:   contour_opts,
				Trace:     trace_opts,
				Rasterize: raster_opts,
			}

			derive_opts := &coloringbook.DeriveObjectImageOptions{
				Reader:  r,
				Outline: outline_opts,
			}

			derived_image, err := coloringbook.DeriveObjectImage(ctx, derive_opts, image_id)

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

		orientation := coloringbook.Orientation(im)

		im_r.Seek(0, 0)

		// Create PDF

		pdf := fpdf.New(orientation, "in", "Letter", "")

		// Add sheet to colouring book

		sheet_opts := &coloringbook.AddSheetOptions{
			Image:           im,
			ImagePath:       object_image,
			ImageReader:     im_r,
			URL:             url,
			Title:           title_rsp.String(),
			Date:            date_rsp.String(),
			CreditLine:      creditline_rsp.String(),
			AccessionNumber: accession_number_rsp.String(),
		}

		err = coloringbook.AddSheet(ctx, pdf, sheet_opts)

		if err != nil {
			return fmt.Errorf("Failed to add sheet, %v", err)
		}

		// Publish PDF file

		if filename == "" {
			filename = fmt.Sprintf("%d-%d-coloringbook.pdf", object_id, image_id)
		}

		if append_tree {

			tree, err := uri.Id2Path(object_id)

			if err != nil {
				return fmt.Errorf("Failed to derive tree for object id %d, %w", object_id, err)
			}

			filename = filepath.Join(tree, filename)
		}

		pdf_wr, err := s3blob.NewWriterWithACL(ctx, bucket, filename, "public-read")

		if err != nil {
			return fmt.Errorf("Failed to create new writer for %s, %v", filename, err)
		}

		err = pdf.OutputAndClose(pdf_wr)

		if err != nil {
			return fmt.Errorf("Failed to write %s, %v", filename, err)
		}

		log.Printf("Wrote %s\n", filename)

		// Publish thumb

		thumb_im := resize.Thumbnail(600, 600, im, resize.Lanczos3)

		thumb_filename := strings.Replace(filename, ".pdf", ".png", 1)

		thumb_wr, err := s3blob.NewWriterWithACL(ctx, bucket, thumb_filename, "public-read")

		if err != nil {
			return fmt.Errorf("Failed to create new writer for %s, %v", thumb_filename, err)
		}

		err = png.Encode(thumb_wr, thumb_im)

		if err != nil {
			return fmt.Errorf("Failed to encode %s, %w", thumb_filename, err)
		}

		err = thumb_wr.Close()

		if err != nil {
			return fmt.Errorf("Failed to close %s, %w", thumb_filename, err)
		}

		log.Printf("Wrote %s\n", thumb_filename)

		// Update object record

		var new_body []byte

		if update_object {

			rsp := gjson.GetBytes(body, "properties.millsfield:has_coloring_book")

			if rsp.Exists() && rsp.String() == "1" {
				update_object = false
			}
		}

		if update_object {

			updates := map[string]interface{}{
				"properties.millsfield:has_coloring_book": 1,
			}

			has_updates, _body, err := export.AssignPropertiesIfChanged(ctx, body, updates)

			if err != nil {
				return fmt.Errorf("Failed to assign updates to object record, %v", err)
			}

			if !has_updates {
				update_object = false
			}

			new_body = _body
		}

		if update_object {

			if access_token_uri != "" {

				writer_uri, err = gh_writer.EnsureGitHubAccessToken(ctx, writer_uri, access_token_uri)

				if err != nil {
					log.Fatalf("Failed to ensure access token, %v", err)
				}

			}

			wr, err := writer.NewWriter(ctx, writer_uri)

			if err != nil {
				log.Fatalf("Failed to create new writer, %v", err)
			}

			_, err = sfom_writer.WriteBytes(ctx, wr, new_body)

			if err != nil {
				return fmt.Errorf("Failed to update object record, %v", err)
			}

			err = wr.Close(ctx)

			if err != nil {
				return fmt.Errorf("Failed to close object update writer, %v", err)
			}
		}

		return nil
	}

	// Finally, run some code

	switch mode {
	case "cli":

		err := run(ctx, object_id)

		if err != nil {
			log.Fatal(err)
		}

	case "lambda":

		handler := func(ctx context.Context, req *coloringbook.ColoringBookRequest) error {
			return run(ctx, req.ObjectId)
		}

		lambda.Start(handler)

	default:
		log.Fatalf("Invalid mode")
	}
}
