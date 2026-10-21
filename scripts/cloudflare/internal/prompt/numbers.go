package prompt

import "sort"

// Numbers is a create run's set-number book: every number accepted for each set,
// so a prompt can offer one past it and refuse a repeat. It covers the run only;
// numbers already in the bucket are outside it, by design.
type Numbers struct {
	bySet map[string][]int
}

func NewNumbers() *Numbers {
	return &Numbers{bySet: make(map[string][]int)}
}

// Default is the number offered for the next item of a set: one past the highest
// number it has taken, so the default is always free. A gap left earlier stays a
// gap — skipping a number is not a collision, and the next item does not fill it.
func (n *Numbers) Default(set string) int {
	highest := 0
	for _, number := range n.bySet[set] {
		if number > highest {
			highest = number
		}
	}

	return highest + 1
}

// Acknowledge records an accepted number, reporting false when the set already
// has it. Only an accepted number is recorded, so a refused answer leaves the
// book as it was.
func (n *Numbers) Acknowledge(set string, number int) bool {
	for _, taken := range n.bySet[set] {
		if taken == number {
			return false
		}
	}

	n.bySet[set] = append(n.bySet[set], number)

	return true
}

// AcknowledgedIn returns the set's accepted numbers, sorted, as a copy the caller
// may keep.
func (n *Numbers) AcknowledgedIn(set string) []int {
	taken := append([]int{}, n.bySet[set]...)
	sort.Ints(taken)

	return taken
}
