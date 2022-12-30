# meta-lib

## Command Car
To build and run:
```shell script
make car
./bin/car --help
```

Usage

```
USAGE:
   car [global options] command [command options] [arguments...]

COMMANDS:
   create, c    Create a car file
   append, c    Append files to a car file
   list, l, ls  List the CIDs in a car
   root         Get the root CID of a car
   verify, v    Verify a CAR is wellformed
   test         test build a car from files
   chunk        Generate CAR files of the specified size
   help, h      Shows a list of commands or help for one command

```