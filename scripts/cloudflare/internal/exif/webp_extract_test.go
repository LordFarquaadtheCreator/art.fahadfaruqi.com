package exif

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

const xmpPacket = `<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:tiff="http://ns.adobe.com/tiff/1.0/"
    xmlns:exif="http://ns.adobe.com/exif/1.0/"
    xmlns:aux="http://ns.adobe.com/exif/1.0/aux/"
    xmlns:xmp="http://ns.adobe.com/xap/1.0/"
    tiff:Make="NIKON CORPORATION"
    tiff:Model="NIKON D3300"
    exif:ExposureTime="1/125"
    exif:FNumber="7/2"
    exif:FocalLength="60/1"
    aux:Lens="60mm f/2.8"
    xmp:CreateDate="2026-10-04T11:35:48.20-05:00"/>
 </rdf:RDF>
</x:xmpmeta>
<?xpacket end="w"?>`

func buildWebP(xmp []byte) []byte {
	var body []byte

	body = append(body, []byte("VP8X")...)
	body = append(body, 10, 0, 0, 0)
	body = append(body, make([]byte, 10)...)

	if xmp != nil {
		body = append(body, []byte("XMP ")...)
		var length [4]byte
		binary.LittleEndian.PutUint32(length[:], uint32(len(xmp)))
		body = append(body, length[:]...)
		body = append(body, xmp...)
		if len(xmp)%2 == 1 {
			body = append(body, 0)
		}
	}

	out := append([]byte("RIFF"), 0, 0, 0, 0)
	binary.LittleEndian.PutUint32(out[4:8], uint32(len(body)+4))
	out = append(out, []byte("WEBP")...)
	out = append(out, body...)

	return out
}

func writeTempFile(t *testing.T, name string, data []byte) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}

	return path
}

func TestExtractWebPReadsTheXMPChunk(t *testing.T) {
	path := writeTempFile(t, "master.webp", buildWebP([]byte(xmpPacket)))

	metadata, err := Extract(path)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}

	want := map[string]string{
		"Make":             "NIKON CORPORATION",
		"Model":            "NIKON D3300",
		"ExposureTime":     "1/125",
		"FNumber":          "7/2",
		"FocalLength":      "60/1",
		"Lens":             "60mm f/2.8",
		"DateTimeOriginal": "2026-10-04T11:35:48.20-05:00",
	}
	for key, value := range want {
		if metadata.Exif[key] != value {
			t.Errorf("%s = %q, want %q", key, metadata.Exif[key], value)
		}
	}
}

func TestExtractWebPWithoutMetadata(t *testing.T) {
	path := writeTempFile(t, "plain.webp", buildWebP(nil))

	metadata, err := Extract(path)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if len(metadata.Exif) != 0 {
		t.Errorf("Exif is %v, want empty", metadata.Exif)
	}
}

func TestExtractWebPThatIsNotOne(t *testing.T) {
	path := writeTempFile(t, "garbage.webp", []byte("this is not a RIFF container"))

	metadata, err := Extract(path)
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if len(metadata.Exif) != 0 {
		t.Errorf("Exif is %v, want empty", metadata.Exif)
	}
}
