// Package storage wraps the Supabase Storage REST API for the bytes-only
// operations we need (upload + delete). Authentication uses the project's
// service_role key, so this client must only be constructed server-side.
package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"

	storage_go "github.com/supabase-community/storage-go"
	"github.com/supabase-community/supabase-go"
)

type Supabase struct {
	bucket  string
	baseURL string
	client  *supabase.Client
}

// NewSupabase builds a client. baseURL is the project URL (no trailing slash),
// e.g. "https://abc.supabase.co".
func NewSupabase(baseURL, serviceKey, bucket string) *Supabase {
	client, err := supabase.NewClient(baseURL, serviceKey, &supabase.ClientOptions{})
	if err != nil {
		log.Fatalf("failed to create Supabase client: %v", err)
	}

	return &Supabase{
		baseURL: baseURL,
		bucket:  bucket,
		client:  client,
	}
}

// Upload streams body into {bucket}/{objectPath} and returns the public URL.
// If the bucket is private, the URL will return 400 for unauthenticated GETs
// — switch to SignedURL in that case.
func (s *Supabase) Upload(ctx context.Context, objectPath, contentType string, body io.Reader) (string, error) {
	isUpsert := true
	result, err := s.client.Storage.UploadFile(s.bucket, encodePath(objectPath), body, storage_go.FileOptions{ContentType: &contentType, Upsert: &isUpsert})
	if err != nil {
		return "", fmt.Errorf("unable to upload to supabase path '%s': %w", encodePath(objectPath), err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("supabase upload: %s", result.Error)
	}

	url := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.baseURL, s.bucket, encodePath(objectPath))
	return url, nil
}

// Delete removes a single object from the bucket. A 404 from the API is
// treated as success so callers can safely retry.
func (s *Supabase) Delete(ctx context.Context, objectPath string) error {
	result, err := s.client.Storage.RemoveFile(s.bucket, []string{encodePath(objectPath)})
	if err != nil {
		return err
	}
	if len(result) > 0 && result[0].Error != "" {
		return fmt.Errorf("supabase delete: %s", result[0].Error)
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
