package phonedata

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"strings"
)

type Version string
type RecordID int64
type CardTypeID int
type RecordOffset int32

// RecordItem 是编码后的文件里的一条记录项。
type RecordItem struct {
	id       RecordID
	province string
	city     string
	zipCode  string
	areaZone string
}

// IndexItem 是编码后的文件里的一条索引项。
type IndexItem struct {
	phonePrefix string     // 号码前缀
	cardTypeID  CardTypeID // 号码类型
	recordID    RecordID   // 归属记录 ID
}

type PhoneFile struct {
	version Version
	records []*RecordItem
	induces []*IndexItem
}

// Extract 读取数据文件并解析
// dataFilePath 原始数据文件的路径。可以是绝对路径或相对路径，但必须能读取。
func Extract(dataFilePath string) (*PhoneFile, error) {
	var content []byte
	if c, err := ioutil.ReadFile(dataFilePath); err != nil {
		panic(err)
	} else {
		content = c
	}

	version := Version(content[:4])

	var indexStartOffset int32
	if err := binary.Read(bytes.NewBuffer(content[4:8]), binary.LittleEndian, &indexStartOffset); err != nil {
		return nil, err
	} else if indexStartOffset > int32(len(content)) {
		return nil, fmt.Errorf("indexStartOffset(%d) is larger than file size(%d)", indexStartOffset, len(content))
	}

	var currentRecordItemStartOffset RecordOffset = 8
	offset2record := make(map[RecordOffset]*RecordItem)
	var records []*RecordItem
	var induces []*IndexItem
	for i := RecordOffset(8); i < RecordOffset(indexStartOffset); i++ {
		if content[i] == 0 { // 遇到一个 record line 的末尾
			currentRecordLine := content[currentRecordItemStartOffset:i]
			if currentRecord, err := parseRecordItem(currentRecordLine); err != nil {
				return nil, fmt.Errorf("failed to parse record data, offset=%d, data=%+v, error=%w", currentRecordItemStartOffset, currentRecordLine, err)
			} else {
				currentRecordID := RecordID(len(records) + 1)
				records = append(records, currentRecord)
				currentRecord.id = currentRecordID
				offset2record[currentRecordItemStartOffset] = currentRecord
			}
			currentRecordItemStartOffset = i + 1
		}
	}

	const IndexLineSize = RecordOffset(9)
	for i := RecordOffset(indexStartOffset); i < RecordOffset(len(content)); i += IndexLineSize {
		if phonePrefix, recordOffset, cardTypeID, err := parseIndexItem(content[i:]); err != nil {
			return nil, fmt.Errorf("failed to parse phone seven segment in index line, index line offset=%d, error=%w", i, err)
		} else if record, ok := offset2record[recordOffset]; !ok {
			return nil, fmt.Errorf("found invalid record offset %d", recordOffset)
		} else {
			induces = append(induces, &IndexItem{
				phonePrefix: phonePrefix,
				cardTypeID:  cardTypeID,
				recordID:    record.id,
			})
		}
	}

	return &PhoneFile{
		version: version,
		records: records,
		induces: induces,
	}, nil
}

// parseRecordItem 解析 record 数据
// b 待解析的 record 数据，形如「山东|济南|250000|0531」，末尾不带 '\0'
func parseRecordItem(b []byte) (*RecordItem, error) {
	segments := strings.Split(string(b), "|")
	if len(segments) != 4 {
		return nil, fmt.Errorf("got %d segments, require 4", len(segments))
	}
	return &RecordItem{
		province: segments[0],
		city:     segments[1],
		zipCode:  segments[2],
		areaZone: segments[3],
	}, nil
}

// parseIndexItem 解析一项索引项
// b: 原始文件中一条索引项
func parseIndexItem(b []byte) (string, RecordOffset, CardTypeID, error) {
	if len(b) < 9 {
		return "", 0, 0, fmt.Errorf("invalid length %d, require 9", len(b))
	}
	var phoneSeven32 int32
	if err := binary.Read(bytes.NewBuffer(b[:4]), binary.LittleEndian, &phoneSeven32); err != nil {
		return "", 0, 0, fmt.Errorf("failed to parse phone seven segment in index line, error=%w", err)
	}
	var recordOffset32 int32
	if err := binary.Read(bytes.NewBuffer(b[4:8]), binary.LittleEndian, &recordOffset32); err != nil {
		return "", 0, 0, fmt.Errorf("failed to parse offset segment in index line, error=%w", err)
	}
	cardTypeID := CardTypeID(b[8])
	return fmt.Sprintf("%d", phoneSeven32), RecordOffset(recordOffset32), cardTypeID, nil
}
