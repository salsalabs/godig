package godig

//Members is a list of supporter keys for a group.  It's a map to make it
//easy to find out if a supporter is a member.
type Members map[int]bool

//Census is a map of groups keys and members. A Census hold the members of a
//group and forms the basis for set logic.
type Census map[int]Members

//ResultInner contains a map of members for a result.
type ResultInner map[int]Members

//Result is a collection of Census records indexed by two levels of groups
//keys. They hold two-way permutations from a list of groups.
type Result map[int]ResultInner

//Analyze shows the overlap for permutations of an inclusive godig.Census without
//the supporters from an exclusinve godig.Census.  Returns a Result, which is a
//map of maps of members.
func Analyze(included Census, excluded Census) (Result, error) {
	result := make(Result)

	for i, m1 := range included {
		for j, m2 := range included {
			// godig.Intersection is reflective.
			if i >= j {
				r := Intersect(m1, m2)
				for _, m3 := range excluded {
					r = Difference(r, m3)
				}
				if len(r) > 0 {
					_, ok := result[i]
					if !ok {
						result[i] = make(ResultInner)
					}
					_, ok = result[i][j]
					if !ok {
						result[i][j] = make(Members)
					}
					result[i][j] = r
				}
			}
		}
	}
	return result, nil
}

//NewCensus creates a census map.
func NewCensus(a []int) Census {
	c := make(Census, len(a))
	for _, k := range a {
		c[int(k)] = make(Members, 1)
	}
	return c
}

//Intersect returns the intersection of two Members.
func Intersect(m1 Members, m2 Members) Members {
	m3 := make(Members)
	for k := range m1 {
		_, ok := m2[k]
		if ok {
			m3[k] = true
		}
	}
	return m3
}

//Difference return the first set without the members of the second set.
func Difference(m1 Members, m2 Members) Members {
	m3 := make(Members)
	for k := range m1 {
		_, ok := m2[k]
		if !ok {
			m3[k] = true
		}

	}
	return m3
}
