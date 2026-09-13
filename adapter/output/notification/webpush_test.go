package notification

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

type webPushHTTPClientStub struct {
	status  int
	called  bool
	request *http.Request
}

func (s *webPushHTTPClientStub) Do(request *http.Request) (*http.Response, error) {
	s.called = true
	s.request = request
	return &http.Response{
		StatusCode: s.status,
		Body:       io.NopCloser(strings.NewReader("")),
		Request:    request,
	}, nil
}

func TestNewWebPushSenderRequiresVAPIDConfiguration(t *testing.T) {
	if _, err := NewWebPushSender("", "private", "mailto:admin@example.com"); err == nil {
		t.Fatal("esperava erro para chave pública vazia")
	}
}

func TestWebPushSenderMapsGoneSubscription(t *testing.T) {
	sender, err := NewWebPushSender("test-public", "test-private", "mailto:admin@example.com")
	if err != nil {
		t.Fatal(err)
	}
	client := &webPushHTTPClientStub{status: http.StatusGone}
	sender.client = client

	err = sender.Send(context.Background(), validPushSubscription(), []byte(`{"title":"Lembrete"}`), 3600)

	if !client.called || !errors.Is(err, domain.ErrPushSubscriptionGone) {
		t.Fatalf("called=%v erro=%v", client.called, err)
	}
}

func TestWebPushSenderAcceptsSuccessfulProviderResponse(t *testing.T) {
	sender, err := NewWebPushSender("test-public", "test-private", "admin@example.com")
	if err != nil {
		t.Fatal(err)
	}
	client := &webPushHTTPClientStub{status: http.StatusCreated}
	sender.client = client

	if err := sender.Send(context.Background(), validPushSubscription(), []byte(`{"title":"Lembrete"}`), 900); err != nil {
		t.Fatal(err)
	}
	if !client.called {
		t.Fatal("cliente HTTP não foi chamado")
	}
	if client.request.Header.Get("TTL") != "900" {
		t.Fatalf("TTL = %q", client.request.Header.Get("TTL"))
	}
}

func TestWebPushSenderRejectsPrivateEndpointBeforeSending(t *testing.T) {
	sender, err := NewWebPushSender("test-public", "test-private", "admin@example.com")
	if err != nil {
		t.Fatal(err)
	}
	client := &webPushHTTPClientStub{status: http.StatusCreated}
	sender.client = client
	subscription := validPushSubscription()
	subscription.Endpoint = "https://127.0.0.1/push"

	err = sender.Send(context.Background(), subscription, []byte(`{"title":"Lembrete"}`), 60)

	if !errors.Is(err, domain.ErrAssinaturaPushInvalida) || client.called {
		t.Fatalf("erro=%v called=%v", err, client.called)
	}
}

type pushResolverStub struct {
	addresses []net.IPAddr
}

func (s pushResolverStub) LookupIPAddr(context.Context, string) ([]net.IPAddr, error) {
	return s.addresses, nil
}

func TestSafePushDialerRejectsHostnameResolvingToPrivateIP(t *testing.T) {
	dialer := &safePushDialer{
		resolver: pushResolverStub{addresses: []net.IPAddr{{IP: net.ParseIP("10.0.0.5")}}},
		dialer:   &net.Dialer{},
	}

	if _, err := dialer.DialContext(context.Background(), "tcp", "push.example:443"); err == nil {
		t.Fatal("esperava bloqueio de IP privado resolvido por DNS")
	}
}

func TestWebPushHTTPClientRejectsRedirects(t *testing.T) {
	sender, err := NewWebPushSender("test-public", "test-private", "admin@example.com")
	if err != nil {
		t.Fatal(err)
	}
	client, ok := sender.client.(*http.Client)
	if !ok || client.CheckRedirect == nil {
		t.Fatal("cliente HTTP sem política de redirecionamento")
	}
	if err := client.CheckRedirect(&http.Request{}, []*http.Request{{}}); err == nil {
		t.Fatal("redirecionamento deveria ser bloqueado")
	}
}

func validPushSubscription() domain.PushSubscription {
	return domain.PushSubscription{
		Endpoint: "https://updates.push.services.mozilla.com/wpush/v2/device",
		P256DH:   "BNNL5ZaTfK81qhXOx23-wewhigUeFb632jN6LvRWCFH1ubQr77FE_9qV1FuojuRmHP42zmf34rXgW80OvUVDgTk",
		Auth:     "zqbxT6JKstKSY9JKibZLSQ",
	}
}
