// Code generated by smithy-go-codegen DO NOT EDIT.

package ssm

import (
	"context"
	"fmt"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"time"
)

// Retrieves a maintenance window.
func (c *Client) GetMaintenanceWindow(ctx context.Context, params *GetMaintenanceWindowInput, optFns ...func(*Options)) (*GetMaintenanceWindowOutput, error) {
	if params == nil {
		params = &GetMaintenanceWindowInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "GetMaintenanceWindow", params, optFns, c.addOperationGetMaintenanceWindowMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*GetMaintenanceWindowOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type GetMaintenanceWindowInput struct {

	// The ID of the maintenance window for which you want to retrieve information.
	//
	// This member is required.
	WindowId *string

	noSmithyDocumentSerde
}

type GetMaintenanceWindowOutput struct {

	// Whether targets must be registered with the maintenance window before tasks can
	// be defined for those targets.
	AllowUnassociatedTargets bool

	// The date the maintenance window was created.
	CreatedDate *time.Time

	// The number of hours before the end of the maintenance window that Amazon Web
	// Services Systems Manager stops scheduling new tasks for execution.
	Cutoff int32

	// The description of the maintenance window.
	Description *string

	// The duration of the maintenance window in hours.
	Duration *int32

	// Indicates whether the maintenance window is enabled.
	Enabled bool

	// The date and time, in ISO-8601 Extended format, for when the maintenance window
	// is scheduled to become inactive. The maintenance window won't run after this
	// specified time.
	EndDate *string

	// The date the maintenance window was last modified.
	ModifiedDate *time.Time

	// The name of the maintenance window.
	Name *string

	// The next time the maintenance window will actually run, taking into account any
	// specified times for the maintenance window to become active or inactive.
	NextExecutionTime *string

	// The schedule of the maintenance window in the form of a cron or rate expression.
	Schedule *string

	// The number of days to wait to run a maintenance window after the scheduled cron
	// expression date and time.
	ScheduleOffset *int32

	// The time zone that the scheduled maintenance window executions are based on, in
	// Internet Assigned Numbers Authority (IANA) format. For example:
	// "America/Los_Angeles", "UTC", or "Asia/Seoul". For more information, see the
	// Time Zone Database (https://www.iana.org/time-zones) on the IANA website.
	ScheduleTimezone *string

	// The date and time, in ISO-8601 Extended format, for when the maintenance window
	// is scheduled to become active. The maintenance window won't run before this
	// specified time.
	StartDate *string

	// The ID of the created maintenance window.
	WindowId *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationGetMaintenanceWindowMiddlewares(stack *middleware.Stack, options Options) (err error) {
	if err := stack.Serialize.Add(&setOperationInputMiddleware{}, middleware.After); err != nil {
		return err
	}
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpGetMaintenanceWindow{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpGetMaintenanceWindow{}, middleware.After)
	if err != nil {
		return err
	}
	if err := addProtocolFinalizerMiddlewares(stack, options, "GetMaintenanceWindow"); err != nil {
		return fmt.Errorf("add protocol finalizers: %v", err)
	}

	if err = addlegacyEndpointContextSetter(stack, options); err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddClientRequestIDMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddComputeContentLengthMiddleware(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = v4.AddComputePayloadSHA256Middleware(stack); err != nil {
		return err
	}
	if err = addRetryMiddlewares(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = addClientUserAgent(stack, options); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = addSetLegacyContextSigningOptionsMiddleware(stack); err != nil {
		return err
	}
	if err = addOpGetMaintenanceWindowValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opGetMaintenanceWindow(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecursionDetection(stack); err != nil {
		return err
	}
	if err = addRequestIDRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	if err = addDisableHTTPSMiddleware(stack, options); err != nil {
		return err
	}
	return nil
}

func newServiceMetadataMiddleware_opGetMaintenanceWindow(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		OperationName: "GetMaintenanceWindow",
	}
}
