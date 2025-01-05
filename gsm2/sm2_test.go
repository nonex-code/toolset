package gsm2

import (
	"log"
	"testing"
)

func TestSm2(t *testing.T) {
	pwd := []byte("")
	text := []byte("123")
	private, public, err := GerenateSM2Key(pwd)
	if err != nil {
		t.Error("生成随机密钥失败", err)
		return
	}
	sign := Sign(text, private, pwd)
	st, err := PublicKeyEncrypt(text, public)
	if err != nil {
		t.Error(err)
		return
	}

	ot, err := PrivateKeyDecrypt(st, private, pwd)
	if err != nil {
		t.Error(err)
		return
	}
	log.Println(string(ot))
	b := Verify(ot, sign, public)
	if !b {
		t.Error("验签名失败")
	}
	t.Log("验签结果：", b)
}
