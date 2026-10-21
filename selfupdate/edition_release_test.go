package selfupdate

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestEditionReleaseVersions(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want int
	}{
		{"0.3.15", "v0.3.15-ko.1", -1},
		{"0.3.15-ko.9", "v0.3.15-global.1", -1},
		{"v0.3.15-global.1", "0.3.15-global.1", 0},
		{"0.3.15-global.2", "0.3.15-global.10", -1},
		{"0.3.16-ko.1", "0.3.15-global.10", 1},
		{"0.3.15-global.1", "0.3.15-ko.1", 1},
	} {
		if got, ok := CompareVersions(c.a, c.b); !ok || got != c.want {
			t.Errorf("版本比较 %q / %q = %d, %v；期望 %d, true", c.a, c.b, got, ok, c.want)
		}
	}
	for _, version := range []string{"dev", "0.3.15-global.1-dirty", "0.3.15-global.1-2-gabc1234", "0.3.15-rc.1", "0.3.15-global.0", "0.3.15-global.+1", "0.3.15-global.1.2", "0.3.15-ko.", "0.3.15-global.999999999999999999999999"} {
		if _, ok := CompareVersions(version, "0.3.16-global.1"); ok {
			t.Errorf("不应自动更新未知或开发版本 %q", version)
		}
	}
}

type editionTransport func(*http.Request) (*http.Response, error)

func (f editionTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestFetchLatestUsesEditionReleaseSource(t *testing.T) {
	client := &http.Client{Transport: editionTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != "https://api.github.com/repos/jiwoochris/artex-global/releases/latest" {
			t.Fatalf("请求到了错误的发行源：%s", r.URL)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"tag_name":"v0.3.15-global.1","assets":[{"name":"artex-0.3.15-global.1-windows-amd64.zip"}]}`)), Header: make(http.Header)}, nil
	})}
	release, err := FetchLatest(context.Background(), client)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := release.FindAsset(AssetName(release.TagName, "windows", "amd64")); !ok {
		t.Fatal("无法按全球版标签定位发布附件")
	}
}
