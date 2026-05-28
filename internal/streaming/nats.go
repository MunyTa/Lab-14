package streaming

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/MunyTa/Lab-14/internal/market"
)

type NATSPublisher struct {
	conn    net.Conn
	reader  *bufio.Reader
	subject string
}

func NewNATSPublisher(ctx context.Context, rawURL, subject string) (*NATSPublisher, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, errors.New("nats url is required")
	}
	if subject == "" {
		subject = "lab14.crypto.candles"
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "nats" {
		return nil, fmt.Errorf("unsupported NATS scheme %q", parsed.Scheme)
	}

	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", parsed.Host)
	if err != nil {
		return nil, err
	}

	publisher := &NATSPublisher{
		conn:    conn,
		reader:  bufio.NewReader(conn),
		subject: subject,
	}
	if err := publisher.handshake(); err != nil {
		conn.Close()
		return nil, err
	}
	return publisher, nil
}

func (p *NATSPublisher) PublishCandles(candles []market.Candle) error {
	for _, candle := range candles {
		payload, err := json.Marshal(candle)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(p.conn, "PUB %s %d\r\n%s\r\n", p.subject, len(payload), payload); err != nil {
			return err
		}
	}
	_, err := p.conn.Write([]byte("PING\r\n"))
	if err != nil {
		return err
	}
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return err
	}
	if strings.TrimSpace(line) != "PONG" {
		return fmt.Errorf("unexpected NATS response %q", strings.TrimSpace(line))
	}
	return nil
}

func (p *NATSPublisher) Close() error {
	if p == nil || p.conn == nil {
		return nil
	}
	return p.conn.Close()
}

func (p *NATSPublisher) handshake() error {
	if err := p.conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return err
	}
	if !strings.HasPrefix(line, "INFO ") {
		return fmt.Errorf("unexpected NATS greeting %q", strings.TrimSpace(line))
	}
	if _, err := p.conn.Write([]byte("CONNECT {\"verbose\":false,\"pedantic\":false}\r\nPING\r\n")); err != nil {
		return err
	}
	line, err = p.reader.ReadString('\n')
	if err != nil {
		return err
	}
	if strings.TrimSpace(line) != "PONG" {
		return fmt.Errorf("unexpected NATS handshake response %q", strings.TrimSpace(line))
	}
	return p.conn.SetDeadline(time.Time{})
}
