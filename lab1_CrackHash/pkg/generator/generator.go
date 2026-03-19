package generator

func CalculateTotalWords(alphabetLen int, maxLength int) int64 {
	var total int64 = 0
	var count int64 = int64(alphabetLen)
	for i := 1; i <= maxLength; i++ {
		total += count
		count *= int64(alphabetLen)
	}
	return total
}

func GetWordAtIndex(idx int64, alphabet []rune) string {
	aLen := int64(len(alphabet))
	length := 1
	count := aLen

	for idx >= count {
		idx -= count
		length++
		count *= aLen
	}

	word := make([]rune, length)
	for i := length - 1; i >= 0; i-- {
		word[i] = alphabet[idx%aLen]
		idx /= aLen
	}

	return string(word)
}
