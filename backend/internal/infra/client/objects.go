package client

import (
	"net/url"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"mixm/backend/internal/config"
)

type ObjectStore struct {
	Config config.Config
	HTTP   *HTTPClient
}

// Client 解析对象存储地址并创建 S3 兼容客户端。
func (s *ObjectStore) Client() (*minio.Client, error) {
	u, err := url.Parse(s.Config.RustEndpoint)
	if err != nil {
		return nil, err
	}

	return minio.New(
		u.Host,
		&minio.Options{
			Creds:  credentials.NewStaticV4(s.Config.RustKey, s.Config.RustSecret, ""),
			Secure: u.Scheme == "https",
		},
	)
}
