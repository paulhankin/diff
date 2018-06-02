package diff

// DynamicDiffer constructs an optimally short diff using a naive dynamic
// programming method. It uses O(NM) space and time where N and M are
// the sizes of the sequences it's asked to diff.
type DynamicDiffer struct{}

// Logically extend the the score table to include scores when one or the other
// sequence has been used up.
func getScore(tbl [][]int, i, j int) int {
	if i == len(tbl) || j == len(tbl[0]) {
		return len(tbl) + len(tbl[0]) - i - j
	}
	return tbl[i][j]
}

// Diff diffs the two sequences using a naive dynamic programming approach.
func (DynamicDiffer) Diff(left, right Sequence) []Edit {
	// Initialise dynamic programming table.
	tbl := make([][]int, left.Length())
	for i := range tbl {
		tbl[i] = make([]int, right.Length())
	}
	// Find optimal path from (0, 0) to (left.Length(), right.Length())
	for i := left.Length() - 1; i >= 0; i-- {
		for j := right.Length() - 1; j >= 0; j-- {
			switch {
			case left.Fingerprint(i) == right.Fingerprint(j):
				tbl[i][j] = 1 + getScore(tbl, i+1, j+1)
			case getScore(tbl, i+1, j) <= getScore(tbl, i, j+1):
				tbl[i][j] = 1 + getScore(tbl, i+1, j)
			default:
				tbl[i][j] = 1 + getScore(tbl, i, j+1)
			}
		}
	}
	// Build an edit list from the optimal path.
	result := make([]Edit, 0, 10)
	i, j := 0, 0
	prev := Edit{0, 0}
	for i != left.Length() || j != right.Length() {
		var ed Edit
		switch {
		case i == left.Length():
			ed = Edit{0, 1}
		case j == right.Length():
			ed = Edit{1, 0}
		case left.Fingerprint(i) == right.Fingerprint(j):
			ed = Edit{1, 1}
		case getScore(tbl, i+1, j) <= getScore(tbl, i, j+1):
			ed = Edit{1, 0}
		case getScore(tbl, i+1, j) > getScore(tbl, i, j+1):
			ed = Edit{0, 1}
		default:
			panic("unreachable")
		}
		// We accumulate edits into the previous edit in the result if possible.
		if ed != prev {
			result = append(result, Edit{})
			prev = ed
		}
		result[len(result)-1].DI += ed.DI
		result[len(result)-1].DJ += ed.DJ
		i += ed.DI
		j += ed.DJ
	}
	return result
}
