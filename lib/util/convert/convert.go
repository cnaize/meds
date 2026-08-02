package convert

func BoolToUint32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}
