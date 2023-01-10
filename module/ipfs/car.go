package ipfs

import (
	"bytes"
	"context"
	"fmt"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-unixfsnode/data"
	"github.com/ipfs/go-unixfsnode/data/builder"
	"github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/blockstore"
	dagpb "github.com/ipld/go-codec-dagpb"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multicodec"
	"github.com/multiformats/go-multihash"
	"io"
	"path"
)

func CreateCarFile(destCar string, srcFiles []string) error {
	var err error

	// make a cid with the right length that we eventually will patch with the root.
	hasher, err := multihash.GetHasher(multihash.SHA2_256)
	if err != nil {
		return err
	}
	digest := hasher.Sum([]byte{})
	hash, err := multihash.Encode(digest, multihash.SHA2_256)
	if err != nil {
		return err
	}
	proxyRoot := cid.NewCidV1(uint64(multicodec.DagPb), hash)

	options := []car.Option{}
	carVer := 2
	switch carVer {
	case 1:
		options = []car.Option{blockstore.WriteAsCarV1(true)}
	case 2:
		// already the default
	default:
		return fmt.Errorf("invalid CAR version")
	}

	cdest, err := blockstore.OpenReadWrite(destCar, []cid.Cid{proxyRoot}, options...)
	if err != nil {
		return err
	}

	// Write the unixfs blocks into the store.
	root, err := writeFiles(cdest, srcFiles)
	if err != nil {
		return err
	}

	if err := cdest.Finalize(); err != nil {
		return err
	}
	// re-open/finalize with the final root.
	return car.ReplaceRootsInFile(destCar, []cid.Cid{root})
}

func writeFiles(bs *blockstore.ReadWrite, srcFiles []string) (cid.Cid, error) {
	ctx := context.Background()

	ls := cidlink.DefaultLinkSystem()
	ls.TrustedStorage = true
	ls.StorageReadOpener = func(_ ipld.LinkContext, l ipld.Link) (io.Reader, error) {
		cl, ok := l.(cidlink.Link)
		if !ok {
			return nil, fmt.Errorf("not a cidlink")
		}
		blk, err := bs.Get(ctx, cl.Cid)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(blk.RawData()), nil
	}

	ls.StorageWriteOpener = func(_ ipld.LinkContext) (io.Writer, ipld.BlockWriteCommitter, error) {
		buf := bytes.NewBuffer(nil)
		return buf, func(l ipld.Link) error {
			cl, ok := l.(cidlink.Link)
			if !ok {
				return fmt.Errorf("not a cidlink")
			}
			blk, err := blocks.NewBlockWithCid(buf.Bytes(), cl.Cid)
			if err != nil {
				return err
			}
			bs.Put(ctx, blk)
			return nil
		}, nil
	}

	topLevel := make([]dagpb.PBLink, 0, len(srcFiles))
	for _, p := range srcFiles {
		l, size, err := builder.BuildUnixFSRecursive(p, &ls)
		if err != nil {
			return cid.Undef, err
		}
		name := path.Base(p)
		entry, err := builder.BuildUnixFSDirectoryEntry(name, int64(size), l)
		if err != nil {
			return cid.Undef, err
		}
		topLevel = append(topLevel, entry)
	}

	// make a directory for the file(s).

	root, _, err := builder.BuildUnixFSDirectory(topLevel, &ls)
	if err != nil {
		return cid.Undef, nil
	}
	rcl, ok := root.(cidlink.Link)
	if !ok {
		return cid.Undef, fmt.Errorf("could not interpret %s", root)
	}

	return rcl.Cid, nil
}

func ExtractCarFile(destDir string, srcCar string) error {
	return nil
}

func printLinksNode(prefix string, node cid.Cid, ls *ipld.LinkSystem, infoList *[]string) error {
	// it might be a raw file (bytes) node. if so, not actually an error.
	if node.Prefix().Codec == cid.Raw {
		return nil
	}

	pbn, err := ls.Load(ipld.LinkContext{}, cidlink.Link{Cid: node}, dagpb.Type.PBNode)
	if err != nil {
		return err
	}

	pbnode := pbn.(dagpb.PBNode)

	ufd, err := data.DecodeUnixFSData(pbnode.Data.Must().Bytes())
	if err != nil {
		return err
	}
	if ufd.FieldDataType().Int() == data.Data_Directory {
		i := pbnode.Links.Iterator()
		for !i.Done() {
			_, l := i.Next()
			name := path.Join(prefix, l.Name.Must().String())
			size := l.Tsize.Must().Int()
			nameLen := len(name)
			uuid := ""
			uuidLen := len("ce547c40-acf9-11e6-80f5-76304dec7eb7")
			// TODO: split uuid string and check it
			if nameLen > uuidLen {
				uuid = name[nameLen-uuidLen+1:]
				name = name[:nameLen-uuidLen]
			}

			// recurse into the file/directory
			cl, err := l.Hash.AsLink()
			if err != nil {
				return err
			}
			if cidl, ok := cl.(cidlink.Link); ok {
				info := ""
				fmt.Printf(info, "FILE:%s     CID:%s     UUID:%s     SIZE:%d\n", name, cidl.Cid, uuid, size)
				*infoList = append(*infoList, info)
				if err := printLinksNode(name, cidl.Cid, ls, infoList); err != nil {
					return err
				}
			}

		}
	} else {
		// file, file chunk, symlink, other un-named entities.
		return nil
	}

	return nil
}
