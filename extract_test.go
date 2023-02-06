package phonedata

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"runtime"
	"testing"
)

func TestExtract(t *testing.T) {
	dir := os.Getenv("PHONE_DATA_DIR")
	if dir == "" {
		_, fulleFilename, _, _ := runtime.Caller(0)
		dir = path.Dir(fulleFilename)
	}
	var err error
	result, err := Extract(path.Join(dir, PHONE_DAT))
	assert.NoError(t, err)
	assert.Equal(t, "2108", result.version)
	assert.Equal(t, 370, len(result.offset2record))
	assert.Equal(t, 454336, len(result.seven2offset))
	assert.Equal(t, 454336, len(result.seven2type))
}
