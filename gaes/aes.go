// Package gaes provides AES encryption helpers.
//
// Security guidance (安全建议):
//
//   - Prefer EncryptGCM / DecryptGCM: AES-GCM is an authenticated cipher (AEAD),
//     which also protects against tampering.
//     优先使用 EncryptGCM / DecryptGCM：AES-GCM 是认证加密，可检测数据篡改。
//   - EncryptV2 / DecryptV2 use AES-CBC with a random IV. The IV is prepended to
//     the ciphertext (format: IV || ciphertext). CBC has no integrity protection.
//     EncryptV2 / DecryptV2 使用带随机 IV 的 AES-CBC，IV 前置在密文中（格式：IV || 密文）。CBC 不提供完整性保护。
//   - Encrypt / Decrypt are kept only for backward compatibility and use a fixed
//     IV derived from the key. They are deprecated and insecure; do not use them
//     for new data.
//     Encrypt / Decrypt 仅为向后兼容保留，使用由密钥派生的固定 IV，已废弃且不安全，请勿用于新数据。
package gaes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// PKCS7 padding (PKCS#7 填充).

// pkcs7Padding appends PKCS#7 padding. The number of padding bytes is always in
// the range [1, blockSize].
// pkcs7Padding 追加 PKCS#7 填充，填充字节数始终在 [1, blockSize] 范围内。
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7UnPadding removes and validates PKCS#7 padding.
// pkcs7UnPadding 移除并校验 PKCS#7 填充。
func pkcs7UnPadding(data []byte, blockSize int) ([]byte, error) {
	length := len(data)
	if length == 0 || length%blockSize != 0 {
		return nil, errors.New("gaes: invalid ciphertext length")
	}
	unPadding := int(data[length-1])
	if unPadding == 0 || unPadding > blockSize || unPadding > length {
		return nil, errors.New("gaes: invalid padding")
	}
	for _, b := range data[length-unPadding:] {
		if int(b) != unPadding {
			return nil, errors.New("gaes: invalid padding")
		}
	}
	return data[:length-unPadding], nil
}

// ---- AES-GCM (recommended / 推荐) ------------------------------------------------------

// EncryptGCM encrypts data with AES-GCM. The returned value is nonce || ciphertext
// and includes an authentication tag, so tampering is detected during decryption.
// EncryptGCM 使用 AES-GCM 加密，返回 nonce || 密文，包含认证标签，解密时可检测篡改。
func EncryptGCM(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("gaes: generate nonce: %w", err)
	}
	return gcm.Seal(nonce, nonce, data, nil), nil
}

// DecryptGCM decrypts data produced by EncryptGCM.
// DecryptGCM 解密由 EncryptGCM 生成的数据。
func DecryptGCM(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("gaes: ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// EncryptGCMBase64 is EncryptGCM followed by standard base64 encoding.
// EncryptGCMBase64 在 EncryptGCM 基础上进行标准 base64 编码。
func EncryptGCMBase64(data, key []byte) (string, error) {
	encrypted, err := EncryptGCM(data, key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptGCMBase64 decodes base64 and decrypts with AES-GCM.
// DecryptGCMBase64 先 base64 解码，再进行 AES-GCM 解密。
func DecryptGCMBase64(encoded string, key []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("gaes: base64 decode: %w", err)
	}
	return DecryptGCM(data, key)
}

// ---- AES-CBC with random IV (V2) -------------------------------------------------------

// EncryptV2 encrypts data with AES-CBC and a random IV. The IV is prepended to
// the ciphertext: the result is IV || ciphertext. Unlike Encrypt, the same
// plaintext encrypts to different ciphertext each time.
//
// Note: CBC does not authenticate the ciphertext. Use EncryptGCM when integrity
// matters.
// EncryptV2 使用随机 IV 的 AES-CBC 加密，返回 IV || 密文。与 Encrypt 不同，相同明文每次加密结果不同。
// 注意：CBC 不认证密文，需要完整性保护时请使用 EncryptGCM。
func EncryptV2(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	iv := make([]byte, blockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("gaes: generate iv: %w", err)
	}

	padded := pkcs7Padding(data, blockSize)
	crypted := make([]byte, blockSize+len(padded))
	copy(crypted, iv)
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(crypted[blockSize:], padded)
	return crypted, nil
}

// DecryptV2 decrypts data produced by EncryptV2. It validates the length and
// padding and never panics on malformed input.
// DecryptV2 解密由 EncryptV2 生成的数据，会校验长度与填充，对非法输入不会 panic。
func DecryptV2(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	if len(data) < blockSize*2 || (len(data)-blockSize)%blockSize != 0 {
		return nil, errors.New("gaes: invalid ciphertext length")
	}

	iv := data[:blockSize]
	ciphertext := data[blockSize:]
	decrypted := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(decrypted, ciphertext)
	return pkcs7UnPadding(decrypted, blockSize)
}

// EncryptV2Base64 is EncryptV2 followed by standard base64 encoding.
// EncryptV2Base64 在 EncryptV2 基础上进行标准 base64 编码。
func EncryptV2Base64(data, key []byte) (string, error) {
	encrypted, err := EncryptV2(data, key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptV2Base64 decodes base64 and decrypts with EncryptV2's format.
// DecryptV2Base64 先 base64 解码，再按 EncryptV2 的格式解密。
func DecryptV2Base64(encoded string, key []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("gaes: base64 decode: %w", err)
	}
	return DecryptV2(data, key)
}

// ---- Legacy (deprecated / 已废弃) ------------------------------------------------------

// Encrypt encrypts data with AES-CBC using a fixed IV derived from the key.
//
// Deprecated: the fixed IV makes encryption deterministic and leaks plaintext
// patterns. Use EncryptGCM (recommended) or EncryptV2 instead.
//
// Encrypt 使用由密钥派生的固定 IV 进行 AES-CBC 加密。
//
// 已废弃：固定 IV 会导致确定性加密并泄露明文模式。请改用 EncryptGCM（推荐）或 EncryptV2。
func Encrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	encryptBytes := pkcs7Padding(data, blockSize)
	crypted := make([]byte, len(encryptBytes))
	cipher.NewCBCEncrypter(block, key[:blockSize]).CryptBlocks(crypted, encryptBytes)
	return crypted, nil
}

// Decrypt decrypts data produced by Encrypt.
//
// Deprecated: kept for backward compatibility with data encrypted by the old
// implementation. Use DecryptGCM or DecryptV2 for new data.
//
// Decrypt 解密由 Encrypt 生成的数据。
//
// 已废弃：仅为兼容旧实现加密的数据而保留。新数据请使用 DecryptGCM 或 DecryptV2。
func Decrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, errors.New("gaes: invalid ciphertext length")
	}
	decrypted := make([]byte, len(data))
	cipher.NewCBCDecrypter(block, key[:blockSize]).CryptBlocks(decrypted, data)
	return pkcs7UnPadding(decrypted, blockSize)
}
