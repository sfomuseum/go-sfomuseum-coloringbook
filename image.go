package colouringbook

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"

	"github.com/jtacoma/uritemplates"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-reader"
	wof_reader "github.com/whosonfirst/go-whosonfirst-reader"
)

func DeriveObjectImage(ctx context.Context, r reader.Reader, image_id int64) (string, error) {

	im_body, err := wof_reader.LoadBytes(ctx, r, image_id)

	if err != nil {
		return "", fmt.Errorf("Failed to load body for image %d, %v", image_id, err)
	}

	o_rsp := gjson.GetBytes(im_body, "properties.media:properties.sizes.o")

	if !o_rsp.Exists() {
		return "", fmt.Errorf("Image %d is missing properties.media:properties.sizes.o property")
	}

	ext_rsp := o_rsp.Get("extension")
	secret_rsp := o_rsp.Get("secret")

	template_rsp := gjson.GetBytes(im_body, "properties.media:uri_template")

	if !template_rsp.Exists() {
		return "", fmt.Errorf("Image %d is missing properties.media:uri_template property")
	}

	uri_template, err := uritemplates.Parse(template_rsp.String())

	if err != nil {
		return "", fmt.Errorf("Failed to create URI template, %v", err)
	}

	template_values := map[string]interface{}{
		"secret":    secret_rsp.String(),
		"extention": ext_rsp.String(),
		"label":     "o",
	}

	im_uri, err := uri_template.Expand(template_values)

	if err != nil {
		return "", fmt.Errorf("Failed to expand URI template, %v", err)
	}

	im_rsp, err := http.Get(im_uri)

	if err != nil {
		return "", fmt.Errorf("Failed to retrieve %s, %w", im_uri, err)
	}

	defer im_rsp.Body.Close()

	im, _, err := image.Decode(im_rsp.Body)

	if err != nil {
		return "", fmt.Errorf("Failed to decode image %d (%s), %v", image_id, im_uri, err)
	}

	log.Println("Generate outline")

	contoured_im, err := GenerateOutline(ctx, im, nil)

	if err != nil {
		return "", fmt.Errorf("Failed to generate outline for image %d, %w", image_id, err)
	}

	im_tmpfile, err := os.CreateTemp("", "*.png")

	if err != nil {
		return "", fmt.Errorf("Failed to create outline file, %v", err)
	}

	object_image := im_tmpfile.Name()

	err = png.Encode(im_tmpfile, contoured_im)

	if err != nil {
		os.Remove(object_image)
		return "", fmt.Errorf("Failed to encode outline file, %v", err)
	}

	err = im_tmpfile.Close()

	if err != nil {
		os.Remove(object_image)
		return "", fmt.Errorf("Failed to close outline file, %v", err)
	}

	return object_image, nil
}