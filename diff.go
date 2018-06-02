// Package diff contains code to produce diffs of arbitrary sequences.
//
// For example, diffing two files (represented as lines of strings),
// producing a unified diff:
//
//   result := diff.UnifiedDiffer(lines1, lines)
//   for _, line := range result { fmt.Println(line) }
//
// Or to diff two strings character by character:
//   ss1 = diff.StringSequence("one as .")
//   ss2 = diff.StringSequence("ike .")
//   d := diff.NewDiffer().Diff(ss1, ss2)
//   ... process d
package diff

// A Sequence is something that can be diffed against another sequence.
// Diffs are done based on hashes rather than the underlying sequence for
// efficiency.
type Sequence interface {
	Length() int
	Fingerprint(idx int) uint64
}

// An Edit describes part of an edit from one sequence to another.
// (x, x) -> x elements that are the same in both sequences.
// (x, 0) -> the removal of x lines from the first sequence.
// (0, x) -> the addition of x lines from the second sequence.
//
// (x, y) where x != 0, y != 0 and x != y is not allowed.
type Edit struct {
	DI, DJ int
}

// A Differ is something that can diff sequences.
type Differ interface {
	Diff(left, right Sequence) []Edit
}

// NewDiffer returns a default good differ.
func NewDiffer() Differ {
	return UniqueRunDiffer{MyersDiffer{}}
}

// UnifiedDiff compares left and right line-by-line, returning a
// unified diff.
func UnifiedDiff(left, right []string) []string {
	s0, s1 := NewLineSequence(left), NewLineSequence(right)
	edit := NewDiffer().Diff(s0, s1)
	var result []string
	i, j := 0, 0
	for _, e := range edit {
		for row := 0; row < e.DI|e.DJ; row++ {
			switch {
			case e.DI != 0 && e.DJ != 0:
				result = append(result, " "+left[i+row])
			case e.DI != 0:
				result = append(result, "-"+left[i+row])
			case e.DJ != 0:
				result = append(result, "+"+right[j+row])
			}
		}
		i, j = i+e.DI, j+e.DJ
	}
	return result
}
