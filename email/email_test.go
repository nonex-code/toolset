package email

import (
	"testing"
)

func TestMailer(t *testing.T) {
	// Example 1: Plain text email (Option pattern)
	// 示例 1：纯文本邮件（Option 模式）
	mailer1 := NewMailer(
		WithSMTPConfig("smtp.example.com", "465", "your-email@example.com", "your-password"),
		WithFrom("your-email@example.com", "Your Name"),
		WithTo([]string{"recipient1@example.com", "recipient2@example.com"}),
		WithSubject("Test Plain Text Email"),
		WithBody("This is a plain text email."),
		WithSSL(true),
	)
	err := mailer1.Send()
	if err != nil {
		t.Error("Error sending email:", err)
	} else {
		t.Log("Plain text email sent successfully")
	}

	// Example 2: HTML email (Fluent interface)
	// 示例 2：HTML 邮件（链式调用模式）
	mailer2 := NewMailer().
		SetSMTPConfig("smtp.example.com", "465", "your-email@example.com", "your-password").
		SetFrom("your-email@example.com", "Your Name").
		SetTo([]string{"recipient1@example.com", "recipient2@example.com"}).
		SetSubject("Test HTML Email").
		SetHTMLBody("<html><body><h1>This is an HTML email</h1></body></html>").
		SetSSL(true)
	err = mailer2.Send()
	if err != nil {
		t.Error("Error sending email:", err)
	} else {
		t.Log("Plain text email sent successfully")
	}

	// Example 3: Email with attachments (Option pattern)
	// 示例 3：带附件的邮件（Option 模式）
	mailer3 := NewMailer(
		WithSMTPConfig("smtp.example.com", "465", "your-email@example.com", "your-password"),
		WithFrom("your-email@example.com", "Your Name"),
		WithTo([]string{"recipient1@example.com", "recipient2@example.com"}),
		WithSubject("Test Email with Attachments"),
		WithBody("This is a plain text email with attachments."),
		WithAttachments([]string{"./file1.txt", "./file2.txt"}),
		WithSSL(true),
	)
	err = mailer3.Send()
	if err != nil {
		t.Error("Error sending email:", err)
	} else {
		t.Log("Plain text email sent successfully")
	}
	// Example 4: Email with inline images (Fluent interface)
	// 示例 4：带内嵌图片的邮件（链式调用模式）
	mailer4 := NewMailer().
		SetSMTPConfig("smtp.example.com", "465", "your-email@example.com", "your-password").
		SetFrom("your-email@example.com", "Your Name").
		SetTo([]string{"recipient1@example.com", "recipient2@example.com"}).
		SetSubject("Test Email with Inline Images").
		SetHTMLBody("<html><body><h1>This is an HTML email with an inline image</h1><img src=\"cid:image1\"></body></html>").
		SetInlineImages(map[string]string{"image1": "./image.jpg"}).
		SetSSL(true)
	err = mailer4.Send()
	if err != nil {
		t.Error("Error sending email:", err)
	} else {
		t.Log("Plain text email sent successfully")
	}

	mailer := NewMailer(
		WithSMTPConfig("smtp.example.com", "465", "your-email@example.com", "your-password"),
		WithFrom("your-email@example.com", "Your Name"),
		WithTo([]string{"recipient1@example.com", "recipient2@example.com"}),
		WithSubject("Test Plain Text Email"),
		WithBody("This is a plain text email."),
		WithSSL(true),
	)
	err = mailer.Send()
	if err != nil {
		t.Error("Error sending email:", err)
	} else {
		t.Log("Plain text email sent successfully")
	}
}

/*

 */
