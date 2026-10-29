package ssh

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/subtle"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	ssh "golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"mixm/backend/internal/config"
	"mixm/backend/internal/infra/cluster"
	"mixm/backend/internal/model"
	"mixm/backend/internal/repository"
)

type Gateway struct {
	Cluster *cluster.Client
	Config  config.Config
}

type terminalSizes struct {
	ch chan remotecommand.TerminalSize
}

// Next 读取下一次终端窗口尺寸变化。
func (t *terminalSizes) Next() *remotecommand.TerminalSize {
	s, ok := <-t.ch
	if !ok {
		return nil
	}

	return &s
}

// hostKey 读取 SSH 主机密钥，不存在时创建并安全保存。
func hostKey(path string) (ssh.Signer, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		_, key, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, err
		}

		der, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			return nil, err
		}

		b = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return nil, err
		}

		if err = os.WriteFile(path, b, 0600); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return ssh.ParsePrivateKey(b)
}

// ServeSSH 监听 SSH 连接，校验实例凭据并分发终端会话。
func (c *Gateway) ServeSSH(ctx context.Context, db *repository.DB) error {
	if c == nil {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 复用持久化主机密钥，使网关重启后保持相同身份。
	key, err := hostKey(c.Config.SSHKeyPath)
	if err != nil {
		return err
	}

	// 通过完整 SSH 用户名定位运行实例，并校验实例访问密码。
	cfg := &ssh.ServerConfig{
		PasswordCallback: func(meta ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			authCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			v, err := repository.RunningSSHInstance(authCtx, db, meta.User())
			if err != nil {
				return nil, fmt.Errorf("authentication failed")
			}

			expected := model.String(model.Object(v["config"]), "password")
			if expected == "" || subtle.ConstantTimeCompare([]byte(expected), password) != 1 {
				return nil, fmt.Errorf("authentication failed")
			}

			return &ssh.Permissions{
				Extensions: map[string]string{"instance": model.String(v, "id"), "namespace": model.String(v, "namespace")},
			}, nil
		},
	}
	cfg.AddHostKey(key)

	// 监听网关端口，退出时关闭监听并等待现有连接结束。
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Config.SSHPort))
	if err != nil {
		return err
	}

	defer listener.Close()

	go func() {
		<-ctx.Done()
		listener.Close()
	}()
	var wg sync.WaitGroup
	defer func() {
		cancel()
		wg.Wait()
	}()

	for {
		raw, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer raw.Close()

			raw.SetDeadline(time.Now().Add(15 * time.Second))
			conn, channels, requests, err := ssh.NewServerConn(raw, cfg)
			if err != nil {
				return
			}

			raw.SetDeadline(time.Time{})
			defer conn.Close()

			done := make(chan struct{})
			defer close(done)

			go func() {
				select {
				case <-ctx.Done():
					conn.Close()
				case <-done:
				}
			}()
			go ssh.DiscardRequests(requests)

			for request := range channels {
				if request.ChannelType() != "session" {
					request.Reject(ssh.UnknownChannelType, "session required")
					continue
				}

				ch, reqs, err := request.Accept()
				if err != nil {
					continue
				}

				go c.sshSession(ctx, conn, ch, reqs)
			}
		}()
	}
}

// sshSession 将 SSH 终端或命令转发到对应实例的容器。
func (c *Gateway) sshSession(
	ctx context.Context,
	conn *ssh.ServerConn,
	ch ssh.Channel,
	requests <-chan *ssh.Request,
) {
	defer ch.Close()

	sessionCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	sizes := &terminalSizes{ch: make(chan remotecommand.TerminalSize, 1)}
	defer close(sizes.ch)

	tty := false
	started := false
	streamDone := make(chan error, 1)

	for {
		select {
		case <-sessionCtx.Done():
			return
		case err := <-streamDone:
			if err != nil {
				slog.Debug("SSH shell ended", "error", err)
			}
			status := uint32(0)
			if err != nil {
				status = 1
			}
			ch.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{status}))
			return
		case req, ok := <-requests:
			if !ok {
				return
			}
			accepted := false
			switch req.Type {
			case "pty-req":
				var p struct {
					Term                      string
					Cols, Rows, Width, Height uint32
					Modes                     string
				}
				if ssh.Unmarshal(req.Payload, &p) == nil {
					tty = true
					accepted = true

					select {
					case sizes.ch <- remotecommand.TerminalSize{Width: uint16(p.Cols), Height: uint16(p.Rows)}:
					default:
					}
				}
			case "window-change":
				var p struct{ Cols, Rows, Width, Height uint32 }
				if ssh.Unmarshal(req.Payload, &p) == nil {
					accepted = true

					select {
					case sizes.ch <- remotecommand.TerminalSize{Width: uint16(p.Cols), Height: uint16(p.Rows)}:
					default:
					}
				}
			case "shell", "exec":
				if !started {
					started = true
					accepted = true

					// 将终端会话和单次命令分别转换为容器内的 Bash 请求。
					command := []string{"/bin/bash"}
					if req.Type == "exec" {
						var p struct{ Command string }
						if ssh.Unmarshal(req.Payload, &p) != nil {
							accepted = false
							break
						}

						command = []string{"/bin/bash", "-lc", p.Command}
					}

					go func(tty bool, command []string) {
						// 按实例标签寻找运行中的 Pod，再建立 Kubernetes exec 数据流。
						pods, err := c.Cluster.K.CoreV1().Pods(conn.Permissions.Extensions["namespace"]).List(
							sessionCtx,
							metav1.ListOptions{LabelSelector: "app=deploy-cs-" + conn.Permissions.Extensions["instance"]},
						)
						if err != nil {
							streamDone <- err
							return
						}

						for _, p := range pods.Items {
							if p.Status.Phase != corev1.PodRunning {
								continue
							}

							req := c.Cluster.K.CoreV1().RESTClient().Post().Resource("pods").Name(p.Name).Namespace(p.Namespace).SubResource("exec").VersionedParams(
								&corev1.PodExecOptions{
									Container: p.Spec.Containers[0].Name,
									Command:   command,
									Stdin:     true,
									Stdout:    true,
									Stderr:    !tty,
									TTY:       tty,
								},
								scheme.ParameterCodec,
							)
							cfg := *c.Cluster.REST
							cfg.Timeout = 0
							executor, err := remotecommand.NewSPDYExecutor(&cfg, "POST", req.URL())
							if err != nil {
								streamDone <- err
								return
							}

							opts := remotecommand.StreamOptions{
								Stdin:  ch,
								Stdout: ch,
								Tty:    tty,
							}
							if tty {
								opts.TerminalSizeQueue = sizes
							} else {
								opts.Stderr = ch.Stderr()
							}

							streamDone <- executor.StreamWithContext(sessionCtx, opts)

							return
						}

						streamDone <- fmt.Errorf("no running pod")
					}(
						tty,
						command,
					)
				}
			}
			if req.WantReply {
				_ = req.Reply(accepted, nil)
			}
		}
	}
}
