package diff

// PatienceDiffer generates diffs using the Patience Diff algorithm. This:
// 1. Removes matching lines from the start and end of a file
// 2. Computes the LCS for lines that occur uniquely in both sequences
// 3. Do step 1 for the sections in between.
// It's described in http://bramcohen.livejournal.com/73318.html
type PatienceDiffer struct{}

type card struct {
	leftidx, rightidx int
	back              *card
}

// TODO: binary search
func findpile(piles []*card, rightidx int) int {
	for i := range piles {
		if piles[i].rightidx > rightidx {
			return i
		}
	}
	return len(piles)
}

func computelcs(left, right Sequence, leftmin, leftmax, rightmin, rightmax int) [][2]int {
	// These maps store the line number if the hash is uniq, else -1.
	leftuniq := make(map[uint64]int)
	rightuniq := make(map[uint64]int)
	for i := leftmin; i < leftmax; i++ {
		fp := left.Fingerprint(i)
		if _, ok := leftuniq[fp]; ok {
			leftuniq[fp] = -1
		} else {
			leftuniq[fp] = i
		}
	}
	for i := rightmin; i < rightmax; i++ {
		fp := right.Fingerprint(i)
		if _, ok := rightuniq[fp]; ok {
			rightuniq[fp] = -1
		} else {
			rightuniq[fp] = i
		}
	}
	var piles []*card
	for i := leftmin; i < leftmax; i++ {
		fp := left.Fingerprint(i)
		leftidx, ok := leftuniq[fp]
		if !ok || leftidx == -1 {
			continue
		}
		rightidx, ok := rightuniq[fp]
		if !ok || rightidx == -1 {
			continue
		}
		// We've found a line that occurs uniquely in both sequences.
		pileidx := findpile(piles, rightidx)
		if pileidx == len(piles) {
			piles = append(piles, nil)
		}
		var prev *card
		if pileidx > 0 {
			prev = piles[pileidx-1]
		}
		piles[pileidx] = &card{leftidx, rightidx, prev}
	}
	if len(piles) == 0 {
		// We had no lines.
		return nil
	}
	result := make([][2]int, len(piles))
	for i, card := 0, piles[len(piles)-1]; card != nil; i, card = i+1, card.back {
		result[len(piles)-i-1] = [2]int{card.leftidx, card.rightidx}
	}
	return result
}

func patience(dolcs bool, left, right Sequence, leftmin, leftmax, rightmin, rightmax int, result *[]Edit) {
	var i int
	// Match the starts of the sequences, writing the edit into the result
	for i = 0; i < leftmax-leftmin && i < rightmax-rightmin; i++ {
		if left.Fingerprint(leftmin+i) != right.Fingerprint(rightmin+i) {
			break
		}
	}
	if i > 0 {
		*result = append(*result, Edit{i, i})
	}
	leftmin, rightmin = leftmin+i, rightmin+i
	// Match the ends of the sequences, storing the number of matched lines
	// so we can add them to the result at the end.
	for i = 0; i < leftmax-leftmin && i < rightmax-rightmin; i++ {
		if left.Fingerprint(leftmax-i-1) != right.Fingerprint(rightmax-i-1) {
			break
		}
	}
	tailMatched := i
	leftmax, rightmax = leftmax-i, rightmax-i
	if dolcs {
		lcs := computelcs(left, right, leftmin, leftmax, rightmin, rightmax)
		for _, line := range lcs {
			patience(false, left, right, leftmin, line[0], rightmin, line[1], result)
			*result = append(*result, Edit{1, 1})
			leftmin, rightmin = line[0]+1, line[1]+1
		}
	}
	if leftmax > leftmin {
		*result = append(*result, Edit{leftmax - leftmin, 0})
	}
	if rightmax > rightmin {
		*result = append(*result, Edit{0, rightmax - rightmin})
	}
	if tailMatched > 0 {
		*result = append(*result, Edit{tailMatched, tailMatched})
	}
}

// Diff diffs left and right, using Bram Cohen's "patience" diffing algorithm.
func (PatienceDiffer) Diff(left, right Sequence) []Edit {
	var result []Edit
	patience(true, left, right, 0, left.Length(), 0, right.Length(), &result)
	return result
}
