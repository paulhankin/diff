package diff

import (
	"hash/crc64"
)

// A StringSequence is a Sequence of bytes in a string.
type StringSequence string

func (ss StringSequence) Length() int {
	return len(ss)
}

func (ss StringSequence) Fingerprint(idx int) uint64 {
	return uint64(ss[idx])
}

// A HashesSequence is a slice of hashes.
type HashesSequence []uint64

func (hs HashesSequence) Length() int {
	return len(hs)
}

func (hs HashesSequence) Fingerprint(idx int) uint64 {
	return hs[idx]
}

// NewLineSequence constructs a Sequence from a slice of lines.
func NewLineSequence(lines []string) Sequence {
	tbl := crc64.MakeTable(crc64.ECMA)
	hashes := make([]uint64, len(lines))
	for i, line := range lines {
		hashes[i] = crc64.Checksum(([]byte)(line), tbl)
	}
	return HashesSequence(hashes)
}

type urdT struct {
	hash uint64
	size int
}

type urdSequence []urdT

func newUrdSequences(left, right Sequence) (urdSequence, urdSequence) {
	hashesLeft := make(map[uint64]struct{})
	hashesRight := make(map[uint64]struct{})
	for i := 0; i < left.Length(); i++ {
		hashesLeft[left.Fingerprint(i)] = struct{}{}
	}
	for j := 0; j < right.Length(); j++ {
		hashesRight[right.Fingerprint(j)] = struct{}{}
	}

	var result [2][]urdT
	cases := []struct {
		seq         Sequence
		otherHashes map[uint64]struct{}
	}{
		{left, hashesRight},
		{right, hashesLeft},
	}
	for k := range cases {
		lastUnique := false
		for i := 0; i < cases[k].seq.Length(); i++ {
			fp := cases[k].seq.Fingerprint(i)
			_, found := cases[k].otherHashes[fp]
			if lastUnique && !found {
				N := len(result[k]) - 1
				result[k][N] = urdT{result[k][N].hash, result[k][N].size + 1}
			} else {
				result[k] = append(result[k], urdT{fp, 1})
			}
			lastUnique = !found
		}
	}
	return result[0], result[1]
}

func (us urdSequence) Fingerprint(i int) uint64 {
	return us[i].hash
}

func (us urdSequence) Length() int {
	return len(us)
}

func (us urdSequence) RangeSize(i, j int) int {
	t := 0
	for k := i; k < j; k++ {
		t += us[k].size
	}
	return t
}

// An UniqueRunDiffer isn't a differ itself. Instead, it replaces runs
// of elements which don't appear in the "other" sequence with a single
// item before using the underlying differ.
type UniqueRunDiffer struct {
	D Differ
}

func (urd UniqueRunDiffer) Diff(left, right Sequence) []Edit {
	urdLeft, urdRight := newUrdSequences(left, right)
	edits := urd.D.Diff(urdLeft, urdRight)
	// Uncompress edits. Snakes expand to themselves (since if the run is
	// in both left and right, they can't be unique-run-compressed).
	// But elements from one side or the other may need decompressing.
	// We uncompress in place.
	i, j := 0, 0
	for k := range edits {
		di, dj := edits[k].DI, edits[k].DJ
		if di != 0 && dj != 0 {
			edits[k] = Edit{di, dj}
		} else if di != 0 {
			edits[k] = Edit{urdLeft.RangeSize(i, i+di), 0}
		} else if dj != 0 {
			edits[k] = Edit{0, urdRight.RangeSize(j, j+dj)}
		}
		i, j = i+di, j+dj
	}
	return edits
}
