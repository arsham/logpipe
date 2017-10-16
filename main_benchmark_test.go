// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"

	"github.com/onsi/gomega/gexec"
)

var (
	port    int
	letters = []byte("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ ")
)

func randBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return b
}

func generateFiles() (*os.File, *os.File, *os.File, error) {

	c, err := ioutil.TempFile("", "main_test_config_")
	if err != nil {
		return nil, nil, nil, err
	}
	file1, err := ioutil.TempFile("", "main_test_file_")
	if err != nil {
		return nil, nil, nil, err
	}
	file2, err := ioutil.TempFile("", "main_test_file_")
	if err != nil {
		return nil, nil, nil, err
	}
	return c, file1, file2, nil
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	program, err := gexec.Build("github.com/arsham/logpipe")
	if err != nil {
		log.Fatalf("should not build the program: %v", err)
	}
	defer func() {
		gexec.CleanupBuildArtifacts()

	}()

	port, err = getRandomPort()
	if err != nil {
		log.Fatal(err)
	}

	c, file1, file2, err := generateFiles()
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.WriteString(fmt.Sprintf(`
app:
  log_level: "info"
writers:
  file1:
    type: file
    location: %s
  file2:
    type: file
    location: %s
`, file1.Name(), file2.Name()))

	if err != nil {
		log.Fatal(err)
	}

	if c.Sync() != nil {
		log.Fatal(err)
	}

	command := exec.Command(program)
	env := os.Environ()
	env = append(env, fmt.Sprintf("LOGLEVEL=info"))
	env = append(env, fmt.Sprintf("PORT=%d", port))
	env = append(env, fmt.Sprintf("CONFIGFILE=%s", c.Name()))

	command.Env = env
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	if err != nil {
		log.Fatal(err)
	}

	done := truncate(file1, file2)
	x := m.Run()
	close(done)
	cleanup(session, c, file1, file2)
	os.Exit(x)
}

func truncate(file1, file2 *os.File) chan struct{} {
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-time.After(time.Millisecond * 30):
				if err := os.Truncate(file1.Name(), 0); err != nil {
					log.Fatal(err)
				}
				if err := os.Truncate(file2.Name(), 0); err != nil {
					log.Fatal(err)
				}
			case <-done:
				return
			}
		}
	}()
	return done
}

func cleanup(session *gexec.Session, c, file1, file2 *os.File) {
	if err := os.Remove(file1.Name()); err != nil {
		log.Fatal(err)
	}

	if err := os.Remove(file2.Name()); err != nil {
		log.Fatal(err)
	}
	if err := os.Remove(c.Name()); err != nil {
		log.Fatal(err)
	}
	session.Interrupt()
}

func BenchmarkMain(b *testing.B) {

	url := "http://127.0.0.1:" + strconv.Itoa(port) + "/"

	client := &http.Client{}
	errCh := make(chan error)

	for i := 0; i < b.N; i++ {
		message := []byte(fmt.Sprintf(`{"message":"%s"}`, randBytes(i+1)))
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(message))
		if err != nil {
			b.Fatal(err)
		}

		go func() {
			_, err := client.Do(req)
			errCh <- err
		}()

		select {
		case e := <-errCh:
			if e != nil {
				b.Log(e)
			}
		case <-time.After(time.Second):
			b.Error("timeout")
		}
	}
}
