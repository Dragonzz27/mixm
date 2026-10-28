package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Address, DatabaseURL, RedisURL, RabbitURL, Queue, Secret, Kubeconfig                    string
	SSHHost, CodeHost, RegistryURL, RegistryPrefix, RegistryUser, RegistryPassword          string
	StorageClass, PVCCapacity, PVCAccessMode, RustEndpoint, RustBucket, RustKey, RustSecret string
	LLMURL, LLMKey, LLMModel, UploadDir, ExportDir, SSHKeyPath                              string
	AccessTTL, RefreshTTL                                                                   time.Duration
	SSHPort, WorkerConcurrency                                                              int
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}
	c := Config{Address: env("HTTP_ADDR", ":8002"), DatabaseURL: env("DATABASE_URL", "postgres://mixm:mixm@localhost:5432/mixm?sslmode=disable"), RedisURL: env("REDIS_URL", "redis://localhost:6379/2"), RabbitURL: env("RABBITMQ_URL", "amqp://mixm:mixm@localhost:5672/"), Queue: env("RABBITMQ_QUEUE", "instance_create"), Secret: env("SECRET_KEY", ""), Kubeconfig: env("K8S_CONFIG_PATH", "kubeconfig/k3s.yaml"), SSHHost: env("K8S_SSH_GATEWAY_HOST", "localhost"), CodeHost: env("K8S_CODE_SERVER_HOST", "localhost"), RegistryURL: env("REGISTRY_URL", "http://localhost:5000"), RegistryPrefix: env("REGISTRY_PREFIX", "localhost:5000"), RegistryUser: env("REGISTRY_USERNAME", ""), RegistryPassword: env("REGISTRY_PASSWORD", ""), StorageClass: env("STORAGE_CLASS_NAME", "local-path"), PVCCapacity: env("USER_PERSONAL_PVC_CAPACITY", "10Gi"), PVCAccessMode: env("PVC_ACCESS_MODE", "ReadWriteMany"), RustEndpoint: env("RUSTFS_ENDPOINT", "http://localhost:19000"), RustBucket: env("RUSTFS_BUCKET", "content"), RustKey: env("RUSTFS_ACCESS_KEY", "rustfsadmin"), RustSecret: env("RUSTFS_SECRET_KEY", "rustfsadmin"), LLMURL: env("LLM_API_URL", "https://api.openai.com/v1"), LLMKey: env("LLM_API_KEY", ""), LLMModel: env("LLM_MODEL", "gpt-4o"), UploadDir: env("UPLOAD_DIR", "../.mixm/backend/uploads/assignments"), ExportDir: env("EXPORT_DIR", "../.mixm/backend/static/export"), SSHKeyPath: env("SSH_HOST_KEY_PATH", "../.mixm/backend/data/ssh_host_key")}
	var err error
	for _, item := range []struct {
		name string
		def  int
		dest *int
	}{{"K8S_SSH_PORT", 30022, &c.SSHPort}, {"WORKER_CONCURRENCY", 4, &c.WorkerConcurrency}} {
		*item.dest, err = integer(item.name, item.def)
		if err != nil {
			return c, err
		}
	}
	for _, item := range []struct {
		name string
		def  int
		unit time.Duration
		dest *time.Duration
	}{{"ACCESS_TOKEN_EXPIRE_MINUTES", 180, time.Minute, &c.AccessTTL}, {"REFRESH_TOKEN_EXPIRE_DAYS", 7, 24 * time.Hour, &c.RefreshTTL}} {
		n, err := integer(item.name, item.def)
		if err != nil {
			return c, err
		}
		*item.dest = time.Duration(n) * item.unit
	}
	if c.Secret == "" {
		return c, fmt.Errorf("SECRET_KEY must be set")
	}
	if env("ALGORITHM", "HS256") != "HS256" {
		return c, fmt.Errorf("only HS256 is supported")
	}
	u, err := url.Parse(c.DatabaseURL)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		return c, fmt.Errorf("DATABASE_URL must use PostgreSQL")
	}
	if c.SSHPort > 65535 || c.WorkerConcurrency > 64 {
		return c, fmt.Errorf("invalid SSH port or worker concurrency")
	}
	return c, nil
}
func env(k, d string) string {
	if v, ok := os.LookupEnv(k); ok {
		return v
	}
	return d
}
func integer(k string, d int) (int, error) {
	n, err := strconv.Atoi(env(k, strconv.Itoa(d)))
	if err != nil || n < 1 {
		return 0, fmt.Errorf("%s must be a positive integer", k)
	}
	return n, nil
}
