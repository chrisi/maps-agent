package camtac

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type FileBundleReader struct {
	embeddedFileDirectory []EmbeddedFileInfo
	rawBytes              []byte
}

func NewFileBundleReader() *FileBundleReader {
	return &FileBundleReader{}
}

func NewFileBundleReaderFromFile(campaignFileBundleFileName string) (*FileBundleReader, error) {
	r := NewFileBundleReader()
	if err := r.Load(campaignFileBundleFileName); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *FileBundleReader) Load(campaignFileBundleFileName string) error {
	f, err := os.Open(campaignFileBundleFileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("file not found: %s", campaignFileBundleFileName)
		}
		return err
	}
	defer f.Close()

	rawBytes, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	if len(rawBytes) < 4 {
		return fmt.Errorf("invalid campaign bundle: file too small")
	}

	directoryStartOffset := binary.LittleEndian.Uint32(rawBytes[0:4])
	if int(directoryStartOffset)+4 > len(rawBytes) {
		return fmt.Errorf("invalid campaign bundle: directory offset out of range")
	}

	numEmbeddedFiles := binary.LittleEndian.Uint32(rawBytes[directoryStartOffset : directoryStartOffset+4])
	curLoc := int(directoryStartOffset) + 4

	embeddedFileDirectory := make([]EmbeddedFileInfo, numEmbeddedFiles)

	for i := uint32(0); i < numEmbeddedFiles; i++ {
		if curLoc >= len(rawBytes) {
			return fmt.Errorf("invalid campaign bundle: unexpected end while reading file entry %d", i)
		}

		thisFileNameLength := int(rawBytes[curLoc])
		curLoc++

		if curLoc+thisFileNameLength > len(rawBytes) {
			return fmt.Errorf("invalid campaign bundle: filename exceeds file size for entry %d", i)
		}

		thisFileName := string(rawBytes[curLoc : curLoc+thisFileNameLength])
		curLoc += thisFileNameLength

		if curLoc+8 > len(rawBytes) {
			return fmt.Errorf("invalid campaign bundle: incomplete metadata for entry %d", i)
		}

		fileOffset := binary.LittleEndian.Uint32(rawBytes[curLoc : curLoc+4])
		curLoc += 4

		fileSizeBytes := binary.LittleEndian.Uint32(rawBytes[curLoc : curLoc+4])
		curLoc += 4

		embeddedFileDirectory[i] = EmbeddedFileInfo{
			FileName:      thisFileName,
			FileOffset:    fileOffset,
			FileSizeBytes: fileSizeBytes,
		}
	}

	r.rawBytes = rawBytes
	r.embeddedFileDirectory = embeddedFileDirectory

	return nil
}

func (r *FileBundleReader) GetEmbeddedFileDirectory() ([]EmbeddedFileInfo, error) {
	if r.embeddedFileDirectory == nil || len(r.rawBytes) == 0 {
		return nil, fmt.Errorf("campaign bundle file not loaded yet")
	}

	result := make([]EmbeddedFileInfo, len(r.embeddedFileDirectory))
	copy(result, r.embeddedFileDirectory)
	return result, nil
}

func (r *FileBundleReader) GetEmbeddedFileContents(embeddedFileName string) ([]byte, error) {
	if r.embeddedFileDirectory == nil || len(r.rawBytes) == 0 {
		return nil, fmt.Errorf("campaign bundle file not loaded yet")
	}

	for _, thisFile := range r.embeddedFileDirectory {
		if strings.EqualFold(thisFile.FileName, embeddedFileName) {
			start := int(thisFile.FileOffset)
			end := start + int(thisFile.FileSizeBytes)

			if start < 0 || end > len(r.rawBytes) || start > end {
				return nil, fmt.Errorf("invalid embedded file range for %q", embeddedFileName)
			}

			toReturn := make([]byte, thisFile.FileSizeBytes)
			copy(toReturn, r.rawBytes[start:end])
			return toReturn, nil
		}
	}

	return nil, fmt.Errorf("embedded file not found: %s", embeddedFileName)
}
