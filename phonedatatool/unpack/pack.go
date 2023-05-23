package unpack

import (
	"bytes"
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
	return nil
}
