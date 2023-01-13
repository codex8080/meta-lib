package main

import (
	log "github.com/FogMeta/meta-lib/logs"
	meta_car "github.com/FogMeta/meta-lib/module/ipfs"
	"github.com/FogMeta/meta-lib/util"
	"golang.org/x/xerrors"
	"os"
)

func main() {

	genCarWithUuidDemo()

	genCarFromFilesDemo()

	genCarFromDirDemo()

	listCarDemo()

	getCarRootDemo()

	genCarFromDirsDemo()

	return
}

func genCarWithUuidDemo() {
	outputDir := "../../test/output"
	srcFiles := []string{
		"../../test/input/test0",
		"../../test/input/test4",
		"../../test/input/dir1/test1",
		"../../test/input/dir1/dir2/test2",
		"../../test/input/dir1/dir2/test3",
	}
	uuid := []string{
		"94d6a0d0-3e76-45b7-9705-4d829e0e3ca8",
		"571e4e2b-d50b-4ac2-a89f-07795b684148",
		"36f4da38-a028-493a-a855-51b07269e709",
		"e99d2819-09a8-4e53-8158-a48d8154e057",
		"6631aa2a-5e89-4f98-b114-86bf4403f1c2",
	}
	sliceSize := 17179869184

	carFileName, err := meta_car.GenerateCarFromFilesWithUuid(outputDir, srcFiles, uuid, int64(sliceSize))
	if err != nil {
		log.GetLog().Error("Test create car file error:", err)
		return
	}

	log.GetLog().Info("create car file is:", carFileName)

}

func genCarFromFilesDemo() {
	outputDir := "../../test/output"
	srcFiles := []string{
		"../../test/input/test0",
		"../../test/input/test4",
		"../../test/input/dir1/test1",
		"../../test/input/dir1/dir2/test2",
		"../../test/input/dir1/dir2/test3",
	}
	sliceSize := 17179869184

	carFileName, err := meta_car.GenerateCarFromFiles(outputDir, srcFiles, int64(sliceSize))
	if err != nil {
		log.GetLog().Error("Create car file error:", err)
		return
	}

	log.GetLog().Info("Create car file is:", carFileName)

}

func genCarFromDirDemo() {
	outputDir := "../../test/output"
	srcDir := "../../test/input/"
	sliceSize := 17179869184

	carFileName, err := meta_car.GenerateCarFromDir(outputDir, srcDir, int64(sliceSize))
	if err != nil {
		log.GetLog().Error("Create car file error:", err)
		return
	}

	log.GetLog().Info("Create car file is:", carFileName)

}

func listCarDemo() {
	destCar := "../../test/output/QmUabWJFQGr1hWxhLikB9eLjfRZcaoTrQZJYTMP6AnozN7.car"
	infoList, err := meta_car.ListCarFile(destCar)
	if err != nil {
		log.GetLog().Error("List car file info error:", err)
	}

	log.GetLog().Info("Car info:\n", infoList)

}

func getCarRootDemo() {
	destCar := "../../test/output/QmUabWJFQGr1hWxhLikB9eLjfRZcaoTrQZJYTMP6AnozN7.car"
	rootCid, err := meta_car.GetCarRoot(destCar)
	if err != nil {
		log.GetLog().Error("List car file info error:", err)
	}

	log.GetLog().Info("Root CID is:", rootCid)

}

func genCarFromDirsDemo() {
	outputDir := "../../test/output"
	srcDir := []string{
		"../../test/input1/",
		"../../test/input2/",
		"../../test/input3/",
	}
	sliceSize := 17179869184

	carInfos, err := GenCarFromDirs(outputDir, srcDir, int64(sliceSize))
	if err != nil {
		log.GetLog().Error("Create car file error:", err)
		return
	}

	log.GetLog().Infof("%+v", carInfos)

}

type CarInfo struct {
	CarFile     string
	TotalSize   int64
	ContainDirs []string
}

func GetFilesSize(args []string) (int64, error) {
	totalSize := int64(0)
	fileList, err := util.GetFileList(args)
	if err != nil {
		return int64(0), err
	}

	for _, path := range fileList {
		finfo, err := os.Stat(path)
		if err != nil {
			return int64(0), err
		}
		totalSize += finfo.Size()
	}

	return totalSize, nil
}

func GenCarFromDirs(outputDir string, srcDir []string, sliceSize int64) ([]CarInfo, error) {

	if !util.ExistDir(outputDir) {
		return nil, xerrors.Errorf("Unexpected! The path of output dir does not exist")
	}

	carInfos := make([]CarInfo, 0)

	accSize := int64(0)
	accDirs := make([]string, 0)
	for _, dir := range srcDir {

		dirSize, err := GetFilesSize([]string{dir})
		if err != nil {
			log.GetLog().Errorf("Get %s size error:%s", dir, err)
			continue
		}

		if dirSize > sliceSize {
			log.GetLog().Errorf("%s size is %d and bigger than :%d", dir, dirSize, sliceSize)
			continue
		}

		if (accSize + dirSize) > sliceSize {
			//to build car
			carFileName, err := meta_car.GenerateCarFromFiles(outputDir, accDirs, sliceSize)
			if err != nil {
				log.GetLog().Errorf("%s size is %d and bigger than :%d", dir, dirSize, sliceSize)
				continue
			}

			carInfos = append(carInfos, CarInfo{
				CarFile:     carFileName,
				TotalSize:   accSize,
				ContainDirs: accDirs,
			})
			log.GetLog().Info("Create CAR:", carFileName)

			accSize = int64(0)
			accDirs = make([]string, 0)
			continue
		}

		accSize += dirSize
		accDirs = append(accDirs, dir)
	}

	return carInfos, nil
}
