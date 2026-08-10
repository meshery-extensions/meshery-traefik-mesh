package config

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestGetLatestReleasesUsesTrustedEndpoint(t *testing.T) {
	originalClient := releaseHTTPClient
	t.Cleanup(func() { releaseHTTPClient = originalClient })

	releaseHTTPClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if got := req.URL.Scheme + "://" + req.URL.Host + req.URL.Path; got != traefikReleasesURL {
			t.Fatalf("unexpected release endpoint: %s", got)
		}
		if got := req.URL.Query().Get("per_page"); got != "3" {
			t.Fatalf("unexpected per_page query: %s", got)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`[{"id":1,"name":"v1.4.8"}]`)),
			Header:     make(http.Header),
		}, nil
	})}

	releases, err := GetLatestReleases(3)
	if err != nil {
		t.Fatalf("GetLatestReleases() error = %v", err)
	}
	if len(releases) != 1 || releases[0].Name != "v1.4.8" {
		t.Fatalf("unexpected releases: %#v", releases)
	}
}

func TestGetLatestReleasesRejectsUnexpectedStatus(t *testing.T) {
	originalClient := releaseHTTPClient
	t.Cleanup(func() { releaseHTTPClient = originalClient })

	releaseHTTPClient = &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Body:       io.NopCloser(strings.NewReader("rate limited")),
			Header:     make(http.Header),
		}, nil
	})}

	if _, err := GetLatestReleases(1); err == nil {
		t.Fatal("GetLatestReleases() expected an error for a non-200 response")
	}
}
