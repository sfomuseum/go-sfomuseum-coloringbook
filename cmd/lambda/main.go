package main

import (
	"context"
	"flag"
	"log"

	"github.com/sfomuseum/go-sfomuseum-coloringbook"
)

func main() {

	var function_uri string
	var object_id int64

	flag.StringVar(&function_uri, "function-uri", coloringbook.GENERATE_COLORING_BOOK_LAMBDA_URI, "")
	flag.Int64Var(&object_id, "object-id", 0, "")

	flag.Parse()

	ctx := context.Background()
	err := coloringbook.GenerateColoringBookLambda(ctx, function_uri, object_id)

	if err != nil {
		log.Fatalf("Failed to invoke Lambda function for %v", err)
	}

}
