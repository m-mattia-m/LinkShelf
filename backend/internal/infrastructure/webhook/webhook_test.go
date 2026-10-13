package webhook

import (
	"backend/internal/config"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testSecret = "webhook-test-secret-0123456789-abcdef"

type received struct {
	header http.Header
	body   []byte
}

func receiver(t *testing.T, failFirst int32) (*httptest.Server, chan received, *atomic.Int32) {
	t.Helper()
	got := make(chan received, 16)
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		body, _ := io.ReadAll(r.Body)
		if n <= failFirst {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		got <- received{header: r.Header.Clone(), body: body}
		w.WriteHeader(http.StatusAccepted)
	}))
	t.Cleanup(srv.Close)
	return srv, got, &calls
}

func newTestDispatcher(t *testing.T, url string, events []string) *Dispatcher {
	t.Helper()
	d := New([]config.WebhookEndpoint{{Url: url, Secret: testSecret, Events: events}})
	d.backoff = []time.Duration{time.Millisecond, time.Millisecond}
	t.Cleanup(d.Close)
	return d
}

func wait(t *testing.T, got chan received) received {
	t.Helper()
	select {
	case r := <-got:
		return r
	case <-time.After(3 * time.Second):
		t.Fatal("no webhook received")
		return received{}
	}
}

func Test_Sign_MatchesDocumentedFormat(t *testing.T) {
	at := time.Unix(1700000000, 0)
	body := []byte(`{"a":1}`)

	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write([]byte("1700000000." + string(body)))

	require.Equal(t, "t=1700000000,v1="+hex.EncodeToString(mac.Sum(nil)), Sign(testSecret, at, body))
}

func Test_Dispatcher_DeliversSignedEnvelope(t *testing.T) {
	srv, got, _ := receiver(t, 0)
	d := newTestDispatcher(t, srv.URL, []string{"shelf.domain.changed"})

	d.Publish("shelf.domain.changed", map[string]string{"domain": "profile.example.com"})
	r := wait(t, got)

	var envelope Envelope
	require.NoError(t, json.Unmarshal(r.body, &envelope))
	require.True(t, strings.HasPrefix(envelope.Id, "evt_"))
	require.Equal(t, "shelf.domain.changed", envelope.Type)
	require.Equal(t, envelope.Id, r.header.Get(EventIdHeader))
	require.Equal(t, "shelf.domain.changed", r.header.Get(EventHeader))

	signature := r.header.Get(SignatureHeader)
	parts := strings.Split(signature, ",")
	require.Len(t, parts, 2)
	stamp, err := strconv.ParseInt(strings.TrimPrefix(parts[0], "t="), 10, 64)
	require.NoError(t, err)
	require.Equal(t, signature, Sign(testSecret, time.Unix(stamp, 0), r.body))
	require.WithinDuration(t, time.Now(), time.Unix(stamp, 0), 5*time.Second)
}

func Test_Dispatcher_RetriesThenSucceeds(t *testing.T) {
	srv, got, calls := receiver(t, 2)
	d := newTestDispatcher(t, srv.URL, []string{"*"})

	d.Publish("shelf.domain.changed", nil)
	wait(t, got)
	require.EqualValues(t, 3, calls.Load())
}

func Test_Dispatcher_DropsAfterRetries(t *testing.T) {
	srv, got, calls := receiver(t, 100)
	d := newTestDispatcher(t, srv.URL, []string{"*"})

	d.Publish("shelf.domain.changed", nil)
	require.Eventually(t, func() bool { return calls.Load() == 3 }, 3*time.Second, 5*time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	require.EqualValues(t, 3, calls.Load())
	require.Empty(t, got)
}

func Test_Dispatcher_FiltersByEventType(t *testing.T) {
	srv, got, calls := receiver(t, 0)
	d := newTestDispatcher(t, srv.URL, []string{"shelf.domain.changed"})

	d.Publish("user.created", nil)
	d.Publish("shelf.domain.changed", nil)
	r := wait(t, got)

	require.Equal(t, "shelf.domain.changed", r.header.Get(EventHeader))
	require.EqualValues(t, 1, calls.Load())
}

func Test_New_WithoutEndpointsIsNilAndSafe(t *testing.T) {
	var d = New(nil)
	require.Nil(t, d)
	d.Publish("x", nil)
	d.Close()
}
