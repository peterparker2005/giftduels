package boc

import (
	"encoding/base64"
	"fmt"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

// EncodeStringAsBOC принимает произвольную строку и возвращает её
// в виде base64-url-закодированного BOC.
func EncodeStringAsBOC(raw string) (string, error) {
	builder := cell.BeginCell().MustStoreStringSnake(raw)
	c := builder.EndCell()

	bocBytes := c.ToBOC()

	encoded := base64.URLEncoding.EncodeToString(bocBytes)
	return encoded, nil
}

// DecodeStringFromBOC принимает base64-url-закодированный BOC
// и извлекает из него строку.
func DecodeStringFromBOC(encoded string) (string, error) {
	bocBytes, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	c, err := cell.FromBOC(bocBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse BOC: %w", err)
	}

	loader := c.BeginParse()
	decoded, err := loader.LoadStringSnake()
	if err != nil {
		return "", fmt.Errorf("failed to read string from BOC: %w", err)
	}

	return decoded, nil
}

func EncodeBOCAsBase64(bocBytes []byte) string {
	return base64.URLEncoding.EncodeToString(bocBytes)
}
