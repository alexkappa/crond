// Copyright 2013-2016 Docker, Inc. All rights reserved. Use of this source 
// code is governed by the Apache License 2.0 that can be found at the projects
// LICENCE file.
// 
// https://github.com/docker/docker/blob/master/LICENSE
package main

import "gopkg.in/fsnotify.v1"

// fsNotify wraps the fsnotify package to satisfy the FileNotifer interface
type fsNotifyWatcher struct {
	*fsnotify.Watcher
}

// GetEvents returns the fsnotify event channel receiver
func (w *fsNotifyWatcher) Events() <-chan fsnotify.Event {
	return w.Watcher.Events
}

// GetErrors returns the fsnotify error channel receiver
func (w *fsNotifyWatcher) Errors() <-chan error {
	return w.Watcher.Errors
}
