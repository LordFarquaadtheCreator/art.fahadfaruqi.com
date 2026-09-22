package exif

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	exif "github.com/dsoprea/go-exif/v3"
)

type Metadata struct {
	Title       string
	AltText     string
	Description string
	Set         string
	Number      int
	Exif        map[string]string
}

// Validate reports the required fields that are still empty.
func (m *Metadata) Validate() error {
	missingFields := []string{}
	if m.Title == "" {
		missingFields = append(missingFields, "title")
	}
	if m.AltText == "" {
		missingFields = append(missingFields, "altText")
	}
	if m.Description == "" {
		missingFields = append(missingFields, "description")
	}
	if m.Set == "" {
		missingFields = append(missingFields, "set")
	}
	if m.Number == 0 {
		missingFields = append(missingFields, "number")
	}
	if len(missingFields) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missingFields, ", "))
	}
	return nil
}

// FromMap rebuilds metadata from the custom metadata an object already carries,
// so a re-encoded file can inherit the fields of the item it replaces instead of
// being prompted for them again. The five named fields are lifted out; every
// other key — the EXIF pairs, a preserved upload time — travels in Exif, which is
// where Map writes them back from.
func FromMap(custom map[string]string) (*Metadata, error) {
	// R2 folds custom metadata keys to lowercase on the way in, so compare in the
	// same case the bucket reports.
	normalised := make(map[string]string, len(custom))
	for k, v := range custom {
		normalised[strings.ToLower(k)] = v
	}

	number, err := strconv.Atoi(normalised["number"])
	if err != nil {
		return nil, fmt.Errorf("number %q is not an integer: %w", normalised["number"], err)
	}

	metadata := &Metadata{
		Title:       normalised["title"],
		AltText:     normalised["alttext"],
		Description: normalised["description"],
		Set:         normalised["set"],
		Number:      number,
		Exif:        map[string]string{},
	}

	for k, v := range normalised {
		switch k {
		case "title", "alttext", "description", "set", "number":
			continue
		}
		metadata.Exif[k] = v
	}

	if err := metadata.Validate(); err != nil {
		return nil, err
	}

	return metadata, nil
}

// Map renders the metadata as the custom metadata R2 stores alongside the object.
func (m *Metadata) Map() map[string]string {
	metaMap := map[string]string{
		"title":       m.Title,
		"altText":     m.AltText,
		"description": m.Description,
		"set":         m.Set,
		"number":      strconv.Itoa(m.Number),
	}

	for k, v := range m.Exif {
		metaMap[k] = v
	}

	return metaMap
}

func Extract(filePath string) (*Metadata, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".jpg", ".jpeg":
		return extractEXIF(filePath)
	case ".png":
		return extractXMP(filePath)
	default:
		return &Metadata{Exif: make(map[string]string)}, nil
	}
}

func extractEXIF(filePath string) (*Metadata, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	metadata := &Metadata{
		Exif: make(map[string]string),
	}

	rawExif, err := exif.SearchAndExtractExif(data)
	if err != nil {
		return metadata, nil
	}

	tags, _, err := exif.GetFlatExifData(rawExif, nil)
	if err != nil {
		return metadata, nil
	}

	for _, tag := range tags {
		metadata.Exif[tag.TagName] = tag.FormattedFirst
	}

	return metadata, nil
}

func extractXMP(filePath string) (*Metadata, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	metadata := &Metadata{
		Exif: make(map[string]string),
	}

	reader := bytes.NewReader(data)
	if err := checkPNGSignature(reader); err != nil {
		return metadata, nil
	}

	chunks, err := readPNGChunks(reader)
	if err != nil {
		return metadata, nil
	}

	for _, chunk := range chunks {
		if chunk.Type == "iTXt" {
			if strings.HasPrefix(string(chunk.Data), "XML:com.adobe.xmp") {
				xmpData := extractXMPData(chunk.Data)
				parseXMP(xmpData, metadata)
			}
		}
	}

	return metadata, nil
}

func checkPNGSignature(r *bytes.Reader) error {
	sig := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	header := make([]byte, 8)
	if _, err := r.Read(header); err != nil {
		return err
	}
	for i := range sig {
		if header[i] != sig[i] {
			return fmt.Errorf("invalid PNG signature")
		}
	}
	return nil
}

type Chunk struct {
	Length uint32
	Type   string
	Data   []byte
	CRC    uint32
}

func readPNGChunks(r *bytes.Reader) ([]Chunk, error) {
	var chunks []Chunk

	for {
		var length uint32
		if err := binaryRead(r, &length); err != nil {
			break
		}

		typeBytes := make([]byte, 4)
		if _, err := r.Read(typeBytes); err != nil {
			return nil, err
		}
		chunkType := string(typeBytes)

		data := make([]byte, length)
		if _, err := r.Read(data); err != nil {
			return nil, err
		}

		var crc uint32
		if err := binaryRead(r, &crc); err != nil {
			return nil, err
		}

		chunks = append(chunks, Chunk{
			Length: length,
			Type:   chunkType,
			Data:   data,
			CRC:    crc,
		})

		if chunkType == "IEND" {
			break
		}
	}

	return chunks, nil
}

func binaryRead(r *bytes.Reader, v interface{}) error {
	return binary.Read(r, binary.BigEndian, v)
}

func extractXMPData(data []byte) string {
	str := string(data)
	idx := strings.Index(str, "<?xpacket")
	if idx == -1 {
		return ""
	}
	return str[idx:]
}

func parseXMP(xmpData string, metadata *Metadata) {
	if strings.Contains(xmpData, "tiff:Make") {
		if idx := strings.Index(xmpData, "tiff:Make=\""); idx != -1 {
			start := idx + len("tiff:Make=\"")
			end := strings.Index(xmpData[start:], "\"")
			if end != -1 {
				metadata.Exif["Make"] = xmpData[start : start+end]
			}
		}
	}
	if strings.Contains(xmpData, "tiff:Model") {
		if idx := strings.Index(xmpData, "tiff:Model=\""); idx != -1 {
			start := idx + len("tiff:Model=\"")
			end := strings.Index(xmpData[start:], "\"")
			if end != -1 {
				metadata.Exif["Model"] = xmpData[start : start+end]
			}
		}
	}
	if strings.Contains(xmpData, "exif:ExposureTime") {
		if idx := strings.Index(xmpData, "exif:ExposureTime=\""); idx != -1 {
			start := idx + len("exif:ExposureTime=\"")
			end := strings.Index(xmpData[start:], "\"")
			if end != -1 {
				metadata.Exif["ExposureTime"] = xmpData[start : start+end]
			}
		}
	}
	if strings.Contains(xmpData, "exif:FNumber") {
		if idx := strings.Index(xmpData, "exif:FNumber=\""); idx != -1 {
			start := idx + len("exif:FNumber=\"")
			end := strings.Index(xmpData[start:], "\"")
			if end != -1 {
				metadata.Exif["FNumber"] = xmpData[start : start+end]
			}
		}
	}
	if strings.Contains(xmpData, "exif:ISOSpeedRatings") {
		if idx := strings.Index(xmpData, "exif:ISOSpeedRatings"); idx != -1 {
			start := idx + len("exif:ISOSpeedRatings")
			for start < len(xmpData) && xmpData[start] != '>' {
				start++
			}
			start++
			end := strings.Index(xmpData[start:], "<")
			if end != -1 {
				metadata.Exif["ISOSpeedRatings"] = strings.TrimSpace(xmpData[start : start+end])
			}
		}
	}
	if strings.Contains(xmpData, "exif:FocalLength") {
		if idx := strings.Index(xmpData, "exif:FocalLength=\""); idx != -1 {
			start := idx + len("exif:FocalLength=\"")
			end := strings.Index(xmpData[start:], "\"")
			if end != -1 {
				metadata.Exif["FocalLength"] = xmpData[start : start+end]
			}
		}
	}
	if strings.Contains(xmpData, "aux:Lens") {
		if idx := strings.Index(xmpData, "aux:Lens=\""); idx != -1 {
			start := idx + len("aux:Lens=\"")
			end := strings.Index(xmpData[start:], "\"")
			if end != -1 {
				metadata.Exif["Lens"] = xmpData[start : start+end]
			}
		}
	}
	if strings.Contains(xmpData, "xmp:CreateDate") {
		if idx := strings.Index(xmpData, "xmp:CreateDate=\""); idx != -1 {
			start := idx + len("xmp:CreateDate=\"")
			end := strings.Index(xmpData[start:], "\"")
			if end != -1 {
				metadata.Exif["DateTimeOriginal"] = xmpData[start : start+end]
			}
		}
	}
}
