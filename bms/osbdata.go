package bms

// OsbStringLength is the size of an OSB label line (OSB_STRING_LENGTH).
const OsbStringLength = 8

// OsbLabel is a single MFD button label.
type OsbLabel struct {
	Line1    [OsbStringLength]byte
	Line2    [OsbStringLength]byte
	Inverted uint8 // C++ bool; use IsInverted
}

func (l *OsbLabel) Text1() string    { return cString(l.Line1[:]) }
func (l *OsbLabel) Text2() string    { return cString(l.Line2[:]) }
func (l *OsbLabel) IsInverted() bool { return l.Inverted != 0 }

// OSBData mirrors the "FalconSharedOsbMemoryArea" layout.
type OSBData struct {
	LeftMFD  [20]OsbLabel
	RightMFD [20]OsbLabel
}
