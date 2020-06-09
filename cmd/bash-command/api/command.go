package api

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"io"
	"os"
	"os/exec"
	"time"
)

type Command struct {
	Name string
}

func (c *Command) Exec(cmdStr string) error {
	cmd := exec.Command("bash", "-c", cmdStr)
	done := make(chan struct{})
	var errStdout, errStderr error
	stdoutIn, _ := cmd.StdoutPipe()
	//stderrIn, _ := cmd.StderrPipe()
	cmd.Start()
	go func() {
		defer close(done)
		copyAndCapture(os.Stdout, stdoutIn)
	}()
	//go func() {
	//	copyAndCapture(os.Stderr, stderrIn)
	//}()
	select {
	case <-done:
		cmd.Wait()
		fmt.Println(cmd.ProcessState.Success())
	case <-time.After(10 * time.Second):
		fmt.Println("Timeout!")
	}
	//err := cmd.Wait()
	//fmt.Println("done: ", cmd.ProcessState.ExitCode())
	//if err != nil {
	//	logs.Error("cmd.Run() failed with ", err)
	//}
	if errStdout != nil || errStderr != nil {
		logs.Error("failed to capture stdout or stderr\n", errStdout.Error(), errStderr.Error())
	}
	//outStr, errStr := string(stdout), string(stderr)
	//fmt.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)

	return nil
}

func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			w.Write(d)
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}

	// never reached
	panic(true)
	return nil, nil
}
