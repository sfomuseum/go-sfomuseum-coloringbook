package colouringbook

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"os/exec"

	"github.com/aaronland/go-image-contour"
	"github.com/sfomuseum/go-svg"
)

type OutlineOptions struct {
}

type TraceOptions struct {
}

func GenerateOutline(ctx context.Context, im image.Image, opts *OutlineOptions) (image.Image, error) {

	vtrace_infile, err := os.CreateTemp("", "vtrace.*.png")

	if err != nil {
		return nil, fmt.Errorf("Failed to create vtrace input file, %w", err)
	}

	infile_uri := vtrace_infile.Name()
	defer os.Remove(infile_uri)

	err = png.Encode(vtrace_infile, im)

	if err != nil {
		return nil, fmt.Errorf("Failed to encode image for tracing, %w", err)
	}

	err = vtrace_infile.Close()

	if err != nil {
		return nil, fmt.Errorf("Failed to close infile after writing, %w", err)
	}

	vtrace_outfile, err := os.CreateTemp("", "vtrace.*.svg")

	if err != nil {
		return nil, fmt.Errorf("Failed to create vtrace outfile file, %w", err)
	}

	err = vtrace_outfile.Close()

	if err != nil {
		return nil, fmt.Errorf("Failed to close outfile, %w", err)
	}

	outfile_uri := vtrace_outfile.Name()
	defer os.Remove(outfile_uri)

	trace_opts := &TraceOptions{}

	log.Println("TRACE")

	traced_im, err := Trace(ctx, infile_uri, outfile_uri, trace_opts)

	if err != nil {
		return nil, fmt.Errorf("Failed to trace image, %w", err)
	}

	log.Println("CONTOUR")

	n := 12

	contoured_im, err := contour.ContourImage(ctx, traced_im, n, 1.0)

	if err != nil {
		return nil, fmt.Errorf("Failed to contour image, %w", err)
	}

	return contoured_im, nil
}

func Trace(ctx context.Context, input string, output string, opts *TraceOptions) (image.Image, error) {

	log.Println("VTRACER")

	err := Vtrace(ctx, input, output)

	if err != nil {
		return nil, fmt.Errorf("Failed to run vtracer, %w", err)
	}

	r, err := os.Open(output)

	if err != nil {
		return nil, fmt.Errorf("Failed to open %s for reading, %w", output, err)
	}

	defer r.Close()

	log.Println("RASTERIZE")

	return svg.Rasterize(ctx, r)
}

func Vtrace(ctx context.Context, input string, output string) error {
	
	cmd := "vtracer"

	args := []string{
		"-i",
		input,
		"-o",
		output,
		"--color_precision",
		"8",
		"--filter_speckle",
		"8",
	}

	return exec.CommandContext(ctx, cmd, args...).Run()
}
