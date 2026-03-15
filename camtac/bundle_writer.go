package camtac

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

type embeddedFileEntry struct {
	FileName string
	Data     []byte
}

type F4CampaignFileBundleWriter struct {
	files []embeddedFileEntry
}

func NewF4CampaignFileBundleWriter() *F4CampaignFileBundleWriter {
	return &F4CampaignFileBundleWriter{
		files: make([]embeddedFileEntry, 0),
	}
}

func (w *F4CampaignFileBundleWriter) AddFile(fileName string, data []byte) error {
	if fileName == "" {
		return fmt.Errorf("file name cannot be empty")
	}
	if len(fileName) > 255 {
		return fmt.Errorf("file name too long: %q exceeds 255 bytes", fileName)
	}

	for i := 0; i < len(fileName); i++ {
		if fileName[i] > 127 {
			return fmt.Errorf("file name must be ASCII: %q", fileName)
		}
	}

	copied := make([]byte, len(data))
	copy(copied, data)

	w.files = append(w.files, embeddedFileEntry{
		FileName: fileName,
		Data:     copied,
	})
	return nil
}

func (w *F4CampaignFileBundleWriter) AddFileFromDisk(bundleFileName string, sourcePath string) error {
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	return w.AddFile(bundleFileName, data)
}

func (w *F4CampaignFileBundleWriter) Build() ([]byte, error) {
	if len(w.files) == 0 {
		return nil, fmt.Errorf("no files added")
	}

	var dataSection bytes.Buffer
	directory := make([]EmbeddedFileInfo, len(w.files))

	for i, file := range w.files {
		offset := uint32(4 + dataSection.Len())
		size := uint32(len(file.Data))

		_, err := dataSection.Write(file.Data)
		if err != nil {
			return nil, err
		}

		directory[i] = EmbeddedFileInfo{
			FileName:      file.FileName,
			FileOffset:    offset,
			FileSizeBytes: size,
		}
	}

	directoryStartOffset := uint32(4 + dataSection.Len())

	var out bytes.Buffer

	err := binary.Write(&out, binary.LittleEndian, directoryStartOffset)
	if err != nil {
		return nil, err
	}

	_, err = out.Write(dataSection.Bytes())
	if err != nil {
		return nil, err
	}

	err = binary.Write(&out, binary.LittleEndian, uint32(len(directory)))
	if err != nil {
		return nil, err
	}

	for _, entry := range directory {
		nameBytes := []byte(entry.FileName)
		if len(nameBytes) > 255 {
			return nil, fmt.Errorf("file name too long: %q exceeds 255 bytes", entry.FileName)
		}

		err = out.WriteByte(byte(len(nameBytes)))
		if err != nil {
			return nil, err
		}

		_, err = out.Write(nameBytes)
		if err != nil {
			return nil, err
		}

		err = binary.Write(&out, binary.LittleEndian, entry.FileOffset)
		if err != nil {
			return nil, err
		}

		err = binary.Write(&out, binary.LittleEndian, entry.FileSizeBytes)
		if err != nil {
			return nil, err
		}
	}

	return out.Bytes(), nil
}

func (w *F4CampaignFileBundleWriter) Save(path string) error {
	raw, err := w.Build()
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}
