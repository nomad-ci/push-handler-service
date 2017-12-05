package push_handler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

    "github.com/Sirupsen/logrus"
)

func TestPushHandler(t *testing.T) {
	RegisterFailHandler(Fail)

    // ginkgo will only output the messages if there's a test failure
    logrus.SetOutput(GinkgoWriter)

    // so we can crank the verbosity
    logrus.SetLevel(logrus.DebugLevel)

    RunSpecs(t, "PushHandler Suite")
}
