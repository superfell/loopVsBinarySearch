package loopVsBinarySearch

import (
	"bytes"
	"math/bits"
)

func (n *nodeLoop) getIndexByte(k byte) interface{} {
	idx := bytes.IndexByte(n.key[:n.count], k)
	if idx < 0 {
		return nil
	}
	return n.val[idx]
}

//go:generate go run asm/asm.go -pkg loopVsBinarySearch -out lookup_amd64.s -stubs "lookup_stub_amd64.go"
func (n *nodeLoop) getLookupAsm(k byte) interface{} {
	idx := Lookup(k, &n.key)
	if idx >= 0 {
		return n.val[idx]
	}
	return nil
}

func (n *nodeLoop) getUnrolledLoop(k byte) interface{} {
	switch n.count {
	case 16:
		if n.key[15] == k {
			return n.val[15]
		}
		fallthrough
	case 15:
		if n.key[14] == k {
			return n.val[14]
		}
		fallthrough
	case 14:
		if n.key[13] == k {
			return n.val[13]
		}
		fallthrough
	case 13:
		if n.key[12] == k {
			return n.val[12]
		}
		fallthrough
	case 12:
		if n.key[11] == k {
			return n.val[11]
		}
		fallthrough
	case 11:
		if n.key[10] == k {
			return n.val[10]
		}
		fallthrough
	case 10:
		if n.key[9] == k {
			return n.val[9]
		}
		fallthrough
	case 9:
		if n.key[8] == k {
			return n.val[8]
		}
		fallthrough
	case 8:
		if n.key[7] == k {
			return n.val[7]
		}
		fallthrough
	case 7:
		if n.key[6] == k {
			return n.val[6]
		}
		fallthrough
	case 6:
		if n.key[5] == k {
			return n.val[5]
		}
		fallthrough
	case 5:
		if n.key[4] == k {
			return n.val[4]
		}
		fallthrough
	case 4:
		if n.key[3] == k {
			return n.val[3]
		}
		fallthrough
	case 3:
		if n.key[2] == k {
			return n.val[2]
		}
		fallthrough
	case 2:
		if n.key[1] == k {
			return n.val[1]
		}
		fallthrough
	case 1:
		if n.key[0] == k {
			return n.val[0]
		}
		fallthrough
	default:
		return nil
	}
}

type nodeMasks struct {
	keys1 uint64
	keys2 uint64
	vals  [16]interface{}
	count int
}

const b1 = 0x00000000000000FF
const b2 = 0x000000000000FF00
const b3 = 0x0000000000FF0000
const b4 = 0x00000000FF000000
const b5 = 0x000000FF00000000
const b6 = 0x0000FF0000000000
const b7 = 0x00FF000000000000
const b8 = 0xFF00000000000000

var masks [256]uint64
var active [16]uint64

func init() {
	for i := 0; i < 256; i++ {
		x := uint64(i)
		masks[i] = x | (x << 8) | (x << 16) | (x << 24) | (x << 32) | (x << 40) | (x << 48) | (x << 56)
	}
	active[0] = 0xFFFFFFFFFFFFFF00
	active[1] = 0xFFFFFFFFFFFF0000
	active[2] = 0xFFFFFFFFFF000000
	active[3] = 0xFFFFFFFF00000000
	active[4] = 0xFFFFFF0000000000
	active[5] = 0xFFFF000000000000
	active[6] = 0xFF00000000000000
	active[7] = 0x0000000000000000
	active[8] = 0x0000000000000000
	active[9] = 0x0000000000000000
	active[10] = 0x0000000000000000
	active[11] = 0x0000000000000000
	active[12] = 0x0000000000000000
	active[13] = 0x0000000000000000
	active[14] = 0x0000000000000000
	active[15] = 0x0000000000000000
}

func (n *nodeMasks) put(k byte, v interface{}) {
	m := &n.keys1
	c := n.count
	if n.count >= 8 {
		m = &n.keys2
		c = c - 8
	}
	*m = *m | (uint64(k) << (c * 8))
	n.vals[n.count] = v
	n.count++
}

func (n *nodeMasks) getWithFinalLoop(k byte) interface{} {
	if n.count == 0 {
		return nil
	}
	// mask is a uint64 with k at each byte position
	mask := masks[k]
	// act has bytes with value FF in positions we don't want to consider
	act := active[n.count-1]
	// XOR the mask and the keys, for any bytes that match the result will be 0
	// for ones that don't match the result will be not 0.
	// OR the result with act so that any key positions we shouldn't consider get
	// set to FF
	r := (mask ^ n.keys1) | act
	// now check each byte in the result for a zero
	for i := 0; i < 8; i++ {
		if r&0xFF == 0 {
			return n.vals[i]
		}
		r = r >> 8
	}
	if n.count < 9 {
		return nil
	}
	// same again for the upper 8 keys
	r = (mask ^ n.keys2) | active[n.count-9]
	for i := 0; i < 8; i++ {
		if r&0xFF == 0 {
			return n.vals[i+8]
		}
		r = r >> 8
	}
	return nil
}

func (n *nodeMasks) get(k byte) interface{} {
	if n.count == 0 {
		return nil
	}
	// mask is a uint64 with k at each byte position
	mask := masks[k]
	// act has bytes with value FF in positions we don't want to consider
	act := active[n.count-1]
	// XOR the mask and the keys, for any bytes that match the result will be 0
	// for ones that don't match the result will be not 0.
	// OR the result with act so that any key positions we shouldn't consider get
	// set to FF
	r := (mask ^ n.keys1) | act
	// now check each byte in the result for a zero
	if (r & b1) == 0 {
		return n.vals[0]
	}
	if (r & b2) == 0 {
		return n.vals[1]
	}
	if (r & b3) == 0 {
		return n.vals[2]
	}
	if (r & b4) == 0 {
		return n.vals[3]
	}
	if (r & b5) == 0 {
		return n.vals[4]
	}
	if (r & b6) == 0 {
		return n.vals[5]
	}
	if (r & b7) == 0 {
		return n.vals[6]
	}
	if (r & b8) == 0 {
		return n.vals[7]
	}
	if n.count < 9 {
		return nil
	}
	// same again for the upper 8 keys
	r = (mask ^ n.keys2) | active[n.count-9]
	if (r & b1) == 0 {
		return n.vals[8]
	}
	if (r & b2) == 0 {
		return n.vals[9]
	}
	if (r & b3) == 0 {
		return n.vals[10]
	}
	if (r & b4) == 0 {
		return n.vals[11]
	}
	if (r & b5) == 0 {
		return n.vals[12]
	}
	if (r & b6) == 0 {
		return n.vals[13]
	}
	if (r & b7) == 0 {
		return n.vals[14]
	}
	if (r & b8) == 0 {
		return n.vals[15]
	}
	return nil
}

func (n *nodeMasks) getMoreBitTwiddling(k byte) interface{} {
	if n.count == 0 {
		return nil
	}
	// This follows the same approach as get. But uses additional
	// bit twiddling to determine if any of the bytes are zero.
	mask := masks[k]
	r := (mask ^ n.keys1) | active[n.count-1]

	// see https://graphics.stanford.edu/~seander/bithacks.html#ZeroInWord
	x := (r - 0x0101010101010101) & ^(r) & 0x8080808080808080
	idx := bits.TrailingZeros64(x) / 8
	if idx < 8 {
		return n.vals[idx]
	}
	if n.count < 9 {
		return nil
	}
	r = (mask ^ n.keys2) | active[n.count-9]
	x = (r - 0x0101010101010101) & ^(r) & 0x8080808080808080
	idx = bits.TrailingZeros64(x) / 8
	if idx < 8 {
		return n.vals[idx+8]
	}
	return nil
}
