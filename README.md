meta-lib
==================

> A library to interact with merkledags stored as a single file

This is a Go implementation of the [CAR specifications](https://ipld.io/specs/transport/car/), both [CARv1](https://ipld.io/specs/transport/car/carv1/) and [CARv2](https://ipld.io/specs/transport/car/carv2/).

Note that there are two major module versions:

* [`go-car/v2`](v2/) is geared towards reading and writing CARv2 files, and also
  supports consuming CARv1 files and using CAR files as an IPFS blockstore.
* `go-car` v0, in the root directory, just supports reading and writing CARv1 files.

Most users should use v2, especially for new software, since the v2 API transparently supports both CAR formats.

## Features

[CARv2](v2) features:
* [Generate index](https://pkg.go.dev/github.com/ipld/go-car/v2#GenerateIndex) from an existing CARv1 file
* [Wrap](https://pkg.go.dev/github.com/ipld/go-car/v2#WrapV1) CARv1 files into a CARv2 with automatic index generation.
* Random-access to blocks in a CAR file given their CID via [Read-Only blockstore](https://pkg.go.dev/github.com/ipld/go-car/v2/blockstore#NewReadOnly) API, with transparent support for both CARv1 and CARv2
* Write CARv2 files via [Read-Write blockstore](https://pkg.go.dev/github.com/ipld/go-car/v2/blockstore#OpenReadWrite) API, with support for appending blocks to an existing CARv2 file, and resumption from a partially written CARv2 files.
* Individual access to [inner CARv1 data payload]((https://pkg.go.dev/github.com/ipld/go-car/v2#Reader.DataReader)) and [index]((https://pkg.go.dev/github.com/ipld/go-car/v2#Reader.IndexReader)) of a CARv2 file via the `Reader` API.

## Install

To install the latest version of `meta-lib` module, run:
```shell script
go get github.com/FogMeta/meta-lib/
```

## API Documentation
See docs on [pkg.go.dev](https://pkg.go.dev/github.com/ipld/go-car).

## Examples
Here is a example for using meta-lib.
```go
package main

import (
  log "github.com/FogMeta/meta-lib/logs"
  meta_car "github.com/FogMeta/meta-lib/module/ipfs"
)

func main() {
    genCarWithUuidDemo()

    genCarFromFilesDemo()

    genCarFromDirDemo()

    return
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
      log.GetLog().Error("Test create CAR file error:", err)
      return
    }
  
    log.GetLog().Info("create CAR file is:", carFileName)

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
      log.GetLog().Error("Create CAR file error:", err)
      return
    }
  
    log.GetLog().Info("Create CAR file is:", carFileName)

}

func genCarFromDirDemo() {
    outputDir := "./test/output"
    srcDir := "./test/input/"
    sliceSize := 17179869184
  
    carFileName, err := meta_car.GenerateCarFromDir(outputDir, srcDir, int64(sliceSize))
    if err != nil {
      log.GetLog().Error("Create CAR file error:", err)
      return
    }
  
    log.GetLog().Info("Create CAR file is:", carFileName)

}
```
## Maintainers

## Contribute

PRs are welcome!

## License

Apache-2.0/MIT © Protocol Labs
