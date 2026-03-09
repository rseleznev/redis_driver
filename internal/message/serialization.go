package message

func SerializeGetCommand(key string) []byte {
	arrLen := 3 + 16 + len(key)
	keyLen := len(key)

	if keyLen > 9 { // если длина ключа - двузначное число (пока как максимум)
		arrLen++
	}

	result := make([]byte, 0, arrLen)

	fixBytes := []byte{
		'*', '2', '\r', '\n',
		'$', '3', '\r', '\n',
		'G', 'E', 'T', '\r', '\n',
		'$', '3', '2', '\r', '\n',
		'd', '4', '1', 'd', '8', 'c', 'd', '9', '8', 'f', '0', '0', 'b', '2', '0', '4', 'e', '9', '8', '0', '0', '9', '9', '8', 'e', 'c', 'f', '8', '4', '2', '7', 'e', '\r', '\n',
		// d41d8cd98f00b204e9800998ecf8427e
	}
	result = append(result, fixBytes...)

	// if keyLen < 9 {
	// 	result = append(result, byte(keyLen))
	// } else {
	// 	fstDg := keyLen / 10
	// 	sndDg := keyLen % 10
	// 	result = append(result, byte(fstDg), byte(sndDg))
	// }
	// result = append(result, '\r', '\n')

	// for range keyLen {
	// 	result = append(result, key)
	// }

	// result = append(result, '\r', '\n')

	return result
}