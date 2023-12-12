package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"

	"github.com/sfomuseum/go-sfomuseum-colouringbook"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-uri"
)

func main() {

	var function_uri string
	var object_id int64
	var iterator_uri string

	flag.StringVar(&function_uri, "function-uri", colouringbook.GENERATE_COLOURING_BOOK_LAMBDA_URI, "")
	flag.Int64Var(&object_id, "object-id", 0, "")
	flag.StringVar(&iterator_uri, "iterator-uri", "", "")

	flag.Parse()

	iterator_sources := flag.Args()

	ctx := context.Background()

	iter_cb := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) error {

		object_id, uri_args, err := uri.ParseURI(path)

		if err != nil {
			return fmt.Errorf("Failed to parse URI for %s, %w", path, err)
		}

		if uri_args.IsAlternate {
			return nil
		}

		err = colouringbook.GenerateColouringBookLambda(ctx, function_uri, object_id)

		if err != nil {
			return fmt.Errorf("Failed to invoke Lambda function for %s, %w", path, err)
		}

		return nil
	}

	iter, err := iterator.NewIterator(ctx, iterator_uri, iter_cb)

	if err != nil {
		log.Fatalf("Failed to create new iterator, %v", err)
	}

	err = iter.IterateURIs(ctx, iterator_sources...)

	if err != nil {
		log.Fatalf("Failed to iterate URIs, %v", err)
	}

}