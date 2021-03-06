package ssh

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type Connection struct {
	*ssh.Client
	password string
}

// Connect connects and returns a connection to a remote host.
func Connect(addr, user, password string, timout ...time.Duration) (*Connection, error) {
	if len(timout) > 1 {
		panic("multiple timout parameters parsed")
	}

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	}

	if len(timout) > 0 {
		sshConfig.Timeout = timout[0]
	}

	conn, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, err
	}

	return &Connection{conn, password}, nil

}

// Send commands executes one or more commands on an existing connection.
func (conn *Connection) SendCommands(cmds ...string) ([]byte, error) {
	if conn == nil {
		var err error
		conn, err = Connect(conn.RemoteAddr().String(), conn.User(), conn.password)
		if err != nil {
			return []byte{}, errors.New("connection is nil")
		}
	}

	session, err := conn.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	err = session.RequestPty("xterm", 80, 40, modes)
	if err != nil {
		return []byte{}, err
	}

	in, err := session.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	out, err := session.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	var output []byte

	go func(in io.WriteCloser, out io.Reader, output *[]byte) {
		var (
			line string
			r    = bufio.NewReader(out)
		)
		for {
			b, err := r.ReadByte()
			if err != nil {
				break
			}

			*output = append(*output, b)

			if b == byte('\n') {
				line = ""
				continue
			}

			line += string(b)

			if strings.HasPrefix(line, "[sudo] password for ") && strings.HasSuffix(line, ": ") {
				_, err = in.Write([]byte(conn.password + "\n"))
				if err != nil {
					break
				}
			}
		}
	}(in, out, &output)

	cmd := strings.Join(cmds, "; ")
	_, err = session.Output(cmd)
	if err != nil {
		return output, err
	}

	return output, nil
}
