//go:generate mockgen -source=mailer.go -destination=mocks/mailer.go -package=mocks

// Package mailer sends transactional email over SMTP (net/smtp) - there is no
// existing mail dependency in this codebase to build on.
package mailer

import (
	"backend/internal/config"
	"crypto/tls"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"
)

type Message struct {
	To       string
	Subject  string
	TextBody string
	HTMLBody string
}

type Mailer interface {
	Send(msg Message) error
}

type smtpMailer struct {
	host     string
	port     int
	username string
	password string
	from     string
	fromName string
	tlsMode  string
}

// New builds a Mailer from the smtp.* config section.
func New() Mailer {
	return &smtpMailer{
		host:     config.String("smtp.host"),
		port:     config.Int("smtp.port"),
		username: config.String("smtp.username"),
		password: config.String("smtp.password"),
		from:     config.String("smtp.from"),
		fromName: config.String("smtp.fromName"),
		tlsMode:  normalizeTlsMode(config.String("smtp.tlsMode")),
	}
}

// fromHeader is what goes in the message's From: header - just the bare
// address, or "Display Name" <address> when smtp.fromName is set. This must
// never be used for the SMTP envelope (MAIL FROM): that command only accepts
// a bare address, and a display name there gets the whole send rejected by
// the server.
func (m *smtpMailer) fromHeader() string {
	return FormatFromHeader(m.from, m.fromName)
}

// FormatFromHeader formats an address and optional display name the same
// way a message's From: header is built, e.g. `"Display Name" <addr>`, or
// just `addr` when name is blank. Exported so the admin-only email delivery
// info endpoint can show exactly what recipients will see, without
// duplicating this formatting.
func FormatFromHeader(address, name string) string {
	if strings.TrimSpace(name) == "" {
		return address
	}
	return (&mail.Address{Name: name, Address: address}).String()
}

// normalizeTlsMode treats anything other than an exact "none"/"starttls" as
// "tls" (implicit TLS) - the secure choice - rather than silently falling
// back to an unencrypted connection on a blank or mistyped value.
func normalizeTlsMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "none":
		return "none"
	case "starttls":
		return "starttls"
	default:
		return "tls"
	}
}

func (m *smtpMailer) Send(msg Message) error {
	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	var auth smtp.Auth
	if m.username != "" {
		auth = smtp.PlainAuth("", m.username, m.password, m.host)
	}

	body := buildMessage(m.fromHeader(), msg)

	switch m.tlsMode {
	case "none":
		return m.sendPlain(addr, auth, msg.To, body, false)
	case "starttls":
		return m.sendPlain(addr, auth, msg.To, body, true)
	default:
		return m.sendImplicitTLS(addr, auth, msg.To, body)
	}
}

func (m *smtpMailer) sendPlain(addr string, auth smtp.Auth, to string, body []byte, startTLS bool) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if startTLS {
		if err := client.StartTLS(&tls.Config{ServerName: m.host}); err != nil {
			return fmt.Errorf("smtp starttls failed: %w", err)
		}
	}

	return sendWithClient(client, auth, m.from, to, body)
}

func (m *smtpMailer) sendImplicitTLS(addr string, auth smtp.Auth, to string, body []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: m.host})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.host)
	if err != nil {
		return err
	}
	defer client.Close()

	return sendWithClient(client, auth, m.from, to, body)
}

func sendWithClient(client *smtp.Client, auth smtp.Auth, from, to string, body []byte) error {
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(body); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

const mimeBoundary = "linkshelf-mail-boundary"

func buildMessage(fromHeader string, msg Message) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", fromHeader)
	fmt.Fprintf(&b, "To: %s\r\n", msg.To)
	fmt.Fprintf(&b, "Subject: %s\r\n", msg.Subject)
	fmt.Fprintf(&b, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: multipart/alternative; boundary=%q\r\n\r\n", mimeBoundary)
	fmt.Fprintf(&b, "--%s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n\r\n", mimeBoundary, msg.TextBody)
	fmt.Fprintf(&b, "--%s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n\r\n", mimeBoundary, msg.HTMLBody)
	fmt.Fprintf(&b, "--%s--\r\n", mimeBoundary)
	return []byte(b.String())
}
