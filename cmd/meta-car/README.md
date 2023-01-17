meta-car - The CLI tool
==================

> A CLI to interact with car files

## Usage

```
USAGE:
   meta-car [global options] command [command options] [arguments...]

COMMANDS:
   compile        compile a car file from a debug patch
   create, c      Create a car file
   debug          debug a car file
   detach-index   Detach an index to a detached file
   extract, x     Extract the contents of a car when the car encodes UnixFS data
   filter, f      Filter the CIDs in a car
   get-block, gb  Get a block out of a car
   get-dag, gd    Get a dag out of a car
   index, i       write out the car with an index
   inspect        verifies a car and prints a basic report about its contents
   list, l, ls    List the CIDs in a car
   root           Get the root CID of a car
   verify, v      Verify a CAR is wellformed
   test           test build a car from files
   build          Generate CAR files of the specified size
   restore        Restore files from CAR files
   help, h        Shows a list of commands or help for one command
```

## Install

To install the latest version of `meta-car` module, run:
```shell script
git clone github.com/FogMeta/meta-lib
make all install
```
