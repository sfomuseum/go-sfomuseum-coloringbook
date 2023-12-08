package main

import (
	"context"
	"flag"
	"log"

	"github.com/go-pdf/fpdf"
	"github.com/sfomuseum/go-sfomuseum-colouringbook"
)

func main() {

	var image string

	flag.StringVar(&image, "image", "", "...")

	flag.Parse()

	ctx := context.Background()

	pdf := fpdf.New("L", "in", "Letter", "")

	//

	url := "https://collection.sfomuseum.org/objects/1511943901/"
	title := "negative: San Francisco Airport, Terminal Building construction"
	date := "1952"
	creditline := "Transfer from San Francisco International Airport"
	accnum := "2011.032.0303"

	sheet_opts := &colouringbook.AddSheetOptions{
		Image:           image,
		Title:           title,
		Date:            date,
		CreditLine:      creditline,
		AccessionNumber: accnum,
		URL:             url,
	}

	err := colouringbook.AddSheet(ctx, pdf, sheet_opts)

	if err != nil {
		log.Fatalf("Failed to add sheet, %v", err)
	}

	// PDF files

	err = pdf.OutputFileAndClose("test.pdf")

	if err != nil {
		panic(err)
	}
}
