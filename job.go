package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
)

var (
	// Stdout specifies where job stdout is written
	Stdout io.Writer = os.Stdout

	// Stderr specifies where job stderr is written
	Stderr io.Writer = os.Stderr
)

type job struct {
	cmd string
	hc  *healthCheck
}

func (j *job) Run() {
	cmd := exec.Command("sh", "-c", j.cmd)
	cmd.Stdout = wrapLog(Stdout)
	cmd.Stderr = wrapLog(Stderr)

	log.Printf("running command %q", cmd.Args)
	err := cmd.Run()
	if err != nil {
		log.Printf("error: %s\n", err)
	} else {
		if j.hc.HasAnnotation(j.cmd) {
			id := j.hc.FromAnnotation(j.cmd)
			if err = j.hc.Ping(id); err != nil {
				log.Printf("error: %s\n", err)
			}
		}
	}
}

func newJob(cmd string) *job {
	return &job{
		cmd: cmd,
		hc:  newHealthCheck(),
	}
}

func wrapLog(w io.Writer) io.Writer {
	return &logWriter{log.New(w, "", log.LstdFlags)}
}

type logWriter struct{ log *log.Logger }

func (l *logWriter) Write(p []byte) (int, error) {
	for _, line := range bytes.Split(p, []byte{'\n'}) {
		if len(line) > 0 {
			l.log.Printf("> %s", line)
		}
	}
	return len(p), nil
}
