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

	cron "gopkg.in/robfig/cron.v2"
)

func usage() {
	fmt.Printf("Usage: %s <path>\n", os.Args[0])
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	run(os.Args[1])
}

func run(dir string) {
	log.Println("crond started")
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalln(err)
	}
	c := new(Cron)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		f, err := os.Open(filepath.Join(dir, file.Name()))
		if err != nil {
			log.Println(err)
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
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill)
	<-ch
	log.Println("crond exiting")
}

type Cron struct{ cron.Cron }

func (c *Cron) Read(r io.Reader) error {
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

func (c *Cron) ReadLine(s string) (int, error) {
	var p []string
	switch {
	case strings.TrimSpace(s) == "" || strings.HasPrefix(s, "#"):
		return -1, nil
	case strings.HasPrefix(s, "@every"):
		p = strings.SplitN(s, " ", 3)
		if len(p) < 3 {
			return -1, fmt.Errorf("malformed entry.")
		}
		p = []string{
			strings.Join(p[0:2], " "),
			strings.Join(p[2:], " "),
		}
	case strings.HasPrefix(s, "@"):
		p = strings.SplitN(s, " ", 2)
	default:
		p = strings.Split(s, " ")
		if len(p) < 5 {
			return -1, fmt.Errorf("malformed entry.")
		}
		p = []string{
			strings.Join(p[0:5], " "),
			strings.Join(p[5:], " "),
		}
	}
	if len(p) != 2 {
		return -1, fmt.Errorf("malformed entry.")
	}
	id, err := c.AddFunc(p[0], func() {
		argv := strings.Split(p[1], " ")
		var cmd *exec.Cmd
		if len(argv) > 1 {
			cmd = exec.Command(argv[0], argv[1:]...)
		} else {
			cmd = exec.Command(argv[0])
		}
		log.Println(p[1])
		cmd.Stdout = wrapLog(os.Stdout)
		cmd.Stderr = wrapLog(os.Stderr)
		err := cmd.Run()
		if err != nil {
			log.Printf("error: %s\n", err)
		}
	})
	return int(id), err
}

func (c *Cron) Entry(id int) cron.Entry {
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
