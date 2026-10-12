package cloudflare

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"

	"github.com/qdm12/ddns-updater/internal/provider/errors"
	"github.com/qdm12/ddns-updater/pkg/publicip/ipversion"
	"github.com/stretchr/testify/require"
)

type rewriteTransport struct {
	host string
	base http.RoundTripper
}

func (t rewriteTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	request.URL.Scheme = "http"
	request.URL.Host = t.host
	return t.base.RoundTrip(request)
}

func newTestClient(server *httptest.Server) *http.Client {
	return &http.Client{Transport: rewriteTransport{
		host: strings.TrimPrefix(server.URL, "http://"),
		base: http.DefaultTransport,
	}}
}

func TestValidateSettingsAcceptsUnderscoreInKey(t *testing.T) {
	t.Parallel()

	err := validateSettings("example.com", "you@example.com", "api_key_with_underscores", "", "zone-id", 300)

	require.NoError(t, err)
}

func TestUpdateUsesCurrentCloudflareListAPI(t *testing.T) {
	t.Parallel()

	type requestObservation struct {
		method        string
		path          string
		recordType    string
		recordName    string
		page          string
		perPage       string
		authorization string
	}
	requestObserved := make(chan requestObservation, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestObserved <- requestObservation{
			method:        r.Method,
			path:          r.URL.Path,
			recordType:    r.URL.Query().Get("type"),
			recordName:    r.URL.Query().Get("name"),
			page:          r.URL.Query().Get("page"),
			perPage:       r.URL.Query().Get("per_page"),
			authorization: r.Header.Get("Authorization"),
		}
		_, _ = io.WriteString(w, `{"success":true,"errors":[],"result":[{"id":"record-id","content":"192.0.2.1"}]}`)
	}))
	defer server.Close()

	provider := &Provider{
		domain:         "example.com",
		owner:          "@",
		ipVersion:      ipversion.IP4,
		token:          "api-token",
		zoneIdentifier: "zone-id",
		ttl:            300,
	}

	gotIP, err := provider.Update(context.Background(), newTestClient(server), netip.MustParseAddr("192.0.2.1"))

	require.NoError(t, err)
	gotRequest := <-requestObserved
	require.Equal(t, http.MethodGet, gotRequest.method)
	require.Equal(t, "/client/v4/zones/zone-id/dns_records", gotRequest.path)
	require.Equal(t, "A", gotRequest.recordType)
	require.Equal(t, "example.com", gotRequest.recordName)
	require.Equal(t, "1", gotRequest.page)
	require.Equal(t, "1", gotRequest.perPage)
	require.Equal(t, "Bearer api-token", gotRequest.authorization)
	require.Equal(t, netip.MustParseAddr("192.0.2.1"), gotIP)
}

func TestUpdateReportsStructuredCloudflareErrors(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"success":false,"errors":[{"code":10000,"message":"Authentication error"}],"result":[]}`)
	}))
	defer server.Close()

	provider := &Provider{
		domain:         "example.com",
		owner:          "@",
		ipVersion:      ipversion.IP4,
		token:          "api-token",
		zoneIdentifier: "zone-id",
		ttl:            300,
	}

	_, err := provider.Update(context.Background(), newTestClient(server), netip.MustParseAddr("192.0.2.1"))

	require.ErrorIs(t, err, errors.ErrUnsuccessful)
	require.ErrorContains(t, err, "error 10000: Authentication error")
}

func TestUpdateRejectsNon2xxMutationResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = io.WriteString(w, `{"success":true,"errors":[],"result":[{"id":"record-id","content":"192.0.2.2"}]}`)
		case http.MethodPut:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"success":false,"errors":[{"code":10000,"message":"Invalid request"}]}`)
		default:
			t.Errorf("unexpected HTTP method %s", r.Method)
		}
	}))
	defer server.Close()

	provider := &Provider{
		domain:         "example.com",
		owner:          "@",
		ipVersion:      ipversion.IP4,
		token:          "api-token",
		zoneIdentifier: "zone-id",
		ttl:            300,
	}

	_, err := provider.Update(context.Background(), newTestClient(server), netip.MustParseAddr("192.0.2.1"))

	require.ErrorIs(t, err, errors.ErrHTTPStatusNotValid)
	require.ErrorContains(t, err, "400")
}
