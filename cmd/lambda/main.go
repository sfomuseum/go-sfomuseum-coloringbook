package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"

	"github.com/aaronland/go-aws-lambda"
	"github.com/sfomuseum/go-sfomuseum-colouringbook"
)

func main() {

	var function_uri string
	var object_id int64

	flag.StringVar(&function_uri, "function-uri", "aws://GenerateColouringBook?region=us-west-2&credentials=session", "")
	flag.Int64Var(&object_id, "object-id", 0, "")

	flag.Parse()

	ctx := context.Background()

	f, err := lambda.NewLambdaFunction(ctx, function_uri)

	if err != nil {
		log.Fatalf("Failed to create new Lambda function, %v", err)
	}

	req := colouringbook.ColouringBookRequest{
		ObjectId: object_id,
	}

	payload, err := json.Marshal(req)

	if err != nil {
		log.Fatalf("Failed to marshal request, %v", err)
	}

	_, err = f.InvokeWithJSON(ctx, payload)

	if err != nil {
		log.Fatalf("Failed to invoke function, %v", err)
	}
}
