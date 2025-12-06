package remote

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (s *SSH) Upload(local, remoteDir string) error {
	f, err := os.Open(local)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	session, err := s.Client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	w, err := session.StdinPipe()
	if err != nil {
		return err
	}
	defer w.Close()

	base := filepath.Base(local)
	go func() {
		fmt.Fprintf(w, "C%#o %d %s\n", info.Mode().Perm(), info.Size(), base)
		io.Copy(w, f)
		fmt.Fprint(w, "\x00")
	}()

	return session.Run(fmt.Sprintf("scp -t %s", remoteDir))
}