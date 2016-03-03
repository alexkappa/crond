// Copyright 2013-2016 Docker, Inc. All rights reserved. Use of this source 
// code is governed by the Apache License 2.0 that can be found at the projects
// LICENCE file.
// 
// https://github.com/docker/docker/blob/master/LICENSE
package main

import "gopkg.in/fsnotify.v1"

// FileWatcher is an interface for implementing file notification watchers
type FileWatcher interface {
	Events() <-chan fsnotify.Event
	Errors() <-chan error
	Add(name string) error
	Remove(name string) error
	Close() error
}

// newFileWatcher tries to use an fs-event watcher, and falls back to the 
// poller if there is an error.
func newFileWatcher() (FileWatcher, error) {
	if watcher, err := newEventWatcher(); err == nil {
		return watcher, nil
	}
	return newPollingWatcher(), nil
}

// newPollingWatcher returns a poll-based file watcher
func newPollingWatcher() FileWatcher {
	return &filePoller{
		events: make(chan fsnotify.Event),
		errors: make(chan error),
	}
}

// newEventWatcher returns an fs-event based file watcher
func newEventWatcher() (FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &fsNotifyWatcher{watcher}, nil
}
