package apiclient_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAPIClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "APIClient Suite")
}
