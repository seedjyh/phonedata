package phonedata

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Version string
type RecordID int64
type CardTypeID int
type Offset32 int32

func (o Offset32) ToLittleEndianByte4() []byte {
	buf := bytes.NewBuffer(nil)
	_ = binary.Write(buf, binary.LittleEndian, int32(o))
	return buf.Bytes()
}

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

type RecordItemList []*RecordItem

func (l RecordItemList) Len() int {
	return len(l)
}
func (l RecordItemList) Less(i, j int) bool {
	return l[i].id < l[j].id
}
func (l RecordItemList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

type IndexItemList []*IndexItem

func (l IndexItemList) Len() int {
	return len(l)
}
func (l IndexItemList) Less(i, j int) bool {
	return l[i].phonePrefix < l[j].phonePrefix
}
func (l IndexItemList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

type FileData struct {
	version Version
	records RecordItemList
	induces IndexItemList
}

// Extract 读取数据文件并解析
// dataFilePath 原始数据文件的路径。可以是绝对路径或相对路径，但必须能读取。
func Extract(dataFilePath string) (*FileData, error) {
	var content []byte
	if c, err := ioutil.ReadFile(dataFilePath); err != nil {
		panic(err)
	} else {
		content = c
	}

	return parseFileData(content)

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

// parseFileData 解析 phone data 文件
// b 是来自 phone data 文件的字节流
func parseFileData(b []byte) (*FileData, error) {
	version := Version(content[:4])
	var indexStartOffset int32
	if err := binary.Read(bytes.NewBuffer(content[4:8]), binary.LittleEndian, &indexStartOffset); err != nil {
		return nil, err
	} else if indexStartOffset > int32(len(content)) {
		return nil, fmt.Errorf("indexStartOffset(%d) is larger than file size(%d)", indexStartOffset, len(content))
	}

	var currentRecordItemStartOffset Offset32 = 8
	offset2record := make(map[Offset32]*RecordItem)
	var records []*RecordItem
	var induces []*IndexItem
	for i := Offset32(8); i < Offset32(indexStartOffset); i++ {
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

	const IndexLineSize = Offset32(9)
	for i := Offset32(indexStartOffset); i < Offset32(len(content)); i += IndexLineSize {
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

	return &FileData{
		version: version,
		records: records,
		induces: induces,
	}, nil
}

// parseIndexItem 解析一项索引项
// b: 原始文件中一条索引项
func parseIndexItem(b []byte) (string, Offset32, CardTypeID, error) {
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
	return fmt.Sprintf("%d", phoneSeven32), Offset32(recordOffset32), cardTypeID, nil
}

func Pack(fileData *FileData, dataFilePath string) error {
	var content []byte
	if c, err := packFileData(fileData); err != nil {
		return fmt.Errorf("failed to pack phone data file, %w", err)
	} else {
		content = c
	}
	return ioutil.WriteFile(dataFilePath, content, os.ModePerm)
}

func packFileData(fileData *FileData) ([]byte, error) {
	sort.Sort(fileData.records)
	sort.Sort(fileData.induces)

	recordID2offset := make(map[RecordID]Offset32)
	buf := bytes.NewBuffer(nil)
	buf.Write([]byte(fileData.version)[:4])
	buf.Write(make([]byte, 4)) // 之后要用来填写 index 的 offset

	for _, recordItem := range fileData.records {
		offset := buf.Len()
		id := recordItem.id
		recordID2offset[id] = Offset32(offset)
		if _, err := buf.WriteString(strings.Join([]string{recordItem.province, recordItem.city, recordItem.zipCode, recordItem.areaZone}, "|")); err != nil {
			return nil, fmt.Errorf("failed to pack file data, %w", err)
		}
		buf.WriteByte(0)
	}

	indexOffset := buf.Len()
	for _, indexItem := range fileData.induces {
		if prefixInt, err := strconv.Atoi(indexItem.phonePrefix); err != nil {
			return nil, fmt.Errorf("invalid phone prefix %s, %w", indexItem.phonePrefix, err)
		} else if err := binary.Write(buf, binary.LittleEndian, int32(prefixInt)); err != nil {
			return nil, fmt.Errorf("failed to pack phone prefix integer %d, %w", prefixInt, err)
		}
		if offset, ok := recordID2offset[indexItem.recordID]; !ok {
			return nil, fmt.Errorf("invalid record ID %d", indexItem.recordID)
		} else if _, err := buf.Write(offset.ToLittleEndianByte4()); err != nil {
			return nil, fmt.Errorf("failed to write record offset %+v, %w", offset, err)
		}
		buf.WriteByte(byte(indexItem.cardTypeID))
	}

	content := buf.Bytes()
	copy(content[4:], Offset32(indexOffset).ToLittleEndianByte4())
	return content, nil
}
