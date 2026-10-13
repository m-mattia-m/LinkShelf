// Package webhook sends best-effort event notifications to external services.
//
// Delivery is asynchronous, retried a few times with backoff and then dropped with a
// log line. Nothing is persisted: a receiver must treat a webhook as a hint and be
// able to resync from the API.
package webhook

import (
	"backend/internal/config"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// SignatureHeader carries "t=<unix>,v1=<hex HMAC-SHA256(secret, t + "." + body)>".
	SignatureHeader = "X-LinkShelf-Signature"
	EventHeader     = "X-LinkShelf-Event"
	EventIdHeader   = "X-LinkShelf-Event-Id"

	queueSize = 256
	workers   = 4
)

// Envelope is the body of every delivery.
type Envelope struct {
	Id         string `json:"id"`
	Type       string `json:"type"`
	OccurredAt string `json:"occurred_at"`
	Data       any    `json:"data"`
}

// Sign returns the value of the signature header for body at time t.
func Sign(secret string, t time.Time, body []byte) string {
	stamp := strconv.FormatInt(t.Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stamp + "."))
	mac.Write(body)
	return "t=" + stamp + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}

type delivery struct {
	endpoint config.WebhookEndpoint
	eventId  string
	typ      string
	body     []byte
}

// Dispatcher implements domain.EventPublisher.
type Dispatcher struct {
	endpoints []config.WebhookEndpoint
	client    *http.Client
	// backoff holds the wait before each retry, so a delivery is tried len(backoff)+1 times.
	backoff []time.Duration
	now     func() time.Time

	queue  chan delivery
	cancel context.CancelFunc
	ctx    context.Context
	wg     sync.WaitGroup
}

// New starts a dispatcher for the configured endpoints, nil when there are none.
func New(endpoints []config.WebhookEndpoint) *Dispatcher {
	if len(endpoints) == 0 {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	d := &Dispatcher{
		endpoints: endpoints,
		client:    &http.Client{Timeout: 5 * time.Second},
		backoff:   []time.Duration{time.Second, 5 * time.Second, 25 * time.Second},
		now:       time.Now,
		queue:     make(chan delivery, queueSize),
		ctx:       ctx,
		cancel:    cancel,
	}
	d.start()
	return d
}

func (d *Dispatcher) start() {
	for i := 0; i < workers; i++ {
		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			for {
				select {
				case <-d.ctx.Done():
					return
				case job := <-d.queue:
					d.deliver(job)
				}
			}
		}()
	}
}

// Close stops the workers, dropping what is still queued.
func (d *Dispatcher) Close() {
	if d == nil {
		return
	}
	d.cancel()
	d.wg.Wait()
}

func matches(events []string, eventType string) bool {
	for _, e := range events {
		if e == "*" || e == eventType {
			return true
		}
	}
	return false
}

// Publish queues the event for every endpoint that subscribed to it. It never blocks.
func (d *Dispatcher) Publish(eventType string, data any) {
	if d == nil {
		return
	}
	eventId := "evt_" + uuid.NewString()
	body, err := json.Marshal(Envelope{Id: eventId, Type: eventType, OccurredAt: d.now().UTC().Format(time.RFC3339), Data: data})
	if err != nil {
		zap.L().Error("webhook payload can not be encoded", zap.String("type", eventType), zap.Error(err))
		return
	}

	for _, endpoint := range d.endpoints {
		if !matches(endpoint.Events, eventType) {
			continue
		}
		select {
		case d.queue <- delivery{endpoint: endpoint, eventId: eventId, typ: eventType, body: body}:
		default:
			zap.L().Warn("webhook queue is full, dropping event", zap.String("type", eventType), zap.String("url", endpoint.Url))
		}
	}
}

func (d *Dispatcher) deliver(job delivery) {
	var lastErr error
	for attempt := 0; ; attempt++ {
		if lastErr = d.send(job); lastErr == nil {
			return
		}
		if attempt >= len(d.backoff) {
			break
		}
		select {
		case <-d.ctx.Done():
			return
		case <-time.After(d.backoff[attempt]):
		}
	}
	zap.L().Warn("webhook dropped after retries",
		zap.String("type", job.typ), zap.String("event_id", job.eventId), zap.String("url", job.endpoint.Url), zap.Error(lastErr))
}

func (d *Dispatcher) send(job delivery) error {
	req, err := http.NewRequestWithContext(d.ctx, http.MethodPost, job.endpoint.Url, bytes.NewReader(job.body))
	if err != nil {
		return err
	}
	// Signed anew on every attempt so a retry stays inside the receiver's timestamp window.
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(EventHeader, job.typ)
	req.Header.Set(EventIdHeader, job.eventId)
	req.Header.Set(SignatureHeader, Sign(job.endpoint.Secret, d.now(), job.body))

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("receiver answered %d", resp.StatusCode)
	}
	return nil
}
