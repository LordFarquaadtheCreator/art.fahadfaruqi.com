package ffmpeg

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Available reports whether the encoder and the probe the uploader shells out
// to are on PATH.
func Available() error {
	for _, binary := range []string{"ffmpeg", "ffprobe"} {
		if _, err := exec.LookPath(binary); err != nil {
			return fmt.Errorf("%s is not on PATH", binary)
		}
	}

	return nil
}

// Dimensions reads an image's width and height as ffprobe reports them.
func Dimensions(path string) (int, int, error) {
	out, err := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0:s=x",
		path,
	).Output()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read the dimensions of %s: %w", path, err)
	}

	fields := strings.Split(strings.TrimSpace(string(out)), "x")
	if len(fields) != 2 {
		return 0, 0, fmt.Errorf("unexpected dimensions from ffprobe for %s: %q", path, out)
	}

	width, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, 0, fmt.Errorf("unexpected width from ffprobe for %s: %q", path, fields[0])
	}

	height, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, fmt.Errorf("unexpected height from ffprobe for %s: %q", path, fields[1])
	}

	return width, height, nil
}

// Compress writes an AVIF no wider than width, keeping the source's ratio. An
// image already narrower than that is not scaled up.
func Compress(src, dst string, width, crf, preset int) error {
	out, err := exec.Command(
		"ffmpeg",
		"-y", "-hide_banner", "-loglevel", "error",
		"-i", src,
		"-vf", fmt.Sprintf("scale=min(%d\\,iw):-2", width),
		"-c:v", "libsvtav1",
		"-crf", strconv.Itoa(crf),
		"-preset", strconv.Itoa(preset),
		"-frames:v", "1",
		dst,
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to compress %s: %w: %s", src, err, strings.TrimSpace(string(out)))
	}

	return nil
}
