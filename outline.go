package colouringbook

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/fogleman/colormap"
	"github.com/fogleman/contourmap"
	"github.com/fogleman/gg"
	"github.com/sfomuseum/go-svg"
)

type OutlineOptions struct {
	Contour   *ContourOptions
	Trace     *TraceOptions
	Rasterize *RasterizeOptions
}

type ContourOptions struct {
	Iterations int
}

type TraceOptions struct {
	Precision int
	Speckle   int
}

type RasterizeOptions struct {
	UseBatik  bool
	PathBatik string
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

	log.Println("TRACE")

	traced_im, err := Trace(ctx, infile_uri, outfile_uri, opts.Trace, opts.Rasterize)

	if err != nil {
		return nil, fmt.Errorf("Failed to trace image, %w", err)
	}

	log.Println("CONTOUR")

	contoured_im, err := Contour(ctx, traced_im, opts.Contour)

	if err != nil {
		return nil, fmt.Errorf("Failed to contour image, %w", err)
	}

	return contoured_im, nil
}

func Contour(ctx context.Context, im image.Image, opts *ContourOptions) (image.Image, error) {

	iterations := opts.Iterations
	scale := 1.0

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

	for i := 0; i < iterations; i++ {

		t := float64(i) / (float64(iterations) - 1)
		z := z0 + (z1-z0)*t
		contours := m.Contours(z + 1e-9)

		// Do line smoothing here?

		for _, c := range contours {

			dc.NewSubPath()

			for _, p := range c {
				dc.LineTo(p.X, p.Y)
			}
		}

		dc.SetRGB(0, 0, 0)

		// z = 1.0
		z = z * float64(i)

		dc.SetLineWidth(z)
		dc.Stroke()
	}

	return dc.Image(), nil
}

func Trace(ctx context.Context, input string, output string, trace_opts *TraceOptions, raster_opts *RasterizeOptions) (image.Image, error) {

	log.Println("VTRACER")

	err := Vtrace(ctx, input, output, trace_opts)

	if err != nil {
		return nil, fmt.Errorf("Failed to run vtracer, %w", err)
	}

	log.Println("RASTERIZE")

	if raster_opts.UseBatik {
		return RasterizeBatik(ctx, raster_opts, output)
	}

	// This is very (very) slow
	return RasterizeNative(ctx, raster_opts, output)
}

func RasterizeNative(ctx context.Context, opts *RasterizeOptions, input string) (image.Image, error) {

	r, err := os.Open(input)

	if err != nil {
		return nil, fmt.Errorf("Failed to open %s for reading, %w", input, err)
	}

	defer r.Close()

	return svg.Rasterize(ctx, r)
}

func RasterizeBatik(ctx context.Context, opts *RasterizeOptions, input string) (image.Image, error) {

	cmd := "java"

	args := []string{
		"-jar",
		opts.PathBatik,
		input,
	}

	err := exec.CommandContext(ctx, cmd, args...).Run()

	if err != nil {
		return nil, fmt.Errorf("Failed to run batik, %w", err)
	}

	// Why can't I specify the output filename in Batik?
	output := strings.Replace(input, ".svg", ".png", 1)

	defer os.Remove(output)

	r, err := os.Open(output)

	if err != nil {
		return nil, fmt.Errorf("Failed to open %s for reading, %w", output, err)
	}

	defer r.Close()

	im, _, err := image.Decode(r)

	if err != nil {
		return nil, fmt.Errorf("Failed to decode image, %w", err)
	}

	return im, nil
}

func Vtrace(ctx context.Context, input string, output string, opts *TraceOptions) error {

	precision := opts.Precision
	speckle := opts.Speckle

	cmd := "vtracer"

	args := []string{
		"-i",
		input,
		"-o",
		output,
		"--color_precision",
		strconv.Itoa(precision),
		"--filter_speckle",
		strconv.Itoa(speckle),
	}

	return exec.CommandContext(ctx, cmd, args...).Run()
}
