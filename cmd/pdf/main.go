package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/go-pdf/fpdf"
	"github.com/sfomuseum/go-sfomuseum-colouringbook"
	"github.com/whosonfirst/go-reader"
	_ "github.com/whosonfirst/go-reader-http"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader"
	"github.com/tidwall/gjson"
)

func main() {

	var image string
	var object_id int64
	var reader_uri string
	
	flag.StringVar(&image, "image", "", "...")
	flag.Int64Var(&object_id, "object-id", 0, "...")	
	flag.StringVar(&reader_uri, "reader-uri", "https://static.sfomuseum.org/data/", "...")
	
	flag.Parse()

	ctx := context.Background()

	r, err := reader.NewReader(ctx, reader_uri)

	if err != nil {
		log.Fatalf("Failed to create reader, %v", err)
	}
		
	// Create PDF
	
	pdf := fpdf.New("L", "in", "Letter", "")

	// Something something something get all the images for object here...
	
	// Get object metadata

	body, err := wof_reader.LoadBytes(ctx, r, object_id)

	if err != nil {
		log.Fatalf("Failed to load feature for object, %v", err)
	}

	title_rsp := gjson.GetBytes(body, "properties.wof:name")
	date_rsp := gjson.GetBytes(body, "properties.sfomuseum:date")
	creditline_rsp := gjson.GetBytes(body, "properties.sfomuseum:creditline")
	accession_number_rsp := gjson.GetBytes(body, "properties.sfomuseum:accession_number")	

	url := fmt.Sprintf("https://collection.sfomuseum.org/objects/%d/", object_id)
	
	sheet_opts := &colouringbook.AddSheetOptions{
		Image:           image,
		URL:             url,		
		Title:           title_rsp.String(),
		Date:            date_rsp.String(),
		CreditLine:      creditline_rsp.String(),
		AccessionNumber: accession_number_rsp.String(),
	}

	err = colouringbook.AddSheet(ctx, pdf, sheet_opts)

	if err != nil {
		log.Fatalf("Failed to add sheet, %v", err)
	}

	// PDF files

	err = pdf.OutputFileAndClose("test.pdf")

	if err != nil {
		panic(err)
	}
}
