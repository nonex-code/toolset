# toolset
## httpd 示例
```go
// 示例代码
client := &http.Client{Timeout: 20 * time.Second}
request := httpd.NewRequest(client, nil, nil)
```

## sm2 示例
```go
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
```
## aes 示例
```go
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
```
## gpool示例
```go
pool := gpool.NewTaskPool(10)
for i := 0; i < 100; i++ {
    v := i

    task := func() {
        log.Println(v)
        //time.Sleep(time.Second * 1)
    }
    err := pool.Submit(task)
    if err != nil {
        return
    }

}
pool.Close()
```
## 验证码示例
```go
secCode := securitycode.NewSimpleSecCode()
sc := secCode.Generate().GetSecCode()
b := securitycode.VerifySecCode(sc, sc)
c := securitycode.VerifySecCodeIgnoreCase(strings.ToLower(sc), sc)
log.Println(b, c)
```