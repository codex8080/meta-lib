package main

import (
	"fmt"
	"github.com/ipfs/go-cid"
	carv2 "github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/index"
	"github.com/multiformats/go-multihash"
	"github.com/urfave/cli/v2"
	"io"
	log "metalib/logs"
	meta_car "metalib/module/ipfs"
	"os"
)

// VerifyCar is a command to check a files validity
func VerifyCar(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("usage: car verify <file.car>")
	}

	// header
	rx, err := carv2.OpenReader(c.Args().First())
	if err != nil {
		return err
	}
	defer rx.Close()
	roots, err := rx.Roots()
	if err != nil {
		return err
	}
	if len(roots) == 0 {
		return fmt.Errorf("no roots listed in car header")
	}
	rootMap := make(map[cid.Cid]struct{})
	for _, r := range roots {
		rootMap[r] = struct{}{}
	}

	if rx.Version == 2 {
		if rx.Header.DataSize == 0 {
			return fmt.Errorf("size of wrapped v1 car listed as '0'")
		}

		flen, err := os.Stat(c.Args().First())
		if err != nil {
			return err
		}
		lengthToIndex := carv2.PragmaSize + carv2.HeaderSize + rx.Header.DataSize
		if uint64(flen.Size()) > lengthToIndex && rx.Header.IndexOffset == 0 {
			return fmt.Errorf("header claims no index, but extra bytes in file beyond data size")
		}
		if rx.Header.DataOffset < carv2.PragmaSize+carv2.HeaderSize {
			return fmt.Errorf("data offset places data within carv2 header")
		}
		if rx.Header.IndexOffset < lengthToIndex {
			return fmt.Errorf("index offset overlaps with data. data ends at %d. index offset of %d", lengthToIndex, rx.Header.IndexOffset)
		}
	}

	// blocks
	fd, err := os.Open(c.Args().First())
	if err != nil {
		return err
	}
	rd, err := carv2.NewBlockReader(fd)
	if err != nil {
		return err
	}

	cidList := make([]cid.Cid, 0)
	for {
		blk, err := rd.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		delete(rootMap, blk.Cid())
		cidList = append(cidList, blk.Cid())
	}

	if len(rootMap) > 0 {
		return fmt.Errorf("header lists root(s) not present as a block: %v", rootMap)
	}

	// index
	if rx.Version == 2 && rx.Header.HasIndex() {
		ir, err := rx.IndexReader()
		if err != nil {
			return err
		}
		idx, err := index.ReadFrom(ir)
		if err != nil {
			return err
		}
		for _, c := range cidList {
			cidHash, err := multihash.Decode(c.Hash())
			if err != nil {
				return err
			}
			if cidHash.Code == multihash.IDENTITY {
				continue
			}
			if err := idx.GetAll(c, func(_ uint64) bool {
				return true
			}); err != nil {
				return fmt.Errorf("could not look up known cid %s in index: %w", c, err)
			}
		}
	}

	return nil
}

func CreateCarFileTest(c *cli.Context) error {
	genCarWithUuidTest()

	return nil
}

func carFileTest() {
	destFile := "./test.car"
	srcFiles := []string{"./dir0/test0.txt", "./dir1/test1.txt", "./dir2/test2.txt"}

	if err := meta_car.CreateCarFile(destFile, srcFiles); err != nil {
		log.GetLog().Error("Test create car file error:", err)
		return
	}

}

func genCarWithUuidTest() {
	outputDir := "/test/output"
	srcFiles := []string{
		"/test/input/test0",
		"/test/input/test4",
		"/test/input/dir1/test1",
		"/test/input/dir1/dir2/test3",
	}
	uuid := []string{
		"uuid-94d6a0d0-3e76-45b7-9705-4d829e0e3ca8",
		"uuid-571e4e2b-d50b-4ac2-a89f-07795b684148",
		"uuid-36f4da38-a028-493a-a855-51b07269e709",
		"uuid-e99d2819-09a8-4e53-8158-a48d8154e057",
	}
	sliceSize := 17179869184

	root, err := meta_car.GenerateCarFileWithUuid(outputDir, srcFiles, uuid, int64(sliceSize))
	if err != nil {
		log.GetLog().Error("Test create car file error:", err)
		return
	}

	log.GetLog().Info("root:", root)

	/*


	 */

}
