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
