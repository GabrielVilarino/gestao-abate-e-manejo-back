package notification

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/SherClockHolmes/webpush-go"
)

type WebPushSender struct {
	publicKey  string
	privateKey string
	subject    string
	client     webpush.HTTPClient
}

type pushIPResolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

type safePushDialer struct {
	resolver pushIPResolver
	dialer   *net.Dialer
}

func NewWebPushSender(publicKey, privateKey, subject string) (*WebPushSender, error) {
	if strings.TrimSpace(publicKey) == "" || strings.TrimSpace(privateKey) == "" || strings.TrimSpace(subject) == "" {
		return nil, errors.New("VAPID_PUBLIC_KEY, VAPID_PRIVATE_KEY e VAPID_SUBJECT são obrigatórios")
	}
	subject = strings.TrimPrefix(strings.TrimSpace(subject), "mailto:")
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = (&safePushDialer{
		resolver: net.DefaultResolver,
		dialer:   &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second},
	}).DialContext
	return &WebPushSender{
		publicKey:  publicKey,
		privateKey: privateKey,
		subject:    subject,
		client: &http.Client{
			Timeout:   20 * time.Second,
			Transport: transport,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return errors.New("redirecionamento de endpoint push bloqueado")
			},
		},
	}, nil
}

func (s *WebPushSender) Send(ctx context.Context, subscription domain.PushSubscription, payload []byte, ttl int) error {
	if err := validatePushEndpoint(subscription.Endpoint); err != nil {
		return err
	}
	if ttl <= 0 {
		return errors.New("TTL da notificação push deve ser positivo")
	}
	if ttl > 60*60 {
		ttl = 60 * 60
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	response, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
		Endpoint: subscription.Endpoint,
		Keys: webpush.Keys{
			P256dh: subscription.P256DH,
			Auth:   subscription.Auth,
		},
	}, &webpush.Options{
		HTTPClient:      s.client,
		Subscriber:      s.subject,
		VAPIDPublicKey:  s.publicKey,
		VAPIDPrivateKey: s.privateKey,
		TTL:             ttl,
		Urgency:         webpush.UrgencyHigh,
	})
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound || response.StatusCode == http.StatusGone {
		return domain.ErrPushSubscriptionGone
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("provedor push respondeu HTTP %d", response.StatusCode)
	}
	return nil
}

func validatePushEndpoint(endpoint string) error {
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil {
		return domain.ErrAssinaturaPushInvalida
	}
	if ip := net.ParseIP(parsed.Hostname()); ip != nil && !isPublicPushIP(ip) {
		return domain.ErrAssinaturaPushInvalida
	}
	return nil
}

func (d *safePushDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	addresses, err := d.resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(addresses) == 0 {
		return nil, errors.New("endpoint push não resolveu para um endereço IP")
	}
	for _, address := range addresses {
		if !isPublicPushIP(address.IP) {
			return nil, errors.New("endpoint push resolveu para endereço local ou privado")
		}
	}
	var lastError error
	for _, address := range addresses {
		connection, err := d.dialer.DialContext(ctx, network, net.JoinHostPort(address.IP.String(), port))
		if err == nil {
			return connection, nil
		}
		lastError = err
	}
	return nil, lastError
}

func isPublicPushIP(ip net.IP) bool {
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return false
	}
	if ipv4 := ip.To4(); ipv4 != nil && ipv4[0] == 100 && ipv4[1] >= 64 && ipv4[1] <= 127 {
		return false
	}
	return true
}
