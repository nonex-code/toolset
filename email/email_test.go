package email

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"strings"
	"testing"
)

// decodeBase64Part decodes a base64 MIME part, ignoring CRLF line breaks.
// decodeBase64Part 解码 base64 MIME part，忽略 CRLF 换行。
func decodeBase64Part(t *testing.T, r io.Reader) []byte {
	t.Helper()
	raw, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	clean := strings.NewReplacer("\r", "", "\n", "", " ", "", "\t", "").Replace(string(raw))
	decoded, err := base64.StdEncoding.DecodeString(clean)
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}
	return decoded
}

func TestBuildMessagePlainText(t *testing.T) {
	m := NewMailer(
		WithSMTPConfig("smtp.example.com", "465", "u@example.com", "p"),
		WithFrom("u@example.com", "Sender"),
		WithTo([]string{"r@example.com"}),
		WithSubject("plain"),
		WithBody("hello world"),
	)

	msg, err := m.buildMessage()
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := mail.ReadMessage(bytes.NewReader(msg))
	if err != nil {
		t.Fatal(err)
	}
	ct := parsed.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("Content-Type = %q, want text/plain", ct)
	}
	if parsed.Header.Get("Date") == "" {
		t.Error("missing Date header")
	}
	if !strings.HasPrefix(parsed.Header.Get("Message-ID"), "<") {
		t.Errorf("bad Message-ID: %q", parsed.Header.Get("Message-ID"))
	}
	body, _ := io.ReadAll(parsed.Body)
	if string(body) != "hello world" {
		t.Errorf("body = %q, want %q", body, "hello world")
	}
}

func TestBuildMessageAlternative(t *testing.T) {
	m := NewMailer(
		WithFrom("u@example.com", "Sender"),
		WithTo([]string{"r@example.com"}),
		WithSubject("hi"),
		WithBody("plain"),
		WithHTMLBody("<b>html</b>"),
	)

	msg, err := m.buildMessage()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := mail.ReadMessage(bytes.NewReader(msg))
	if err != nil {
		t.Fatal(err)
	}
	_, params, err := mime.ParseMediaType(parsed.Header.Get("Content-Type"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(parsed.Header.Get("Content-Type"), "multipart/alternative") {
		t.Fatalf("Content-Type = %q, want multipart/alternative", parsed.Header.Get("Content-Type"))
	}

	mr := multipart.NewReader(parsed.Body, params["boundary"])
	var types []string
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		types = append(types, p.Header.Get("Content-Type"))
	}
	if len(types) != 2 || !strings.HasPrefix(types[0], "text/plain") || !strings.HasPrefix(types[1], "text/html") {
		t.Errorf("parts = %v, want [text/plain, text/html]", types)
	}
}

func TestBuildMessageMixedWithAttachment(t *testing.T) {
	want, err := os.ReadFile("testdata/sample.bin")
	if err != nil {
		t.Fatal(err)
	}

	m := NewMailer(
		WithSMTPConfig("smtp.example.com", "465", "u@example.com", "p"),
		WithFrom("u@example.com", "发件人"),
		WithTo([]string{"r@example.com"}),
		WithCc([]string{"cc@example.com"}),
		WithSubject("带附件"),
		WithBody("plain"),
		WithHTMLBody("<b>html</b>"),
		WithAttachments([]string{"testdata/sample.bin"}),
	)

	msg, err := m.buildMessage()
	if err != nil {
		t.Fatal(err)
	}
	raw := string(msg)
	for _, wantHeader := range []string{
		"multipart/mixed",
		"multipart/alternative",
		"Content-Transfer-Encoding: base64",
		"Content-Disposition: attachment",
		"Date:",
		"Message-ID:",
		"From: ",
		"Cc: cc@example.com",
	} {
		if !strings.Contains(raw, wantHeader) {
			t.Errorf("message missing %q", wantHeader)
		}
	}

	parsed, err := mail.ReadMessage(bytes.NewReader(msg))
	if err != nil {
		t.Fatal(err)
	}
	_, params, err := mime.ParseMediaType(parsed.Header.Get("Content-Type"))
	if err != nil {
		t.Fatal(err)
	}
	mr := multipart.NewReader(parsed.Body, params["boundary"])

	var found bool
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(p.Header.Get("Content-Disposition"), "sample.bin") {
			got := decodeBase64Part(t, p)
			if !bytes.Equal(got, want) {
				t.Errorf("attachment bytes mismatch: got %x want %x", got, want)
			}
			found = true
		}
	}
	if !found {
		t.Error("attachment part not found")
	}
}

func TestBuildMessageInlineImage(t *testing.T) {
	want, err := os.ReadFile("testdata/sample.bin")
	if err != nil {
		t.Fatal(err)
	}

	m := NewMailer(
		WithFrom("u@example.com", "Sender"),
		WithTo([]string{"r@example.com"}),
		WithSubject("inline"),
		WithHTMLBody(`<img src="cid:image1">`),
		WithInlineImages(map[string]string{"image1": "testdata/sample.bin"}),
	)

	msg, err := m.buildMessage()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := mail.ReadMessage(bytes.NewReader(msg))
	if err != nil {
		t.Fatal(err)
	}
	_, params, err := mime.ParseMediaType(parsed.Header.Get("Content-Type"))
	if err != nil {
		t.Fatal(err)
	}
	mr := multipart.NewReader(parsed.Body, params["boundary"])

	var found bool
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if p.Header.Get("Content-ID") == "<image1>" {
			got := decodeBase64Part(t, p)
			if !bytes.Equal(got, want) {
				t.Errorf("inline image bytes mismatch")
			}
			found = true
		}
	}
	if !found {
		t.Error("inline image part not found")
	}
}

func TestOptionsAndFluent(t *testing.T) {
	m := NewMailer(
		WithSMTPConfig("h", "465", "u", "p"),
		WithSSL(false),
		WithSkipTLSVerify(true),
	)
	if m.SMTPHost != "h" || m.SMTPPort != "465" || m.SMTPUser != "u" || m.SMTPPassword != "p" {
		t.Error("option pattern did not apply SMTP config")
	}
	if m.UseSSL || !m.SkipTLSVerify {
		t.Error("option pattern did not apply SSL flags")
	}

	f := NewMailer().
		SetSMTPConfig("h2", "587", "u2", "p2").
		SetFrom("a@b.com", "Alias").
		SetTo([]string{"c@d.com"}).
		SetCc([]string{"e@f.com"}).
		SetBcc([]string{"g@h.com"}).
		SetSubject("s").
		SetBody("b").
		SetHTMLBody("<i>h</i>").
		SetSSL(false).
		SetSkipTLSVerify(true)
	if f.SMTPHost != "h2" || f.SMTPPort != "587" || f.From != "a@b.com" || f.FromAlias != "Alias" {
		t.Error("fluent interface did not apply config")
	}
	if len(f.To) != 1 || len(f.Cc) != 1 || len(f.Bcc) != 1 {
		t.Error("fluent interface did not apply recipients")
	}
}

func TestRandomBoundaryUnique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		b := randomBoundary()
		if b == "" || seen[b] {
			t.Fatalf("duplicate or empty boundary: %q", b)
		}
		seen[b] = true
	}
}
