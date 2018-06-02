package diff

// MyersDiffer constructs an optimally short diff using the algorithm described in
// "An O(ND) Difference Algorithm and Its Variations" by Eugene Myers.
type MyersDiffer struct{}

// TODO: clean this mess up.
func makeEdit(VS [][]int, k int, left, right Sequence) []Edit {
	// pairs will store (in reverse order) the ends of the snakes along
	// the optimal path.
	var pairs [][2]int
	for D := len(VS) - 1; D > 0; D-- {
		x := VS[D][k]
		y := x + D - k*2
		pairs = append(pairs, [2]int{x, y})
		snake := 0
		for x-snake > 0 && y-snake > 0 && left.Fingerprint(x-snake-1) == right.Fingerprint(y-snake-1) {
			snake++
		}
		if k > 0 && VS[D-1][k-1]+1+snake >= x {
			k--
		}
	}
	pairs = append(pairs, [2]int{VS[0][k], VS[0][k] - k*2})
	var ed []Edit
	x, y := 0, 0
	for i := 0; i < len(pairs); i++ {
		nxy := pairs[len(pairs)-i-1]
		nx, ny := nxy[0], nxy[1]
		if nx-ny > x-y {
			if len(ed) == 0 || ed[len(ed)-1].DJ != 0 {
				ed = append(ed, Edit{})
			}
			ed[len(ed)-1].DI++
			x++
		} else if nx-ny < x-y {
			if len(ed) == 0 || ed[len(ed)-1].DI != 0 {
				ed = append(ed, Edit{})
			}
			ed[len(ed)-1].DJ++
			y++
		}
		if x != nx {
			ed = append(ed, Edit{nx - x, nx - x})
		}
		x, y = nx, ny
	}
	return ed
}

// Diff diffs left and right, producting a minimal set of edits.
func (MyersDiffer) Diff(left, right Sequence) []Edit {
	MAX := left.Length() + right.Length()
	var VS [][]int
	for D := 0; D <= MAX; D++ {
		VS = append(VS, make([]int, D+1))
		for k := 0; k <= D; k++ {
			var x int
			var PV []int
			if len(VS) >= 2 {
				PV = VS[len(VS)-2]
			}
			if k == 0 || k < D && PV[k] > PV[k-1] {
				if D == 0 {
					x = 0
				} else {
					x = PV[k]
				}
			} else {
				x = PV[k-1] + 1
			}
			y := x + D - k*2
			for x < left.Length() && y < right.Length() && left.Fingerprint(x) == right.Fingerprint(y) {
				x, y = x+1, y+1
			}
			VS[len(VS)-1][k] = x
			if x == left.Length() && y == right.Length() {
				return makeEdit(VS, k, left, right)
			}
		}
	}
	panic("Unreachable")
	return nil
}
