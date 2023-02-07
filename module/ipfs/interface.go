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
	"path/filepath"
	"runtime"
)

type DetailInfo struct {
	FilePath string `json:"file_path"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	CID      string `json:"cid"`
	UUID     string `json:"uuid"`
}

type CarInfo struct {
	CarFilePath string       `json:"car_file_path"`
	CarFileName string       `json:"car_file_name"`
	RootCid     string       `json:"root_cid"`
	PieceCID    string       `json:"piece_cid"`
	PieceSize   int64        `json:"piece_size"`
	Details     []DetailInfo `json:"details"`
}

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

func GenerateCarFromDirEx(outputDir string, srcDir string, sliceSize int64, withUUID bool) ([]CarInfo, error) {

	if !util.ExistDir(outputDir) {
		return nil, xerrors.Errorf("Unexpected! The path of output dir does not exist")
	}

	files := util.GetFileListAsync([]string{srcDir}, withUUID)
	accSize := int64(0)
	accFiles := make([]string, 0)
	accUUIDs := make([]string, 0)
	remainFiles := make([]string, 0)
	buildCars := make([]CarInfo, 0)
	for item := range files {
		fileSize := item.Info.Size()
		if fileSize > sliceSize {
			log.GetLog().Errorf("%s size is %d and bigger than: %d", item.Path, fileSize, sliceSize)
			remainFiles = append(remainFiles, item.Path)
			continue
		}

		if (accSize + fileSize) > sliceSize {
			if len(accFiles) != len(accUUIDs) {
				log.GetLog().Error("The length of accFiles should be the same as the length of accUUIDs.")
				continue
			}

			carFile, rootCid, detailStr, detaiInfo, err := doGenerateCarWithUuidEx(outputDir, accFiles, accUUIDs)
			if err != nil {
				log.GetLog().Error("generate CAR file error:", err)
				//TODO: move accFiles to remainFiles
				continue
			}

			//one CAR generated
			log.GetLog().Debug("Create CAR: ", carFile)
			log.GetLog().Debug("Create Detail: ", detailStr)

			buildCars = append(buildCars, CarInfo{
				CarFilePath: carFile,
				CarFileName: filepath.Base(carFile),
				RootCid:     rootCid,
				Details:     detaiInfo,
			})

			accSize = int64(0)
			accFiles = accFiles[:0] //make([]string, 0)
			accUUIDs = accUUIDs[:0]
			continue
		}

		accSize += fileSize
		accFiles = append(accFiles, item.Path)
		accUUIDs = append(accUUIDs, item.Uuid)
	}

	if accSize > 0 {
		if len(accFiles) != len(accUUIDs) {
			log.GetLog().Error("The length of accFiles should be the same as the length of accUUIDs.")
		}

		carFile, rootCid, detailStr, detaiInfo, err := doGenerateCarWithUuidEx(outputDir, accFiles, accUUIDs)
		if err != nil {
			log.GetLog().Error("generate CAR file error:", err)
			//TODO: move accFiles to remainFiles
		}
		//one CAR generated
		log.GetLog().Debug("Create CAR: ", carFile)
		log.GetLog().Debug("Create Detail: ", detailStr)

		buildCars = append(buildCars, CarInfo{
			CarFilePath: carFile,
			CarFileName: filepath.Base(carFile),
			RootCid:     rootCid,
			Details:     detaiInfo,
		})

	}

	//TODO: write json file to output dir
	log.GetLog().Debug("Build CARs Info:", buildCars)
	return buildCars, nil
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

func RestoreCar(outputDir string, srcCar string) error {

	parallel := runtime.NumCPU()
	carTo(srcCar, outputDir, parallel)
	merge(outputDir, parallel)
	fmt.Println("completed!")

	return nil
}

func ExtractFileFromCar(outputDir string, srcCar string, inFileName string) error {
	extractFromCar(srcCar, outputDir, inFileName)
	fmt.Println("completed!")
	return nil
}
