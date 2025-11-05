package logger_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Te8va/GophKeeper/pkg/logger"
)

func TestLogger(t *testing.T) {
	str := "abc_tmp.txt"
	logger.SetLogFile(str)

	n, err := logger.Logger().Write([]byte(str))
	require.NoError(t, err)
	require.Equal(t, len(str), n)
}
