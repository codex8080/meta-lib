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
### CreateCarFile
Create a car file from source files.

## Examples
Here is a example for use CreateCarFile of meta-lib.
```go
package main

import (
    meta_car "github.com/FogMeta/meta-lib"
)

func main () {
    destFile := "./dest.car"
    srcFiles := []string{"./src0.txt", "./src1.txt", "./src2.txt"}
	
    if err := meta_car.CreateCarFile(destFile, srcFiles); err != nil {
        log.GetLog().Error("Test create car file error:", err)
    }

    return
}

```
## Maintainers

## Contribute

PRs are welcome!

## License

Apache-2.0/MIT © Protocol Labs
