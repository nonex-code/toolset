package gsm2

import (
	"crypto"
	"crypto/rand"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"
)

/*
	基于 "github.com/tjfoc/gmsm/gsm2" 简单的封装
*/
// GerenateSM2Key 生成公私钥
func GerenateSM2Key(pwd []byte) (private, public []byte, err error) {
	//1.生成sm2密钥对
	privateKey, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	//2.通过x509将私钥反序列化并进行pem编码
	privateKeyToPem, err := x509.WritePrivateKeyToPem(privateKey, pwd)
	if err != nil {
		return nil, nil, err
	}
	//3.进行SM2公钥断言
	publicKey := privateKey.Public().(*sm2.PublicKey)
	//4.将公钥通过x509序列化并进行pem编码
	publicKeyToPem, err := x509.WritePublicKeyToPem(publicKey)
	if err != nil {
		return nil, nil, err
	}
	// 5.返回公私钥 []byte 是否需要写文件自己决定
	return privateKeyToPem, publicKeyToPem, nil
}

// PublicKeyEncrypt 公钥加密 originalText 加密原文 publicKey 公钥  return  密文
func PublicKeyEncrypt(originalText []byte, publicKey []byte) ([]byte, error) {
	//1.将pem格式公钥解码并反序列化
	publicKeyFromPem, err := x509.ReadPublicKeyFromPem(publicKey)
	if err != nil {
		return nil, err
	}
	//2.加密
	secretText, err := publicKeyFromPem.EncryptAsn1(originalText, rand.Reader)
	if err != nil {
		return nil, err
	}
	return secretText, err
}

// PrivateKeyDecrypt 私钥解密
func PrivateKeyDecrypt(secretText []byte, privateKey []byte, pwd []byte) ([]byte, error) {
	//1.将pem格式私钥文件解码并反序列话
	privateKeyFromPem, err := x509.ReadPrivateKeyFromPem(privateKey, pwd)
	if err != nil {
		return nil, err
	}
	//2.解密
	originalText, err := privateKeyFromPem.DecryptAsn1(secretText)
	if err != nil {
		return nil, err
	}
	return originalText, err
}

// Sign 签名 originalText 签名原文 privateKey 私钥
func Sign(originalText []byte, privateKey []byte, pwd []byte) []byte {
	//1.将pem格式私钥文件解码并反序列话
	privateKeyFromPem, err := x509.ReadPrivateKeyFromPem(privateKey, pwd)
	if err != nil {
		panic(err)
	}
	//2.签名
	sign, err := privateKeyFromPem.Sign(rand.Reader, originalText, crypto.SHA256)
	if err != nil {
		panic(err)
	}
	return sign
}

// Verify 验签  originalText 密文 sign 摘要 publicKey 公钥
func Verify(secretText, sign, publicKey []byte) bool {
	//1.将pem格式公钥解码并反序列化
	publicKeyFromPem, err := x509.ReadPublicKeyFromPem(publicKey)
	if err != nil {
		panic(err)
	}
	//3.验签
	verify := publicKeyFromPem.Verify(secretText, sign)
	return verify
}
