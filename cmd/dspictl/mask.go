package main

import (
	"fmt"
	"strconv"
	"strings"
)

// formatMask returns a human-readable list of active bits from a uint16 mask.
// Examples: "none", "0, 1", "0-3, 7", "all (16)".
func formatMaskU16(mask uint16, maxBits int) string {
	return formatMask(uint64(mask), maxBits)
}

// formatMaskU8 returns a human-readable list of active bits from a uint8 mask.
func formatMaskU8(mask uint8, maxBits int) string {
	return formatMask(uint64(mask), maxBits)
}

func formatMask(m uint64, maxBits int) string {
	bits := maskBits(m, maxBits)
	if len(bits) == 0 {
		return "none"
	}
	if len(bits) == maxBits {
		return fmt.Sprintf("all (%d)", maxBits)
	}

	// Collapse consecutive ranges for compact display.
	return collapseRanges(bits)
}

// maskBits returns the sorted list of set bit positions.
func maskBits(m uint64, maxBits int) []int {
	var bits []int
	for i := range maxBits {
		if m&(1<<uint(i)) != 0 {
			bits = append(bits, i)
		}
	}
	return bits
}

// collapseRanges runs a slice of sorted ints into compact ranges.
// E.g. [0,1,2,5,6,8] → "0-2, 5-6, 8".
func collapseRanges(nums []int) string {
	if len(nums) == 0 {
		return "none"
	}

	var parts []string
	start := nums[0]
	end := nums[0]

	for _, n := range nums[1:] {
		if n == end+1 {
			end = n
		} else {
			parts = append(parts, rangeStr(start, end))
			start = n
			end = n
		}
	}
	parts = append(parts, rangeStr(start, end))

	return strings.Join(parts, ", ")
}

func rangeStr(a, b int) string {
	if a == b {
		return strconv.Itoa(a)
	}
	return fmt.Sprintf("%d-%d", a, b)
}

// parseMaskBits parses channel/pair numbers from args and returns a uint16 mask.
// Each arg must be a decimal integer in [0, maxBits).
func parseMaskBits(args []string, maxBits int) (uint16, error) {
	var mask uint16
	seen := make(map[int]bool)

	for _, arg := range args {
		n, err := strconv.Atoi(arg)
		if err != nil {
			return 0, fmt.Errorf("invalid channel number %q: %w", arg, err)
		}
		if n < 0 || n >= maxBits {
			return 0, fmt.Errorf("channel %d out of range (0-%d)", n, maxBits-1)
		}
		if seen[n] {
			continue // silently ignore duplicates
		}
		seen[n] = true
		mask |= 1 << uint(n)
	}

	return mask, nil
}

// parseMaskBitsU8 parses channel/pair numbers from args and returns a uint8 mask.
func parseMaskBitsU8(args []string, maxBits int) (uint8, error) {
	v, err := parseMaskBits(args, maxBits)
	return uint8(v), err
}

// maskSetBits returns a new mask with the given bits set.
func maskSetBits(current uint16, args []string, maxBits int) (uint16, error) {
	bits, err := parseMaskBits(args, maxBits)
	if err != nil {
		return 0, err
	}
	return current | bits, nil
}

// maskClearBits returns a new mask with the given bits cleared.
func maskClearBits(current uint16, args []string, maxBits int) (uint16, error) {
	bits, err := parseMaskBits(args, maxBits)
	if err != nil {
		return 0, err
	}
	return current &^ bits, nil
}

// maskSetBitsU8 returns a new uint8 mask with the given bits set.
func maskSetBitsU8(current uint8, args []string, maxBits int) (uint8, error) {
	bits, err := parseMaskBitsU8(args, maxBits)
	if err != nil {
		return 0, err
	}
	return current | bits, nil
}

// maskClearBitsU8 returns a new uint8 mask with the given bits cleared.
func maskClearBitsU8(current uint8, args []string, maxBits int) (uint8, error) {
	bits, err := parseMaskBitsU8(args, maxBits)
	if err != nil {
		return 0, err
	}
	return current &^ bits, nil
}
