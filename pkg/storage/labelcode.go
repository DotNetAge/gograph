package storage

import (
	"fmt"
	"strings"
	"sync"
)

// LabelCoder manages short codes for label names to reduce index key size.
// It assigns sequential single-character codes (a-z, A-Z, 0-9, then multi-char)
// to labels as they are first encountered.
type LabelCoder struct {
	mu      sync.RWMutex
	codeToLabel map[string]string
	labelToCode map[string]string
	counter     int
}

// NewLabelCoder creates a new LabelCoder with optional pre-defined mappings.
func NewLabelCoder() *LabelCoder {
	return &LabelCoder{
		codeToLabel: make(map[string]string),
		labelToCode: make(map[string]string),
	}
}

// nextCode generates the next available short code.
func (lc *LabelCoder) nextCode() string {
	lc.counter++
	n := lc.counter

	// Use single chars first: a-z (26), A-Z (26), 0-9 (10) = 62 total
	if n <= 26 {
		return string(rune('a' + n - 1))
	}
	n -= 26
	if n <= 26 {
		return string(rune('A' + n - 1))
	}
	n -= 26
	if n <= 10 {
		return string(rune('0' + n - 1))
	}

	// Fall back to base-62 multi-char encoding
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var sb strings.Builder
	n-- // zero-based for multi-char
	for n >= 0 {
		sb.WriteByte(chars[n%62])
		n = n/62 - 1
		if n < 0 {
			break
		}
	}
	// Reverse since we built least-significant first
	runes := []rune(sb.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Encode returns the short code for a label, creating one if necessary.
func (lc *LabelCoder) Encode(label string) string {
	lc.mu.RLock()
	code, ok := lc.labelToCode[label]
	lc.mu.RUnlock()
	if ok {
		return code
	}

	lc.mu.Lock()
	defer lc.mu.Unlock()
	// Double-check after acquiring write lock
	if code, ok := lc.labelToCode[label]; ok {
		return code
	}
	code = lc.nextCode()
	lc.labelToCode[label] = code
	lc.codeToLabel[code] = label
	return code
}

// Decode returns the original label for a short code.
func (lc *LabelCoder) Decode(code string) (string, bool) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	label, ok := lc.codeToLabel[code]
	return label, ok
}

// globalCoder is the package-level label coder used by all key functions.
var globalCoder = NewLabelCoder()

// SetLabelCoder replaces the global label coder (mainly for testing).
func SetLabelCoder(coder *LabelCoder) {
	globalCoder = coder
}

// EncodeLabel returns the short code for a label using the global coder.
func EncodeLabel(label string) string {
	return globalCoder.Encode(label)
}

// DecodeLabel returns the original label for a short code using the global coder.
func DecodeLabel(code string) (string, bool) {
	return globalCoder.Decode(code)
}

// ResetLabelCoder resets the global label coder to a fresh state (for testing).
func ResetLabelCoder() {
	globalCoder = NewLabelCoder()
}

// EncodeLabelKey encodes a label into a key-safe short code with a prefix.
// This ensures encoded keys are distinguishable from raw labels.
func EncodeLabelKey(label string) string {
	return "L" + EncodeLabel(label)
}

// DecodeLabelKey decodes a key-safe short code back to the original label.
func DecodeLabelKey(key string) (string, bool) {
	if !strings.HasPrefix(key, "L") {
		return "", false
	}
	return DecodeLabel(key[1:])
}

// MustDecodeLabelKey is like DecodeLabelKey but panics on failure.
func MustDecodeLabelKey(key string) string {
	label, ok := DecodeLabelKey(key)
	if !ok {
		panic(fmt.Sprintf("invalid encoded label key: %s", key))
	}
	return label
}
