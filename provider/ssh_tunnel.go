package provider

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Endpoint struct {
	Host string
	Port int
	User string
}

func NewEndpoint(s string) *Endpoint {
	endpoint := &Endpoint{
		Host: s,
	}

	if parts := strings.Split(endpoint.Host, "@"); len(parts) > 1 {
		endpoint.User = parts[0]
		endpoint.Host = parts[1]
	}

	if parts := strings.Split(endpoint.Host, ":"); len(parts) > 1 {
		endpoint.Host = parts[0]
		endpoint.Port, _ = strconv.Atoi(parts[1])
	}

	return endpoint
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

type SSHTunnel struct {
	Local  *Endpoint
	Server *Endpoint
	Remote *Endpoint
	Config *ssh.ClientConfig
	Log    *log.Logger
}

func (tunnel *SSHTunnel) logf(fmt string, args ...interface{}) {
	if tunnel.Log != nil {
		tunnel.Log.Printf(fmt, args...)
	}
}

func (tunnel *SSHTunnel) Start() error {
	listener, err := net.Listen("tcp", tunnel.Local.String())
	if err != nil {
		return err
	}

	defer listener.Close()

	tunnel.Local.Port = listener.Addr().(*net.TCPAddr).Port
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		tunnel.logf("accepted connection")
		go tunnel.forward(conn)
	}
}

func (tunnel *SSHTunnel) forward(localConn net.Conn) {
	serverConn, err := ssh.Dial("tcp", tunnel.Server.String(), tunnel.Config)
	if err != nil {
		tunnel.logf("server dial error: %s", err)
		return
	}

	tunnel.logf("connected to %s (1 of 2)\n", tunnel.Server.String())

	remoteConn, err := serverConn.Dial("tcp", tunnel.Remote.String())
	if err != nil {
		tunnel.logf("remote dial error: %s", err)
		return
	}

	tunnel.logf("connected to %s (2 of 2)\n", tunnel.Remote.String())

	copyConn := func(writer, reader net.Conn) {
		_, err := io.Copy(writer, reader)
		if err != nil {
			tunnel.logf("io.Copy error: %s", err)
		}
	}

	go copyConn(localConn, remoteConn)
	go copyConn(remoteConn, localConn)
}

func PrivateKey(privateKey string) ssh.AuthMethod {
	buffer := []byte(privateKey)

	key, err := ssh.ParsePrivateKey(buffer)

	if err != nil {
		return nil
	}

	return ssh.PublicKeys(key)
}

func NewSSHTunnel(gateway_host string, gateway_port int, gateway_username string,
	auth ssh.AuthMethod,
	remote_host string, remote_port int) *SSHTunnel {
	// A random port will be chosen for us.
	localEndpoint := NewEndpoint("localhost:0")
	server := NewEndpoint(fmt.Sprintf("%s@%s:%d", gateway_username, gateway_host, gateway_port))

	if server.Port <= 0 {
		server.Port = 22
	}

	sshTunnel := &SSHTunnel{
		Config: &ssh.ClientConfig{
			User: server.User,
			Auth: []ssh.AuthMethod{auth},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				// Always accept key.
				return nil
			},
		},
		Local:  localEndpoint,
		Server: server,
		Remote: NewEndpoint(fmt.Sprintf("%s:%d", remote_host, remote_port)),
	}

	return sshTunnel
}
