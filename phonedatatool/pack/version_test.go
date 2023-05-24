package pack

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVersionPart_Bytes(t *testing.T) {
	assert.Equal(t, []byte{'2', '3', '0', '6'}, (&VersionPart{version: "2306"}).Bytes())
}

func TestVersionPart_ParsePlainText(t *testing.T) {
	reader := bytes.NewReader([]byte("2306\n"))
	versionPart := new(VersionPart)
	assert.NoError(t, versionPart.ParsePlainText(reader))
	assert.Equal(t, "2306", versionPart.version)
}