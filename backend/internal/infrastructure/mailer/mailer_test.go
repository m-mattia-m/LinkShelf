package mailer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_NormalizeTlsMode(t *testing.T) {
	require.Equal(t, "none", normalizeTlsMode("none"))
	require.Equal(t, "none", normalizeTlsMode("None"))
	require.Equal(t, "starttls", normalizeTlsMode("starttls"))
	require.Equal(t, "starttls", normalizeTlsMode("STARTTLS"))
	require.Equal(t, "tls", normalizeTlsMode("tls"))
	require.Equal(t, "tls", normalizeTlsMode(""))
	require.Equal(t, "tls", normalizeTlsMode("   "))
	require.Equal(t, "tls", normalizeTlsMode("bogus"))
}

func Test_BuildMessage(t *testing.T) {
	body := string(buildMessage("no-reply@example.com", Message{
		To:       "user@example.com",
		Subject:  "Verify your email",
		TextBody: "click here: https://example.com/verify?token=abc",
		HTMLBody: "<p>click <a href=\"https://example.com/verify?token=abc\">here</a></p>",
	}))

	require.Contains(t, body, "From: no-reply@example.com")
	require.Contains(t, body, "To: user@example.com")
	require.Contains(t, body, "Subject: Verify your email")
	require.Contains(t, body, "Content-Type: multipart/alternative")
	require.Contains(t, body, "click here: https://example.com/verify?token=abc")
	require.Contains(t, body, "<a href=\"https://example.com/verify?token=abc\">here</a>")
	require.True(t, strings.HasSuffix(body, "--"+mimeBoundary+"--\r\n"))
}

func Test_FromHeader(t *testing.T) {
	// No display name configured: the header is just the bare address, same
	// as the envelope.
	m := &smtpMailer{from: "no-reply@example.com"}
	require.Equal(t, "no-reply@example.com", m.fromHeader())

	// A display name is quoted/escaped per RFC 5322 and only ever affects the
	// header - the envelope address (m.from, used separately for MAIL FROM)
	// stays untouched.
	m = &smtpMailer{from: "no-reply@example.com", fromName: "LinkShelf by Fermion"}
	require.Equal(t, `"LinkShelf by Fermion" <no-reply@example.com>`, m.fromHeader())
	require.Equal(t, "no-reply@example.com", m.from)

	// Blank/whitespace-only fromName behaves like it wasn't set at all.
	m = &smtpMailer{from: "no-reply@example.com", fromName: "   "}
	require.Equal(t, "no-reply@example.com", m.fromHeader())
}
