# Porting a new Falcon BMS `FlightData.h` to Go

The package `bms` is a pure-Go port (no cgo) of the shared memory structures in
`include/FlightData.h`. The Go structs are copied byte for byte from the mapped memory, so their
**memory layout must match the C++ classes as compiled by MSVC for Windows x64 exactly**. A mismatch
does not cause an error; every value after the bad field is just silently wrong.

Do not reintroduce cgo or a C copy of the header (the old `include/FlightData_c.h`). The point of
the port is that the agent cross-compiles from Linux with `CGO_ENABLED=0` (to run it in Wine/Proton
next to BMS).

## Files

| File                       | Content                                                                 |
|----------------------------|-------------------------------------------------------------------------|
| `flightdata.go`            | `FlightData` ("FalconSharedMemoryArea"), LightBits/2/3, HsiBits         |
| `flightdata2.go`           | `FlightData2` ("FalconSharedMemoryArea2") and all its enums/bit types   |
| `osbdata.go`               | `OsbLabel`, `OSBData` ("FalconSharedOsbMemoryArea")                     |
| `stringdata.go`            | `StringID` enum + parser for "FalconSharedMemoryAreaString"             |
| `drawingdata.go`           | parser for "FalconSharedMemoryAreaDrawing"                              |
| `reader.go`                | `Reader` with snapshot getters, area names                              |
| `area_windows.go`          | OpenFileMappingW / MapViewOfFile / VirtualQuery                         |
| `area_other.go`            | non-Windows stub (lets everything build and test on Linux)              |
| `layout_test.go`           | compares Go structs with `testdata/layout.txt`                          |
| `testdata/gen_layout.py`   | generates `layout.txt` from the header with g++                         |
| `../sharedmem.go`          | agent-side adapter (`SharedMemReader`); only touch it if the agent needs new data |

## Workflow

1. Look at what changed. The new header is usually not committed yet, so diff it against the
   committed one: `git diff include/FlightData.h`. Read the changelog comments at the top of each
   class (`FLIGHTDATA_VERSION`, `FLIGHTDATA2_VERSION`, `STRINGDATA_VERSION`, `DRAWINGDATA_VERSION`),
   but **do not rely on them alone**. Fields have been added without a changelog entry before
   (e.g. `sideSlipdeg` in FlightData2 has no version comment).
2. Regenerate the reference layout: `python3 bms/testdata/gen_layout.py` (needs g++).
3. Run `go test ./bms`. It lists every field whose name, offset or size differs, a wrong field
   count, and a wrong total size.
4. Change the Go structs until the test passes. Then also update the parts the test cannot check
   (see the checklist below).
5. Verify everything:
   ```
   gofmt -l .
   go vet ./... && GOOS=windows go vet ./...
   go test ./...
   GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o /dev/null .
   ```

## Type mapping (MSVC, Windows x64)

| C++                                | Go                                  | Note                                           |
|------------------------------------|-------------------------------------|------------------------------------------------|
| `float` / `double`                 | `float32` / `float64`               |                                                |
| `int` / `unsigned int`             | `int32` / `uint32`                  | never Go `int`/`uint` (8 bytes)                |
| `long` / `unsigned long`           | `int32` / `uint32`                  | **4 bytes on Windows**, unlike Linux           |
| `long long`                        | `int64` / `uint64`                  |                                                |
| `short` / `unsigned short`         | `int16` / `uint16`                  |                                                |
| `char` used as a number            | `int8`                              | `char` is signed on MSVC                       |
| `char[]` used as text              | `[N]byte`                           | add a getter using `cString()`                 |
| `bool`                             | `uint8` + `IsX() bool` method       | Go `bool` must not hold values other than 0/1  |
| `enum X : unsigned char` / `: int` | named Go type with that base type   | e.g. `type NavMode uint8`                      |
| plain `enum X` used as a field     | 4-byte type (`int32`/`uint32`)      |                                                |
| bitfield `unsigned int xBits`      | named `uint32` type with `Has()`    |                                                |

- Go inserts padding the same way MSVC does (natural alignment). Never add padding fields by hand;
  if the test reports an offset difference, a type is wrong.
- `int a, b;` declares two fields. Port each one as its own Go field.
- Array sizes come from `#define`s (`MAX_RWR_OBJECTS`, `RWRINFO_SIZE`, ...). Use or update the
  matching Go constants.
- Members containing `std::string`/`std::vector` cannot be mapped 1:1. Those areas (StringData,
  DrawingData) are parsed manually from a byte buffer, so update `ParseStringData` /
  `ParseDrawingData` if their serialized format changes.

## Naming

- Go field = C name in exported form. The layout test compares names after lowercasing and
  removing `_`, so `uhf_panel_preset` → `UHFPanelPreset`, `x` → `X`, `sideSlipdeg` → `SideSlipDeg`.
  Do not rename fields beyond that, or the test fails.
- Keep the C field order and keep the header's comments (units, ranges) as Go comments.
- Bit/enum constants get a prefix of their type because the C++ enums reuse names (`Launch`,
  `OuterMarker`, `PROBEHEAT`, ... appear in several enums): `Light`, `Light2`, `Light3`, `Hsi`,
  `Blink`, `Power`, `Alt`, `Betty`, `Misc`, `Ecm`, `Tacan`, `Cmds`, `NavMode`, `FlyState`, `Str`, ...
- `StringID` uses `iota`. New IDs only get appended (in the header, *above* `StringIdentifier_DIM`),
  so append them at the end of the Go const block in the same order.

## Checklist (not covered by the layout test)

- [ ] Version constants: `FlightDataVersion`, `FlightData2Version`, `StringDataVersion`,
      `DrawingDataVersion`, and `#define` sizes (`RwrInfoSize`, `MaxCallsigns`, ...).
- [ ] New or changed enum values and bit flags, including the `AllLampBits*On` masks.
- [ ] New `StringID`s (the test does not see enums at all).
- [ ] Changes to the StringData/DrawingData serialization (`GetStringData()` in the header shows
      the format). Extend `parse_test.go` accordingly.
- [ ] Comments in the Go structs still match the header (units, value ranges, "Version N" groups).
- [ ] **A completely new shared memory area:** add the struct, an `Area...` name constant and a
      getter in `reader.go`, add the class name to `STRUCTS` in `gen_layout.py`, and add the type to
      the map in `layout_test.go`. Otherwise it is not checked at all.

## Pitfalls

- **gen_layout.py** compiles the header natively on Linux with `#define long int` so that `long` is
  4 bytes like on Windows. Everything else (`int`, `short`, `float`, alignment) is identical to
  MSVC x64. If the header ever uses `long long`, `wchar_t` (2 bytes on Windows, 4 on Linux) or
  `#pragma pack`, the generated layout can no longer be trusted. Check those fields by hand.
- **gen_layout.py** finds members with a simple parser: top-level `type name[...];` statements,
  skipping `enum {...}` blocks and methods. If the field count in the test looks wrong, first check
  whether the parser missed or invented a member (e.g. new syntax like bitfields `int a : 3;`, nested
  structs, or default initializers `int a = 0;`).
- `FlightData` is declared FINAL by BMS ("no further changes allowed"); new data goes into
  `FlightData2`. If `FlightData` ever changes anyway, treat it as a breaking change and look at
  the diff very carefully.
- BMS only appends fields. Older BMS versions provide shorter areas, so `Reader` copies
  `min(area size, struct size)` and leaves the rest zero. Never reorder or remove Go fields to
  "clean up".
- Never return pointers into the mapped memory. Getters return copies; BMS writes concurrently.
