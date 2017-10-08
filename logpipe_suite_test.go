package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLogpipe(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logpipe Suite")
}
