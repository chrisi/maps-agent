package bms

import (
	"encoding/binary"
	"errors"
	"testing"
)

type builder []byte

func (b builder) u32(v uint32) builder { return binary.LittleEndian.AppendUint32(b, v) }

// str appends length, data and NUL terminator.
func (b builder) str(s string) builder { return append(b.u32(uint32(len(s))), append([]byte(s), 0)...) }

func TestParseStringData(t *testing.T) {
	buf := builder{}.u32(5).u32(3).u32(0).
		u32(uint32(StrThrName)).str("Korea KTO").
		u32(uint32(StrNavPoint)).str("NP:1,WP,1.0,2.0,3.0,0.0;").
		u32(uint32(StrNavPoint)).str("NP:2,WP,4.0,5.0,6.0,0.0;")

	sd, err := ParseStringData(buf)
	if err != nil {
		t.Fatal(err)
	}
	if sd.VersionNum != 5 || len(sd.Strings) != 3 {
		t.Fatalf("got version %d, %d strings", sd.VersionNum, len(sd.Strings))
	}
	if v, ok := sd.Get(StrThrName); !ok || v != "Korea KTO" {
		t.Errorf("ThrName = %q, %v", v, ok)
	}
	if _, ok := sd.Get(StrAcName); ok {
		t.Errorf("AcName should not be present")
	}
	if nps := sd.GetAll(StrNavPoint); len(nps) != 2 || nps[1] != "NP:2,WP,4.0,5.0,6.0,0.0;" {
		t.Errorf("NavPoints = %q", nps)
	}
}

func TestParseStringDataTruncated(t *testing.T) {
	full := builder{}.u32(5).u32(1).u32(0).u32(uint32(StrAcName)).str("F-16C-50")
	for n := range len(full) {
		if _, err := ParseStringData(full[:n]); !errors.Is(err, errTruncated) {
			t.Fatalf("len %d: expected truncation error, got %v", n, err)
		}
	}
}

func TestParseDrawingData(t *testing.T) {
	buf := builder{}.u32(1).str("R:100,100;L:0,0,1,1;").str("").str("P:1,2;")
	dd, err := ParseDrawingData(buf)
	if err != nil {
		t.Fatal(err)
	}
	want := DrawingData{VersionNum: 1, HUDCommands: "R:100,100;L:0,0,1,1;", HMSCommands: "P:1,2;"}
	if *dd != want {
		t.Errorf("got %+v, want %+v", *dd, want)
	}
	if _, err := ParseDrawingData(buf[:len(buf)-1]); !errors.Is(err, errTruncated) {
		t.Errorf("expected truncation error, got %v", err)
	}
}

func TestCString(t *testing.T) {
	var fd FlightData
	copy(fd.DEDLines[1][:], "  UHF  242.00  STPT  1\x00garbage")
	if got := fd.DEDLine(1); got != "  UHF  242.00  STPT  1" {
		t.Errorf("DEDLine = %q", got)
	}
}
