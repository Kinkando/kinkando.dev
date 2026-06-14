// Package storage wraps the Supabase Storage REST API for the bytes-only
// operations we need (upload + delete). Authentication uses the project's
// service_role key, so this client must only be constructed server-side.
package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Supabase struct {
	baseURL    string
	serviceKey string
	bucket     string
	httpClient *http.Client
}

// NewSupabase builds a client. baseURL is the project URL (no trailing slash),
// e.g. "https://abc.supabase.co".
func NewSupabase(baseURL, serviceKey, bucket string) *Supabase {
	return &Supabase{
		baseURL:    strings.TrimRight(baseURL, "/"),
		serviceKey: serviceKey,
		bucket:     bucket,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// Upload streams body into {bucket}/{objectPath} and returns the public URL.
// If the bucket is private, the URL will return 400 for unauthenticated GETs
// — switch to SignedURL in that case.
func (s *Supabase) Upload(ctx context.Context, objectPath, contentType string, body io.Reader) (string, error) {
	endpoint := fmt.Sprintf("%s/storage/v1/object/%s/%s",
		s.baseURL, s.bucket, encodePath(objectPath))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	// Without this header Supabase rejects duplicate paths with 409; we want
	// upload-or-replace semantics since paths embed a timestamp.
	req.Header.Set("x-upsert", "true")

	res, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close() //nolint:errcheck

	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("supabase upload %d: %s", res.StatusCode, string(b))
	}
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s",
		s.baseURL, s.bucket, encodePath(objectPath)), nil
}

// Delete removes a single object from the bucket. A 404 from the API is
// treated as success so callers can safely retry.
func (s *Supabase) Delete(ctx context.Context, objectPath string) error {
	endpoint := fmt.Sprintf("%s/storage/v1/object/%s/%s",
		s.baseURL, s.bucket, encodePath(objectPath))
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)

	res, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close() //nolint:errcheck

	if res.StatusCode == http.StatusNotFound {
		return nil
	}
	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("supabase delete %d: %s", res.StatusCode, string(b))
	}
	return nil
}

// encodePath percent-encodes each path segment while keeping the slashes.
func encodePath(p string) string {
	parts := strings.Split(p, "/")
	for i, seg := range parts {
		parts[i] = url.PathEscape(seg)
	}
	return strings.Join(parts, "/")
}
