package ipfs

import (
	"bytes"
	"context"
	"fmt"
	carv2 "github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/blockstore"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"golang.org/x/xerrors"
	"io"
	log "metalib/logs"
	"metalib/util"
	"os"
	"runtime"
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
		if err := printLinksNode("", r, &ls, infoList); err != nil {
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

func GenerateCarFile(destCar string, srcFiles []string) error {
	parallel := runtime.NumCPU()
	sliceSize := (1 << 30) * 16 // 16G
	parentPath := ""
	outputDir := ""
	isUuid := false
	if !util.ExistDir(outputDir) {
		return xerrors.Errorf("Unexpected! The path of output dir does not exist")
	}
	graphName := "graph-name"
	if sliceSize == 0 {
		return xerrors.Errorf("Unexpected! Slice size has been set as 0")
	}
	inputPath := ""

	doGenerateCar(int64(sliceSize), parentPath, inputPath, outputDir, graphName, int(parallel), isUuid)

	return nil
}

func GenerateCarFileWithUuid(outputDir string, srcFiles []string, uuid []string, sliceSize int64) (string, error) {

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
