package manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootPath(t *testing.T) {
	assert.Equal(t, "", rootPath(""))
	assert.Equal(t, "", rootPath("/"))
	assert.Equal(t, "", rootPath("/abc"))
	assert.Equal(t, "..", rootPath("/abc/"))
	assert.Equal(t, "..", rootPath("/abc/ddd"))
}
