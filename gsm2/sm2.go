package gsm2

import (
	"crypto"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"
)

// Package gsm2 is a small wrapper around "github.com/tjfoc/gmsm".
// 基于 "github.com/tjfoc/gmsm" 的简单封装。

// GenerateSM2Key generates an SM2 key pair and returns PEM-encoded private and
// public keys. pwd is an optional password used to encrypt the private key; pass
// nil or an empty slice for no encryption.
// GenerateSM2Key 生成 SM2 密钥对，返回 PEM 编码的私钥与公钥。pwd 为可选的私钥加密口令，nil 或空表示不加密。
func GenerateSM2Key(pwd []byte) (private, public []byte, err error) {
	// 1. Generate the SM2 key pair.
	privateKey, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("gsm2: generate key: %w", err)
	}
	// 2. Serialize and PEM-encode the private key.
	privateKeyToPem, err := x509.WritePrivateKeyToPem(privateKey, pwd)
	if err != nil {
		return nil, nil, fmt.Errorf("gsm2: encode private key: %w", err)
	}
	// 3. Derive the public key.
	publicKey, ok := privateKey.Public().(*sm2.PublicKey)
	if !ok {
		return nil, nil, errors.New("gsm2: unexpected public key type")
	}
	// 4. Serialize and PEM-encode the public key.
	publicKeyToPem, err := x509.WritePublicKeyToPem(publicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("gsm2: encode public key: %w", err)
	}
	return privateKeyToPem, publicKeyToPem, nil
}

// GerenateSM2Key is a deprecated, misspelled alias of GenerateSM2Key.
//
// Deprecated: use GenerateSM2Key.
//
// GerenateSM2Key 是拼写错误的 GenerateSM2Key 别名，已废弃，请使用 GenerateSM2Key。
func GerenateSM2Key(pwd []byte) (private, public []byte, err error) {
	return GenerateSM2Key(pwd)
}

// PublicKeyEncrypt encrypts originalText with the PEM-encoded public key.
// PublicKeyEncrypt 使用 PEM 编码的公钥加密 originalText。
func PublicKeyEncrypt(originalText, publicKey []byte) ([]byte, error) {
	publicKeyFromPem, err := x509.ReadPublicKeyFromPem(publicKey)
	if err != nil {
		return nil, fmt.Errorf("gsm2: read public key: %w", err)
	}
	secretText, err := publicKeyFromPem.EncryptAsn1(originalText, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("gsm2: encrypt: %w", err)
	}
	return secretText, nil
}

// PrivateKeyDecrypt decrypts secretText with the PEM-encoded private key.
// PrivateKeyDecrypt 使用 PEM 编码的私钥解密 secretText。
func PrivateKeyDecrypt(secretText, privateKey, pwd []byte) ([]byte, error) {
	privateKeyFromPem, err := x509.ReadPrivateKeyFromPem(privateKey, pwd)
	if err != nil {
		return nil, fmt.Errorf("gsm2: read private key: %w", err)
	}
	originalText, err := privateKeyFromPem.DecryptAsn1(secretText)
	if err != nil {
		return nil, fmt.Errorf("gsm2: decrypt: %w", err)
	}
	return originalText, nil
}

// Sign signs originalText with the PEM-encoded private key using SHA-256.
// Unlike the previous version it returns an error instead of panicking.
// Sign 使用 PEM 编码的私钥并采用 SHA-256 对 originalText 签名。出错时返回 error 而不是 panic。
func Sign(originalText, privateKey, pwd []byte) ([]byte, error) {
	privateKeyFromPem, err := x509.ReadPrivateKeyFromPem(privateKey, pwd)
	if err != nil {
		return nil, fmt.Errorf("gsm2: read private key: %w", err)
	}
	sign, err := privateKeyFromPem.Sign(rand.Reader, originalText, crypto.SHA256)
	if err != nil {
		return nil, fmt.Errorf("gsm2: sign: %w", err)
	}
	return sign, nil
}

// Verify verifies sign against originalText with the PEM-encoded public key.
// It returns an error instead of panicking when the key cannot be parsed.
// Verify 使用 PEM 编码的公钥验证 originalText 的签名 sign。公钥无法解析时返回 error 而不是 panic。
func Verify(originalText, sign, publicKey []byte) (bool, error) {
	publicKeyFromPem, err := x509.ReadPublicKeyFromPem(publicKey)
	if err != nil {
		return false, fmt.Errorf("gsm2: read public key: %w", err)
	}
	return publicKeyFromPem.Verify(originalText, sign), nil
}
