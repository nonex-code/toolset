package gsm2

import (
	"bytes"
	"testing"
)

func TestGenerateEncryptDecrypt(t *testing.T) {
	pwd := []byte("")
	text := []byte("123")

	private, public, err := GenerateSM2Key(pwd)
	if err != nil {
		t.Fatalf("GenerateSM2Key: %v", err)
	}
	if len(private) == 0 || len(public) == 0 {
		t.Fatal("generated empty key material")
	}

	encrypted, err := PublicKeyEncrypt(text, public)
	if err != nil {
		t.Fatalf("PublicKeyEncrypt: %v", err)
	}
	decrypted, err := PrivateKeyDecrypt(encrypted, private, pwd)
	if err != nil {
		t.Fatalf("PrivateKeyDecrypt: %v", err)
	}
	if !bytes.Equal(decrypted, text) {
		t.Errorf("decrypted = %q, want %q", decrypted, text)
	}
}

func TestSignVerify(t *testing.T) {
	pwd := []byte("")
	text := []byte("123")

	private, public, err := GenerateSM2Key(pwd)
	if err != nil {
		t.Fatal(err)
	}

	sign, err := Sign(text, private, pwd)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	ok, err := Verify(text, sign, public)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !ok {
		t.Error("valid signature rejected")
	}

	// A different message must not verify against the same signature.
	ok, err = Verify([]byte("124"), sign, public)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if ok {
		t.Error("signature verified for a different message")
	}
}

func TestSignWithPassword(t *testing.T) {
	pwd := []byte("secret")
	text := []byte("payload")

	private, public, err := GenerateSM2Key(pwd)
	if err != nil {
		t.Fatal(err)
	}
	sign, err := Sign(text, private, pwd)
	if err != nil {
		t.Fatalf("Sign with password: %v", err)
	}
	ok, err := Verify(text, sign, public)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("signature with password rejected")
	}

	// Wrong password must produce an error, not a panic.
	if _, err := Sign(text, private, []byte("wrong")); err == nil {
		t.Error("expected error when signing with a wrong password")
	}
}

func TestInvalidKeysReturnErrors(t *testing.T) {
	if _, err := Sign([]byte("x"), []byte("not-a-pem"), nil); err == nil {
		t.Error("Sign accepted an invalid private key")
	}
	if _, err := Verify([]byte("x"), []byte("sig"), []byte("not-a-pem")); err == nil {
		t.Error("Verify accepted an invalid public key")
	}
	if _, err := PublicKeyEncrypt([]byte("x"), []byte("not-a-pem")); err == nil {
		t.Error("PublicKeyEncrypt accepted an invalid public key")
	}
	if _, err := PrivateKeyDecrypt([]byte("x"), []byte("not-a-pem"), nil); err == nil {
		t.Error("PrivateKeyDecrypt accepted an invalid private key")
	}
}

func TestDeprecatedAlias(t *testing.T) {
	private, public, err := GerenateSM2Key(nil)
	if err != nil {
		t.Fatalf("GerenateSM2Key: %v", err)
	}
	if len(private) == 0 || len(public) == 0 {
		t.Fatal("deprecated alias generated empty key material")
	}
}
