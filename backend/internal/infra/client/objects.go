package client

import (
	"net/url"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"mixm-edu/backend/internal/config"
)

type ObjectStore struct {
	Config config.Config
	HTTP   *HTTPClient
}

func (s *ObjectStore) Client() (*minio.Client, error) {
	u, err := url.Parse(s.Config.RustEndpoint)
	if err != nil {
		return nil, err
	}
	return minio.New(u.Host, &minio.Options{Creds: credentials.NewStaticV4(s.Config.RustKey, s.Config.RustSecret, ""), Secure: u.Scheme == "https"})
}
