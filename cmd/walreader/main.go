package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/harhitosw/distrikiwi/internal/wal"
)

func main() {
	walPath := flag.String("wal", "distrikiwi.wal", "path to the WAL file")
	flag.Parse()

	entries, err := wal.ReadWALFile(*walPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read WAL %q: %v\n", *walPath, err)
		os.Exit(1)
	}

	fmt.Printf("WAL: %s\n", *walPath)
	fmt.Printf("Entries: %d\n", len(entries))
	if len(entries) == 0 {
		return
	}

	for index, entry := range entries {
		fmt.Printf("\n[%04d] %s\n", index+1, operationName(entry.OpType))
		fmt.Printf("  Key:   %q\n", entry.Key)
		fmt.Printf("  Value: %s\n", formatValue(entry.Value))
	}
}

func operationName(opType uint8) string {
	switch opType {
	case 0:
		return "PUT"
	case 1:
		return "DELETE"
	default:
		return fmt.Sprintf("UNKNOWN (op type %d)", opType)
	}
}

func formatValue(value []byte) string {
	if len(value) == 0 {
		return "<empty> (0 bytes)"
	}
	if isPrintableUTF8(value) {
		return fmt.Sprintf("%s (%d bytes)", strconv.Quote(string(value)), len(value))
	}
	return fmt.Sprintf("0x%s (%d bytes)", hex.EncodeToString(value), len(value))
}

func isPrintableUTF8(value []byte) bool {
	if !utf8.Valid(value) {
		return false
	}
	for _, character := range string(value) {
		if unicode.IsControl(character) && !strings.ContainsRune("\n\r\t", character) {
			return false
		}
	}
	return true
}