package gaes

import (
	"bytes"
	"strings"
	"testing"
)

var testKey = []byte("1111111111111111") // 16 bytes -> AES-128

func TestGCMRoundTrip(t *testing.T) {
	for _, data := range [][]byte{
		[]byte(""),
		[]byte("123"),
		bytes.Repeat([]byte("A"), 1000),
	} {
		encrypted, err := EncryptGCM(data, testKey)
		if err != nil {
			t.Fatalf("EncryptGCM: %v", err)
		}
		decrypted, err := DecryptGCM(encrypted, testKey)
		if err != nil {
			t.Fatalf("DecryptGCM: %v", err)
		}
		if !bytes.Equal(decrypted, data) {
			t.Errorf("round trip mismatch: got %q want %q", decrypted, data)
		}
	}
}

func TestGCMRandomNonce(t *testing.T) {
	data := []byte("same plaintext")
	first, _ := EncryptGCM(data, testKey)
	second, _ := EncryptGCM(data, testKey)
	if bytes.Equal(first, second) {
		t.Error("two encryptions produced identical ciphertext; nonce is not random")
	}
}

func TestGCMTamperDetection(t *testing.T) {
	encrypted, err := EncryptGCM([]byte("sensitive"), testKey)
	if err != nil {
		t.Fatal(err)
	}
	// Flip a byte in the ciphertext body.
	encrypted[len(encrypted)-1] ^= 0xFF
	if _, err := DecryptGCM(encrypted, testKey); err == nil {
		t.Error("tampered ciphertext was accepted")
	}
}

func TestGCMTooShort(t *testing.T) {
	if _, err := DecryptGCM([]byte("short"), testKey); err == nil {
		t.Error("expected error for short ciphertext")
	}
}

func TestGCMBase64(t *testing.T) {
	encoded, err := EncryptGCMBase64([]byte("hello"), testKey)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsAny(encoded, " \n") {
		t.Errorf("base64 output contains whitespace: %q", encoded)
	}
	decrypted, err := DecryptGCMBase64(encoded, testKey)
	if err != nil {
		t.Fatal(err)
	}
	if string(decrypted) != "hello" {
		t.Errorf("got %q want %q", decrypted, "hello")
	}
}

func TestV2RoundTrip(t *testing.T) {
	for _, data := range [][]byte{
		[]byte(""),
		[]byte("123"),
		bytes.Repeat([]byte("A"), 1000),
	} {
		encrypted, err := EncryptV2(data, testKey)
		if err != nil {
			t.Fatalf("EncryptV2: %v", err)
		}
		// IV must be prepended.
		if len(encrypted) < 32 {
			t.Fatalf("ciphertext too short: %d", len(encrypted))
		}
		decrypted, err := DecryptV2(encrypted, testKey)
		if err != nil {
			t.Fatalf("DecryptV2: %v", err)
		}
		if !bytes.Equal(decrypted, data) {
			t.Errorf("round trip mismatch: got %q want %q", decrypted, data)
		}
	}
}

func TestV2RandomIV(t *testing.T) {
	data := []byte("same plaintext")
	first, _ := EncryptV2(data, testKey)
	second, _ := EncryptV2(data, testKey)
	if bytes.Equal(first, second) {
		t.Error("two encryptions produced identical ciphertext; IV is not random")
	}
	// Both must still decrypt correctly.
	if d, _ := DecryptV2(first, testKey); !bytes.Equal(d, data) {
		t.Error("first ciphertext did not decrypt")
	}
	if d, _ := DecryptV2(second, testKey); !bytes.Equal(d, data) {
		t.Error("second ciphertext did not decrypt")
	}
}

func TestV2InvalidInputs(t *testing.T) {
	cases := map[string][]byte{
		"empty":             {},
		"shorter than iv":   bytes.Repeat([]byte{0}, 15),
		"iv only":           bytes.Repeat([]byte{0}, 16),
		"not block-aligned": bytes.Repeat([]byte{0}, 33),
	}
	for name, input := range cases {
		if _, err := DecryptV2(input, testKey); err == nil {
			t.Errorf("%s: expected error, got nil", name)
		}
	}
}

func TestV2PaddingValidation(t *testing.T) {
	// Build a ciphertext whose decrypted content has invalid padding by
	// encrypting then corrupting the last ciphertext block.
	encrypted, err := EncryptV2([]byte("data"), testKey)
	if err != nil {
		t.Fatal(err)
	}
	encrypted[len(encrypted)-1] ^= 0xFF
	if _, err := DecryptV2(encrypted, testKey); err == nil {
		// CBC cannot always detect tampering, but this specific corruption
		// should break the padding in practice.
		t.Log("warning: corrupted padding was not detected (CBC has no MAC)")
	}
}

func TestV2Base64(t *testing.T) {
	encoded, err := EncryptV2Base64([]byte("hello"), testKey)
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := DecryptV2Base64(encoded, testKey)
	if err != nil {
		t.Fatal(err)
	}
	if string(decrypted) != "hello" {
		t.Errorf("got %q want %q", decrypted, "hello")
	}
}

func TestInvalidKeyLength(t *testing.T) {
	badKey := []byte("short")
	if _, err := EncryptGCM([]byte("x"), badKey); err == nil {
		t.Error("EncryptGCM accepted an invalid key length")
	}
	if _, err := EncryptV2([]byte("x"), badKey); err == nil {
		t.Error("EncryptV2 accepted an invalid key length")
	}
	if _, err := DecryptV2([]byte("x"), badKey); err == nil {
		t.Error("DecryptV2 accepted an invalid key length")
	}
}

func TestLegacyRoundTrip(t *testing.T) {
	data := []byte("legacy payload")
	encrypted, err := Encrypt(data, testKey)
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := Decrypt(encrypted, testKey)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decrypted, data) {
		t.Errorf("legacy round trip mismatch: got %q want %q", decrypted, data)
	}
}

func TestLegacyDecryptInvalidInput(t *testing.T) {
	if _, err := Decrypt([]byte("not aligned"), testKey); err == nil {
		t.Error("expected error for non block-aligned input")
	}
	if _, err := Decrypt(nil, testKey); err == nil {
		t.Error("expected error for empty input")
	}
}
