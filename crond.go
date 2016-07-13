package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	cron "gopkg.in/robfig/cron.v2"
)

var Version string

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <dir>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	run(os.Args[1])
}

func run(dir string) {
	sigint := make(chan os.Signal)
	signal.Notify(sigint, os.Interrupt, os.Kill)

	sighup := make(chan os.Signal)
	signal.Notify(sighup, syscall.SIGHUP)

	watcher, err := newFileWatcher()
	if err != nil {
		log.Fatalln(err)
	}

	err = watcher.Add(dir)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("crond started")
	for {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			log.Fatalf("failed to open directory: %s\n", err)
		}
		c := new(Crond)
		for _, file := range files {
			log.Printf("crond parsing file %s", filepath.Join(dir, file.Name()))
			if file.IsDir() {
				continue
			}
			f, err := os.Open(filepath.Join(dir, file.Name()))
			if err != nil {
				log.Printf("failed to open file %s: %s", file.Name(), err)
				continue
			}
			defer f.Close()
			err = c.Read(f)
			if err != nil {
				log.Printf("error in file %s %s\n", f.Name(), err)
				continue
			}
		}
		c.Start()
		select {
		case <-sigint:
			log.Println("crond exiting")
			return
		case <-sighup:
			log.Println("crond reloading")
		case <-watcher.Events():
			log.Printf("crond detected changes in %s. reloading\n", dir)
		}
		c.Stop()
	}
}

// Crond provides functionality for reading and parsing cron entries.
type Crond struct{ cron.Cron }

// Read reads and parses cron entries from a reader.
func (c *Crond) Read(r io.Reader) error {
	l := 1
	s := bufio.NewScanner(r)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		_, err := c.ReadLine(s.Text())
		if err != nil {
			return fmt.Errorf("line %d. %s", l, err)
		}
		l++
	}
	return s.Err()
}

// ReadLine reads and parses an individual entry from a string.
func (c *Crond) ReadLine(line string) (int, error) {
	// We'll split each line on spaces and try to identify which part is the
	// schedule and which is the command.
	var parts []string

	switch {
	// Empty lines or lines that start with a pound (#) are treated as comments
	// and therefore are ignored.
	//
	// 	# This line is ignored.
	case strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#"):
		return -1, nil
	// The duration expression is used to schedule jobs to execute at fixed
	// intervals.
	//
	// 	@every 3m <command>
	case strings.HasPrefix(line, "@every"):
		parts = strings.SplitN(line, " ", 3)
		if len(parts) < 3 {
			return -1, fmt.Errorf("malformed entry")
		}
		parts = []string{
			strings.Join(parts[0:2], " "),
			strings.Join(parts[2:], " "),
		}
	// The predefined schedules are used to schedule jobs using some common cron
	// expressions.
	//
	// 	@monthly <command>
	case strings.HasPrefix(line, "@"):
		parts = strings.SplitN(line, " ", 2)
	// The cron expression format represents a set of times, using 5
	// space-separated fields.
	//
	// 	0 0 * * * <command>
	default:
		parts = strings.Split(line, " ")
		if len(parts) < 5 {
			return -1, fmt.Errorf("malformed entry")
		}
		parts = []string{
			strings.Join(parts[0:5], " "),
			strings.Join(parts[5:], " "),
		}
	}
	if len(parts) != 2 {
		return -1, fmt.Errorf("malformed entry")
	}
	id, err := c.AddFunc(parts[0], func() {
		cmd := exec.Command("sh", "-c", parts[1])
		cmd.Stdout = wrapLog(os.Stdout)
		cmd.Stderr = wrapLog(os.Stderr)
		log.Printf("running command %q", cmd.Args)
		err := cmd.Run()
		if err != nil {
			log.Printf("error: %s\n", err)
		}
		// If the command is followed by a comment matching a particular pattern
		// it is assumed to be an annotation related to health checking.
		//
		// The value of the annotation is extracted and used to ping the health
		// check service.
		//
		// 	@every <duration> <command> # hc:<check>
		hc := newHealthCheck()
		if hc.HasAnnotation(parts[1]) {
			id := hc.FromAnnotation(parts[1])
			if err = hc.Ping(id); err != nil {
				log.Printf("error: %s\n", err)
			}
		}
	})
	return int(id), err
}

// Entry retrieves a cron entry by its id.
func (c *Crond) Entry(id int) cron.Entry {
	return c.Cron.Entry(cron.EntryID(id))
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
