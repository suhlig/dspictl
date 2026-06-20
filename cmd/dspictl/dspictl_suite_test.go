package main

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDSPiCtl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "dspictl Suite")
}
