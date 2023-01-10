package ipfs

import (
	"bytes"
	"context"
	"fmt"
	log "github.com/FogMeta/meta-lib/logs"
	"github.com/FogMeta/meta-lib/util"
	carv2 "github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/blockstore"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"golang.org/x/xerrors"
	"io"
	"os"
)

func ListCarFile(destCar string) ([]string, error) {
	infoList := make([]string, 0)

	bs, err := blockstore.OpenReadOnly(destCar)
	if err != nil {
		return infoList, err
	}
	ls := cidlink.DefaultLinkSystem()
	ls.TrustedStorage = true
	ls.StorageReadOpener = func(_ ipld.LinkContext, l ipld.Link) (io.Reader, error) {
		cl, ok := l.(cidlink.Link)
		if !ok {
			return nil, fmt.Errorf("not a cidlink")
		}

		blk, err := bs.Get(context.Background(), cl.Cid)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(blk.RawData()), nil
	}

	roots, err := bs.Roots()
	if err != nil {
		return infoList, err
	}

	for _, r := range roots {
		if err := printLinksNode("", r, &ls, &infoList); err != nil {
			return infoList, err
		}
	}

	return infoList, nil
}

func GetCarRoot(destCar string) (string, error) {
	root := ""
	inStream, err := os.Open(destCar)
	if err != nil {
		return root, err
	}

	rd, err := carv2.NewBlockReader(inStream)
	if err != nil {
		return root, err
	}
	for _, r := range rd.Roots {
		root = r.String()
	}

	return root, nil
}

func GenerateCarFromFiles(outputDir string, srcFiles []string, sliceSize int64) (string, error) {

	if !util.ExistDir(outputDir) {
		return "", xerrors.Errorf("Unexpected! The path of output dir does not exist")
	}

	var totalSize int64 = 0
	files := util.GetFileListAsync(srcFiles, false)
	for item := range files {
		totalSize += item.Info.Size()
	}
	if totalSize > sliceSize {
		return "", xerrors.Errorf("Total files size has been bigger than sliceSize(%u)", sliceSize)
	}

	carFileName, _, err := doGenerateCarFrom(outputDir, srcFiles)
	if err != nil {
		return "", err
	}

	return carFileName, nil
}

func GenerateCarFromDir(outputDir string, srcDir string, sliceSize int64) (string, error) {

	if !util.ExistDir(outputDir) {
		return "", xerrors.Errorf("Unexpected! The path of output dir does not exist")
	}

	var totalSize int64 = 0
	files := util.GetFileListAsync([]string{srcDir}, false)
	for item := range files {
		totalSize += item.Info.Size()
	}
	if totalSize > sliceSize {
		return "", xerrors.Errorf("Total files size has been bigger than sliceSize(%u)", sliceSize)
	}

	carFileName, _, err := doGenerateCarFrom(outputDir, []string{srcDir})
	if err != nil {
		return "", err
	}

	return carFileName, nil
}

func GenerateCarFromFilesWithUuid(outputDir string, srcFiles []string, uuid []string, sliceSize int64) (string, error) {

	if len(srcFiles) != len(uuid) {
		return "", xerrors.Errorf("The len of source files and uuids do not match.")
	}

	if !checkFiles(srcFiles, sliceSize) {
		return "", xerrors.Errorf("Total files size has been bigger than sliceSize(%u)", sliceSize)
	}

	carFileName, detailJson, err := doGenerateCarWithUuid(outputDir, srcFiles, uuid)
	if err != nil {
		return "", err
	}

	log.GetLog().Info(detailJson)
	return carFileName, nil
}
