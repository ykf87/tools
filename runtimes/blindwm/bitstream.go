package blindwm

// bytes <-> bits
func BytesToBits(data []byte) []int {
	bits := make([]int, len(data)*8)
	for i, b := range data {
		for j := 0; j < 8; j++ {
			if b&(1<<(7-j)) != 0 {
				bits[i*8+j] = 1
			} else {
				bits[i*8+j] = 0
			}
		}
	}
	return bits
}

func BitsToBytesLen(bits []int, length int) []byte {
	data := make([]byte, length)
	for i := 0; i < length; i++ {
		var b byte
		for j := 0; j < 8; j++ {
			b <<= 1
			idx := i*8 + j
			if idx < len(bits) && bits[idx] == 1 {
				b |= 1
			}
		}
		data[i] = b
	}
	return data
}

// uint32 <-> bits
func Uint32ToBits(v uint32) []int {
	bits := make([]int, 32)
	for i := 0; i < 32; i++ {
		if (v & (1 << (31 - i))) != 0 {
			bits[i] = 1
		} else {
			bits[i] = 0
		}
	}
	return bits
}

func BitsToUint32(bits []int) uint32 {
	var v uint32
	for i := 0; i < 32; i++ {
		v <<= 1
		if bits[i] == 1 {
			v |= 1
		}
	}
	return v
}

// 简单 ECC，重复 n 次
func RepeatBits(bits []int, n int) []int {
	out := []int{}
	for _, b := range bits {
		for i := 0; i < n; i++ {
			out = append(out, b)
		}
	}
	return out
}

// 解码重复码（多数投票）
func DecodeRepeatedBits(bits []int, n int) []int {
	out := []int{}
	for i := 0; i < len(bits); i += n {
		cnt := 0
		for j := 0; j < n && i+j < len(bits); j++ {
			cnt += bits[i+j]
		}
		if cnt*2 >= n {
			out = append(out, 1)
		} else {
			out = append(out, 0)
		}
	}
	return out
}
