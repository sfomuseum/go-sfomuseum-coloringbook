package coloringbook

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aaronland/go-aws-lambda"
)

const GENERATE_COLORING_BOOK_LAMBDA_URI string = "aws://GenerateColouringBook?region=us-west-2&credentials=session"

type ColoringBookRequest struct {
	ObjectId     int64 `json:"object_id"`
	UpdateObject bool  `json:"update_object"`
	// ContourIterations int `json:"contour_iterations"`
}

func GenerateColoringBookLambda(ctx context.Context, function_uri string, object_id int64) error {

	f, err := lambda.NewLambdaFunction(ctx, function_uri)

	if err != nil {
		return fmt.Errorf("Failed to create new Lambda function, %v", err)
	}

	req := ColoringBookRequest{
		ObjectId: object_id,
	}

	payload, err := json.Marshal(req)

	if err != nil {
		return fmt.Errorf("Failed to marshal request, %v", err)
	}

	_, err = f.InvokeWithJSON(ctx, payload)

	if err != nil {
		return fmt.Errorf("Failed to invoke function, %v", err)
	}

	return nil
}
