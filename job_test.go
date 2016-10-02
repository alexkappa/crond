package main

import (
	"bytes"
	"log"
	"os"
	"testing"
	"time"
)

func TestJob(t *testing.T) {
	newJob("false")
}

func TestJobOutput(t *testing.T) {
	var b bytes.Buffer

	Stdout = &b
	Stderr = &b
	log.SetOutput(&b)

	defer func() {
		Stdout = os.Stdout
		Stderr = os.Stderr
		log.SetOutput(os.Stderr)
	}()

	job := newJob("echo foo")
	job.Run()

	logTime := time.Now().Format("2006/01/02 15:04:05")

	var e bytes.Buffer
	e.WriteString(logTime + " running command [\"sh\" \"-c\" \"echo foo\"]\n")
	e.WriteString(logTime + " > foo\n")

	if e.String() != b.String() {
		t.Errorf("unexpected output %q, expected %q", e, b)
	}
}
