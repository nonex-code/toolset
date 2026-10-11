# toolset

Go 常用工具集：邮件发送、AES 加解密、SM2 国密、协程池、图片验证码。

| 模块 | 说明 |
| --- | --- |
| [email](./email) | SMTP 邮件发送（SSL/STARTTLS、HTML、附件、内嵌图片） |
| [gaes](./gaes) | AES 加解密（推荐 GCM，兼容随机 IV 的 CBC V2） |
| [gsm2](./gsm2) | SM2 密钥生成、加解密、签名验签 |
| [gpool](./gpool) | 固定大小协程池 |
| [securitycode](./securitycode) | 图片验证码 |

## sm2 示例

```go
pwd := []byte("")
text := []byte("123")
private, public, err := gsm2.GenerateSM2Key(pwd)
if err != nil {
    log.Println("生成随机密钥失败", err)
    return
}
sign, err := gsm2.Sign(text, private, pwd)
if err != nil {
    log.Println("签名失败", err)
    return
}
st, err := gsm2.PublicKeyEncrypt(text, public)
if err != nil {
    return
}

ot, err := gsm2.PrivateKeyDecrypt(st, private, pwd)
if err != nil {
    return
}
log.Println(string(ot))
b, err := gsm2.Verify(ot, sign, public)
if err != nil {
    log.Println("验签失败", err)
    return
}
if !b {
    log.Println("验签名失败")
}
log.Println("验签结果：", b)
```

## aes 示例

> 推荐使用 AES-GCM（认证加密，可检测篡改）。`Encrypt` / `Decrypt` 为兼容旧数据的废弃接口，使用固定 IV，请勿用于新数据。

### AES-GCM（推荐）

```go
data := []byte("123")
key := []byte("1111111111111111")
ed, err := gaes.EncryptGCM(data, key)
if err != nil {
    return
}
dd, err := gaes.DecryptGCM(ed, key)
if err != nil {
    return
}
log.Println(string(dd))
```

### AES-CBC + 随机 IV（V2）

```go
data := []byte("123")
key := []byte("1111111111111111")
ed, err := gaes.EncryptV2(data, key)
if err != nil {
    return
}
dd, err := gaes.DecryptV2(ed, key)
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

    task := func () {
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

```text
secCode := securitycode.NewSimpleSecCode()
sc := secCode.Generate().GetSecCode()
b := securitycode.VerifySecCode(sc, sc)                          // 不区分大小写
c := securitycode.VerifySecCodeCaseSensitive(sc, sc)             // 区分大小写
log.Println(b, c)
```

## 邮件示例

> 以下示例来自 `email` 包，实际使用时请加上包名前缀，例如 `email.NewMailer(...)`。
> SMTP 默认使用 SSL（端口 465）；如需明文连接并按需 STARTTLS，可设置 `WithSSL(false)`。
> 默认校验 TLS 证书，仅在明确需要时使用 `WithSkipTLSVerify(true)` 跳过。

### 1. 纯文本邮件

#### Option 模式

```go
mailer := NewMailer(
    WithSMTPConfig("smtp.example.com", "465", "your-email@example.com", "your-password"),
    WithFrom("your-email@example.com", "Your Name"),
    WithTo([]string{"recipient1@example.com", "recipient2@example.com"}),
    WithSubject("Test Plain Text Email"),
    WithBody("This is a plain text email."),
    WithSSL(true),
)
err := mailer.Send()
if err != nil {
fmt.Println("Error sending email:", err)
} else {
fmt.Println("Plain text email sent successfully")
}
```

#### 链式调用模式

```go
mailer := NewMailer().
    SetFrom("your-email@example.com", "Your Name").
    SetTo([]string{"recipient1@example.com", "recipient2@example.com"}).
    SetSubject("Test Plain Text Email").
    SetBody("This is a plain text email.").
    SetSSL(true)
err := mailer.Send()
if err != nil {
    fmt.Println("Error sending email:", err)
} else {
    fmt.Println("Plain text email sent successfully")
}

```

### 2. HTML 邮件

#### Option 模式

```go
mailer := NewMailer(
    WithSMTPConfig("smtp.example.com", "465", "your-email@example.com", "your-password"),
    WithFrom("your-email@example.com", "Your Name"),
    WithTo([]string{"recipient1@example.com", "recipient2@example.com"}),
    WithSubject("Test HTML Email"),
    WithHTMLBody("<html><body><h1>This is an HTML email</h1></body></html>"),
    WithSSL(true),
)
err := mailer.Send()
if err != nil {
    fmt.Println("Error sending email:", err)
} else {
    fmt.Println("HTML email sent successfully")
}

```

#### 链式调用模式

```go
mailer := NewMailer().
    SetFrom("your-email@example.com", "Your Name").
    SetTo([]string{"recipient1@example.com", "recipient2@example.com"}).
    SetSubject("Test HTML Email").
    SetHTMLBody("<html><body><h1>This is an HTML email</h1></body></html>").
    SetSSL(true)
err := mailer.Send()
if err != nil {
    fmt.Println("Error sending email:", err)
} else {
    fmt.Println("HTML email sent successfully")
}

```

### 3.带附件的邮件

#### Option 模式

```go
mailer := NewMailer(
    WithSMTPConfig("smtp.example.com", "465", "your-email@example.com", "your-password"),
    WithFrom("your-email@example.com", "Your Name"),
    WithTo([]string{"recipient1@example.com", "recipient2@example.com"}),
    WithSubject("Test Email with Attachments"),
    WithBody("This is a plain text email with attachments."),
    WithAttachments([]string{"./file1.txt", "./file2.txt"}),
    WithSSL(true),
)
err := mailer.Send()
if err != nil {
    fmt.Println("Error sending email:", err)
} else {
    fmt.Println("Email with attachments sent successfully")
}

```

#### 链式调用模式

```go
mailer := NewMailer().
    SetFrom("your-email@example.com", "Your Name").
    SetTo([]string{"recipient1@example.com", "recipient2@example.com"}).
    SetSubject("Test Email with Attachments").
    SetBody("This is a plain text email with attachments.").
    SetAttachments([]string{"./file1.txt", "./file2.txt"}).
    SetSSL(true)
err := mailer.Send()
if err != nil {
    fmt.Println("Error sending email:", err)
} else {
    fmt.Println("Email with attachments sent successfully")
}

```

### 4. 带内嵌图片的邮件

#### Option 模式

```go
mailer := NewMailer(
    WithSMTPConfig("smtp.example.com", "465", "your-email@example.com", "your-password"),
    WithFrom("your-email@example.com", "Your Name"),
    WithTo([]string{"recipient1@example.com", "recipient2@example.com"}),
    WithSubject("Test Email with Inline Images"),
    WithHTMLBody("<html><body><h1>This is an HTML email with an inline image</h1><img src=\"cid:image1\"></body></html>"),
    WithInlineImages(map[string]string{"image1": "./image.jpg"}),
    WithSSL(true),
)
err := mailer.Send()
if err != nil {
    fmt.Println("Error sending email:", err)
} else {
    fmt.Println("Email with inline images sent successfully")
}

```
#### 链式调用模式

```go
mailer := NewMailer().
    SetFrom("your-email@example.com", "Your Name").
    SetTo([]string{"recipient1@example.com", "recipient2@example.com"}).
    SetSubject("Test Email with Inline Images").
    SetHTMLBody("<html><body><h1>This is an HTML email with an inline image</h1><img src=\"cid:image1\"></body></html>").
    SetInlineImages(map[string]string{"image1": "./image.jpg"}).
    SetSSL(true)
err := mailer.Send()
if err != nil {
    fmt.Println("Error sending email:", err)
} else {
    fmt.Println("Email with inline images sent successfully")
}
```
