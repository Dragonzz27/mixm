package cluster

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

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"mixm-edu/backend/internal/store"
)

type terminalSizes struct {
	ch chan remotecommand.TerminalSize
}

func (t *terminalSizes) Next() *remotecommand.TerminalSize {
	s, ok := <-t.ch
	if !ok {
		return nil
	}
	return &s
}
func hostKey(path string) (ssh.Signer, error) {
	b, e := os.ReadFile(path)
	if os.IsNotExist(e) {
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
	} else if e != nil {
		return nil, e
	}
	return ssh.ParsePrivateKey(b)
}
func (c *Client) ServeSSH(ctx context.Context, db *pgxpool.Pool) error {
	if c == nil {
		return nil
	}
	key, e := hostKey(c.Config.SSHKeyPath)
	if e != nil {
		return e
	}
	cfg := &ssh.ServerConfig{PasswordCallback: func(meta ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
		authCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		v, e := store.One(authCtx, db, "SELECT * FROM instances WHERE ssh_username=$1 AND status='running'", meta.User())
		if e != nil {
			return nil, fmt.Errorf("authentication failed")
		}
		expected := store.String(store.Object(v["config"]), "password")
		if expected == "" || subtle.ConstantTimeCompare([]byte(expected), password) != 1 {
			return nil, fmt.Errorf("authentication failed")
		}
		return &ssh.Permissions{Extensions: map[string]string{"instance": store.String(v, "id"), "namespace": store.String(v, "namespace")}}, nil
	}}
	cfg.AddHostKey(key)
	listener, e := net.Listen("tcp", fmt.Sprintf(":%d", c.Config.SSHPort))
	if e != nil {
		return e
	}
	defer listener.Close()
	go func() { <-ctx.Done(); listener.Close() }()
	var wg sync.WaitGroup
	defer wg.Wait()
	for {
		raw, e := listener.Accept()
		if e != nil {
			if ctx.Err() != nil {
				return nil
			}
			return e
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer raw.Close()
			raw.SetDeadline(time.Now().Add(15 * time.Second))
			conn, channels, requests, e := ssh.NewServerConn(raw, cfg)
			if e != nil {
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
				ch, reqs, e := request.Accept()
				if e != nil {
					continue
				}
				go c.sshSession(ctx, conn, ch, reqs)
			}
		}()
	}
}
func (c *Client) sshSession(ctx context.Context, conn *ssh.ServerConn, ch ssh.Channel, requests <-chan *ssh.Request) {
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
		case e := <-streamDone:
			if e != nil {
				slog.Debug("SSH shell ended", "error", e)
			}
			status := uint32(0)
			if e != nil {
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
						pods, e := c.K.CoreV1().Pods(conn.Permissions.Extensions["namespace"]).List(sessionCtx, metav1.ListOptions{LabelSelector: "app=deploy-cs-" + conn.Permissions.Extensions["instance"]})
						if e != nil {
							streamDone <- e
							return
						}
						for _, p := range pods.Items {
							if p.Status.Phase != corev1.PodRunning {
								continue
							}
							req := c.K.CoreV1().RESTClient().Post().Resource("pods").Name(p.Name).Namespace(p.Namespace).SubResource("exec").VersionedParams(&corev1.PodExecOptions{Container: p.Spec.Containers[0].Name, Command: command, Stdin: true, Stdout: true, Stderr: !tty, TTY: tty}, scheme.ParameterCodec)
							cfg := *c.REST
							cfg.Timeout = 0
							executor, e := remotecommand.NewSPDYExecutor(&cfg, "POST", req.URL())
							if e != nil {
								streamDone <- e
								return
							}
							opts := remotecommand.StreamOptions{Stdin: ch, Stdout: ch, Tty: tty}
							if tty {
								opts.TerminalSizeQueue = sizes
							} else {
								opts.Stderr = ch.Stderr()
							}
							streamDone <- executor.StreamWithContext(sessionCtx, opts)
							return
						}
						streamDone <- fmt.Errorf("no running pod")
					}(tty, command)
				}
			}
			if req.WantReply {
				_ = req.Reply(accepted, nil)
			}
		}
	}
}
