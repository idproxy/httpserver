package pathsegment

import "fmt"

type PathSegments interface {
	Add(PathSegment)
	Get(index int) PathSegment
	Size() int
}

type pathSegments []PathSegment

// PathSegment is a segment in the url path, which can be of a certain kind
type PathSegment struct {
	Value string
	Kind  PathSegmentKind
}

func New(path string) (PathSegments, bool) {
	fmt.Printf("getPathSegments -> path: %s\n", path)
	ps := &pathSegments{}

	//we always start with a "/" pathSegment
	ps.Add(PathSegment{
		Value: string([]byte(path)[0]),
		Kind:  Root,
	})

	valid := true
	begin := 0
	kind := Normal
	lastPathSegment := false
	// when there is a single "/" we dont need to find the pathSegments as it is "/"
	for idx, c := range []byte(path) {
		// ignore the first char as it is always "/"
		if idx > 0 {
			// if char is a / or we are done we append the pathSegment to the slice
			if c == '/' || idx == len(path)-1 {
				end := idx
				if idx == len(path)-1 { // if we are at the end we include the last char
					if c == '/' {
						// the last char is "/" so we need to add create a new segment
						lastPathSegment = true
					} else {
						// the last char is not a "/", so we need to add the last char to the
						// path segment value
						end = end + 1
					}
				}
				fmt.Printf("getPathSegments -> append begin %d, end %d\n", begin, end)
				ps.Add(PathSegment{
					Value: string([]byte(path[begin+1 : end])),
					Kind:  kind,
				})
				if lastPathSegment {
					ps.Add(PathSegment{
						Value: string(c),
						Kind:  kind,
					})
				}
				begin = idx
				kind = Normal
			}
			switch c {
			case ':':
				// this means there is a : char in the pathSegment that is not the first one
				if kind != Normal {
					valid = false
				}
				kind = Param
			case '*':
				// this means there is a : char in the pathSegment that is not the first one
				if kind != Normal {
					valid = false
				}
				kind = CatchAll
			}
		}
	}
	return ps, valid
}

func (r *pathSegments) Add(ps PathSegment) {
	*r = append(*r, ps)
}

func (r *pathSegments) Get(index int) PathSegment {
	return (*r)[index]
}

func (r *pathSegments) Size() int {
	return len(*r)
}

type PathSegmentKind uint32

const (
	Normal PathSegmentKind = iota
	Root
	Param
	CatchAll
	Query
)

func (r PathSegmentKind) String() string {
	switch r {
	case Normal:
		return "normal"
	case Root:
		return "root"
	case Param:
		return "param"
	case CatchAll:
		return "catchAll"
	case Query:
		return "query"
	default:
		return "unknown"
	}
}
