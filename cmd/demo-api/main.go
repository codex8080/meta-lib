package main

import (
	log "github.com/codex8080/meta-lib/logs"
	meta_car "github.com/codex8080/meta-lib/module/ipfs"
)

func main() {

	genCarWithUuidDemo()

	genCarFromFilesDemo()

	genCarFromDirDemo()

	listCarDemo()

	GetCarRootDemo()

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

func GetCarRootDemo() {
	destCar := "../../test/output/QmUabWJFQGr1hWxhLikB9eLjfRZcaoTrQZJYTMP6AnozN7.car"
	rootCid, err := meta_car.GetCarRoot(destCar)
	if err != nil {
		log.GetLog().Error("List car file info error:", err)
	}

	log.GetLog().Info("Root CID is:", rootCid)

}
