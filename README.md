meta-lib
==================

## Features

[meta-lib](v1.0.0) features:
* A CLI tools to interact with CAR file. [Usage](https://github.com/FogMeta/meta-lib/blob/main/cmd/meta-car/README.md#usage)
* [Generate CAR](https://github.com/FogMeta/meta-lib#api-documentation) from files or folders. 
* [Get root CID](https://github.com/FogMeta/meta-lib/blob/main/README.md#L71) of a CAR file.
* [List](https://github.com/FogMeta/meta-lib/blob/main/README.md#L85) original file(s) information in the CAR.
* [Restore](https://github.com/FogMeta/meta-lib/blob/main/README.md#L99) the original file(s) in the CAR.

## Install

To install the latest version of `meta-lib` module:
```shell script
go get github.com/FogMeta/meta-lib/
```

## API Documentation

**func [GenerateCarFromFilesWithUuid](https://github.com/FogMeta/meta-lib/blob/main/module/ipfs/interface.go#L119)**
```go
func GenerateCarFromFilesWithUuid(outputDir string, srcFiles []string, uuid []string, sliceSize int64) (carFile string, err error)

PARAMETER:
    --outputDir    directory where CAR file(s) will be generated.
    --srcFiles     file(s) where source file(s) is(are).
    --uuid         uuid(s) that corresponds to srcFiles
    --sliceSize    bytes of each piece (default: 17179869184)

RETURN:
    --carFile      the CAR which generated.
    --err          error information returned.
```
`GenerateCarFromFilesWithUuid` returns the CAR which generated from file(s) and uuid(s) which are specified by the input parameter `srcFiles` `uuid` and limited by `sliceSize` then output CAR to the specified directory `outputDir`.


**func [GenerateCarFromFiles](https://github.com/FogMeta/meta-lib/blob/main/module/ipfs/interface.go#L73)**
```go
func GenerateCarFromFiles(outputDir string, srcFiles []string, sliceSize int64) (carFile string, err error)

PARAMETER:
    --outputDir    directory where CAR file(s) will be generated.
    --srcFiles     file(s) where source file(s) is(are).
    --sliceSize    bytes of each piece (default: 17179869184)

RETURN:
    --carFile      the CAR which generated.
    --err          error information returned.
```
`GenerateCarFromFiles` returns the CAR which generated from the file(s) specified by the input parameter `srcDir` and limited by `sliceSize` then output CAR to the specified directory `outputDir`.


**func [GenerateCarFromDir](https://github.com/FogMeta/meta-lib/blob/main/module/ipfs/interface.go#L96)**
```go
func GenerateCarFromDir(outputDir string, srcDir string, sliceSize int64) (carFile string, err error)

PARAMETER:
    --outputDir    directory where CAR file(s) will be generated.
    --srcDir       folder where source file(s) is(are) in.
    --sliceSize    bytes of each piece (default: 17179869184)

RETURN:
    --carFile      the CAR which generated.
    --err          error information returned.
```
`GenerateCarFromDir` returns the CAR which generated from the folder specified by the input parameter `srcDir` and limited by `sliceSize` then output CAR to the specified directory `outputDir`.


**func [GetCarRoot](https://github.com/FogMeta/meta-lib/blob/main/module/ipfs/interface.go#L55)**
```go
func GetCarRoot(destCar string) (cid string, err error)

PARAMETER:
    --destCar      the dest CAR file which want to get the root CID string.

RETURN:
    --cid          the root CID string of the destCar.
    --err          error information returned.
```
`GetCarRoot` returns the root CIDs of the CAR which is specified by the input parameter `destCar`.


**func [ListCarFile](https://github.com/FogMeta/meta-lib/blob/main/module/ipfs/interface.go#L19)**
```go
func ListCarFile(destCar string) (info []string, err error)

PARAMETER:
    --destCar      the dest CAR file which want to get the root CID string.

RETURN:
    --info         list of FILE/CID/UUID/SIZE string(s).
    --err          error information returned.
```
`ListCarFile` returns list of FILE/CID/UUID/SIZE information in the CAR which is specified by the input parameter `destCar`.


**func [RestoreCar](https://github.com/FogMeta/meta-lib/blob/main/module/ipfs/interface.go#L138)**
```go
func RestoreCar(outputDir string, srcCar string) (err error)

PARAMETER:
    --outputDir    directory where the original file(s) will be generated.
    --srcCar       the source CAR file witch restore from.

RETURN:
    --err          error information returned.
```
`RestoreCar` returns the original file(s) in the CAR which is specified by the input parameter `srcCar`, and output original file(s) to `outputDir` where specified by the parameter.


## Examples
Here are examples for using meta-lib.
* Generate CAR from file(s) and uuids which is(are) specified by the input parameter. [Example](https://github.com/FogMeta/meta-lib/blob/main/cmd/demo-api/main.go#L28)
* Generate CAR from file(s) which is(are) specified by the input parameter. [Example](https://github.com/FogMeta/meta-lib/blob/main/cmd/demo-api/main.go#L56)
* Generate CAR from a folder where source file(s) is(are) in. [Example](https://github.com/FogMeta/meta-lib/blob/main/cmd/demo-api/main.go#L77)
* Get root CID of a CAR. [Example](https://github.com/FogMeta/meta-lib/blob/main/cmd/demo-api/main.go#L103)
* List FILE/CID/UUID/SIZE in a CAR. [Example](https://github.com/FogMeta/meta-lib/blob/main/cmd/demo-api/main.go#L92)

## Contribute

PRs are welcome!

## License

Apache-2.0/MIT © Protocol Labs
