package gaes

import (
	"testing"
)

func TestGaes(t *testing.T) {
	data := []byte("123")
	key := []byte("1111111111111111")
	ed, err := Encrypt(data, key)
	if err != nil {
		t.Error(err)
		return
	}
	dd, err := Decrypt(ed, key)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(dd))
}
