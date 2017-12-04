package push_handler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"

    "github.com/Sirupsen/logrus"
)

func TestPushHandler(t *testing.T) {
	RegisterFailHandler(Fail)
    logrus.SetLevel(logrus.PanicLevel)
	RunSpecs(t, "PushHandler Suite")
}
