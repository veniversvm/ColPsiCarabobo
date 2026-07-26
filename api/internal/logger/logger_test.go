package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit_Development(t *testing.T) {
	require.NotPanics(t, func() {
		Init("development")
	})
}

func TestInit_EmptyString(t *testing.T) {
	require.NotPanics(t, func() {
		Init("")
	})
}

func TestInit_Production(t *testing.T) {
	require.NotPanics(t, func() {
		Init("production")
	})
}

func TestInit_Staging(t *testing.T) {
	require.NotPanics(t, func() {
		Init("staging")
	})
}
