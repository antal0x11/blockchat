package lib

import (
	"math/rand"
	"sort"

	"github.com/antal0x11/blockchat/dst"
)

func MineBlock(seed *string, neighboors *dst.Neighboors) string {

	hashSum := 0

	for _, _char := range *seed {
		hashSum += int(_char)
	}

	random := rand.New(rand.NewSource(int64(hashSum)))

	// Each node will have n amount of space allocated in the lottery
	// slice based on the (int) amount of stake it has.
	// The random number will point the position that each node owns
	// with its corresponding id in the lottery slice.
	lottery := make([]int, 0)
	keysID := make([]int, 0, len(neighboors.DSNodes))
	for key := range neighboors.DSNodes {
		keysID = append(keysID, int(key))
	}

	sort.Ints(keysID)

	for key := range keysID {
		for range uint(neighboors.DSNodes[uint32(key)].Stake) {
			lottery = append(lottery, key)
		}
	}

	selected := random.Intn(len(lottery))

	return neighboors.DSNodes[uint32(lottery[selected])].PublicKey
}
