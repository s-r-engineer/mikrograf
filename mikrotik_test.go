package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/s-r-engineer/library/lineProtocol"
	"github.com/stretchr/testify/require"
)

const username = "testOfTheMikrotik"

func generateHandler() func(http.ResponseWriter, *http.Request) {
	crackers := map[string]string{}
	for k, v := range modules {
		crackers[v] = k
	}
	crackers["/rest/system/routerboard"] = "system_routerboard"
	return func(w http.ResponseWriter, r *http.Request) {
		filePath := "./testData/" + crackers[r.URL.Path] + ".json"
		data, err := os.ReadFile(filePath)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			_, err = w.Write(data)
			if err != nil {
				panic(err)
			}
		}
	}
}

func getShit() (*url.URL, *lineProtocol.Accumulator, *httptest.Server) {
	fakeServer := httptest.NewServer(http.HandlerFunc(generateHandler()))
	newUrl, _ := url.Parse(fakeServer.URL)
	acc := lineProtocol.NewAccumulator()
	return newUrl, &acc, fakeServer
}

func TestUptimeConverter(t *testing.T) {
	values := map[string]time.Duration{
		"1d0h0m0s":  time.Duration(24) * time.Hour,
		"0d0m0s":    0,
		"1s0m0s":    0,
		"0d0a0s":    0,
		"2m29s":     time.Duration(29)*time.Second + time.Duration(2)*time.Minute,
		"23m35s":    time.Duration(35)*time.Second + time.Duration(23)*time.Minute,
		"1d0h52m0s": time.Duration(24)*time.Hour + time.Duration(52)*time.Minute,
		"12s":       time.Duration(12) * time.Second,
		"121d12h12m0s": (time.Duration(24*121) * time.Hour) +
			(time.Duration(12) * time.Hour) +
			(time.Duration(12) * time.Minute) +
			(time.Duration(0) * time.Second),
	}
	for uptime, result := range values {
		d, err := parseUptimeIntoDuration(uptime)
		require.NoError(t, err)
		require.Equal(t, d, int64(result/time.Second))
	}
}

func TestMikrotikBase(t *testing.T) {
	newUrl, acc, fServer := getShit()
	defer fServer.Close()
	plugin, err := newMikrotik(fmt.Sprintf("%s://%s:@%s?modules=interface", newUrl.Scheme, username, newUrl.Host), acc)
	require.NoError(t, err)
	require.NoError(t, plugin.Run())
	var metric = string(acc.GetBytes())
	require.Equal(t, "mikrotik", metric[:8])
}

func TestMikrotikDisabled(t *testing.T) {
	newUrl, acc, fServer1 := getShit()
	defer fServer1.Close()
	plugin, err := newMikrotik(fmt.Sprintf("%s://%s:@%s?modules=interface", newUrl.Scheme, username, newUrl.Host), acc)
	require.NoError(t, err)
	require.NoError(t, plugin.Run())
	parsedData, err := lineProtocol.LineProtocolParser(string(acc.GetBytes()))
	require.NoError(t, err)
	require.Len(t, parsedData, 2)
	
	newUrl, acc, fServer2 := getShit()
	defer fServer2.Close()
	plugin, err = newMikrotik(fmt.Sprintf("%s://%s:@%s?modules=interface&ignoreDisabled=false", newUrl.Scheme, username, newUrl.Host), acc)
	require.NoError(t, err)
	require.NoError(t, plugin.Run())
	parsedData, err = lineProtocol.LineProtocolParser(string(acc.GetBytes()))
	require.NoError(t, err)
	require.Len(t, parsedData, 3)
}

func TestMikrotikCheckCorrectAmountOfMetricsWithOneEmptyAndOneDisabled(t *testing.T) {
	newUrl, acc, fServer := getShit()
	defer fServer.Close()
	plugin, err := newMikrotik(fmt.Sprintf("%s://%s:@%s?modules=interface", newUrl.Scheme, username, newUrl.Host), acc)
	require.NoError(t, err)
	require.NoError(t, plugin.Run())
	parsedData, err := lineProtocol.LineProtocolParser(string(acc.GetBytes()))
	require.NoError(t, err)
	require.Len(t, parsedData, 2)
}

func TestMikrotikAllDataPoints(t *testing.T) {
	newUrl, acc, fServer := getShit()
	defer fServer.Close()
	includeModules := "interface,interface_wireguard_peers,interface_wireless_registration,ip_dhcp_server_lease,ip_firewall_connection,ip_firewall_filter,ip_firewall_nat,ip_firewall_mangle,ipv6_firewall_connection,ipv6_firewall_filter,ipv6_firewall_nat,ipv6_firewall_mangle,system_script,system_resourses"
	plugin, err := newMikrotik(fmt.Sprintf("%s://%s:@%s?modules=%s", newUrl.Scheme, username, newUrl.Host, includeModules), acc)
	require.NoError(t, err)
	require.NoError(t, plugin.Run())
	parsedData, err := lineProtocol.LineProtocolParser(string(acc.GetBytes()))
	require.NoError(t, err)
	require.Len(t, parsedData, 16)
}

func TestMikrotikConfigurationErrors(t *testing.T) {
	newUrl, acc, fServer := getShit()
	defer fServer.Close()
	includeModules := "interfacessss"
	_, err := newMikrotik(fmt.Sprintf("%s://%s:@%s?modules=%s", newUrl.Scheme, username, newUrl.Host, includeModules), acc)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "mikrotik init ->"))
}

func TestMikrotikDataIsCorrect(t *testing.T) {
	newUrl, acc, fServer := getShit()
	defer fServer.Close()
	includeModules := "interface"
	plugin, err := newMikrotik(fmt.Sprintf("%s://%s:@%s?modules=%s", newUrl.Scheme, username, newUrl.Host, includeModules), acc)
	require.NoError(t, err)
	require.NoError(t, plugin.Run())
	parsedData, err := lineProtocol.LineProtocolParser(string(acc.GetBytes()))
	require.NoError(t, err)
	require.Len(t, parsedData, 2)
	requiredFields := map[string]int64{
		"tx-packet":     56850710,
		"tx-error":      0,
		"fp-rx-packet":  144047662,
		"fp-tx-byte":    0,
		"rx-byte":       194632346152,
		"rx-error":      0,
		"rx-packet":     144047662,
		"fp-rx-byte":    194056155504,
		"tx-queue-drop": 0,
		"rx-drop":       0,
		"tx-byte":       15355309685,
		"tx-drop":       0,
		"fp-tx-packet":  0,
	}

	requiredTags := map[string]string{
		".id":               "*1",
		"architecture-name": "arm",
		"board-name":        "hAP",
		"cpu":               "ARM",
		"current-firmware":  "7.15.3",
		"default-name":      "ether1",
		"disabled":          "false",
		"firmware-type":     "ipq4000L",
		"mac-address":       "00:11:22:33:44:55",
		"model":             "RBD52G-5HacD2HnD",
		"name":              "ether1",
		"platform":          "MikroTik",
		"version":           "7.16 (stable)",
		"running":           "true",
		"serial-number":     "123456789",
		"type":              "ether",
	}

	fields := parsedData[0].Fields
	for k := range fields {
		require.Equal(t, fmt.Sprintf("%di", requiredFields[k]), fields[k])
		delete(requiredFields, k)
	}
	require.Empty(t, requiredFields)

	tags := parsedData[0].Tags
	for k := range tags {
		if _, ok := requiredTags[k]; ok {
			require.Equal(t, requiredTags[k], tags[k])
			delete(requiredTags, k)
		}
	}
	require.Empty(t, requiredTags)
}

func TestMikrotikCommentExclusion(t *testing.T) {
	newUrl, acc, fServer := getShit()
	defer fServer.Close()
	includeModules := "interface,interface_wireguard_peers,interface_wireless_registration,ip_dhcp_server_lease,ip_firewall_connection,ip_firewall_filter,ip_firewall_nat,ip_firewall_mangle,ipv6_firewall_connection,ipv6_firewall_filter,ipv6_firewall_nat,ipv6_firewall_mangle,system_script,system_resourses"
	plugin, err := newMikrotik(fmt.Sprintf("%s://%s:@%s?modules=%s&ignoreComments=%s", newUrl.Scheme, username, newUrl.Host, includeModules, "ignoreThis,ignoreThat"), acc)
	require.NoError(t, err)
	require.NoError(t, plugin.Run())
	parsedData, err := lineProtocol.LineProtocolParser(string(acc.GetBytes()))
	require.NoError(t, err)
	require.Len(t, parsedData, 15)
}
