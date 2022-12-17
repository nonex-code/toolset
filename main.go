package main

import (
	"github.com/nonex-code/toolset/gaes"
	"github.com/nonex-code/toolset/gpool"
	"github.com/nonex-code/toolset/gsm2"
	"log"
	"time"
)

func main() {
	//test_sm2()
	//test_aes()
	tast_gpool()
}
func test_sm2() {
	pwd := []byte("")
	text := []byte("123")
	private, public, err := gsm2.GerenateSM2Key(pwd)
	if err != nil {
		log.Println("生成随机密钥失败", err)
		return
	}
	sign := gsm2.Sign(text, private, pwd)
	st, err := gsm2.PublicKeyEncrypt(text, public)
	if err != nil {
		return
	}

	ot, err := gsm2.PrivateKeyDecrypt(st, private, pwd)
	if err != nil {
		return
	}
	log.Println(string(ot))
	b := gsm2.Verify(ot, sign, public)
	if !b {
		log.Println("验签名失败")
	}
	log.Println("验签结果：", b)
}

func test_aes() {
	data := []byte("123")
	key := []byte("1111111111111111")
	ed, err := gaes.Encrypt(data, key)
	if err != nil {
		return
	}
	dd, err := gaes.Decrypt(ed, key)
	if err != nil {
		return
	}
	log.Println(string(dd))
}

func tast_gpool() {
	pool := gpool.NewTaskPool(1000)
	for i := 0; i < 100000; i++ {
		v := i

		task := func() {
			log.Println(v)
			time.Sleep(time.Second * 1)
		}
		err := pool.Submit(task)
		if err != nil {
			return
		}

	}
	pool.Close()

}
