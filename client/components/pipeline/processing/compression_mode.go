package processing

import (
	"fmt"
	"strings"
)

// CompressionMode represents a compression mode.
//
// A compression mode is a hint to the compression type's constructor,
// it is not guaranteed that the returned compressor is as ideal as you'll hope,
// the implementation should however try to fit the requested mode
// to a possible configuration which respects the particalar mode as much as possible.
type CompressionMode uint8

const (
	// CompressionModeDefault is the enum constant which identifies
	// the default compression mode. What this means exactly
	// is up to the compression algorithm to define.
	CompressionModeDefault CompressionMode = iota
	// CompressionModeBestSpeed is the enum constant which identifies
	// the request to use the compression with a configuration,
	// which aims for the best possible speed, using that algorithm.
	//
	// What this exactly means is up to the compression type,
	// and it is entirely free to ignore this compression mode all together,
	// if it doesn't make sense for that type in particular.
	CompressionModeBestSpeed
	// CompressionModeBestCompression is the enum constant which identifies
	// the request to use the compression with a configuration,
	// which aims for the best possible compression (ratio), using that algorithm.
	//
	// What this exactly means is up to the compression type,
	// and it is entirely free to ignore this compression mode all together,
	// if it doesn't make sense for that type in particular.
	CompressionModeBestCompression
)

// String implements Stringer.String
func (cm CompressionMode) String() string {
	return _CompressionModeValueToStringMapping[cm]
}

var (
	_CompressionModeValueToStringMapping = map[CompressionMode]string{
		CompressionModeDefault:         _CompressionModeStrings[:7],
		CompressionModeBestSpeed:       _CompressionModeStrings[7:17],
		CompressionModeBestCompression: _CompressionModeStrings[17:],
	}
	_CompressionModeStringToValueMapping = map[string]CompressionMode{
		_CompressionModeStrings[:7]:   CompressionModeDefault,
		_CompressionModeStrings[7:17]: CompressionModeBestSpeed,
		_CompressionModeStrings[17:]:  CompressionModeBestCompression,
	}
)

const _CompressionModeStrings = "defaultbest_speedbest_compression"

// MarshalText implements encoding.TextMarshaler.MarshalText
func (cm CompressionMode) MarshalText() ([]byte, error) {
	str := cm.String()
	if str == "" {
		return nil, fmt.Errorf("'%s' is not a valid CompressionMode value", cm)
	}
	return []byte(str), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.UnmarshalText
func (cm *CompressionMode) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*cm = CompressionModeDefault
		return nil
	}

	var ok bool
	*cm, ok = _CompressionModeStringToValueMapping[strings.ToLower(string(text))]
	if !ok {
		return fmt.Errorf("'%s' is not a valid CompressionMode string", text)
	}
	return nil
}
