package diff

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type easyCase struct {
	a, b string
	lcs  int
}

var easyCases = []easyCase{
	{"abc", "abc", 3},
	{"abcd", "abc", 3},
	{"xabc", "yabc", 3},
	{"xabcx", "yabcy", 3},
	{"aaaa", "aaaaa", 4},
	{"aaaxyzaaa", "aaastuaaa", 6},
	{"aaa", "bbb", 0},
	{"I wandered lonely as a cloud.", "I wonder why I wander like a cloud.", 20},
	{"", "a", 0},
	{"", "", 0},
	{"ababababababab", "babababababab", 13},
	{"aaa", "bacabac", 3},
	{"one as .", "ike .", 3},
}

func reconstitute(t *testing.T, ed []Edit, a, b, header string) (string, string, int) {
	var ra, rb []byte
	lcs := 0
	i, j := 0, 0
	for _, e := range ed {
		if e.DI != 0 && e.DJ != 0 {
			if e.DI != e.DJ {
				t.Errorf("%s: DI != DJ. %d, %d", header, e.DI, e.DJ)
			}
			lcs += e.DI
			ra = append(ra, []byte(a[i:i+e.DI])...)
			rb = append(rb, []byte(a[i:i+e.DI])...)
		} else if e.DI != 0 {
			ra = append(ra, []byte(a[i:i+e.DI])...)
		} else if e.DJ != 0 {
			rb = append(rb, []byte(b[j:j+e.DJ])...)
		}
		i, j = i+e.DI, j+e.DJ
	}
	return string(ra), string(rb), lcs
}

var easyTestDiffers = []struct {
	d       Differ
	optimal bool
}{
	{MyersDiffer{}, true},
	{DynamicDiffer{}, true},
	{UniqueRunDiffer{MyersDiffer{}}, true},
	{UniqueRunDiffer{DynamicDiffer{}}, true},
	{PatienceDiffer{}, false},
	{UniqueRunDiffer{PatienceDiffer{}}, false},
}

func TestEasy(t *testing.T) {
	for _, cs0 := range easyCases {
		for _, cs := range []easyCase{cs0, easyCase{cs0.b, cs0.a, cs0.lcs}} {
			seqa := StringSequence(cs.a)
			seqb := StringSequence(cs.b)
			for _, differ := range easyTestDiffers {
				ed := differ.d.Diff(seqa, seqb)
				header := fmt.Sprintf(`%s: "%s" vs "%s."`, reflect.TypeOf(differ), cs.a, cs.b)
				a, b, lcs := reconstitute(t, ed, cs.a, cs.b, header)
				if a != cs.a {
					t.Errorf("%s: produced %v => %s, %s", header, ed, a, b)
				}
				if b != cs.b {
					t.Errorf("%s: produced %v => %s, %s", header, ed, a, b)
				}
				if differ.optimal && lcs != cs.lcs {
					t.Errorf("%s: produced suboptimal edit %v. LCS %d vs %d", header, ed, cs.lcs, lcs)
				}
			}
		}
	}
}

func TestUnifiedDiff(t *testing.T) {
	left := []string{
		"This is line one",
		"This is line two",
		"Line three which will be deleted.",
		"Line four",
	}
	right := []string{
		"This is line one",
		"This is line two",
		"Line four",
		"Line five: new",
		"Line six -- also new",
	}
	expected := strings.Join([]string{
		" This is line one",
		" This is line two",
		"-Line three which will be deleted.",
		" Line four",
		"+Line five: new",
		"+Line six -- also new",
	}, "\n")
	ud := strings.Join(UnifiedDiff(left, right), "\n")
	if expected != ud {
		t.Errorf("Diff wrong. Expected\n\n%s\n\nbut got\n\n%s", expected, ud)
	}
}

func TestUniquedRuns(t *testing.T) {
	seqa := StringSequence("abcg")
	seqb := StringSequence("defg")
	ua, ub := newUrdSequences(seqa, seqb)
	d := MyersDiffer{}.Diff(ua, ub)
	if len(d) != 3 {
		t.Errorf("Expected urd'ed edit length to be 3, but found %d: %v", len(d), d)
	}
}
