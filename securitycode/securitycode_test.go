package securitycode

import (
	"strings"
	"testing"
)

func TestSecuritycode(t *testing.T) {
	secCode := NewSimpleSecCode()
	sc := secCode.Generate().GetSecCode()
	b := VerifySecCode(sc, sc)
	c := VerifySecCodeIgnoreCase(strings.ToLower(sc), sc)
	t.Log(b, c)
}
