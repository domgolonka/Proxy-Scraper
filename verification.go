package proxy

import (
	"bytes"
	"encoding/json"
	"github.com/domgolonka/proxy-scraper/models"
	"github.com/domgolonka/proxy-scraper/providers"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

type checkIP struct {
	IP string
}

func verifyProxy(logger *slog.Logger, proxy models.Proxy) bool {
	req, err := http.NewRequest("GET", "https://api.ipify.org/?format=json", nil)
	if err != nil {
		logger.Error("cannot create new request for verify err: %s", slog.String("error", err.Error()))
		return false
	}

	proxyURL, err := url.Parse(proxy.ToString())
	if err != nil {
		logger.Error("cannot parse proxy %q err: %s", slog.Any("error", proxy), slog.String("error", err.Error()))
		return false
	}

	client := providers.NewClient()
	client.Transport.(*http.Transport).Proxy = http.ProxyURL(proxyURL)

	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		logger.Debug("cannot verify proxy %q err:%s", slog.Any("error", proxy), slog.String("error", err.Error()))
		return false
	}

	var body bytes.Buffer
	if _, err := io.Copy(&body, resp.Body); err != nil {
		logger.Error("cannot copy resp.Body err: %s", slog.String("error", err.Error()))
		return false
	}

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var check checkIP
	if err := json.Unmarshal(body.Bytes(), &check); err != nil {
		logger.Error("%d cannot unmarshal %q to checkIP struct err: %s", slog.Int("status", resp.StatusCode), slog.String("body", body.String()), slog.String("error", err.Error()))
		return false
	}

	return strings.HasPrefix(proxy.ToString(), check.IP)
}
