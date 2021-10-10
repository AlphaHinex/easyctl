package deny

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/weiliang-ms/easyctl/pkg/util/command"
	"testing"
)

func TestSelinux(t *testing.T) {
	err := Selinux(command.OperationItem{
		B:          nil,
		Logger:     logrus.New(),
		OptionFunc: nil,
	})

	assert.Equal(t, command.RunErr{}, err)
}
