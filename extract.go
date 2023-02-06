package phonedata

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"strings"
)

type Offset int32
type PhoneSeven int32
type CardTypeIndex byte

type ExtractResult struct {
	version       string
	offset2record map[Offset]*PhoneRecord
	seven2offset  map[PhoneSeven]Offset
	seven2type    map[PhoneSeven]CardTypeIndex
}

func NewExtractResult() *ExtractResult {
	return &ExtractResult{
		version:       "",
		offset2record: make(map[Offset]*PhoneRecord),
		seven2offset:  make(map[PhoneSeven]Offset),
		seven2type:    make(map[PhoneSeven]CardTypeIndex),
	}
}

// Extract 读取数据文件并解析
// dataFilePath 原始数据文件的路径。可以是绝对路径或相对路径，但必须能读取。
func Extract(dataFilePath string) (*ExtractResult, error) {
	var content []byte
	if c, err := ioutil.ReadFile(dataFilePath); err != nil {
		panic(err)
	} else {
		content = c
	}

	result := NewExtractResult()

	result.version = string(content[:4])

	var indexStartOffset int32
	if err := binary.Read(bytes.NewBuffer(content[4:8]), binary.LittleEndian, &indexStartOffset); err != nil {
		return nil, err
	} else if indexStartOffset > int32(len(content)) {
		return nil, fmt.Errorf("indexStartOffset(%d) is larger than file size(%d)", indexStartOffset, len(content))
	}

	var currentOffset Offset = 8
	for i := Offset(8); i < Offset(indexStartOffset); i++ {
		if content[i] == 0 {
			currentRecordLine := content[currentOffset:i]
			if currentRecord, err := parseRecord(currentRecordLine); err != nil {
				return nil, fmt.Errorf("failed to parse record data, offset=%d, data=%+v, error=%w", currentOffset, currentRecordLine, err)
			} else {
				result.offset2record[currentOffset] = currentRecord
			}
			currentOffset = i + 1
		}
	}

	const IndexLineSize = Offset(9)
	for i := Offset(indexStartOffset); i < Offset(len(content)); i += IndexLineSize {
		var line [IndexLineSize]byte
		copy(line[:], content[i:])
		var phoneSeven32 PhoneSeven
		var recordOffset32 Offset
		if err := binary.Read(bytes.NewBuffer(line[:4]), binary.LittleEndian, &phoneSeven32); err != nil {
			return nil, fmt.Errorf("failed to parse phone seven segment in index line, index line offset=%d, error=%w", i, err)
		}
		if err := binary.Read(bytes.NewBuffer(line[4:8]), binary.LittleEndian, &recordOffset32); err != nil {
			return nil, fmt.Errorf("failed to parse offset segment in index line, index line offset=%d, error=%w", i, err)
		}
		cardTypeIndex := CardTypeIndex(line[8])
		result.seven2offset[phoneSeven32] = recordOffset32
		result.seven2type[phoneSeven32] = cardTypeIndex
	}

	return result, nil
}

// parseRecord 解析 record 数据
// b 待解析的 record 数据，形如「山东|济南|250000|0531」，末尾不带 '\0'
func parseRecord(b []byte) (*PhoneRecord, error) {
	segments := strings.Split(string(b), "|")
	if len(segments) != 4 {
		return nil, fmt.Errorf("got %d segments, require 4", len(segments))
	}
	return &PhoneRecord{
		PhoneNum: "",
		Province: segments[0],
		City:     segments[1],
		ZipCode:  segments[2],
		AreaZone: segments[3],
		CardType: "",
	}, nil
}
