package pwextractor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDomain = "dns.google"
var testResult = []string{
	"2001:4860:4860::8844", "2001:4860:4860::8888", "8.8.8.8", "8.8.4.4",
}

func TestGetIPs(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedIPStrings []string
		wantErr           bool
	}{
		{
			name:              "valid IPv4",
			input:             "192.0.2.1",
			expectedIPStrings: []string{"192.0.2.1"},
			wantErr:           false,
		},
		{
			name:              "valid IPv6",
			input:             "2001:db8::1",
			expectedIPStrings: []string{"2001:db8::1"},
			wantErr:           false,
		},
		{
			name:              "URL with IPv4 host",
			input:             "http://192.0.2.1",
			expectedIPStrings: []string{"192.0.2.1"},
			wantErr:           false,
		},
		{
			name:              "URL with IPv6 host",
			input:             "http://[2001:db8::1]",
			expectedIPStrings: []string{"2001:db8::1"},
			wantErr:           false,
		},
		{
			name:              "URL with hostname",
			input:             fmt.Sprintf("https://%s:8080", testDomain),
			expectedIPStrings: testResult,
			wantErr:           false,
		},
		{
			name:              "hostname",
			input:             testDomain,
			expectedIPStrings: testResult,
			wantErr:           false,
		},
		{
			name:    "invalid IP address",
			input:   "256.0.0.0",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid URL format",
			input:   "://invalid",
			wantErr: true,
		},
		{
			name:    "unresolvable hostname",
			input:   "nonexistent.invalid",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ips, err := getIPs(tc.input)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, ips)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, ips, "result slice should not be empty when error is nil")

				if tc.expectedIPStrings != nil {
					ipStrings := make([]string, len(ips))
					for i, ip := range ips {
						ipStrings[i] = ip.String()
					}
					assert.ElementsMatch(t, tc.expectedIPStrings, ipStrings, "IPs do not match expected")
				}
			}
		})
	}
	t.Run("cache set", func(t *testing.T) {
		dnsCache.DeleteAll()
		require.False(t, dnsCache.Has(testDomain))
		ips, _ := getIPs(testDomain)
		require.True(t, dnsCache.Has(testDomain))
		require.ElementsMatch(t, ips, dnsCache.Get(testDomain).Value())
	})
}

func Test_parseBaseDomain(t *testing.T) {
	tests := []struct {
		name           string
		urlStr         string
		expectedDomain string
		expectedScheme string
		expectErr      bool
	}{
		{
			name:           "valid https with subdomain",
			urlStr:         "https://kek.example.com/lol",
			expectedDomain: "example.com",
			expectedScheme: "https",
		},
		{
			name:           "valid http with www subdomain",
			urlStr:         "http://www.example.com/path",
			expectedDomain: "example.com",
			expectedScheme: "http",
		},
		{
			name:           "valid http with no subdomain",
			urlStr:         "http://example.com",
			expectedDomain: "example.com",
			expectedScheme: "http",
		},
		{
			name:           "url with port in host",
			urlStr:         "http://example.com:8080/path",
			expectedDomain: "example.com",
			expectedScheme: "http",
		},
		{
			name:           "url with ip address host",
			urlStr:         "http://192.168.1.1",
			expectedDomain: "192.168.1.1",
			expectedScheme: "http",
		},
		{
			name:           "url with uppercase http scheme",
			urlStr:         "HTTP://EXAMPLE.COM",
			expectedDomain: "example.com",
			expectedScheme: "http",
		},
		{
			name:      "invalid scheme (ftp)",
			urlStr:    "ftp://example.com",
			expectErr: true,
		},
		{
			name:      "no scheme",
			urlStr:    "example.com",
			expectErr: true,
		},
		{
			name:      "invalid url format",
			urlStr:    "http//example.com",
			expectErr: true,
		},
		{
			name:      "empty url string",
			urlStr:    "",
			expectErr: true,
		},
		{
			name:           "url with user info",
			urlStr:         "http://user:pass@example.com",
			expectedDomain: "example.com",
			expectedScheme: "http",
		},
		{
			name:           "url with multiple subdomains",
			urlStr:         "https://a.b.c.example.com",
			expectedDomain: "example.com",
			expectedScheme: "https",
		},
		{
			name:      "url with leading/trailing whitespace",
			urlStr:    " https://example.com ",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain, scheme, err := parseBaseDomain(tt.urlStr)

			if tt.expectErr {
				require.Error(t, err)
				assert.Empty(t, domain)
				assert.Empty(t, scheme)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedDomain, domain)
				assert.Equal(t, tt.expectedScheme, scheme)
			}
		})
	}
}
