package bms

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// StringDataVersion is the StringData version these definitions were ported from (STRINGDATA_VERSION).
const StringDataVersion = 5

// StringID identifies an entry of the "FalconSharedMemoryAreaString" area.
type StringID uint32

const (
	// Version 1
	StrBmsExe  StringID = iota // BMS exe name, full path
	StrKeyFile                 // key file in use, full path

	StrBmsBasedir // BMS directories in use
	StrBmsBinDirectory
	StrBmsDataDirectory
	StrBmsUIArtDirectory
	StrBmsUserDirectory
	StrBmsAcmiDirectory
	StrBmsBriefingsDirectory
	StrBmsConfigDirectory
	StrBmsLogsDirectory
	StrBmsPatchDirectory
	StrBmsPictureDirectory

	StrThrName        // current theater name
	StrThrCampaigndir // current theater directories in use
	StrThrTerraindir
	StrThrArtdir
	StrThrMoviedir
	StrThrUisounddir
	StrThrObjectdir
	StrThr3ddatadir
	StrThrMisctexdir
	StrThrSounddir
	StrThrTacrefdir
	StrThrSplashdir
	StrThrCockpitdir
	StrThrSimdatadir
	StrThrSubtitlesdir
	StrThrTacrefpicsdir

	StrAcName // current aircraft name
	StrAcNCTR // current aircraft NCTR

	// Version 2
	StrButtonsFile // current 3dbuttons.dat, full path
	StrCockpitFile // current 3dckpit.dat, full path

	// Version 3
	StrNavPoint // one entry per NavPoint, see FlightData.h for the format

	// Version 4
	StrThrTerrdatadir

	// Version 5
	StrVoiceHelpers // "<player>|<seats>,<tanker>,<awacs>,<departure>,<arrival>,<alternate>"
)

// StringEntry is one string of the string area.
type StringEntry struct {
	ID    StringID
	Value string
}

// StringData is the parsed content of the "FalconSharedMemoryAreaString" area.
type StringData struct {
	VersionNum uint32
	Strings    []StringEntry
}

// Get returns the first string with the given ID.
func (sd *StringData) Get(id StringID) (string, bool) {
	for _, s := range sd.Strings {
		if s.ID == id {
			return s.Value, true
		}
	}
	return "", false
}

// GetAll returns all strings with the given ID (e.g. StrNavPoint).
func (sd *StringData) GetAll(id StringID) []string {
	var res []string
	for _, s := range sd.Strings {
		if s.ID == id {
			res = append(res, s.Value)
		}
	}
	return res
}

var errTruncated = errors.New("shared memory data truncated")

// ParseStringData parses the raw "FalconSharedMemoryAreaString" area.
//
// Layout: VersionNum, NoOfStrings, dataSize (uint32 each), followed by NoOfStrings
// entries of strId (uint32), strLength (uint32) and strLength+1 bytes (NUL terminated).
func ParseStringData(buf []byte) (*StringData, error) {
	p := parser{buf: buf}
	version := p.u32()
	count := p.u32()
	_ = p.u32() // dataSize
	if p.err != nil {
		return nil, p.err
	}

	sd := &StringData{VersionNum: version}
	for i := uint32(0); i < count; i++ {
		id := p.u32()
		value := p.str(p.u32())
		if p.err != nil {
			return nil, fmt.Errorf("string %d of %d: %w", i, count, p.err)
		}
		sd.Strings = append(sd.Strings, StringEntry{ID: StringID(id), Value: value})
	}
	return sd, nil
}

// parser reads little-endian values from a byte slice and records the first error.
type parser struct {
	buf []byte
	off int
	err error
}

func (p *parser) u32() uint32 {
	if p.err != nil {
		return 0
	}
	if len(p.buf)-p.off < 4 {
		p.err = errTruncated
		return 0
	}
	v := binary.LittleEndian.Uint32(p.buf[p.off:])
	p.off += 4
	return v
}

// str reads a string of length n followed by a NUL terminator.
func (p *parser) str(n uint32) string {
	if p.err != nil {
		return ""
	}
	if uint64(len(p.buf)-p.off) < uint64(n)+1 {
		p.err = errTruncated
		return ""
	}
	s := string(p.buf[p.off : p.off+int(n)])
	p.off += int(n) + 1
	return s
}

// cString converts a NUL terminated byte array to a string.
func cString(b []byte) string {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		b = b[:i]
	}
	return string(b)
}
