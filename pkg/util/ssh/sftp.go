package ssh

import (
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"github.com/tmc/scp"
)

func (c *connection) sftp() (*sftp.Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sshclient == nil {
		return nil, errors.New("connection closed")
	}

	if c.sftpclient == nil {
		s, err := sftp.NewClient(c.sshclient)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to get sftp.Client")
		}
		c.sftpclient = s
	}

	return c.sftpclient, nil
}

func (c *connection) Scp(src, dst string) error {
	session, err := c.session()

	err = scp.CopyPath(src, dst, session)
	if err != nil {
		return err
	}

	return nil
}
