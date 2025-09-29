package youdao

import (
	"cmp"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lftk/anki-vocab/internal/dict"
)

type Config struct {
	UserAgent string `yaml:"user_agent"`
}

func New(cfg *Config) *dict.Dict {
	d := &Dict{
		client:    http.DefaultClient,
		userAgent: cmp.Or(cfg.UserAgent, defaultUserAgent),
	}
	caps := &dict.Capabilities{
		Query: &dict.QueryCapabilities{
			AI: false,
		},
		Pronounce: &dict.PronounceCapabilities{
			Accents: []string{"us", "uk"},
			Formats: []string{"mp3"},
		},
	}
	return &dict.Dict{Queryer: d, Pronouncer: d, Capabilities: caps}
}

var defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36"

type Dict struct {
	client    *http.Client
	userAgent string
}

func (d *Dict) Query(ctx context.Context, word string) ([]byte, error) {
	t := len(word+"webdict") % 10

	vals := make(url.Values)
	vals.Set("q", word)
	vals.Set("le", "en")
	vals.Set("t", fmt.Sprint(t))
	vals.Set("client", "web")
	vals.Set("keyfrom", "webdict")

	s := fmt.Sprintf("web%s%d%s%s",
		word, t, "Mk6hqtUp33DGGtoS63tTJbMUYjRrG1Lu", md5Sum(word+"webdict"),
	)
	vals.Set("sign", md5Sum(s))

	url := "https://dict.youdao.com/jsonapi_s?doctype=json&jsonversion=4"
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(vals.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := d.invoke(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func accentType(accent string) string {
	switch accent {
	case "us":
		return "2"
	case "uk":
		return "1"
	default:
		return "2"
	}
}

func (d *Dict) Pronounce(ctx context.Context, word, accent, format string) (io.ReadCloser, error) {
	typ := accentType(accent)
	audio, err := d.pronounce(ctx, word, typ)
	if err != nil {
		// Fallback to the second pronounce method
		audio, err = d.pronounce2(ctx, word, typ)
		if err != nil {
			return nil, err
		}
	}
	return audio, nil
}

func (d *Dict) pronounce(ctx context.Context, word, typ string) (io.ReadCloser, error) {
	vals := make(url.Values)
	vals.Set("audio", word)
	vals.Set("type", typ)

	url := "https://dict.youdao.com/dictvoice?" + vals.Encode()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := d.invoke(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (d *Dict) pronounce2(ctx context.Context, word, typ string) (io.ReadCloser, error) {
	params := []struct {
		key, val string
	}{
		{"appVersion", "1"},
		{"client", "web"},
		{"imei", "1"},
		{"keyfrom", "dict"},
		{"keyid", "voiceDictWeb"},
		{"mid", "1"},
		{"model", "1"},
		{"mysticTime", fmt.Sprint(time.Now().UnixMilli())},
		{"network", "wifi"},
		{"product", "webdict"},
		{"rate", "4"},
		{"screen", "1"},
		{"type", typ},
		{"vendor", "web"},
		{"word", word},
		{"yduuid", "abcdefg"},
		{"key", "U3uACNRWSDWdcsKm"},
	}

	vals := make(url.Values)
	keys := make([]string, 0, len(params))
	parts := make([]string, 0, len(params))
	for _, kv := range params {
		vals.Set(kv.key, kv.val)
		keys = append(keys, kv.key)
		parts = append(parts, fmt.Sprintf("%s=%s", kv.key, kv.val))
	}

	vals.Set("sign", md5Sum(strings.Join(parts, "&")))
	vals.Set("pointParam", strings.Join(keys, ","))
	vals.Del("key")

	url := "https://dict.youdao.com/pronounce/base?" + vals.Encode()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := d.invoke(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (d *Dict) invoke(req *http.Request) (*http.Response, error) {
	req.Header.Set("Origin", "https://www.youdao.com")
	req.Header.Set("Referer", "https://www.youdao.com/")
	req.Header.Set("User-Agent", d.userAgent)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp, nil
}

func md5Sum(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}
