package securitycode

import (
	"encoding/base64"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateAndCode(t *testing.T) {
	sc := NewSimpleSecCode()
	if _, err := sc.GenerateWithError(); err != nil {
		t.Fatalf("GenerateWithError: %v", err)
	}
	code := sc.GetSecCode()
	if len(code) != sc.SecCodeLength {
		t.Errorf("code length = %d, want %d", len(code), sc.SecCodeLength)
	}
	if !isAlphanumeric(code) {
		t.Errorf("code %q contains characters outside [a-zA-Z0-9]", code)
	}
	if len(sc.GetSecCodeKey()) != 32 {
		t.Errorf("key = %q, want 32-char md5 hex", sc.GetSecCodeKey())
	}
}

func TestGenerateProducesRandomCodes(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 20; i++ {
		code := NewSimpleSecCode().Generate().GetSecCode()
		seen[code] = true
	}
	// 20 draws of a 4-char [a-zA-Z0-9] code should not all be identical.
	if len(seen) == 1 {
		t.Errorf("captcha generation appears deterministic: %v", seen)
	}
}

func TestGetImgBase64(t *testing.T) {
	sc := NewSimpleSecCode()
	if _, err := sc.GenerateWithError(); err != nil {
		t.Fatal(err)
	}
	encoded, err := sc.GetImgBase64()
	if err != nil {
		t.Fatalf("GetImgBase64: %v", err)
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}
	img, err := png.Decode(strings.NewReader(string(raw)))
	if err != nil {
		t.Fatalf("png decode: %v", err)
	}
	if img.Bounds().Dx() != sc.imgWidth || img.Bounds().Dy() != sc.imgHeight {
		t.Errorf("image size = %dx%d, want %dx%d",
			img.Bounds().Dx(), img.Bounds().Dy(), sc.imgWidth, sc.imgHeight)
	}
}

func TestCreateImgFile(t *testing.T) {
	sc := NewSimpleSecCode()
	if _, err := sc.GenerateWithError(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "captcha.png")
	if err := sc.CreateImgFile(path); err != nil {
		t.Fatalf("CreateImgFile: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("output file missing: %v", err)
	}
}

func TestGetImgBase64BeforeGenerate(t *testing.T) {
	sc := NewSimpleSecCode()
	if _, err := sc.GetImgBase64(); err == nil {
		t.Error("expected error before Generate")
	}
	if err := sc.CreateImgFile(filepath.Join(t.TempDir(), "x.png")); err == nil {
		t.Error("expected error before Generate")
	}
}

func TestVerifyFunctions(t *testing.T) {
	// VerifySecCode ignores case.
	if !VerifySecCode("AbC", "aBc") {
		t.Error("VerifySecCode should ignore case")
	}
	// Different strings must not match.
	if VerifySecCode("abc", "abcd") {
		t.Error("VerifySecCode matched different strings")
	}
	if VerifySecCode("abc", "abd") {
		t.Error("VerifySecCode matched different strings")
	}
	// Case-sensitive comparison.
	if !VerifySecCodeCaseSensitive("AbC", "AbC") {
		t.Error("VerifySecCodeCaseSensitive should match identical strings")
	}
	if VerifySecCodeCaseSensitive("AbC", "abc") {
		t.Error("VerifySecCodeCaseSensitive should not ignore case")
	}
	// Deprecated alias now matches its name (case-insensitive).
	if !VerifySecCodeIgnoreCase("AbC", "aBc") {
		t.Error("VerifySecCodeIgnoreCase should ignore case")
	}
}

func TestCustomFontPath(t *testing.T) {
	sc := NewSimpleSecCode()
	sc.FontFilePath = filepath.Join(t.TempDir(), "missing.ttf")
	// A missing custom font must fall back to the embedded font, not fail.
	if _, err := sc.GenerateWithError(); err != nil {
		t.Fatalf("GenerateWithError with missing custom font: %v", err)
	}
}

func TestCharacterMap(t *testing.T) {
	sc := NewSimpleSecCode()
	if len(sc.characterMap) != 3 {
		t.Fatalf("characterMap groups = %d, want 3", len(sc.characterMap))
	}
	for i, group := range sc.characterMap {
		if len(group) == 0 {
			t.Errorf("characterMap[%d] is empty", i)
		}
	}
}

func isAlphanumeric(s string) bool {
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		default:
			return false
		}
	}
	return true
}
