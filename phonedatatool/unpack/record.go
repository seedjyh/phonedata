package unpack

import (
	"bytes"
	"fmt"
	"github.com/xluohome/phonedata/phonedatatool"
	"github.com/xluohome/phonedata/phonedatatool/util"
	"sort"
	"strconv"
	"strings"
)

type RecordItem struct {
	Province phonedatatool.ProvinceName
	City     phonedatatool.CityName
	ZipCode  phonedatatool.ZipCode
	AreaCode phonedatatool.AreaCode
}

func (ri *RecordItem) Parse(reader *bytes.Reader) error {
	if buf, err := util.ReadUntil(reader, 0); err != nil {
		return fmt.Errorf("no term char for record item: %v", err)
	} else {
		words := strings.Split(string(buf), "|")
		if len(words) != 4 {
			return fmt.Errorf("invalid item bytes, %v", string(buf))
		}
		ri.Province = phonedatatool.ProvinceName(words[0])
		ri.City = phonedatatool.CityName(words[1])
		ri.ZipCode = phonedatatool.ZipCode(words[2])
		ri.AreaCode = phonedatatool.AreaCode(words[3])
		return nil
	}
}

type RecordOffset int64

func (ro RecordOffset) String() string {
	return strconv.FormatInt(int64(ro), 10)
}

type RecordID int64

func (rid RecordID) String() string {
	return strconv.FormatInt(int64(rid), 10)
}

type RecordIDList []RecordID

func (l RecordIDList) Len() int {
	return len(l)
}
func (l RecordIDList) Less(i, j int) bool {
	return l[i] < l[j]
}
func (l RecordIDList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

type RecordPart struct {
	offset2item map[RecordOffset]*RecordItem
	offset2id   map[RecordOffset]RecordID
	id2offset   map[RecordID]RecordOffset
}

func NewRecordPart() *RecordPart {
	return &RecordPart{
		offset2item: make(map[RecordOffset]*RecordItem),
		offset2id:   make(map[RecordOffset]RecordID),
		id2offset:   make(map[RecordID]RecordOffset),
	}
}

func (rp *RecordPart) BytesPlainText() []byte {
	w := bytes.NewBuffer(nil)
	var idList RecordIDList
	for k := range rp.id2offset {
		idList = append(idList, k)
	}
	sort.Sort(idList)
	for _, id := range idList {
		item := rp.offset2item[rp.id2offset[id]]
		w.WriteString(strings.Join([]string{
			id.String(),
			item.Province.String(),
			item.City.String(),
			item.ZipCode.String(),
			item.AreaCode.String(),
		}, "|"))
		w.WriteByte('\n')
	}
	return w.Bytes()
}

func (rp *RecordPart) Parse(reader *bytes.Reader, endOffset RecordOffset) error {
	for nowID := RecordID(1); ; nowID++ {
		startOffset := RecordOffset(reader.Size() - int64(reader.Len()))
		if startOffset >= endOffset {
			break
		}
		item := new(RecordItem)
		if err := item.Parse(reader); err != nil {
			return err
		}
		rp.offset2item[startOffset] = item
		rp.offset2id[startOffset] = nowID
		rp.id2offset[nowID] = startOffset
	}
	return nil
}
