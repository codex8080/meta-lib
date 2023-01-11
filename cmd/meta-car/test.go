package main

import (
	log "github.com/FogMeta/meta-lib/logs"
	meta_car "github.com/FogMeta/meta-lib/module/ipfs"
	"github.com/urfave/cli/v2"
)

func CreateCarFileTest(c *cli.Context) error {

	genCarWithUuidDemo()

	genCarFromFilesDemo()

	genCarFromDirDemo()

	listCarDemo()

	GetCarRootDemo()

	return nil
}

func carFileTest() {
	destFile := "./test/output/test.car"
	srcFiles := []string{
		"./test/input/test0",
		"./test/input/dir1/test1",
		"./test/input/dir1/dir2/test2",
	}

	if err := meta_car.CreateCarFile(destFile, srcFiles); err != nil {
		log.GetLog().Error("Test create car file error:", err)
		return
	}

}

func genCarWithUuidDemo() {
	outputDir := "./test/output"
	srcFiles := []string{
		"./test/input/test0",
		"./test/input/test4",
		"./test/input/dir1/test1",
		"./test/input/dir1/dir2/test2",
		"./test/input/dir1/dir2/test3",
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

	/*
		OUTPUT:
		2023-01-10T07:48:03.788Z        INFO    meta    ipfs/gencar.go:694      FILE:./test/input/dir1/dir2/test3    CID:QmcA1M4cUFeGZGTwTHAMPZrt6yXyRXkgDoBNeYQ3bhbuJD    UUID:6631aa2a-5e89-4f98-b114-86bf4403f1c2      SIZE:49

		2023-01-10T07:48:03.787Z        INFO    meta    ipfs/gencar.go:694      FILE:./test/input/dir1/dir2/test2    CID:QmeRAKJCjykxuU8NTjtWgeX59Zn8xcCp4NLF6jw3dCrAnX    UUID:e99d2819-09a8-4e53-8158-a48d8154e057      SIZE:57

		2023-01-10T07:48:03.788Z        INFO    meta    ipfs/gencar.go:694      FILE:./test/input/dir1/test1    CID:QmTGVzSq5v5mzYUt9jpvQYLaPjsEFotUqCeLCp54p3PkSz    UUID:36f4da38-a028-493a-a855-51b07269e709      SIZE:57

		2023-01-10T07:48:03.789Z        INFO    meta    ipfs/gencar.go:694      FILE:./test/input/test4    CID:QmcA1M4cUFeGZGTwTHAMPZrt6yXyRXkgDoBNeYQ3bhbuJD    UUID:571e4e2b-d50b-4ac2-a89f-07795b684148      SIZE:49

		2023-01-10T07:48:03.789Z        INFO    meta    ipfs/gencar.go:694      FILE:./test/input/test0    CID:QmTvhGdaTkpSWGjQGKcqrLRQqF6LJrfqMv9BWPYJ5tZ9Zp    UUID:94d6a0d0-3e76-45b7-9705-4d829e0e3ca8      SIZE:57

		2023-01-10T07:48:03.790Z        INFO    meta    ipfs/interface.go:133   {"Name":"","Hash":"QmdYHTLyw6WkWERej5HaC4NfxmwUynebEGq4NuVQ7reuGM","Size":0,"Link":[{"Name":"test","Hash":"QmagwmWgQnGTnpqgDEkeCirVTw5igJeZ91WiHVTd3iyxYQ","Size":842,"Link":[{"Name":"input","Hash":"QmPAU4QWdZQgWtVr33uafgNYNezqC57AWJPnSj1VieVAmY","Size":790,"Link":[{"Name":"dir1","Hash":"QmYRd1yyncZUTCg9v12QMQuov83LYLxPrVVEUKtK3XwYst","Size":467,"Link":[{"Name":"dir2","Hash":"QmRNTAX6uQKKsp94qMdcbHBV4RgDNTUM35VD9iT6DA1YSy","Size":276,"Link":[{"Name":"test2e99d2819-09a8-4e53-8158-a48d8154e057","Hash":"QmeRAKJCjykxuU8NTjtWgeX59Zn8xcCp4NLF6jw3dCrAnX","Size":57,"Link":null},{"Name":"test36631aa2a-5e89-4f98-b114-86bf4403f1c2","Hash":"QmcA1M4cUFeGZGTwTHAMPZrt6yXyRXkgDoBNeYQ3bhbuJD","Size":49,"Link":null}]},{"Name":"test136f4da38-a028-493a-a855-51b07269e709","Hash":"QmTGVzSq5v5mzYUt9jpvQYLaPjsEFotUqCeLCp54p3PkSz","Size":57,"Link":null}]},{"Name":"test094d6a0d0-3e76-45b7-9705-4d829e0e3ca8","Hash":"QmTvhGdaTkpSWGjQGKcqrLRQqF6LJrfqMv9BWPYJ5tZ9Zp","Size":57,"Link":null},{"Name":"test4571e4e2b-d50b-4ac2-a89f-07795b684148","Hash":"QmcA1M4cUFeGZGTwTHAMPZrt6yXyRXkgDoBNeYQ3bhbuJD","Size":49,"Link":null}]}]}]}
		2023-01-10T07:48:03.790Z        INFO    meta    meta-car/test.go:63     create car file is:test/output/QmdYHTLyw6WkWERej5HaC4NfxmwUynebEGq4NuVQ7reuGM.car
	*/

}

func genCarFromFilesDemo() {
	outputDir := "./test/output"
	srcFiles := []string{
		"./test/input/test0",
		"./test/input/test4",
		"./test/input/dir1/test1",
		"./test/input/dir1/dir2/test2",
		"./test/input/dir1/dir2/test3",
	}
	sliceSize := 17179869184

	carFileName, err := meta_car.GenerateCarFromFiles(outputDir, srcFiles, int64(sliceSize))
	if err != nil {
		log.GetLog().Error("Create car file error:", err)
		return
	}

	log.GetLog().Info("Create car file is:", carFileName)

	/*
		OUTPUT:
		2023-01-10T07:48:03.790Z        INFO    meta    ipfs/gencar.go:694      FILE:./test/input/dir1/dir2/test3    CID:QmcA1M4cUFeGZGTwTHAMPZrt6yXyRXkgDoBNeYQ3bhbuJD    UUID:      SIZE:49

		2023-01-10T07:48:03.790Z        INFO    meta    ipfs/gencar.go:694      FILE:./test/input/test4    CID:QmcA1M4cUFeGZGTwTHAMPZrt6yXyRXkgDoBNeYQ3bhbuJD    UUID:      SIZE:49

		2023-01-10T07:48:03.790Z        INFO    meta    ipfs/gencar.go:694      FILE:./test/input/dir1/test1    CID:QmTGVzSq5v5mzYUt9jpvQYLaPjsEFotUqCeLCp54p3PkSz    UUID:      SIZE:57

		2023-01-10T07:48:03.791Z        INFO    meta    ipfs/gencar.go:694      FILE:./test/input/dir1/dir2/test2    CID:QmeRAKJCjykxuU8NTjtWgeX59Zn8xcCp4NLF6jw3dCrAnX    UUID:      SIZE:57

		2023-01-10T07:48:03.791Z        INFO    meta    ipfs/gencar.go:694      FILE:./test/input/test0    CID:QmTvhGdaTkpSWGjQGKcqrLRQqF6LJrfqMv9BWPYJ5tZ9Zp    UUID:      SIZE:57

		2023-01-10T07:48:03.792Z        INFO    meta    meta-car/test.go:102    Create car file is:test/output/QmNMkmt5qQMYBhc1b3gUbkF3L7nNvLPKMnaJ4WP5jaYrMu.car
	*/

}

func genCarFromDirDemo() {
	outputDir := "./test/output"
	srcDir := "./test/input/"
	sliceSize := 17179869184

	carFileName, err := meta_car.GenerateCarFromDir(outputDir, srcDir, int64(sliceSize))
	if err != nil {
		log.GetLog().Error("Create car file error:", err)
		return
	}

	log.GetLog().Info("Create car file is:", carFileName)

	/*
		OUTPUT:
		2023-01-10T07:48:03.792Z        INFO    meta    ipfs/gencar.go:694      FILE:./test/input//test4    CID:QmcA1M4cUFeGZGTwTHAMPZrt6yXyRXkgDoBNeYQ3bhbuJD    UUID:      SIZE:49

		2023-01-10T07:48:03.792Z        INFO    meta    ipfs/gencar.go:694      FILE:./test/input//dir1/dir2/test3    CID:QmcA1M4cUFeGZGTwTHAMPZrt6yXyRXkgDoBNeYQ3bhbuJD    UUID:      SIZE:49

		2023-01-10T07:48:03.792Z        INFO    meta    ipfs/gencar.go:694      FILE:./test/input//dir1/test1    CID:QmTGVzSq5v5mzYUt9jpvQYLaPjsEFotUqCeLCp54p3PkSz    UUID:      SIZE:57

		2023-01-10T07:48:03.792Z        INFO    meta    ipfs/gencar.go:694      FILE:./test/input//test0    CID:QmTvhGdaTkpSWGjQGKcqrLRQqF6LJrfqMv9BWPYJ5tZ9Zp    UUID:      SIZE:57

		2023-01-10T07:48:03.792Z        INFO    meta    ipfs/gencar.go:694      FILE:./test/input//dir1/dir2/test2    CID:QmeRAKJCjykxuU8NTjtWgeX59Zn8xcCp4NLF6jw3dCrAnX    UUID:      SIZE:57

		2023-01-10T07:48:03.793Z        INFO    meta    meta-car/test.go:132    Create car file is:test/output/QmNMkmt5qQMYBhc1b3gUbkF3L7nNvLPKMnaJ4WP5jaYrMu.car
	*/

}

func listCarDemo() {
	destCar := "./test/output/QmUabWJFQGr1hWxhLikB9eLjfRZcaoTrQZJYTMP6AnozN7.car"
	infoList, err := meta_car.ListCarFile(destCar)
	if err != nil {
		log.GetLog().Error("List car file info error:", err)
	}

	log.GetLog().Info("Car info:\n", infoList)

	/*
		OUTPUT:
		2023-01-10T07:48:03.794Z        INFO    meta    meta-car/test.go:158    Car info:
		[FILE:test     CID:QmagwmWgQnGTnpqgDEkeCirVTw5igJeZ91WiHVTd3iyxYQ     UUID:     SIZE:842
		 FILE:test/input     CID:QmPAU4QWdZQgWtVr33uafgNYNezqC57AWJPnSj1VieVAmY     UUID:     SIZE:790
		 FILE:test/input/dir1     CID:QmYRd1yyncZUTCg9v12QMQuov83LYLxPrVVEUKtK3XwYst     UUID:     SIZE:467
		 FILE:test/input/dir1/dir2     CID:QmRNTAX6uQKKsp94qMdcbHBV4RgDNTUM35VD9iT6DA1YSy     UUID:     SIZE:276
		 FILE:test/input/dir1/dir2/test2     CID:QmeRAKJCjykxuU8NTjtWgeX59Zn8xcCp4NLF6jw3dCrAnX     UUID:99d2819-09a8-4e53-8158-a48d8154e057     SIZE:57
		 FILE:test/input/dir1/dir2/test3     CID:QmcA1M4cUFeGZGTwTHAMPZrt6yXyRXkgDoBNeYQ3bhbuJD     UUID:631aa2a-5e89-4f98-b114-86bf4403f1c2     SIZE:49
		 FILE:test/input/dir1/test1     CID:QmTGVzSq5v5mzYUt9jpvQYLaPjsEFotUqCeLCp54p3PkSz     UUID:6f4da38-a028-493a-a855-51b07269e709     SIZE:57
		 FILE:test/input/test0     CID:QmTvhGdaTkpSWGjQGKcqrLRQqF6LJrfqMv9BWPYJ5tZ9Zp     UUID:4d6a0d0-3e76-45b7-9705-4d829e0e3ca8     SIZE:57
		 FILE:test/input/test4     CID:QmcA1M4cUFeGZGTwTHAMPZrt6yXyRXkgDoBNeYQ3bhbuJD     UUID:71e4e2b-d50b-4ac2-a89f-07795b684148     SIZE:49
		]
	*/

}

func GetCarRootDemo() {
	destCar := "./test/output/QmUabWJFQGr1hWxhLikB9eLjfRZcaoTrQZJYTMP6AnozN7.car"
	rootCid, err := meta_car.GetCarRoot(destCar)
	if err != nil {
		log.GetLog().Error("List car file info error:", err)
	}

	log.GetLog().Info("Root CID is:", rootCid)

	/*
		OUTPUT:
		2023-01-10T07:48:03.794Z        INFO    meta    meta-car/test.go:184    Root CID is:QmdYHTLyw6WkWERej5HaC4NfxmwUynebEGq4NuVQ7reuGM
	*/

}
