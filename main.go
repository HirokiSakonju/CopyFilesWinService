// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !plan9

package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/kardianos/service"
)

func main() {
	svcConfig := &service.Config{
		Name:        "CopyFilesWinService",
		DisplayName: "CopyFilesWinService",
		Description: "CopyFilesWinService.",
	}
	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) > 1 {
		var err error
		verb := os.Args[1]
		switch verb {
		case "install":
			err = s.Install()
			if err != nil {
				logger.Error(err)
			}
		}
	}

}

var logger service.Logger

type program struct{}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}
func (p *program) run() {
	ExampleNewWatcher()
}
func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	return nil
}
func ExampleNewWatcher() {

	setting := ReadSettingFile()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)

				}
				src, err := os.Open(event.Name)
				if err != nil {
					//panic(err)
					log.Println("error:", err)
				}
				defer src.Close()
				dest, err := os.Create(filepath.Join(setting.DestDir, filepath.Base(event.Name)))
				if err != nil {
					//panic(err)
					log.Println("error:", err)
				}
				defer dest.Close()
				defer log.Println("called defer")
				_, err = io.Copy(dest, src)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(setting.SourceDir)
	if err != nil {
		log.Fatal(err)
	}
	// Block until chan is closed
	<-done
}

func ReadSettingFile() Setting {
	bytes, err := ioutil.ReadFile("copyfiles_settings.json")
	if err != nil {
		log.Fatal(err)
		//panic("Setting file not Found")
	}
	var settingfile Setting
	if err := json.Unmarshal(bytes, &settingfile); err != nil {
		log.Fatal(err)
		//panic("Setting file not Unmarshal")
		log.Println("error:", err)
	}
	return settingfile
}

type Setting struct {
	SourceDir string `json:"SourceDir"`
	DestDir   string `json:"DestDir"`
}
