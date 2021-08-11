package ssh

import (
	"context"
	"github.com/kubesphere/kubekey/experiment/core/connector"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Cfg struct {
	Username    string
	Password    string
	Address     string
	Port        int
	PrivateKey  string
	KeyFile     string
	AgentSocket string
	Timeout     time.Duration
	Bastion     string
	BastionPort int
	BastionUser string
}

const socketEnvPrefix = "env:"

type connection struct {
	mu         sync.Mutex
	sftpclient *sftp.Client
	sshclient  *ssh.Client
	dialer     *Dialer
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewConnection(dialer *Dialer, cfg Cfg) (connector.Connection, error) {
	cfg, err := validateOptions(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to validate ssh connection parameters")
	}

	authMethods := make([]ssh.AuthMethod, 0)

	if len(cfg.Password) > 0 {
		authMethods = append(authMethods, ssh.Password(cfg.Password))
	}

	if len(cfg.PrivateKey) > 0 {
		signer, parseErr := ssh.ParsePrivateKey([]byte(cfg.PrivateKey))
		if parseErr != nil {
			return nil, errors.Wrap(parseErr, "The given SSH key could not be parsed")
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
			_ = socket.Close()
			return nil, errors.Wrap(signersErr, "error when creating signer for SSH agent")
		}

		authMethods = append(authMethods, ssh.PublicKeys(signers...))
	}

	sshConfig := &ssh.ClientConfig{
		User:            cfg.Username,
		Timeout:         cfg.Timeout,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	targetHost := cfg.Address
	targetPort := strconv.Itoa(cfg.Port)

	if cfg.Bastion != "" {
		targetHost = cfg.Bastion
		targetPort = strconv.Itoa(cfg.BastionPort)
		sshConfig.User = cfg.BastionUser
	}

	endpoint := net.JoinHostPort(targetHost, targetPort)

	client, err := ssh.Dial("tcp", endpoint, sshConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "could not establish connection to %s", endpoint)
	}

	ctx, cancelFn := context.WithCancel(context.Background())
	sshConn := &connection{
		ctx:    ctx,
		cancel: cancelFn,
		dialer: dialer,
	}

	if cfg.Bastion == "" {
		sshConn.sshclient = client
		return sshConn, nil
	}

	endpointBehindBastion := net.JoinHostPort(cfg.Address, strconv.Itoa(cfg.Port))

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

func validateOptions(cfg Cfg) (Cfg, error) {
	if len(cfg.Username) == 0 {
		return cfg, errors.New("No username specified for SSH connection")
	}

	if len(cfg.Address) == 0 {
		return cfg, errors.New("No address specified for SSH connection")
	}

	if len(cfg.Password) == 0 && len(cfg.PrivateKey) == 0 && len(cfg.KeyFile) == 0 && len(cfg.AgentSocket) == 0 {
		return cfg, errors.New("Must specify at least one of password, private key, keyfile or agent socket")
	}

	if len(cfg.PrivateKey) == 0 && len(cfg.KeyFile) > 0 {
		content, err := ioutil.ReadFile(cfg.KeyFile)
		if err != nil {
			return cfg, errors.Wrapf(err, "Failed to read keyfile %q", cfg.KeyFile)
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

func (c *connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sshclient == nil && c.sftpclient == nil {
		return
	}
	c.cancel()
	c.dialer.forgetConnection(c)

	if c.sshclient != nil {
		c.sshclient.Close()
		c.sshclient = nil
	}
	if c.sftpclient != nil {
		c.sftpclient.Close()
		c.sftpclient = nil
	}
}

func (c *connection) PExec(cmd string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
	sess, err := c.session()
	if err != nil {
		return 1, errors.Wrap(err, "failed to get SSH session")
	}
	defer sess.Close()

	sess.Stdin = stdin
	sess.Stdout = stdout
	sess.Stderr = stderr

	exitCode := 0
	if err = sess.Run(cmd); err != nil {
		exitCode = -1
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		}
	}

	// preserve original error
	return exitCode, err
}

func (c *connection) Exec(cmd string) (stdout string, stderr string, code int, err error) {
	var stdoutBuf, stderrBuf strings.Builder

	exitCode, err := c.PExec(cmd, nil, &stdoutBuf, &stderrBuf)

	return strings.TrimSpace(stdoutBuf.String()), stderrBuf.String(), exitCode, err
}

func (c *connection) Scp(src, dst string) error {
	session, err := c.session()

	err = scp.CopyPath(src, dst, session)
	if err != nil {
		return err
	}

	return nil
}

func (c *connection) session() (*ssh.Session, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sshclient == nil {
		return nil, errors.New("connection closed")
	}

	sess, err := c.sshclient.NewSession()
	if err != nil {
		return nil, err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	err = sess.RequestPty("xterm", 100, 50, modes)
	if err != nil {
		return nil, err
	}

	return sess, nil
}
