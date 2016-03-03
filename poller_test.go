// Copyright 2013-2016 Docker, Inc. All rights reserved. Use of this source
// code is governed by the Apache License 2.0 that can be found at the projects
// LICENCE file.
//
// https://github.com/docker/docker/blob/master/LICENSE
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"testing"
	"time"

	"gopkg.in/fsnotify.v1"
)

func TestPollerAddRemove(t *testing.T) {
	w := newPollingWatcher()

	if err := w.Add("no-such-file"); err == nil {
		t.Fatal("should have gotten error when adding a non-existent file")
	}
	if err := w.Remove("no-such-file"); err == nil {
		t.Fatal("should have gotten error when removing non-existent watch")
	}

	f, err := ioutil.TempFile("", "asdf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(f.Name())

	if err := w.Add(f.Name()); err != nil {
		t.Fatal(err)
	}

	if err := w.Remove(f.Name()); err != nil {
		t.Fatal(err)
	}
}

func TestPollerEvent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("No chmod on Windows")
	}
	w := newPollingWatcher()

	f, err := ioutil.TempFile("", "test-poller")
	if err != nil {
		t.Fatal("error creating temp file")
	}
	defer os.RemoveAll(f.Name())
	f.Close()

	if err := w.Add(f.Name()); err != nil {
		t.Fatal(err)
	}

	select {
	case <-w.Events():
		t.Fatal("got event before anything happened")
	case <-w.Errors():
		t.Fatal("got error before anything happened")
	default:
	}

	if err := ioutil.WriteFile(f.Name(), []byte("hello"), 644); err != nil {
		t.Fatal(err)
	}
	if err := assertEvent(w, fsnotify.Write); err != nil {
		t.Fatal(err)
	}

	if err := os.Chmod(f.Name(), 600); err != nil {
		t.Fatal(err)
	}
	if err := assertEvent(w, fsnotify.Chmod); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(f.Name()); err != nil {
		t.Fatal(err)
	}
	if err := assertEvent(w, fsnotify.Remove); err != nil {
		t.Fatal(err)
	}
}

func TestPollerClose(t *testing.T) {
	w := newPollingWatcher()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	// test double-close
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	f, err := ioutil.TempFile("", "asdf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(f.Name())
	if err := w.Add(f.Name()); err == nil {
		t.Fatal("should have gotten error adding watch for closed watcher")
	}
}

func assertEvent(w FileWatcher, eType fsnotify.Op) error {
	var err error
	select {
	case e := <-w.Events():
		if e.Op != eType {
			err = fmt.Errorf("got wrong event type, expected %q: %v", eType, e)
		}
	case e := <-w.Errors():
		err = fmt.Errorf("got unexpected error waiting for events %v: %v", eType, e)
	case <-time.After(watchWaitTime * 3):
		err = fmt.Errorf("timeout waiting for event %v", eType)
	}
	return err
}
