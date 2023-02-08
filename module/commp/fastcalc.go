package commp

import (
	"fmt"
	"github.com/FogMeta/meta-lib/module/commp/calpiece"
	"github.com/FogMeta/meta-lib/module/commp/calunseal"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	"io"
	"os"
)

func FastCommP(carFileName string) (pieceCid cid.Cid, pieceSize abi.PaddedPieceSize, err error) {
	carFile, err := os.Open(carFileName)
	if err != nil {
		return pieceCid, pieceSize, err
	}
	defer carFile.Close()

	carlData, err := io.ReadAll(carFile)
	if err != nil {
		return pieceCid, pieceSize, err
	}

	pieceSize, unsealData, err := calunseal.NewUnsealData(abi.PaddedPieceSize(32<<30), carlData)
	if err != nil {
		panic(err)
	}
	genFactory, err := calpiece.NewGenPieceFactory(int(pieceSize), unsealData.Fr32Data, 1.2)
	if err != nil {
		panic(err)
	}
	defer genFactory.Close()
	pieceCid, err = genFactory.Sum()
	if err != nil {
		panic(err)
	}
	fmt.Println("pieceCid:", pieceCid)

	return pieceCid, pieceSize, nil
}
