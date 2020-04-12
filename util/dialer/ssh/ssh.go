package ssh

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const socketEnvPrefix = "env:"

var (
	_ Connection = &connection{}
	_ Tunneler   = &connection{}
)

// Connection represents an established connection to an SSH server.
type Connection interface {
	Exec(cmd string) (stdout string, stderr string, exitCode int, err error)
	File(filename string, flags int) (io.ReadWriteCloser, error)
	Stream(cmd string, stdout io.Writer, stderr io.Writer) (exitCode int, err error)
	io.Closer
}

// Tunneler interface creates net.Conn originating from the remote ssh host to
// target `addr`
type Tunneler interface {
	// `network` can be tcp, tcp4, tcp6, unix
	TunnelTo(ctx context.Context, network, addr string) (net.Conn, error)
}

// SSHCfg represents all the possible options for connecting to
// a remote server via SSH.
type SSHCfg struct {
	Username    string
	Password    string
	Hostname    string
	Port        int
	PrivateKey  string
	KeyFile     string
	AgentSocket string
	Timeout     time.Duration
	Bastion     string
	BastionPort int
	BastionUser string
}

func validateOptions(cfg SSHCfg) (SSHCfg, error) {
	if len(cfg.Username) == 0 {
		return cfg, errors.New("no username specified for SSH connection")
	}

	if len(cfg.Hostname) == 0 {
		return cfg, errors.New("no hostname specified for SSH connection")
	}

	if len(cfg.Password) == 0 && len(cfg.PrivateKey) == 0 && len(cfg.KeyFile) == 0 && len(cfg.AgentSocket) == 0 {
		return cfg, errors.New("must specify at least one of password, private key, keyfile or agent socket")
	}

	if len(cfg.KeyFile) > 0 {
		content, err := ioutil.ReadFile(cfg.KeyFile)
		if err != nil {
			return cfg, errors.Wrapf(err, "failed to read keyfile %q", cfg.KeyFile)
		}

		cfg.PrivateKey = string(content)
		cfg.KeyFile = ""
	}

	if cfg.Port <= 0 {
		cfg.Port = 22
	}

	if cfg.BastionPort <= 0 {
		cfg.BastionPort = 22
	}

	if cfg.BastionUser == "" {
		cfg.BastionUser = cfg.Username
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 60 * time.Second
	}

	return cfg, nil
}

type connection struct {
	mu         sync.Mutex
	sftpclient *sftp.Client
	sshclient  *ssh.Client
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewConnection attempts to create a new SSH connection to the host
// specified via the given options.
func NewConnection(cfg SSHCfg) (Connection, error) {
	cfg, err := validateOptions(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to validate ssh connection options")
	}

	authMethods := make([]ssh.AuthMethod, 0)

	if len(cfg.Password) > 0 {
		authMethods = append(authMethods, ssh.Password(cfg.Password))
	}

	if len(cfg.PrivateKey) > 0 {
		signer, parseErr := ssh.ParsePrivateKey([]byte(cfg.PrivateKey))
		if parseErr != nil {
			return nil, errors.Wrap(parseErr, "the given SSH key could not be parsed (note that password-protected keys are not supported)")
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(cfg.AgentSocket) > 0 {
		addr := cfg.AgentSocket

		if strings.HasPrefix(cfg.AgentSocket, socketEnvPrefix) {
			envName := strings.TrimPrefix(cfg.AgentSocket, socketEnvPrefix)

			if envAddr := os.Getenv(envName); len(envAddr) > 0 {
				addr = envAddr
			}
		}

		socket, dialErr := net.Dial("unix", addr)
		if dialErr != nil {
			return nil, errors.Wrapf(dialErr, "could not open socket %q", addr)
		}

		agentClient := agent.NewClient(socket)

		signers, signersErr := agentClient.Signers()
		if signersErr != nil {
			socket.Close()
			return nil, errors.Wrap(signersErr, "error when creating signer for SSH agent")
		}

		authMethods = append(authMethods, ssh.PublicKeys(signers...))
	}

	sshConfig := &ssh.ClientConfig{
		User:            cfg.Username,
		Timeout:         cfg.Timeout,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
	}

	targetHost := cfg.Hostname
	targetPort := strconv.Itoa(cfg.Port)

	if cfg.Bastion != "" {
		targetHost = cfg.Bastion
		targetPort = strconv.Itoa(cfg.BastionPort)
		sshConfig.User = cfg.BastionUser
	}

	// do not use fmt.Sprintf() to allow proper IPv6 handling if hostname is an IP address
	endpoint := net.JoinHostPort(targetHost, targetPort)

	client, err := ssh.Dial("tcp", endpoint, sshConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "could not establish connection to %s", endpoint)
	}

	ctx, cancelFn := context.WithCancel(context.Background())
	sshConn := &connection{
		ctx:    ctx,
		cancel: cancelFn,
	}

	if cfg.Bastion == "" {
		sshConn.sshclient = client
		// connection established
		return sshConn, nil
	}

	// continue to setup if we are running over bastion
	endpointBehindBastion := net.JoinHostPort(cfg.Hostname, strconv.Itoa(cfg.Port))

	// Dial a connection to the service host, from the bastion
	conn, err := client.Dial("tcp", endpointBehindBastion)
	if err != nil {
		return nil, errors.Wrapf(err, "could not establish connection to %s", endpointBehindBastion)
	}

	sshConfig.User = cfg.Username
	ncc, chans, reqs, err := ssh.NewClientConn(conn, endpointBehindBastion, sshConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "could not establish connection to %s", endpointBehindBastion)
	}

	sshConn.sshclient = ssh.NewClient(ncc, chans, reqs)
	return sshConn, nil
}

// File return remote file (as an io.ReadWriteCloser).
//
// mode is os package file modes: https://golang.org/pkg/os/#pkg-constants
// returned file optionally implement
func (c *connection) File(filename string, flags int) (io.ReadWriteCloser, error) {
	sftpClient, err := c.sftp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to open SFTP")
	}

	return sftpClient.OpenFile(filename, flags)
}

func (c *connection) TunnelTo(_ context.Context, network, addr string) (net.Conn, error) {
	netconn, err := c.sshclient.Dial(network, addr)
	if err == nil {
		go func() {
			<-c.ctx.Done()
			netconn.Close()
		}()
	}
	return netconn, err
}

func (c *connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sshclient == nil {
		return nil
	}
	c.cancel()

	defer func() { c.sshclient = nil }()
	defer func() { c.sftpclient = nil }()

	return c.sshclient.Close()
}

func (c *connection) Stream(cmd string, stdout io.Writer, stderr io.Writer) (int, error) {
	sess, err := c.session()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get SSH session")
	}
	defer sess.Close()

	sess.Stdout = stdout
	sess.Stderr = stderr

	exitCode := 0
	err = sess.Run(strings.TrimSpace(cmd))
	if err != nil {
		exitCode = 1
	}

	return exitCode, errors.Wrapf(err, "failed to exec command: %s", cmd)
}

func (c *connection) Exec(cmd string) (string, string, int, error) {
	var stdoutBuf, stderrBuf bytes.Buffer

	exitCode, err := c.Stream(cmd, &stdoutBuf, &stderrBuf)

	return strings.TrimSpace(stdoutBuf.String()), strings.TrimSpace(stderrBuf.String()), exitCode, err
}

func (c *connection) session() (*ssh.Session, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sshclient == nil {
		return nil, errors.New("connection closed")
	}

	return c.sshclient.NewSession()
}

func (c *connection) sftp() (*sftp.Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sshclient == nil {
		return nil, errors.New("connection closed")
	}

	if c.sftpclient == nil {
		s, err := sftp.NewClient(c.sshclient)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get sftp.Client")
		}
		c.sftpclient = s
	}

	return c.sftpclient, nil
}
