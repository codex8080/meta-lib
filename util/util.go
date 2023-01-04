package util

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/pborman/uuid"
	"io"
	"io/ioutil"
	log "metalib/logs"
	"os"
	"strings"

	cid "github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

// MaxAllowedSectionSize dictates the maximum number of bytes that a CARv1 header
// or section is allowed to occupy without causing a decode to error.
// This cannot be supplied as an option, only adjusted as a global. You should
// use v2#NewReader instead since it allows for options to be passed in.
var MaxAllowedSectionSize uint = 32 << 20 // 32MiB

var cidv0Pref = []byte{0x12, 0x20}

type BytesReader interface {
	io.Reader
	io.ByteReader
}

// Deprecated: ReadCid shouldn't be used directly, use CidFromReader from go-cid
func ReadCid(buf []byte) (cid.Cid, int, error) {
	if len(buf) >= 2 && bytes.Equal(buf[:2], cidv0Pref) {
		i := 34
		if len(buf) < i {
			i = len(buf)
		}
		c, err := cid.Cast(buf[:i])
		return c, i, err
	}

	br := bytes.NewReader(buf)

	// assume cidv1
	vers, err := binary.ReadUvarint(br)
	if err != nil {
		return cid.Cid{}, 0, err
	}

	// TODO: the go-cid package allows version 0 here as well
	if vers != 1 {
		return cid.Cid{}, 0, fmt.Errorf("invalid cid version number")
	}

	codec, err := binary.ReadUvarint(br)
	if err != nil {
		return cid.Cid{}, 0, err
	}

	mhr := mh.NewReader(br)
	h, err := mhr.ReadMultihash()
	if err != nil {
		return cid.Cid{}, 0, err
	}

	return cid.NewCidV1(codec, h), len(buf) - br.Len(), nil
}

func ReadNode(br *bufio.Reader) (cid.Cid, []byte, error) {
	data, err := LdRead(br)
	if err != nil {
		return cid.Cid{}, nil, err
	}

	n, c, err := cid.CidFromReader(bytes.NewReader(data))
	if err != nil {
		return cid.Cid{}, nil, err
	}

	return c, data[n:], nil
}

func LdWrite(w io.Writer, d ...[]byte) error {
	var sum uint64
	for _, s := range d {
		sum += uint64(len(s))
	}

	buf := make([]byte, 8)
	n := binary.PutUvarint(buf, sum)
	_, err := w.Write(buf[:n])
	if err != nil {
		return err
	}

	for _, s := range d {
		_, err = w.Write(s)
		if err != nil {
			return err
		}
	}

	return nil
}

func LdSize(d ...[]byte) uint64 {
	var sum uint64
	for _, s := range d {
		sum += uint64(len(s))
	}
	buf := make([]byte, 8)
	n := binary.PutUvarint(buf, sum)
	return sum + uint64(n)
}

func LdRead(r *bufio.Reader) ([]byte, error) {
	if _, err := r.Peek(1); err != nil { // no more blocks, likely clean io.EOF
		return nil, err
	}

	l, err := binary.ReadUvarint(r)
	if err != nil {
		if err == io.EOF {
			return nil, io.ErrUnexpectedEOF // don't silently pretend this is a clean EOF
		}
		return nil, err
	}

	if l > uint64(MaxAllowedSectionSize) { // Don't OOM
		return nil, errors.New("malformed car; header is bigger than util.MaxAllowedSectionSize")
	}

	buf := make([]byte, l)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	return buf, nil
}

func ExistDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

type Finfo struct {
	Path      string
	Name      string
	Uuid      string
	Info      os.FileInfo
	SeekStart int64
	SeekEnd   int64
}

func GetFileListAsync(args []string) chan Finfo {
	fichan := make(chan Finfo, 0)
	go func() {
		defer close(fichan)
		for _, path := range args {
			finfo, err := os.Stat(path)
			if err != nil {
				log.GetLog().Warn(err)
				return
			}
			//Ignore hidden directories
			if strings.HasPrefix(finfo.Name(), ".") {
				continue
			}
			if finfo.IsDir() {
				files, err := ioutil.ReadDir(path)
				if err != nil {
					log.GetLog().Warn(err)
					return
				}
				templist := make([]string, 0)
				for _, n := range files {
					templist = append(templist, fmt.Sprintf("%s/%s", path, n.Name()))
				}
				embededChan := GetFileListAsync(templist)
				if err != nil {
					log.GetLog().Warn(err)
					return
				}

				for item := range embededChan {
					fichan <- item
				}
			} else {
				fichan <- Finfo{
					Path: path,
					Name: finfo.Name(),
					Uuid: uuid.New(),
					Info: finfo,
				}
			}
		}
	}()

	return fichan
}

func GetFileList(args []string) (fileList []string, err error) {
	fileList = make([]string, 0)
	for _, path := range args {
		finfo, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(finfo.Name(), ".") {
			continue
		}
		if finfo.IsDir() {
			files, err := ioutil.ReadDir(path)
			if err != nil {
				return nil, err
			}
			templist := make([]string, 0)
			for _, n := range files {
				templist = append(templist, fmt.Sprintf("%s/%s", path, n.Name()))
			}
			list, err := GetFileList(templist)
			if err != nil {
				return nil, err
			}
			fileList = append(fileList, list...)
		} else {
			fileList = append(fileList, path)
		}
	}

	return
}
