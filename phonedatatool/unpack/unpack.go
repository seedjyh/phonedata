package unpack

import (
	"bytes"
	"fmt"
	"github.com/xluohome/phonedata/phonedatatool"
	"github.com/xluohome/phonedata/phonedatatool/util"
	"os"
	"path"
)

const (
	VersionFileName = "version.txt"
	RecordFileName  = "record.txt"
	IndexFileName   = "index.txt"
)

type Unpacker struct {
}

func NewUnpacker() phonedatatool.Unpacker {
	return new(Unpacker)
}

func (u *Unpacker) Unpack(phoneDataFilePath string, plainDirectoryPath string) error {
	if err := os.MkdirAll(plainDirectoryPath, 0); err != nil {
		return fmt.Errorf("target directory %v not exist and can't be created: %v", plainDirectoryPath, err)
	}

	versionFilePath := path.Join(plainDirectoryPath, VersionFileName)
	recordFilePath := path.Join(plainDirectoryPath, RecordFileName)
	indexFilePath := path.Join(plainDirectoryPath, IndexFileName)

	if err := util.AssureAllFileNotExist(versionFilePath, recordFilePath, indexFilePath); err != nil {
		return err
	}

	var rawBuf []byte
	if b, err := os.ReadFile(phoneDataFilePath); err != nil {
		return err
	} else {
		rawBuf = b
	}

	if res, err := u.parse(rawBuf); err != nil {
		return fmt.Errorf("failed to parse raw file data: %v", err)
	} else {
		if err := os.WriteFile(versionFilePath, res.versionPart.BytesPlainText(), 0); err != nil {
			return err
		}
		if err := os.WriteFile(recordFilePath, res.recordPart.BytesPlainText(), 0); err != nil {
			return err
		}
		if err := os.WriteFile(indexFilePath, res.indexPart.BytesPlainText(), 0); err != nil {
			return err
		}
		return nil
	}
}

type ParseResult struct {
	versionPart *VersionPart
	recordPart  *RecordPart
	indexPart   *IndexPart
}

func (u *Unpacker) parse(buf []byte) (*ParseResult, error) {
	reader := bytes.NewReader(buf)
	versionPart := new(VersionPart)
	if err := versionPart.Parse(reader); err != nil {
		return nil, fmt.Errorf("failed to read version part: %v", err)
	}
	offsetPart := new(IndexPartOffsetPart)
	if err := offsetPart.Parse(reader); err != nil {
		return nil, fmt.Errorf("failed to read index-part-offset part: %v", err)
	}
	recordPart := NewRecordPart()
	if err := recordPart.Parse(reader, offsetPart.IndexPartOffset); err != nil {
		return nil, fmt.Errorf("failed to read record part: %v", err)
	}
	indexPart := NewIndexPart()
	if err := indexPart.Parse(reader); err != nil {
		return nil, fmt.Errorf("failed to read index part: %v", err)
	}
	if err := indexPart.MatchRecordOffsetToRecordID(recordPart.offset2id); err != nil {
		return nil, fmt.Errorf("failed to match offset to record id: %v", err)
	}

	return &ParseResult{
		versionPart: versionPart,
		recordPart:  recordPart,
		indexPart:   indexPart,
	}, nil
}
