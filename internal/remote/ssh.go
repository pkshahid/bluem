package remote

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"bluem/internal/logx"
	"golang.org/x/crypto/ssh"
)

type SSH struct {
	Client *ssh.Client
}

func expandPath(p string) string {
	if len(p) == 0 {
		return p
	}
	if p[0] != '~' {
		return p
	}
	usr, err := user.Current()
	if err != nil {
		return p
	}
	return filepath.Join(usr.HomeDir, p[1:])
}

func NewSSH(userHost, keyPath string) (*SSH, error) {
	keyPath = expandPath(keyPath)
	b, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(b)
	if err != nil {
		return nil, err
	}

	cfg := &ssh.ClientConfig{
		User:            userHostUser(userHost),
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := userHostAddr(userHost)
	logx.L().Infow("dialing ssh", "addr", addr)
	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return nil, err
	}
	return &SSH{Client: client}, nil
}

func userHostUser(uh string) string {
	// "user@host:22" -> "user"
	for i := 0; i < len(uh); i++ {
		if uh[i] == '@' {
			return uh[:i]
		}
	}
	return os.Getenv("USER")
}

func userHostAddr(uh string) string {
	// "user@host:22" -> "host:22"
	at := -1
	for i := 0; i < len(uh); i++ {
		if uh[i] == '@' {
			at = i
			break
		}
	}
	if at == -1 {
		return uh
	}
	return uh[at+1:]
}

func (s *SSH) Run(cmd string) (string, error) {
	session, err := s.Client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	out, err := session.CombinedOutput(cmd)
	return string(out), err
}