package lambda

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/aaronland/go-aws-session"
	"github.com/aws/aws-sdk-go/aws"
	aws_session "github.com/aws/aws-sdk-go/aws/session"
	aws_lambda "github.com/aws/aws-sdk-go/service/lambda"
)

type LambdaFunction struct {
	service   *aws_lambda.Lambda
	func_name string
	func_type string
}

func NewLambdaFunction(ctx context.Context, uri string) (*LambdaFunction, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	func_name := u.Host
	func_type := "Event"

	if q.Get("type") != "" {
		func_type = q.Get("type")
	}

	sess, err := session.NewSession(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new session, %w", err)
	}

	return NewLambdaFunctionWithSession(sess, func_name, func_type)
}

func NewLambdaFunctionWithDSN(dsn string, func_name string, func_type string) (*LambdaFunction, error) {

	sess, err := session.NewSessionWithDSN(dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new session, %w", err)
	}

	return NewLambdaFunctionWithSession(sess, func_name, func_type)
}

func NewLambdaFunctionWithSession(sess *aws_session.Session, func_name string, func_type string) (*LambdaFunction, error) {

	svc := aws_lambda.New(sess)
	return NewLambdaFunctionWithService(svc, func_name, func_type)
}

func NewLambdaFunctionWithService(svc *aws_lambda.Lambda, func_name string, func_type string) (*LambdaFunction, error) {

	f := &LambdaFunction{
		service:   svc,
		func_name: func_name,
		func_type: func_type,
	}

	return f, nil
}

func (f *LambdaFunction) Invoke(ctx context.Context, payload interface{}) (*aws_lambda.InvokeOutput, error) {

	enc_payload, err := json.Marshal(payload)

	if err != nil {
		return nil, fmt.Errorf("Failed to marshal payload, %w", err)
	}

	return f.InvokeWithJSON(ctx, enc_payload)
}

func (f *LambdaFunction) InvokeWithJSON(ctx context.Context, payload []byte) (*aws_lambda.InvokeOutput, error) {

	input := &aws_lambda.InvokeInput{
		FunctionName:   aws.String(f.func_name),
		InvocationType: aws.String(f.func_type),
		Payload:        payload,
	}

	if *input.InvocationType == "RequestResponse" {
		input.LogType = aws.String("Tail")
	}

	rsp, err := f.service.Invoke(input)

	if err != nil {
		return nil, fmt.Errorf("Failed to invoke function %s (%s), %w", f.func_name, f.func_type, err)
	}

	if *input.InvocationType != "RequestResponse" {
		return nil, nil
	}

	enc_result := *rsp.LogResult

	result, err := base64.StdEncoding.DecodeString(enc_result)

	if err != nil {
		return nil, fmt.Errorf("Failed to decode result, %w", err)
	}

	if *rsp.StatusCode != 200 {
		return nil, fmt.Errorf("Unexpected status code  %d (%s)", *rsp.StatusCode, string(result))
	}

	return rsp, nil
}
