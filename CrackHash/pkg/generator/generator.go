package generator

type Generator struct {
	alphabet []byte
	State    []int
	aLen     int
}

func NewGenerator(alphabet []byte, startIdx int64) Generator {
	aLen := int64(len(alphabet))
	wordLength, adjustedIdx := getLengthAndAdjustedIndex(startIdx, int(aLen))

	state := make([]int, wordLength)
	for i := wordLength - 1; i >= 0; i-- {
		state[i] = int(adjustedIdx % aLen)
		adjustedIdx /= aLen
	}

	return Generator{
		alphabet: alphabet,
		State:    state,
		aLen:     len(alphabet),
	}
}

func (g *Generator) NextState() {
	for i := len(g.State) - 1; i >= 0; i-- {
		g.State[i]++
		if g.State[i] < g.aLen {
			return
		}
		g.State[i] = 0
	}
	g.State = make([]int, len(g.State)+1)
}

func (g *Generator) CurrentWordBytes(wordBuf []byte) []byte {
	if cap(wordBuf) < len(g.State) {
		wordBuf = make([]byte, len(g.State))
	} else {
		wordBuf = wordBuf[:len(g.State)]
	}

	for i, symbolIdx := range g.State {
		wordBuf[i] = g.alphabet[symbolIdx]
	}
	return wordBuf
}

func TotalWords(maxWordLength int, alphabetLength int) int64 {
	var total int64 = 0
	var count int64 = int64(alphabetLength)
	for i := 1; i <= maxWordLength; i++ {
		total += count
		count *= int64(alphabetLength)
	}
	return total
}

// возвращает длину слова и локальный индекс среди слов этой длины
func getLengthAndAdjustedIndex(idx int64, alphabetLength int) (int, int64) {
	length := 1
	countWordsOfLength := int64(alphabetLength)

	for idx >= countWordsOfLength {
		idx -= countWordsOfLength
		length++
		countWordsOfLength *= int64(alphabetLength)
	}
	return length, idx
}
