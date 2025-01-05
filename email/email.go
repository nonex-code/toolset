package email

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
)

// Mailer represents an email client for sending emails.
// Mailer 表示用于发送邮件的邮件客户端。
type Mailer struct {
	SMTPHost     string            // SMTP server address (SMTP 服务器地址)
	SMTPPort     string            // SMTP server port (SMTP 服务器端口)
	SMTPUser     string            // Email account (邮箱账号)
	SMTPPassword string            // Email password (邮箱密码)
	From         string            // Sender email address (发件人邮箱地址)
	FromAlias    string            // Sender alias (发件人别名)
	To           []string          // List of recipient email addresses (收件人邮箱地址列表)
	Cc           []string          // List of CC email addresses (抄送邮箱地址列表)
	Bcc          []string          // List of BCC email addresses (密送邮箱地址列表)
	Subject      string            // Email subject (邮件主题)
	Body         string            // Plain text body (纯文本正文)
	HTMLBody     string            // HTML body (HTML 正文)
	Attachments  []string          // List of attachment file paths (附件文件路径列表)
	InlineImages map[string]string // Inline images: key is Content-ID, value is file path (内嵌图片：key 为 Content-ID，value 为文件路径)
	UseSSL       bool              // Whether to use SSL (是否使用 SSL)
}

// NewMailer creates a new Mailer with optional configurations.
// NewMailer 创建一个新的 Mailer，支持可选配置。
//
// Example (Option pattern):
// 示例（Option 模式）：
// mailer := NewMailer(
//
//	WithSMTPConfig("smtp.example.com", "465", "your-email@example.com", "your-password"),
//	WithFrom("your-email@example.com", "Your Name"),
//	WithTo([]string{"recipient1@example.com", "recipient2@example.com"}),
//	WithCc([]string{"cc@example.com"}),
//	WithBcc([]string{"bcc@example.com"}),
//	WithSubject("Test Email"),
//	WithBody("This is a plain text email."),
//	WithHTMLBody("<html><body><h1>This is an HTML email</h1></body></html>"),
//	WithAttachments([]string{"./file1.txt", "./file2.txt"}),
//	WithInlineImages(map[string]string{"image1": "./image.jpg"}),
//	WithSSL(true),
//
// )
//
// Example (Fluent interface):
// 示例（链式调用模式）：
// mailer := NewMailer().
//
//	SetFrom("your-email@example.com", "Your Name").
//	SetTo([]string{"recipient1@example.com", "recipient2@example.com"}).
//	SetCc([]string{"cc@example.com"}).
//	SetBcc([]string{"bcc@example.com"}).
//	SetSubject("Test Email").
//	SetBody("This is a plain text email.").
//	SetHTMLBody("<html><body><h1>This is an HTML email</h1></body></html>").
//	SetAttachments([]string{"./file1.txt", "./file2.txt"}).
//	SetInlineImages(map[string]string{"image1": "./image.jpg"}).
//	SetSSL(true)
func NewMailer(options ...func(*Mailer)) *Mailer {
	mailer := &Mailer{
		SMTPHost:     "smtp.example.com", // Default SMTP server (默认 SMTP 服务器)
		SMTPPort:     "465",              // Default SSL port (默认 SSL 端口)
		SMTPUser:     "your-email@example.com",
		SMTPPassword: "your-password",
		InlineImages: make(map[string]string),
		UseSSL:       true, // Default to use SSL (默认使用 SSL)
	}
	for _, option := range options {
		option(mailer)
	}
	return mailer
}

// WithSMTPConfig sets the SMTP configuration (Option pattern).
// WithSMTPConfig 设置 SMTP 配置（Option 模式）。
//
// Parameters:
// 参数：
// - host: SMTP server address (SMTP 服务器地址)
// - port: SMTP server port (SMTP 服务器端口)
// - user: Email account (邮箱账号)
// - password: Email password (邮箱密码)
func WithSMTPConfig(host, port, user, password string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.SMTPHost = host
		mailer.SMTPPort = port
		mailer.SMTPUser = user
		mailer.SMTPPassword = password
	}
}

// WithFrom sets the sender (Option pattern).
// WithFrom 设置发件人（Option 模式）。
//
// Parameters:
// 参数：
// - from: Sender email address (发件人邮箱地址)
// - alias: Sender alias (发件人别名)
func WithFrom(from, alias string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.From = from
		mailer.FromAlias = alias
	}
}

// WithTo sets the recipients (Option pattern).
// WithTo 设置收件人（Option 模式）。
//
// Parameters:
// 参数：
// - to: List of recipient email addresses (收件人邮箱地址列表)
func WithTo(to []string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.To = to
	}
}

// WithCc sets the CC recipients (Option pattern).
// WithCc 设置抄送（Option 模式）。
//
// Parameters:
// 参数：
// - cc: List of CC email addresses (抄送邮箱地址列表)
func WithCc(cc []string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.Cc = cc
	}
}

// WithBcc sets the BCC recipients (Option pattern).
// WithBcc 设置密送（Option 模式）。
//
// Parameters:
// 参数：
// - bcc: List of BCC email addresses (密送邮箱地址列表)
func WithBcc(bcc []string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.Bcc = bcc
	}
}

// WithSubject sets the email subject (Option pattern).
// WithSubject 设置邮件主题（Option 模式）。
//
// Parameters:
// 参数：
// - subject: Email subject (邮件主题)
func WithSubject(subject string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.Subject = subject
	}
}

// WithBody sets the plain text body (Option pattern).
// WithBody 设置纯文本正文（Option 模式）。
//
// Parameters:
// 参数：
// - body: Plain text body (纯文本正文)
func WithBody(body string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.Body = body
	}
}

// WithHTMLBody sets the HTML body (Option pattern).
// WithHTMLBody 设置 HTML 正文（Option 模式）。
//
// Parameters:
// 参数：
// - htmlBody: HTML body (HTML 正文)
func WithHTMLBody(htmlBody string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.HTMLBody = htmlBody
	}
}

// WithAttachments sets the attachments (Option pattern).
// WithAttachments 设置附件（Option 模式）。
//
// Parameters:
// 参数：
// - attachments: List of attachment file paths (附件文件路径列表)
func WithAttachments(attachments []string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.Attachments = attachments
	}
}

// WithInlineImages sets the inline images (Option pattern).
// WithInlineImages 设置内嵌图片（Option 模式）。
//
// Parameters:
// 参数：
// - images: Map of inline images, key is Content-ID, value is file path (内嵌图片映射，key 为 Content-ID，value 为文件路径)
func WithInlineImages(images map[string]string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.InlineImages = images
	}
}

// WithSSL sets whether to use SSL (Option pattern).
// WithSSL 设置是否使用 SSL（Option 模式）。
//
// Parameters:
// 参数：
// - useSSL: Whether to use SSL (是否使用 SSL)
func WithSSL(useSSL bool) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.UseSSL = useSSL
	}
}

// SetSMTPConfig sets the SMTP configuration (Fluent interface).
// SetSMTPConfig 设置 SMTP 配置（链式调用模式）。
//
// Parameters:
// 参数：
// - host: SMTP server address (SMTP 服务器地址)
// - port: SMTP server port (SMTP 服务器端口)
// - user: Email account (邮箱账号)
// - password: Email password (邮箱密码)
func (mailer *Mailer) SetSMTPConfig(host, port, user, password string) *Mailer {
	mailer.SMTPHost = host
	mailer.SMTPPort = port
	mailer.SMTPUser = user
	mailer.SMTPPassword = password
	return mailer
}

// SetFrom sets the sender (Fluent interface).
// SetFrom 设置发件人（链式调用模式）。
//
// Parameters:
// 参数：
// - from: Sender email address (发件人邮箱地址)
// - alias: Sender alias (发件人别名)
func (mailer *Mailer) SetFrom(from, alias string) *Mailer {
	mailer.From = from
	mailer.FromAlias = alias
	return mailer
}

// SetTo sets the recipients (Fluent interface).
// SetTo 设置收件人（链式调用模式）。
//
// Parameters:
// 参数：
// - to: List of recipient email addresses (收件人邮箱地址列表)
func (mailer *Mailer) SetTo(to []string) *Mailer {
	mailer.To = to
	return mailer
}

// SetCc sets the CC recipients (Fluent interface).
// SetCc 设置抄送（链式调用模式）。
//
// Parameters:
// 参数：
// - cc: List of CC email addresses (抄送邮箱地址列表)
func (mailer *Mailer) SetCc(cc []string) *Mailer {
	mailer.Cc = cc
	return mailer
}

// SetBcc sets the BCC recipients (Fluent interface).
// SetBcc 设置密送（链式调用模式）。
//
// Parameters:
// 参数：
// - bcc: List of BCC email addresses (密送邮箱地址列表)
func (mailer *Mailer) SetBcc(bcc []string) *Mailer {
	mailer.Bcc = bcc
	return mailer
}

// SetSubject sets the email subject (Fluent interface).
// SetSubject 设置邮件主题（链式调用模式）。
//
// Parameters:
// 参数：
// - subject: Email subject (邮件主题)
func (mailer *Mailer) SetSubject(subject string) *Mailer {
	mailer.Subject = subject
	return mailer
}

// SetBody sets the plain text body (Fluent interface).
// SetBody 设置纯文本正文（链式调用模式）。
//
// Parameters:
// 参数：
// - body: Plain text body (纯文本正文)
func (mailer *Mailer) SetBody(body string) *Mailer {
	mailer.Body = body
	return mailer
}

// SetHTMLBody sets the HTML body (Fluent interface).
// SetHTMLBody 设置 HTML 正文（链式调用模式）。
//
// Parameters:
// 参数：
// - htmlBody: HTML body (HTML 正文)
func (mailer *Mailer) SetHTMLBody(htmlBody string) *Mailer {
	mailer.HTMLBody = htmlBody
	return mailer
}

// SetAttachments sets the attachments (Fluent interface).
// SetAttachments 设置附件（链式调用模式）。
//
// Parameters:
// 参数：
// - attachments: List of attachment file paths (附件文件路径列表)
func (mailer *Mailer) SetAttachments(attachments []string) *Mailer {
	mailer.Attachments = attachments
	return mailer
}

// SetInlineImages sets the inline images (Fluent interface).
// SetInlineImages 设置内嵌图片（链式调用模式）。
//
// Parameters:
// 参数：
// - images: Map of inline images, key is Content-ID, value is file path (内嵌图片映射，key 为 Content-ID，value 为文件路径)
func (mailer *Mailer) SetInlineImages(images map[string]string) *Mailer {
	mailer.InlineImages = images
	return mailer
}

// SetSSL sets whether to use SSL (Fluent interface).
// SetSSL 设置是否使用 SSL（链式调用模式）。
//
// Parameters:
// 参数：
// - useSSL: Whether to use SSL (是否使用 SSL)
func (mailer *Mailer) SetSSL(useSSL bool) *Mailer {
	mailer.UseSSL = useSSL
	return mailer
}

// Send sends the email.
// Send 发送邮件。
//
// Returns:
// 返回值：
// - error: Returns an error if sending fails. (如果发送失败，返回错误信息)
func (mailer *Mailer) Send() error {
	// Connect to the SMTP server
	// 连接到 SMTP 服务器
	var conn *tls.Conn
	var err error

	if mailer.UseSSL {
		// Use SSL connection
		// 使用 SSL 连接
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // Skip certificate verification (not recommended for production) (跳过证书验证，生产环境不建议使用)
			ServerName:         mailer.SMTPHost,
		}
		conn, err = tls.Dial("tcp", mailer.SMTPHost+":"+mailer.SMTPPort, tlsConfig)
	} else {
		// Non-SSL connection
		// 非 SSL 连接
		conn, err = tls.Dial("tcp", mailer.SMTPHost+":"+mailer.SMTPPort, nil)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer conn.Close()

	// Create SMTP client
	// 创建 SMTP 客户端
	smtpClient, err := smtp.NewClient(conn, mailer.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}
	defer smtpClient.Close()

	// Authenticate
	// 身份验证
	auth := smtp.PlainAuth("", mailer.SMTPUser, mailer.SMTPPassword, mailer.SMTPHost)
	if err := smtpClient.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}

	// Set sender
	// 设置发件人
	if err := smtpClient.Mail(mailer.SMTPUser); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}

	// Set recipients (to, cc, bcc)
	// 设置收件人（to, cc, bcc）
	recipients := append(mailer.To, mailer.Cc...)
	recipients = append(recipients, mailer.Bcc...)
	for _, recipient := range recipients {
		if err := smtpClient.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %v", recipient, err)
		}
	}

	// Create email content
	// 创建邮件内容
	w, err := smtpClient.Data()
	if err != nil {
		return fmt.Errorf("failed to open data connection: %v", err)
	}
	defer w.Close()

	// Build email headers
	// 构建邮件头部
	headers := make(map[string]string)
	headers["From"] = mime.QEncoding.Encode("UTF-8", mailer.FromAlias) + " <" + mailer.From + ">"
	headers["To"] = strings.Join(mailer.To, ", ")
	if len(mailer.Cc) > 0 {
		headers["Cc"] = strings.Join(mailer.Cc, ", ")
	}
	headers["Subject"] = mime.QEncoding.Encode("UTF-8", mailer.Subject)
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "multipart/mixed; boundary=unique-boundary"

	// Write email headers
	// 写入邮件头部
	for k, v := range headers {
		if _, err := w.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v))); err != nil {
			return fmt.Errorf("failed to write header %s: %v", k, err)
		}
	}
	if _, err := w.Write([]byte("\r\n")); err != nil {
		return fmt.Errorf("failed to write header end: %v", err)
	}

	// Create multipart writer
	// 创建 multipart writer
	multipartWriter := multipart.NewWriter(w)
	multipartWriter.SetBoundary("unique-boundary")

	// Add plain text body
	// 添加纯文本正文
	if mailer.Body != "" {
		textHeader := make(textproto.MIMEHeader)
		textHeader.Set("Content-Type", "text/plain; charset=UTF-8")
		textPart, err := multipartWriter.CreatePart(textHeader)
		if err != nil {
			return fmt.Errorf("failed to create text part: %v", err)
		}
		if _, err := textPart.Write([]byte(mailer.Body)); err != nil {
			return fmt.Errorf("failed to write text part: %v", err)
		}
	}

	// Add HTML body
	// 添加 HTML 正文
	if mailer.HTMLBody != "" {
		htmlHeader := make(textproto.MIMEHeader)
		htmlHeader.Set("Content-Type", "text/html; charset=UTF-8")
		htmlPart, err := multipartWriter.CreatePart(htmlHeader)
		if err != nil {
			return fmt.Errorf("failed to create HTML part: %v", err)
		}
		if _, err := htmlPart.Write([]byte(mailer.HTMLBody)); err != nil {
			return fmt.Errorf("failed to write HTML part: %v", err)
		}
	}

	// Add inline images
	// 添加内嵌图片
	for cid, filePath := range mailer.InlineImages {
		fileData, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read inline image %s: %v", filePath, err)
		}

		imageHeader := make(textproto.MIMEHeader)
		imageHeader.Set("Content-Type", "image/jpeg")
		imageHeader.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filepath.Base(filePath)))
		imageHeader.Set("Content-ID", fmt.Sprintf("<%s>", cid))
		imageHeader.Set("Content-Transfer-Encoding", "base64")

		imagePart, err := multipartWriter.CreatePart(imageHeader)
		if err != nil {
			return fmt.Errorf("failed to create inline image part: %v", err)
		}
		if _, err := imagePart.Write(fileData); err != nil {
			return fmt.Errorf("failed to write inline image part: %v", err)
		}
	}

	// Add attachments
	// 添加附件
	for _, attachment := range mailer.Attachments {
		fileData, err := os.ReadFile(attachment)
		if err != nil {
			return fmt.Errorf("failed to read attachment %s: %v", attachment, err)
		}

		attachmentHeader := make(textproto.MIMEHeader)
		attachmentHeader.Set("Content-Type", "application/octet-stream")
		attachmentHeader.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(attachment)))
		attachmentHeader.Set("Content-Transfer-Encoding", "base64")

		attachmentPart, err := multipartWriter.CreatePart(attachmentHeader)
		if err != nil {
			return fmt.Errorf("failed to create attachment part: %v", err)
		}
		if _, err := attachmentPart.Write(fileData); err != nil {
			return fmt.Errorf("failed to write attachment part: %v", err)
		}
	}

	// Close multipart writer
	// 关闭 multipart writer
	if err := multipartWriter.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %v", err)
	}

	return nil
}
