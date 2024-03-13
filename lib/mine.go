package lib

import (
	"math/rand"

	"github.com/antal0x11/blockchat/dst"
)

func MineBlock(seed *string, neighboors *dst.Neighboors) string {

	// TODO take into consideration the stake of each node

	hashSum := 0

	for _, _char := range *seed {
		hashSum += int(_char)
	}

	random := rand.New(rand.NewSource(int64(hashSum)))
	selected := random.Intn(len(neighboors.DSNodes))

	return neighboors.DSNodes[uint32(selected)].PublicKey
}
