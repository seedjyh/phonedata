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
	assert.Equal(t, Version("2108"), result.version)
	assert.Equal(t, 370, len(result.records))
	assert.Equal(t, 454336, len(result.induces))
}

func TestPack(t *testing.T) {
	fileData := &FileData{
		version: "1234",
		records: []*RecordItem{{
			id:       123,
			province: "p1",
			city:     "c1",
			zipCode:  "z1",
			areaZone: "a1",
		}, {
			id:       456,
			province: "p2",
			city:     "c2",
			zipCode:  "z2",
			areaZone: "a2",
		}},
		induces: []*IndexItem{{
			phonePrefix: "1333606",
			cardTypeID:  1,
			recordID:    123,
		}, {
			phonePrefix: "1333607",
			cardTypeID:  2,
			recordID:    456,
		}, {
			phonePrefix: "1333608",
			cardTypeID:  3,
			recordID:    123,
		}},
	}
	filePath := "tmp.dat"
	assert.NoError(t, Pack(fileData, filePath))
}
