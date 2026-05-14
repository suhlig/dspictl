package dspi_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDSPi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DSPi Suite")
}
