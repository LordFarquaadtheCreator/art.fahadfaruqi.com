package exif

import (
	"reflect"
	"testing"
)

func base() map[string]string {
	return map[string]string{
		"title":       "Morning light",
		"altText":     "Morning light",
		"description": "A description",
		"set":         "Sam",
		"number":      "3",
		"make":        "NIKON CORPORATION",
	}
}

func TestZoomRoundTrip(t *testing.T) {
	custom := base()
	custom["zoom_x"] = "0.22"
	custom["zoom_y"] = "0.71"
	custom["zoom_level"] = "2.5"
	custom["zoom_duration"] = "1500"

	metadata, err := FromMap(custom)
	if err != nil {
		t.Fatalf("FromMap: %v", err)
	}
	if metadata.ZoomX != 0.22 || metadata.ZoomY != 0.71 || metadata.ZoomLevel != 2.5 || metadata.ZoomDuration != 1500 {
		t.Fatalf("parsed %v/%v/%v/%v, want 0.22/0.71/2.5/1500",
			metadata.ZoomX, metadata.ZoomY, metadata.ZoomLevel, metadata.ZoomDuration)
	}

	out := metadata.Map()
	for _, key := range []string{"zoom_x", "zoom_y", "zoom_level", "zoom_duration"} {
		if out[key] != custom[key] {
			t.Errorf("%s round-tripped as %q, want %q", key, out[key], custom[key])
		}
	}

	// And the EXIF must survive alongside, not be swallowed by the zoom case.
	if out["make"] != "NIKON CORPORATION" {
		t.Errorf("make round-tripped as %q", out["make"])
	}
	if metadata.Exif["make"] != "NIKON CORPORATION" {
		t.Errorf("make landed in Exif as %q", metadata.Exif["make"])
	}
}

// The zoom keys are lifted into the four fields, so they must not also ride along in
// Exif — a re-upload would then write each twice.
func TestZoomKeysDoNotLandInExif(t *testing.T) {
	custom := base()
	custom["zoom_x"] = "0.5"
	custom["zoom_y"] = "0.5"
	custom["zoom_level"] = "2"
	custom["zoom_duration"] = "900"

	metadata, err := FromMap(custom)
	if err != nil {
		t.Fatalf("FromMap: %v", err)
	}

	for _, key := range []string{"zoom_x", "zoom_y", "zoom_level", "zoom_duration"} {
		if _, present := metadata.Exif[key]; present {
			t.Errorf("%s leaked into Exif", key)
		}
	}
	if len(metadata.Exif) != 1 || metadata.Exif["make"] != "NIKON CORPORATION" {
		t.Errorf("Exif is %v, want only make", metadata.Exif)
	}
}

func TestZoomAbsentIsNotAnError(t *testing.T) {
	metadata, err := FromMap(base())
	if err != nil {
		t.Fatalf("FromMap: %v", err)
	}
	if metadata.ZoomLevel != 0 {
		t.Fatalf("ZoomLevel is %v, want 0", metadata.ZoomLevel)
	}
	if metadata.zoomEnabled() {
		t.Error("zoomEnabled is true for an object with no zoom")
	}
	if out := metadata.Map(); len(out) != len(base()) {
		t.Errorf("Map wrote %d keys, want %d: %v", len(out), len(base()), out)
	}
}

// 0 is the absent marker, so a zoom-out at level 0.5 must still be written.
func TestZoomOutIsWritten(t *testing.T) {
	metadata, err := FromMap(base())
	if err != nil {
		t.Fatalf("FromMap: %v", err)
	}
	metadata.ZoomX, metadata.ZoomY = 0, 0
	metadata.ZoomLevel, metadata.ZoomDuration = 0.5, 2000

	out := metadata.Map()
	if !reflect.DeepEqual(out["zoom_level"], "0.5") {
		t.Fatalf("zoom_level is %q, want \"0.5\"", out["zoom_level"])
	}
	if out["zoom_x"] != "0" || out["zoom_y"] != "0" {
		t.Errorf("a corner origin wrote %q/%q, want \"0\"/\"0\"", out["zoom_x"], out["zoom_y"])
	}

	// And the far edge, since 1 is the other end of the range and must not be read as
	// "unset": the marker is 0, not a falsy-looking value.
	metadata.ZoomX, metadata.ZoomY, metadata.ZoomLevel = 1, 1, 1
	out = metadata.Map()
	if out["zoom_x"] != "1" || out["zoom_y"] != "1" || out["zoom_level"] != "1" {
		t.Errorf("the far corner wrote %q/%q/%q", out["zoom_x"], out["zoom_y"], out["zoom_level"])
	}
}

func TestZoomUnparseableIsRefused(t *testing.T) {
	for _, key := range []string{"zoom_x", "zoom_y", "zoom_level", "zoom_duration"} {
		custom := base()
		custom["zoom_x"] = "0.5"
		custom["zoom_y"] = "0.5"
		custom["zoom_level"] = "2"
		custom["zoom_duration"] = "900"
		custom[key] = "halfway"

		if _, err := FromMap(custom); err == nil {
			t.Errorf("FromMap accepted %s=%q", key, custom[key])
		}
	}
}
