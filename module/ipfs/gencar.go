package ipfs

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/codex8080/metalib/logs"
	"github.com/codex8080/metalib/util"
	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	dss "github.com/ipfs/go-datastore/sync"
	bstore "github.com/ipfs/go-ipfs-blockstore"
	chunker "github.com/ipfs/go-ipfs-chunker"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	format "github.com/ipfs/go-ipld-format"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	dag "github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-unixfs"
	"github.com/ipfs/go-unixfs/importer/balanced"
	ihelper "github.com/ipfs/go-unixfs/importer/helpers"
	"github.com/ipld/go-car"
	ipldprime "github.com/ipld/go-ipld-prime"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/traversal/selector"
	"github.com/ipld/go-ipld-prime/traversal/selector/builder"
	"golang.org/x/xerrors"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
)

func doGenerateCar(sliceSize int64, parentPath, targetPath, carDir, graphName string, parallel int, isUuid bool) error {
	var cumuSize int64 = 0
	graphSliceCount := 0
	graphFiles := make([]util.Finfo, 0)
	if sliceSize == 0 {
		return xerrors.Errorf("Unexpected! Slice size has been set as 0")
	}
	if parallel <= 0 {
		return xerrors.Errorf("Unexpected! Parallel has to be greater than 0")
	}
	if parentPath == "" {
		parentPath = targetPath
	}

	args := []string{targetPath}
	sliceTotal := GetGraphCount(args, sliceSize)
	if sliceTotal == 0 {
		log.GetLog().Warn("Empty folder or file!")
		return nil
	}
	files := util.GetFileListAsync(args, isUuid)
	for item := range files {
		fileSize := item.Info.Size()
		switch {
		case cumuSize+fileSize < sliceSize:
			cumuSize += fileSize
			graphFiles = append(graphFiles, item)
		case cumuSize+fileSize == sliceSize:
			cumuSize += fileSize
			graphFiles = append(graphFiles, item)
			// todo build ipld from graphFiles
			BuildIpldGraph(graphFiles, GenGraphName(graphName, graphSliceCount, sliceTotal), parentPath, carDir, parallel)
			fmt.Printf("cumu-size: %d\n", cumuSize)
			// fmt.Printf(GenGraphName(graphName, graphSliceCount, sliceTotal))
			// fmt.Printf("=================\n")
			cumuSize = 0
			graphFiles = make([]util.Finfo, 0)
			graphSliceCount++
		case cumuSize+fileSize > sliceSize:
			fileSliceCount := 0
			// need to split item to fit graph slice
			//
			// first cut
			firstCut := sliceSize - cumuSize
			var seekStart int64 = 0
			var seekEnd int64 = seekStart + firstCut - 1
			fmt.Printf("first cut %d, seek start at %d, end at %d", firstCut, seekStart, seekEnd)
			fmt.Printf("----------------\n")
			graphFiles = append(graphFiles, util.Finfo{
				Path:      item.Path,
				Name:      fmt.Sprintf("%s.%08d", item.Info.Name(), fileSliceCount),
				Info:      item.Info,
				SeekStart: seekStart,
				SeekEnd:   seekEnd,
			})
			fileSliceCount++
			// todo build ipld from graphFiles
			BuildIpldGraph(graphFiles, GenGraphName(graphName, graphSliceCount, sliceTotal), parentPath, carDir, parallel)
			fmt.Printf("cumu-size: %d\n", cumuSize+firstCut)
			// fmt.Printf(GenGraphName(graphName, graphSliceCount, sliceTotal))
			// fmt.Printf("=================\n")
			cumuSize = 0
			graphFiles = make([]util.Finfo, 0)
			graphSliceCount++
			for seekEnd < fileSize-1 {
				seekStart = seekEnd + 1
				seekEnd = seekStart + sliceSize - 1
				if seekEnd >= fileSize-1 {
					seekEnd = fileSize - 1
				}
				fmt.Printf("following cut %d, seek start at %d, end at %d", seekEnd-seekStart+1, seekStart, seekEnd)
				// fmt.Printf("----------------\n")
				cumuSize += seekEnd - seekStart + 1
				graphFiles = append(graphFiles, util.Finfo{
					Path:      item.Path,
					Name:      fmt.Sprintf("%s.%08d", item.Info.Name(), fileSliceCount),
					Info:      item.Info,
					SeekStart: seekStart,
					SeekEnd:   seekEnd,
				})
				fileSliceCount++
				if seekEnd-seekStart == sliceSize-1 {
					// todo build ipld from graphFiles
					BuildIpldGraph(graphFiles, GenGraphName(graphName, graphSliceCount, sliceTotal), parentPath, carDir, parallel)
					fmt.Printf("cumu-size: %d\n", sliceSize)
					// fmt.Printf(GenGraphName(graphName, graphSliceCount, sliceTotal))
					// fmt.Printf("=================\n")
					cumuSize = 0
					graphFiles = make([]util.Finfo, 0)
					graphSliceCount++
				}
			}

		}
	}
	if cumuSize > 0 {
		// todo build ipld from graphFiles
		BuildIpldGraph(graphFiles, GenGraphName(graphName, graphSliceCount, sliceTotal), parentPath, carDir, parallel)
		fmt.Printf("cumu-size: %d\n", cumuSize)
		// fmt.Printf(GenGraphName(graphName, graphSliceCount, sliceTotal))
		// fmt.Printf("=================\n")
	}
	return nil
}

// 1K 1024
const UnixfsLinksPerLevel = 1 << 10

// 1M 1024*1024
const UnixfsChunkSize uint64 = 1 << 20

// file system tree node
type fsNode struct {
	Name string
	Hash string
	Size uint64
	Link []fsNode
}

type FSBuilder struct {
	root *dag.ProtoNode
	ds   ipld.DAGService
}

func NewFSBuilder(root *dag.ProtoNode, ds ipld.DAGService) *FSBuilder {
	return &FSBuilder{root, ds}
}

func (b *FSBuilder) Build() (*fsNode, error) {
	fsn, err := unixfs.FSNodeFromBytes(b.root.Data())
	if err != nil {
		return nil, xerrors.Errorf("input dag is not a unixfs node: %s", err)
	}

	rootn := &fsNode{
		Hash: b.root.Cid().String(),
		Size: fsn.FileSize(),
		Link: []fsNode{},
	}
	if !fsn.IsDir() {
		return rootn, nil
	}
	for _, ln := range b.root.Links() {
		fn, err := b.getNodeByLink(ln)
		if err != nil {
			return nil, err
		}
		rootn.Link = append(rootn.Link, fn)
	}

	return rootn, nil
}

func (b *FSBuilder) getNodeByLink(ln *format.Link) (fn fsNode, err error) {
	ctx := context.Background()
	fn = fsNode{
		Name: ln.Name,
		Hash: ln.Cid.String(),
		Size: ln.Size,
	}
	nd, err := b.ds.Get(ctx, ln.Cid)
	if err != nil {

		log.GetLog().Warn(err)
		return
	}

	nnd, ok := nd.(*dag.ProtoNode)
	if !ok {
		err = xerrors.Errorf("failed to transformed to dag.ProtoNode")
		return
	}
	fsn, err := unixfs.FSNodeFromBytes(nnd.Data())
	if err != nil {
		log.GetLog().Warnf("input dag is not a unixfs node: %s", err)
		return
	}
	if !fsn.IsDir() {
		return
	}
	for _, ln := range nnd.Links() {
		node, err := b.getNodeByLink(ln)
		if err != nil {
			return node, err
		}
		fn.Link = append(fn.Link, node)
	}
	return
}

func GenGraphName(graphName string, sliceCount, sliceTotal int) string {
	if sliceTotal == 1 {
		return fmt.Sprintf("%s.car", graphName)
	}
	return fmt.Sprintf("%s-total-%d-part-%d.car", graphName, sliceTotal, sliceCount+1)
}

func GetGraphCount(args []string, sliceSize int64) int {
	list, err := util.GetFileList(args)
	if err != nil {
		panic(err)
	}
	var totalSize int64 = 0
	for _, path := range list {
		finfo, err := os.Stat(path)
		if err != nil {
			panic(err)
		}
		totalSize += finfo.Size()
	}
	if totalSize == 0 {
		return 0
	}
	count := (totalSize / sliceSize) + 1
	return int(count)
}

func BuildIpldGraph(fileList []util.Finfo, graphName, parentPath, carDir string, parallel int) {
	node, fsDetail, err := buildIpldGraph(fileList, parentPath, carDir, parallel)
	if err != nil {
		log.GetLog().Fatal(err)
		return
	}
	SaveToCsv(carDir, node, graphName, fsDetail)
	//log.GetLog().Info("Build ipld graph result:", "Cid=", node.Cid().String(), " Detail=", fsDetail)
}

func buildIpldGraph(fileList []util.Finfo, parentPath, carDir string, parallel int) (ipld.Node, string, error) {

	ctx := context.Background()

	bs2 := bstore.NewBlockstore(dss.MutexWrap(datastore.NewMapDatastore()))
	dagServ := merkledag.NewDAGService(blockservice.New(bs2, offline.Exchange(bs2)))

	cidBuilder, err := merkledag.PrefixForCidVersion(0)
	if err != nil {
		return nil, "", err
	}
	fileNodeMap := make(map[string]*dag.ProtoNode)
	dirNodeMap := make(map[string]*dag.ProtoNode)

	var rootNode *dag.ProtoNode
	rootNode = unixfs.EmptyDirNode()
	rootNode.SetCidBuilder(cidBuilder)
	var rootKey = "root"
	dirNodeMap[rootKey] = rootNode

	// fmt.Println("************ start to build **************")
	// build file node
	// parallel build
	cpun := runtime.NumCPU()
	if parallel > cpun {
		parallel = cpun
	}
	pchan := make(chan struct{}, parallel)
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}
	for i, item := range fileList {
		wg.Add(1)
		go func(i int, item util.Finfo) {
			defer func() {
				<-pchan
				wg.Done()
			}()
			pchan <- struct{}{}
			fileNode, err := BuildFileNode(item, dagServ, cidBuilder)
			if err != nil {
				log.GetLog().Warn(err)
				return
			}
			fn, ok := fileNode.(*dag.ProtoNode)
			if !ok {
				emsg := "file node should be *dag.ProtoNode"
				log.GetLog().Warn(emsg)
				return
			}
			lock.Lock()
			fileNodeMap[item.Path] = fn
			lock.Unlock()
			// fmt.Println(item.Path)
			stat, _ := fileNode.Stat()
			log.GetLog().Infof("FILE:%s    CID:%s    UUID:uuid-%s      SIZE:%d\n", item.Path, fileNode, item.Uuid, stat.CumulativeSize)
		}(i, item)
	}
	wg.Wait()

	// build dir tree
	for _, item := range fileList {
		// log.GetLog().Info(item.Path)
		// log.Infof("file name: %s, file size: %d, item size: %d, seek-start:%d, seek-end:%d", item.Name, item.Info.Size(), item.SeekEnd-item.SeekStart, item.SeekStart, item.SeekEnd)
		dirStr := path.Dir(item.Path)
		parentPath = path.Clean(parentPath)
		// when parent path equal target path, and the parent path is also a file path
		if parentPath == path.Clean(item.Path) {
			dirStr = ""
		} else if parentPath != "" && strings.HasPrefix(dirStr, parentPath) {
			dirStr = dirStr[len(parentPath):]
		}

		if strings.HasPrefix(dirStr, "/") {
			dirStr = dirStr[1:]
		}
		var dirList []string
		if dirStr == "" {
			dirList = []string{}
		} else {
			dirList = strings.Split(dirStr, "/")
		}
		fileNode, ok := fileNodeMap[item.Path]
		if !ok {
			panic("unexpected, missing file node")
		}
		if len(dirList) == 0 {
			dirNodeMap[rootKey].AddNodeLink(item.Name+"-"+item.Uuid, fileNode)
			continue
		}
		//log.Info(item.Path)
		log.GetLog().Info(dirList)
		i := len(dirList) - 1
		for ; i >= 0; i-- {
			// get dirNodeMap by index
			var ok bool
			var dirNode *dag.ProtoNode
			var parentNode *dag.ProtoNode
			var parentKey string
			dir := dirList[i]
			dirKey := getDirKey(dirList, i)
			log.GetLog().Info(dirList)
			log.GetLog().Infof("dirKey: %s", dirKey)
			dirNode, ok = dirNodeMap[dirKey]
			if !ok {
				dirNode = unixfs.EmptyDirNode()
				dirNode.SetCidBuilder(cidBuilder)
				dirNodeMap[dirKey] = dirNode
			}
			// add file node to its nearest parent node
			if i == len(dirList)-1 {
				dirNode.AddNodeLink(item.Name+"-"+item.Uuid, fileNode)
			}
			if i == 0 {
				parentKey = rootKey
			} else {
				parentKey = getDirKey(dirList, i-1)
			}
			log.GetLog().Infof("parentKey: %s", parentKey)
			parentNode, ok = dirNodeMap[parentKey]
			if !ok {
				parentNode = unixfs.EmptyDirNode()
				parentNode.SetCidBuilder(cidBuilder)
				dirNodeMap[parentKey] = parentNode
			}
			if isLinked(parentNode, dir) {
				parentNode, err = parentNode.UpdateNodeLink(dir, dirNode)
				if err != nil {
					return nil, "", err
				}
				dirNodeMap[parentKey] = parentNode
			} else {
				parentNode.AddNodeLink(dir, dirNode)
			}
		}
	}

	for _, node := range dirNodeMap {
		// fmt.Printf("add node to store: %v\n", node)
		// fmt.Printf("key: %s, links: %v\n", key, len(node.Links()))
		dagServ.Add(ctx, node)
	}

	rootNode = dirNodeMap[rootKey]
	//fmt.Printf("root node cid: %s\n", rootNode.Cid())
	// log.GetLog().Infof("start to generate car for %s", rootNode.Cid())
	// genCarStartTime := time.Now()
	//car
	carF, err := os.Create(path.Join(carDir, rootNode.Cid().String()+".car"))
	if err != nil {
		return nil, "", err
	}
	defer carF.Close()
	selector := allSelector()
	sc := car.NewSelectiveCar(ctx, bs2, []car.Dag{{Root: rootNode.Cid(), Selector: selector}})
	err = sc.Write(carF)
	// cario := cario.NewCarIO()
	// err = cario.WriteCar(context.Background(), bs2, rootNode.Cid(), selector, carF)
	if err != nil {
		return nil, "", err
	}
	//log.GetLog().Infof("generate car file completed, time elapsed: %s", time.Now().Sub(genCarStartTime))

	fsBuilder := NewFSBuilder(rootNode, dagServ)
	fsNode, err := fsBuilder.Build()
	if err != nil {
		return nil, "", err
	}
	fsNodeBytes, err := json.Marshal(fsNode)
	if err != nil {
		return nil, "", err
	}
	// log.GetLog().Info("File Node Map:", fileNodeMap)
	// log.GetLog().Info("Dir  Node Map:", dirNodeMap)
	// fmt.Println("++++++++++++ finished to build +++++++++++++")
	return rootNode, fmt.Sprintf("%s", fsNodeBytes), nil
}

func allSelector() ipldprime.Node {
	ssb := builder.NewSelectorSpecBuilder(basicnode.Prototype.Any)
	return ssb.ExploreRecursive(selector.RecursionLimitNone(),
		ssb.ExploreAll(ssb.ExploreRecursiveEdge())).
		Node()
}

func getDirKey(dirList []string, i int) (key string) {
	for j := 0; j <= i; j++ {
		key += dirList[j]
		if j < i {
			key += "."
		}
	}
	return
}

func isLinked(node *dag.ProtoNode, name string) bool {
	for _, lk := range node.Links() {
		if lk.Name == name {
			return true
		}
	}
	return false
}

type fileSlice struct {
	r        *os.File
	offset   int64
	start    int64
	end      int64
	fileSize int64
}

func (fs *fileSlice) Read(p []byte) (n int, err error) {
	if fs.end == 0 {
		fs.end = fs.fileSize - 1
	}
	if fs.offset == 0 && fs.start > 0 {
		_, err = fs.r.Seek(fs.start, 0)
		if err != nil {
			log.GetLog().Warn(err)
			return 0, err
		}
		fs.offset = fs.start
	}
	//fmt.Printf("offset: %d, end: %d, start: %d, size: %d\n", fs.offset, fs.end, fs.start, fs.fileSize)
	if fs.end-fs.offset+1 == 0 {
		return 0, io.EOF
	}
	if fs.end-fs.offset+1 < 0 {
		return 0, xerrors.Errorf("read data out bound of the slice")
	}
	plen := len(p)
	leftLen := fs.end - fs.offset + 1
	if leftLen > int64(plen) {
		n, err = fs.r.Read(p)
		if err != nil {
			log.GetLog().Warn(err)
			return
		}
		//fmt.Printf("read num: %d\n", n)
		fs.offset += int64(n)
		return
	}
	b := make([]byte, leftLen)
	n, err = fs.r.Read(b)
	if err != nil {
		return
	}
	//fmt.Printf("read num: %d\n", n)
	fs.offset += int64(n)

	return copy(p, b), io.EOF
}

func BuildFileNode(item util.Finfo, bufDs ipld.DAGService, cidBuilder cid.Builder) (node ipld.Node, err error) {
	var r io.Reader
	f, err := os.Open(item.Path)
	if err != nil {
		return nil, err
	}
	r = f

	// read all data of item
	if item.SeekStart > 0 || item.SeekEnd > 0 {
		r = &fileSlice{
			r:        f,
			start:    item.SeekStart,
			end:      item.SeekEnd,
			fileSize: item.Info.Size(),
		}
	}

	params := ihelper.DagBuilderParams{
		Maxlinks:   UnixfsLinksPerLevel,
		RawLeaves:  false,
		CidBuilder: cidBuilder,
		Dagserv:    bufDs,
		NoCopy:     false,
	}
	db, err := params.New(chunker.NewSizeSplitter(r, int64(UnixfsChunkSize)))
	if err != nil {
		return nil, err
	}
	node, err = balanced.Layout(db)
	if err != nil {
		return nil, err
	}
	return
}

func SaveToCsv(carDir string, node ipld.Node, graphName, fsDetail string) {
	// Add node inof to manifest.csv
	manifestPath := path.Join(carDir, "manifest.csv")
	_, err := os.Stat(manifestPath)
	if err != nil && !os.IsNotExist(err) {
		log.GetLog().Fatal(err)
	}
	var isCreateAction bool
	if err != nil && os.IsNotExist(err) {
		isCreateAction = true
	}
	f, err := os.OpenFile(manifestPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.GetLog().Fatal(err)
	}
	defer f.Close()
	if isCreateAction {
		if _, err := f.Write([]byte("playload_cid,filename,detail\n")); err != nil {
			log.GetLog().Fatal(err)
		}
	}
	if _, err := f.Write([]byte(fmt.Sprintf("%s,%s,%s\n", node.Cid(), graphName, fsDetail))); err != nil {
		log.GetLog().Fatal(err)
	}
}

func checkFiles(srcFiles []string, sliceSize int64) bool {
	var totalSize int64 = 0
	for _, path := range srcFiles {
		finfo, err := os.Stat(path)
		if err != nil {
			log.GetLog().Fatal("check slice size error:", err)
			return false
		}
		//TODO: absolute path
		totalSize += finfo.Size()
	}

	if totalSize > sliceSize {
		return false
	}

	return true
}

func getFileInfoWithUuidAsync(srcFiles []string, uuidStr []string) chan util.Finfo {
	fichan := make(chan util.Finfo, 0)
	go func() {
		defer close(fichan)
		for index, path := range srcFiles {
			finfo, err := os.Stat(path)
			if err != nil {
				log.GetLog().Warn(err)
				continue
			}
			fichan <- util.Finfo{
				Path: path,
				Name: finfo.Name(),
				Uuid: uuidStr[index],
				Info: finfo,
			}
		}
	}()

	return fichan
}

func doGenerateCarFrom(outputPath string, srcFiles []string) (string, string, error) {

	graphFiles := make([]util.Finfo, 0)
	files := util.GetFileListAsync(srcFiles, false)
	for item := range files {
		graphFiles = append(graphFiles, item)
	}

	carName, detail, err := buildGraph(graphFiles, outputPath)
	if err != nil {
		return "", "", err
	}

	return carName, detail, nil
}

func doGenerateCarWithUuid(outputPath string, srcFiles []string, uuidStr []string) (string, string, error) {
	graphFiles := make([]util.Finfo, 0)
	files := getFileInfoWithUuidAsync(srcFiles, uuidStr)
	for item := range files {
		graphFiles = append(graphFiles, item)
	}

	return buildGraph(graphFiles, outputPath)
}

func buildGraph(fileList []util.Finfo, outputPath string) (string, string, error) {

	parentPath := "/"
	ctx := context.Background()

	bs2 := bstore.NewBlockstore(dss.MutexWrap(datastore.NewMapDatastore()))
	dagServ := merkledag.NewDAGService(blockservice.New(bs2, offline.Exchange(bs2)))

	cidBuilder, err := merkledag.PrefixForCidVersion(0)
	if err != nil {
		return "", "", err
	}
	fileNodeMap := make(map[string]*dag.ProtoNode)
	dirNodeMap := make(map[string]*dag.ProtoNode)

	var rootNode *dag.ProtoNode
	rootNode = unixfs.EmptyDirNode()
	rootNode.SetCidBuilder(cidBuilder)
	var rootKey = "root"
	dirNodeMap[rootKey] = rootNode

	parallel := runtime.NumCPU()
	pchan := make(chan struct{}, parallel)
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}
	for i, item := range fileList {
		wg.Add(1)
		go func(i int, item util.Finfo) {
			defer func() {
				<-pchan
				wg.Done()
			}()
			pchan <- struct{}{}
			fileNode, err := BuildFileNode(item, dagServ, cidBuilder)
			if err != nil {
				log.GetLog().Warn(err)
				return
			}
			fn, ok := fileNode.(*dag.ProtoNode)
			if !ok {
				emsg := "file node should be *dag.ProtoNode"
				log.GetLog().Warn(emsg)
				return
			}
			lock.Lock()
			fileNodeMap[item.Path] = fn
			lock.Unlock()
			stat, _ := fileNode.Stat()
			log.GetLog().Infof("FILE:%s    CID:%s    UUID:%s      SIZE:%d\n", item.Path, fileNode, item.Uuid, stat.CumulativeSize)
		}(i, item)
	}
	wg.Wait()

	// build dir tree
	for _, item := range fileList {
		// log.GetLog().Info(item.Path)
		// log.Infof("file name: %s, file size: %d, item size: %d, seek-start:%d, seek-end:%d", item.Name, item.Info.Size(), item.SeekEnd-item.SeekStart, item.SeekStart, item.SeekEnd)
		dirStr := path.Dir(item.Path)
		parentPath = path.Clean(parentPath)
		// when parent path equal target path, and the parent path is also a file path
		if parentPath == path.Clean(item.Path) {
			dirStr = ""
		} else if parentPath != "" && strings.HasPrefix(dirStr, parentPath) {
			dirStr = dirStr[len(parentPath):]
		}

		if strings.HasPrefix(dirStr, "/") {
			dirStr = dirStr[1:]
		}
		var dirList []string
		if dirStr == "" {
			dirList = []string{}
		} else {
			dirList = strings.Split(dirStr, "/")
		}
		fileNode, ok := fileNodeMap[item.Path]
		if !ok {
			panic("unexpected, missing file node")
		}
		if len(dirList) == 0 {
			dirNodeMap[rootKey].AddNodeLink(item.Name+item.Uuid, fileNode)
			continue
		}
		//log.Info(item.Path)
		//log.GetLog().Info(dirList)
		i := len(dirList) - 1
		for ; i >= 0; i-- {
			// get dirNodeMap by index
			var ok bool
			var dirNode *dag.ProtoNode
			var parentNode *dag.ProtoNode
			var parentKey string
			dir := dirList[i]
			dirKey := getDirKey(dirList, i)
			//log.GetLog().Info(dirList)
			//log.GetLog().Infof("dirKey: %s", dirKey)
			dirNode, ok = dirNodeMap[dirKey]
			if !ok {
				dirNode = unixfs.EmptyDirNode()
				dirNode.SetCidBuilder(cidBuilder)
				dirNodeMap[dirKey] = dirNode
			}
			// add file node to its nearest parent node
			if i == len(dirList)-1 {
				dirNode.AddNodeLink(item.Name+item.Uuid, fileNode)
			}
			if i == 0 {
				parentKey = rootKey
			} else {
				parentKey = getDirKey(dirList, i-1)
			}
			//log.GetLog().Infof("parentKey: %s", parentKey)
			parentNode, ok = dirNodeMap[parentKey]
			if !ok {
				parentNode = unixfs.EmptyDirNode()
				parentNode.SetCidBuilder(cidBuilder)
				dirNodeMap[parentKey] = parentNode
			}
			if isLinked(parentNode, dir) {
				parentNode, err = parentNode.UpdateNodeLink(dir, dirNode)
				if err != nil {
					return "", "", err
				}
				dirNodeMap[parentKey] = parentNode
			} else {
				parentNode.AddNodeLink(dir, dirNode)
			}
		}
	}

	for _, node := range dirNodeMap {
		// fmt.Printf("add node to store: %v\n", node)
		// fmt.Printf("key: %s, links: %v\n", key, len(node.Links()))
		dagServ.Add(ctx, node)
	}

	rootNode = dirNodeMap[rootKey]
	carFileName := path.Join(outputPath, rootNode.Cid().String()+".car")
	carF, err := os.Create(carFileName)
	if err != nil {
		return "", "", err
	}
	defer carF.Close()
	selector := allSelector()
	sc := car.NewSelectiveCar(ctx, bs2, []car.Dag{{Root: rootNode.Cid(), Selector: selector}})
	err = sc.Write(carF)

	if err != nil {
		return "", "", err
	}
	//log.GetLog().Infof("generate car file completed, time elapsed: %s", time.Now().Sub(genCarStartTime))

	fsBuilder := NewFSBuilder(rootNode, dagServ)
	fsNode, err := fsBuilder.Build()
	if err != nil {
		return "", "", err
	}
	fsNodeBytes, err := json.Marshal(fsNode)
	if err != nil {
		return "", "", err
	}
	detail := fmt.Sprintf("%s", fsNodeBytes)

	return carFileName, detail, nil
}
