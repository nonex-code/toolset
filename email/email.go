package email

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Mailer represents an email client for sending emails.
// Mailer 表示用于发送邮件的邮件客户端。
type Mailer struct {
	SMTPHost      string            // SMTP server address (SMTP 服务器地址)
	SMTPPort      string            // SMTP server port (SMTP 服务器端口)
	SMTPUser      string            // Email account (邮箱账号)
	SMTPPassword  string            // Email password (邮箱密码)
	From          string            // Sender email address (发件人邮箱地址)
	FromAlias     string            // Sender alias (发件人别名)
	To            []string          // List of recipient email addresses (收件人邮箱地址列表)
	Cc            []string          // List of CC email addresses (抄送邮箱地址列表)
	Bcc           []string          // List of BCC email addresses (密送邮箱地址列表)
	Subject       string            // Email subject (邮件主题)
	Body          string            // Plain text body (纯文本正文)
	HTMLBody      string            // HTML body (HTML 正文)
	Attachments   []string          // List of attachment file paths (附件文件路径列表)
	InlineImages  map[string]string // Inline images: key is Content-ID, value is file path (内嵌图片：key 为 Content-ID，value 为文件路径)
	UseSSL        bool              // Whether to use SSL (是否使用 SSL)
	SkipTLSVerify bool              // Whether to skip TLS certificate verification (是否跳过 TLS 证书校验，默认 false)
}

// NewMailer creates a new Mailer with optional configurations.
// NewMailer 创建一个新的 Mailer，支持可选配置。
//
// Example (Option pattern):
// 示例（Option 模式）：
//
//	mailer := NewMailer(
//		WithSMTPConfig("smtp.example.com", "465", "your-email@example.com", "your-password"),
//		WithFrom("your-email@example.com", "Your Name"),
//		WithTo([]string{"recipient1@example.com"}),
//		WithSubject("Test Email"),
//		WithBody("This is a plain text email."),
//		WithSSL(true),
//	)
//
// Example (Fluent interface):
// 示例（链式调用模式）：
//
//	mailer := NewMailer().
//		SetFrom("your-email@example.com", "Your Name").
//		SetTo([]string{"recipient1@example.com"}).
//		SetSubject("Test Email").
//		SetBody("This is a plain text email.").
//		SetSSL(true)
func NewMailer(options ...func(*Mailer)) *Mailer {
	mailer := &Mailer{
		SMTPHost:      "smtp.example.com", // Default SMTP server (默认 SMTP 服务器)
		SMTPPort:      "465",              // Default SSL port (默认 SSL 端口)
		SMTPUser:      "your-email@example.com",
		SMTPPassword:  "your-password",
		InlineImages:  make(map[string]string),
		UseSSL:        true,  // Default to use SSL (默认使用 SSL)
		SkipTLSVerify: false, // Verify certificate by default (默认校验证书)
	}
	for _, option := range options {
		option(mailer)
	}
	return mailer
}

// ---- Option pattern ------------------------------------------------------------------

// WithSMTPConfig sets the SMTP configuration (Option pattern).
func WithSMTPConfig(host, port, user, password string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.SMTPHost = host
		mailer.SMTPPort = port
		mailer.SMTPUser = user
		mailer.SMTPPassword = password
	}
}

// WithFrom sets the sender (Option pattern).
func WithFrom(from, alias string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.From = from
		mailer.FromAlias = alias
	}
}

// WithTo sets the recipients (Option pattern).
func WithTo(to []string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.To = to
	}
}

// WithCc sets the CC recipients (Option pattern).
func WithCc(cc []string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.Cc = cc
	}
}

// WithBcc sets the BCC recipients (Option pattern).
func WithBcc(bcc []string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.Bcc = bcc
	}
}

// WithSubject sets the email subject (Option pattern).
func WithSubject(subject string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.Subject = subject
	}
}

// WithBody sets the plain text body (Option pattern).
func WithBody(body string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.Body = body
	}
}

// WithHTMLBody sets the HTML body (Option pattern).
func WithHTMLBody(htmlBody string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.HTMLBody = htmlBody
	}
}

// WithAttachments sets the attachments (Option pattern).
func WithAttachments(attachments []string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.Attachments = attachments
	}
}

// WithInlineImages sets the inline images (Option pattern).
func WithInlineImages(images map[string]string) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.InlineImages = images
	}
}

// WithSSL sets whether to use SSL (Option pattern).
func WithSSL(useSSL bool) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.UseSSL = useSSL
	}
}

// WithSkipTLSVerify sets whether to skip TLS certificate verification (Option pattern).
// WithSkipTLSVerify 设置是否跳过 TLS 证书校验（Option 模式）。
func WithSkipTLSVerify(skip bool) func(*Mailer) {
	return func(mailer *Mailer) {
		mailer.SkipTLSVerify = skip
	}
}

// ---- Fluent interface -----------------------------------------------------------------

// SetSMTPConfig sets the SMTP configuration (Fluent interface).
func (mailer *Mailer) SetSMTPConfig(host, port, user, password string) *Mailer {
	mailer.SMTPHost = host
	mailer.SMTPPort = port
	mailer.SMTPUser = user
	mailer.SMTPPassword = password
	return mailer
}

// SetFrom sets the sender (Fluent interface).
func (mailer *Mailer) SetFrom(from, alias string) *Mailer {
	mailer.From = from
	mailer.FromAlias = alias
	return mailer
}

// SetTo sets the recipients (Fluent interface).
func (mailer *Mailer) SetTo(to []string) *Mailer {
	mailer.To = to
	return mailer
}

// SetCc sets the CC recipients (Fluent interface).
func (mailer *Mailer) SetCc(cc []string) *Mailer {
	mailer.Cc = cc
	return mailer
}

// SetBcc sets the BCC recipients (Fluent interface).
func (mailer *Mailer) SetBcc(bcc []string) *Mailer {
	mailer.Bcc = bcc
	return mailer
}

// SetSubject sets the email subject (Fluent interface).
func (mailer *Mailer) SetSubject(subject string) *Mailer {
	mailer.Subject = subject
	return mailer
}

// SetBody sets the plain text body (Fluent interface).
func (mailer *Mailer) SetBody(body string) *Mailer {
	mailer.Body = body
	return mailer
}

// SetHTMLBody sets the HTML body (Fluent interface).
func (mailer *Mailer) SetHTMLBody(htmlBody string) *Mailer {
	mailer.HTMLBody = htmlBody
	return mailer
}

// SetAttachments sets the attachments (Fluent interface).
func (mailer *Mailer) SetAttachments(attachments []string) *Mailer {
	mailer.Attachments = attachments
	return mailer
}

// SetInlineImages sets the inline images (Fluent interface).
func (mailer *Mailer) SetInlineImages(images map[string]string) *Mailer {
	mailer.InlineImages = images
	return mailer
}

// SetSSL sets whether to use SSL (Fluent interface).
func (mailer *Mailer) SetSSL(useSSL bool) *Mailer {
	mailer.UseSSL = useSSL
	return mailer
}

// SetSkipTLSVerify sets whether to skip TLS certificate verification (Fluent interface).
// SetSkipTLSVerify 设置是否跳过 TLS 证书校验（链式调用模式）。
func (mailer *Mailer) SetSkipTLSVerify(skip bool) *Mailer {
	mailer.SkipTLSVerify = skip
	return mailer
}

// ---- Sending --------------------------------------------------------------------------

// Send sends the email.
// Send 发送邮件。
//
// Returns:
// 返回值：
// - error: Returns an error if sending fails. (如果发送失败，返回错误信息)
func (mailer *Mailer) Send() error {
	msg, err := mailer.buildMessage()
	if err != nil {
		return fmt.Errorf("failed to build message: %w", err)
	}

	conn, err := mailer.dial()
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	smtpClient, err := smtp.NewClient(conn, mailer.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer smtpClient.Close()

	// Upgrade to TLS when the server supports STARTTLS and we are on a plain connection.
	// 明文连接时，若服务器支持 STARTTLS 则升级为加密连接。
	if !mailer.UseSSL {
		if ok, _ := smtpClient.Extension("STARTTLS"); ok {
			tlsConfig := &tls.Config{
				ServerName:         mailer.SMTPHost,
				InsecureSkipVerify: mailer.SkipTLSVerify,
			}
			if err := smtpClient.StartTLS(tlsConfig); err != nil {
				return fmt.Errorf("starttls failed: %w", err)
			}
		}
	}

	// Authenticate. smtp.PlainAuth itself refuses to send credentials over an
	// unencrypted connection, which is the desired behavior.
	// 身份验证。smtp.PlainAuth 会拒绝在未加密连接上发送凭据，这是期望的行为。
	auth := smtp.PlainAuth("", mailer.SMTPUser, mailer.SMTPPassword, mailer.SMTPHost)
	if err := smtpClient.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Envelope sender: prefer the configured From address, fall back to SMTPUser.
	// 信封发件人：优先使用配置的 From 地址，否则退回 SMTPUser。
	from := mailer.From
	if from == "" {
		from = mailer.SMTPUser
	}
	if err := smtpClient.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Recipients (to, cc, bcc). Copy before append to avoid mutating the caller's slice.
	// 收件人（to, cc, bcc）。先复制再追加，避免修改调用方的切片。
	recipients := make([]string, 0, len(mailer.To)+len(mailer.Cc)+len(mailer.Bcc))
	recipients = append(recipients, mailer.To...)
	recipients = append(recipients, mailer.Cc...)
	recipients = append(recipients, mailer.Bcc...)
	for _, recipient := range recipients {
		if err := smtpClient.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	w, err := smtpClient.Data()
	if err != nil {
		return fmt.Errorf("failed to open data connection: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		w.Close()
		return fmt.Errorf("failed to write message: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to finish message: %w", err)
	}
	return nil
}

// dial opens a connection to the SMTP server, using implicit TLS for SSL (port 465)
// or a plain TCP connection (later upgraded via STARTTLS when available).
// dial 连接 SMTP 服务器：SSL 使用隐式 TLS，否则使用明文 TCP（后续按需 STARTTLS）。
func (mailer *Mailer) dial() (net.Conn, error) {
	addr := net.JoinHostPort(mailer.SMTPHost, mailer.SMTPPort)
	if mailer.UseSSL {
		tlsConfig := &tls.Config{
			ServerName:         mailer.SMTPHost,
			InsecureSkipVerify: mailer.SkipTLSVerify,
		}
		return tls.Dial("tcp", addr, tlsConfig)
	}
	return net.Dial("tcp", addr)
}

// ---- Message construction -------------------------------------------------------------

// buildMessage renders the full RFC 5322 message (headers + body).
// buildMessage 生成完整的 RFC 5322 邮件内容（头部 + 正文）。
//
// Structure:
//   - no attachments + text & html -> multipart/alternative
//   - attachments/inline images    -> multipart/mixed
//     (text & html wrapped in a nested multipart/alternative)
//   - single body only             -> single part text/plain or text/html
func (mailer *Mailer) buildMessage() ([]byte, error) {
	hasPlain := mailer.Body != ""
	hasHTML := mailer.HTMLBody != ""
	hasBody := hasPlain || hasHTML
	hasAttachment := len(mailer.Attachments) > 0 || len(mailer.InlineImages) > 0

	var buf bytes.Buffer

	// Determine the top-level content type and whether a top-level multipart
	// container is needed.
	// 确定顶层 Content-Type，以及是否需要顶层 multipart 容器。
	needTopMultipart := hasAttachment || (hasPlain && hasHTML)
	var topContentType string
	var topWriter *multipart.Writer

	if needTopMultipart {
		topWriter = multipart.NewWriter(&buf)
		if hasAttachment {
			topContentType = `multipart/mixed; boundary="` + topWriter.Boundary() + `"`
		} else {
			topContentType = `multipart/alternative; boundary="` + topWriter.Boundary() + `"`
		}
	} else {
		switch {
		case hasHTML:
			topContentType = "text/html; charset=UTF-8"
		case hasPlain:
			topContentType = "text/plain; charset=UTF-8"
		default:
			topContentType = "text/plain; charset=UTF-8"
		}
	}

	if err := mailer.writeHeaders(&buf, topContentType); err != nil {
		return nil, err
	}

	// Single-part message: write the body directly after the headers.
	// 单部分邮件：头部之后直接写入正文。
	if !needTopMultipart {
		if hasHTML {
			io.WriteString(&buf, mailer.HTMLBody)
		} else {
			io.WriteString(&buf, mailer.Body)
		}
		return buf.Bytes(), nil
	}

	// Build the body part(s).
	// 构建正文部分。
	if hasAttachment {
		if hasBody {
			switch {
			case hasPlain && hasHTML:
				// Nested multipart/alternative.
				// 嵌套 multipart/alternative。
				b := randomBoundary()
				header := textproto.MIMEHeader{}
				header.Set("Content-Type", `multipart/alternative; boundary="`+b+`"`)
				part, err := topWriter.CreatePart(header)
				if err != nil {
					return nil, fmt.Errorf("failed to create alternative part: %w", err)
				}
				sub := multipart.NewWriter(part)
				_ = sub.SetBoundary(b)
				if err := writeTextPart(sub, "text/plain; charset=UTF-8", mailer.Body); err != nil {
					return nil, err
				}
				if err := writeTextPart(sub, "text/html; charset=UTF-8", mailer.HTMLBody); err != nil {
					return nil, err
				}
				if err := sub.Close(); err != nil {
					return nil, err
				}
			case hasHTML:
				if err := writeTextPart(topWriter, "text/html; charset=UTF-8", mailer.HTMLBody); err != nil {
					return nil, err
				}
			default:
				if err := writeTextPart(topWriter, "text/plain; charset=UTF-8", mailer.Body); err != nil {
					return nil, err
				}
			}
		}
		if err := mailer.writeInlineImages(topWriter); err != nil {
			return nil, err
		}
		if err := mailer.writeAttachments(topWriter); err != nil {
			return nil, err
		}
	} else {
		// Top-level multipart/alternative.
		// 顶层 multipart/alternative。
		if err := writeTextPart(topWriter, "text/plain; charset=UTF-8", mailer.Body); err != nil {
			return nil, err
		}
		if err := writeTextPart(topWriter, "text/html; charset=UTF-8", mailer.HTMLBody); err != nil {
			return nil, err
		}
	}

	if err := topWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}
	return buf.Bytes(), nil
}

// writeHeaders writes the message headers in a deterministic order.
// writeHeaders 以固定顺序写入邮件头部。
func (mailer *Mailer) writeHeaders(w io.Writer, contentType string) error {
	from := mailer.From
	if from == "" {
		from = mailer.SMTPUser
	}
	fromHeader := from
	if mailer.FromAlias != "" {
		fromHeader = mime.QEncoding.Encode("UTF-8", mailer.FromAlias) + " <" + from + ">"
	}

	headers := [][2]string{
		{"From", fromHeader},
	}
	if len(mailer.To) > 0 {
		headers = append(headers, [2]string{"To", strings.Join(mailer.To, ", ")})
	}
	if len(mailer.Cc) > 0 {
		headers = append(headers, [2]string{"Cc", strings.Join(mailer.Cc, ", ")})
	}
	headers = append(headers,
		[2]string{"Subject", mime.QEncoding.Encode("UTF-8", mailer.Subject)},
		[2]string{"Date", time.Now().Format(time.RFC1123Z)},
		[2]string{"Message-ID", newMessageID(from)},
		[2]string{"MIME-Version", "1.0"},
		[2]string{"Content-Type", contentType},
	)

	for _, h := range headers {
		if _, err := fmt.Fprintf(w, "%s: %s\r\n", h[0], h[1]); err != nil {
			return fmt.Errorf("failed to write header %s: %w", h[0], err)
		}
	}
	if _, err := io.WriteString(w, "\r\n"); err != nil {
		return fmt.Errorf("failed to write header end: %w", err)
	}
	return nil
}

// writeTextPart creates a text part with the given content type.
// writeTextPart 以指定的 Content-Type 创建一个文本 part。
func writeTextPart(mw *multipart.Writer, contentType, body string) error {
	header := textproto.MIMEHeader{}
	header.Set("Content-Type", contentType)
	header.Set("Content-Transfer-Encoding", "8bit")
	part, err := mw.CreatePart(header)
	if err != nil {
		return fmt.Errorf("failed to create text part: %w", err)
	}
	if _, err := io.WriteString(part, body); err != nil {
		return fmt.Errorf("failed to write text part: %w", err)
	}
	return nil
}

// writeInlineImages writes inline images as base64 parts.
// writeInlineImages 将内嵌图片以 base64 编码写入。
func (mailer *Mailer) writeInlineImages(mw *multipart.Writer) error {
	for cid, filePath := range mailer.InlineImages {
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read inline image %s: %w", filePath, err)
		}

		header := textproto.MIMEHeader{}
		header.Set("Content-Type", detectContentType(filePath, fileData))
		header.Set("Content-Transfer-Encoding", "base64")
		header.Set("Content-ID", "<"+cid+">")
		header.Set("Content-Disposition", `inline; filename="`+encodeFilename(filepath.Base(filePath))+`"`)

		part, err := mw.CreatePart(header)
		if err != nil {
			return fmt.Errorf("failed to create inline image part: %w", err)
		}
		if err := writeBase64(part, fileData); err != nil {
			return fmt.Errorf("failed to write inline image part: %w", err)
		}
	}
	return nil
}

// writeAttachments writes attachments as base64 parts.
// writeAttachments 将附件以 base64 编码写入。
func (mailer *Mailer) writeAttachments(mw *multipart.Writer) error {
	for _, attachment := range mailer.Attachments {
		fileData, err := os.ReadFile(attachment)
		if err != nil {
			return fmt.Errorf("failed to read attachment %s: %w", attachment, err)
		}

		header := textproto.MIMEHeader{}
		header.Set("Content-Type", detectContentType(attachment, fileData))
		header.Set("Content-Transfer-Encoding", "base64")
		header.Set("Content-Disposition", `attachment; filename="`+encodeFilename(filepath.Base(attachment))+`"`)

		part, err := mw.CreatePart(header)
		if err != nil {
			return fmt.Errorf("failed to create attachment part: %w", err)
		}
		if err := writeBase64(part, fileData); err != nil {
			return fmt.Errorf("failed to write attachment part: %w", err)
		}
	}
	return nil
}

// ---- Helpers --------------------------------------------------------------------------

// writeBase64 encodes data as MIME base64 (76-char lines) into w.
// writeBase64 将数据编码为 MIME base64（每行 76 字符）写入 w。
func writeBase64(w io.Writer, data []byte) error {
	enc := base64.NewEncoder(base64.StdEncoding, &lineWrapWriter{w: w, width: 76})
	if _, err := enc.Write(data); err != nil {
		return err
	}
	return enc.Close()
}

// lineWrapWriter inserts CRLF every width bytes.
// lineWrapWriter 每写入 width 字节插入一个 CRLF。
type lineWrapWriter struct {
	w     io.Writer
	width int
	n     int
}

func (l *lineWrapWriter) Write(p []byte) (int, error) {
	total := 0
	for len(p) > 0 {
		if l.n == l.width {
			if _, err := io.WriteString(l.w, "\r\n"); err != nil {
				return total, err
			}
			l.n = 0
		}
		space := l.width - l.n
		chunk := p
		if len(chunk) > space {
			chunk = chunk[:space]
		}
		if _, err := l.w.Write(chunk); err != nil {
			return total, err
		}
		total += len(chunk)
		l.n += len(chunk)
		p = p[len(chunk):]
	}
	return total, nil
}

// detectContentType guesses the MIME type from the file extension, falling back to
// content sniffing and finally to application/octet-stream.
// detectContentType 优先根据扩展名推断 MIME 类型，其次嗅探内容，最后回退为 application/octet-stream。
func detectContentType(path string, data []byte) string {
	if ct := mime.TypeByExtension(filepath.Ext(path)); ct != "" {
		return ct
	}
	if len(data) > 0 {
		if len(data) > 512 {
			data = data[:512]
		}
		return http.DetectContentType(data)
	}
	return "application/octet-stream"
}

// encodeFilename MIME-encodes a filename only when it contains non-ASCII characters.
// encodeFilename 仅当文件名包含非 ASCII 字符时才进行 MIME 编码。
func encodeFilename(name string) string {
	for _, r := range name {
		if r > 127 {
			return mime.QEncoding.Encode("UTF-8", name)
		}
	}
	return name
}

// randomBoundary returns a random multipart boundary.
// randomBoundary 返回一个随机的 multipart 边界字符串。
func randomBoundary() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("boundary-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

// newMessageID builds a unique Message-ID header value.
// newMessageID 生成唯一的 Message-ID 头部值。
func newMessageID(from string) string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("<%d.%s@%s>", time.Now().UnixNano(), hex.EncodeToString(b[:]), domainOf(from))
}

// domainOf extracts the domain part of an email address.
// domainOf 提取邮箱地址中的域名部分。
func domainOf(addr string) string {
	if i := strings.LastIndex(addr, "@"); i >= 0 && i+1 < len(addr) {
		return addr[i+1:]
	}
	return "localhost"
}
