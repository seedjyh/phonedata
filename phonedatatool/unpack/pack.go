package unpack

import (
	"bytes"
	"fmt"
	"github.com/xluohome/phonedata/phonedatatool"
	"github.com/xluohome/phonedata/phonedatatool/util"
	"os"
	"path"
)

type Packer struct{}

func NewPacker() phonedatatool.Packer {
	return new(Packer)
}

func (p *Packer) Pack(plainDirectoryPath string, phoneDataFilePath string) error {
	if err := util.AssureFileNotExist(phoneDataFilePath); err != nil {
		return err
	}

	versionPart := new(VersionPart)
	if buf, err := os.ReadFile(path.Join(plainDirectoryPath, VersionFileName)); err != nil {
		return err
	} else if err := versionPart.ParsePlainText(bytes.NewReader(buf)); err != nil {
		return err
	}
	versionPartBuf := versionPart.Bytes()

	recordPartOffset := RecordOffset(new(VersionPart).EncodedLen() + new(IndexPartOffsetPart).EncodedLen())
	recordPart := NewRecordPart()
	if buf, err := os.ReadFile(path.Join(plainDirectoryPath, RecordFileName)); err != nil {
		return err
	} else if err := recordPart.ParsePlainText(bytes.NewReader(buf)); err != nil {
		return err
	}
	recordPartBuf, recordID2Offset := recordPart.Bytes()

	indexPartOffsetPart := &IndexPartOffsetPart{
		IndexPartOffset: recordPartOffset + RecordOffset(len(recordPartBuf)),
	}

	indexPart := NewIndexPart()
	if buf, err := os.ReadFile(path.Join(plainDirectoryPath, IndexFileName)); err != nil {
		return err
	} else if err := indexPart.ParsePlainText(bytes.NewReader(buf)); err != nil {
		return err
	}
	indexPartBuf := indexPart.Bytes(recordID2Offset, recordPartOffset)

	var outFile *os.File
	if f, err := os.OpenFile(phoneDataFilePath, os.O_CREATE|os.O_WRONLY, 0); err != nil {
		return fmt.Errorf("failed to open file %v to write, %v", phoneDataFilePath, err)
	} else {
		outFile = f
	}
	defer outFile.Close()

	if _, err := outFile.Write(versionPartBuf); err != nil {
		return err
	}
	if _, err := outFile.Write(indexPartOffsetPart.Bytes()); err != nil {
		return err
	}
	if _, err := outFile.Write(recordPartBuf); err != nil {
		return err
	}
	if _, err := outFile.Write(indexPartBuf); err != nil {
		return err
	}
	return nil
}
