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

	iterations := 8
	scale := 1.0

	contoured_im, err := Contour(ctx, traced_im, iterations, scale)

	if err != nil {
		return nil, fmt.Errorf("Failed to contour image, %w", err)
	}

	return contoured_im, nil
}

func Contour(ctx context.Context, im image.Image, iterations int, scale float64) (image.Image, error) {

	return contour.ContourImage(ctx, im, iterations, scale)

	/*

		"github.com/fogleman/colormap"
		"github.com/fogleman/contourmap"
		"github.com/fogleman/gg"

		m := contourmap.FromImage(im).Closed()
		z0 := m.Min
		z1 := m.Max

		w := int(float64(m.W) * scale)
		h := int(float64(m.H) * scale)

		dc := gg.NewContext(w, h)
		dc.SetRGB(1, 1, 1)
		dc.SetColor(colormap.ParseColor("FFFFFF"))
		dc.Clear()
		dc.Scale(scale, scale)

		for i := 0; i < n; i++ {

			t := float64(i) / (float64(n) - 1)
			z := z0 + (z1-z0)*t
			contours := m.Contours(z + 1e-9)

			for _, c := range contours {

				dc.NewSubPath()

				for _, p := range c {
					dc.LineTo(p.X, p.Y)
				}
			}

			dc.SetRGB(0, 0, 0)
			dc.SetLineWidth(z)
			dc.Stroke()
		}

		return dc.Image(), nil
	*/
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
