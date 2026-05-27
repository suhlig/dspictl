package dspi

import (
	"encoding/binary"
	"fmt"
	"os"
)

const (
	// UF2BlockSize is the fixed block size for UF2 files.
	UF2BlockSize  = 512
	uf2MagicStart = 0x0A324655 // "UF2\n" in little-endian
	uf2MagicEnd   = 0x0AB16F30
)

// UF2 board family identifiers.
const (
	UF2FamilyRP2040   = 0xe48bff56
	UF2FamilyRP2350   = 0xe48bff57
	UF2FamilyRP2350V2 = 0xe48bff59
)

// UF2Info holds metadata extracted from a UF2 firmware file.
type UF2Info struct {
	BoardFamily uint32
	NumBlocks   uint32
}

// ParseUF2 reads a UF2 firmware file and extracts its metadata.
func ParseUF2(path string) (*UF2Info, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("reading UF2 file: %w", err)
	}

	if len(data) < UF2BlockSize {
		return nil, fmt.Errorf("UF2 file too small (%d bytes)", len(data))
	}

	if len(data)%UF2BlockSize != 0 {
		return nil, fmt.Errorf("UF2 file size %d is not a multiple of block size %d", len(data), UF2BlockSize)
	}

	blockCount := len(data) / UF2BlockSize

	// Validate magic numbers in every block.
	for i := range blockCount {
		off := i * UF2BlockSize

		if binary.LittleEndian.Uint32(data[off:off+4]) != uf2MagicStart {
			return nil, fmt.Errorf("block %d: invalid magic start", i)
		}

		if binary.LittleEndian.Uint32(data[off+508:off+512]) != uf2MagicEnd {
			return nil, fmt.Errorf("block %d: invalid magic end", i)
		}
	}

	numBlocks := binary.LittleEndian.Uint32(data[24:28])

	// The board family sits at offset 28–31 of every block in Pico SDK UF2 files.
	// Read it from the first block.
	family := binary.LittleEndian.Uint32(data[28:32])

	return &UF2Info{
		BoardFamily: family,
		NumBlocks:   numBlocks,
	}, nil
}

// PlatformForFamily maps a UF2 board family to a Platform.
func PlatformForFamily(family uint32) (Platform, error) {
	switch family {
	case UF2FamilyRP2040:
		return PlatformRP2040, nil
	case UF2FamilyRP2350, UF2FamilyRP2350V2:
		return PlatformRP2350, nil
	default:
		return 0, fmt.Errorf("unknown UF2 board family: 0x%08x", family)
	}
}

// FamilyForPlatform maps a Platform to its primary UF2 board family.
func FamilyForPlatform(p Platform) (uint32, error) {
	switch p {
	case PlatformRP2040:
		return UF2FamilyRP2040, nil
	case PlatformRP2350:
		return UF2FamilyRP2350V2, nil
	default:
		return 0, fmt.Errorf("unknown platform: %s", p)
	}
}
