package bms

import (
	"bufio"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type fieldLayout struct {
	name         string
	offset, size uintptr
}

// loadLayout reads testdata/layout.txt, generated from FlightData.h by gen_layout.py.
func loadLayout(t *testing.T) (map[string][]fieldLayout, map[string]uintptr) {
	t.Helper()
	f, err := os.Open("testdata/layout.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	fields := map[string][]fieldLayout{}
	sizes := map[string]uintptr{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		p := strings.Fields(line)
		num := func(s string) uintptr {
			n, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				t.Fatalf("bad line %q: %v", line, err)
			}
			return uintptr(n)
		}
		if p[1] == "size" {
			sizes[p[0]] = num(p[2])
			continue
		}
		fields[p[0]] = append(fields[p[0]], fieldLayout{p[1], num(p[2]), num(p[3])})
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return fields, sizes
}

// normalize makes C and Go field names comparable, e.g. "uhf_panel_preset" and "UHFPanelPreset".
func normalize(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, "_", ""))
}

// TestLayoutMatchesHeader ensures the Go structs have exactly the memory layout
// of the C++ structs in FlightData.h (Windows x64), field by field.
func TestLayoutMatchesHeader(t *testing.T) {
	fields, sizes := loadLayout(t)

	for name, typ := range map[string]reflect.Type{
		"FlightData":  reflect.TypeFor[FlightData](),
		"FlightData2": reflect.TypeFor[FlightData2](),
		"OsbLabel":    reflect.TypeFor[OsbLabel](),
		"OSBData":     reflect.TypeFor[OSBData](),
	} {
		t.Run(name, func(t *testing.T) {
			want := fields[name]
			if len(want) == 0 {
				t.Fatalf("no layout for %s in testdata/layout.txt", name)
			}
			if typ.NumField() != len(want) {
				t.Errorf("field count: got %d, header has %d", typ.NumField(), len(want))
			}
			for i := range min(typ.NumField(), len(want)) {
				got, w := typ.Field(i), want[i]
				if normalize(got.Name) != normalize(w.name) {
					t.Errorf("field %d: got %s, header has %s", i, got.Name, w.name)
					continue
				}
				if got.Offset != w.offset || got.Type.Size() != w.size {
					t.Errorf("%s: got offset %d size %d, header has offset %d size %d",
						got.Name, got.Offset, got.Type.Size(), w.offset, w.size)
				}
			}
			if typ.Size() != sizes[name] {
				t.Errorf("size: got %d, header has %d", typ.Size(), sizes[name])
			}
		})
	}
}
