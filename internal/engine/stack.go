package engine

type (
	callStack struct {
		maxSize int
		size    int
		head    *node
	}

	node struct {
		frame stackFrame
		next  *node
	}

	stackFrame struct {
		name string
	}
)

func newCallStack() *callStack {
	return &callStack{}
}

func (s *callStack) Push(frame stackFrame) bool {
	if s.maxSize != 0 && s.size+1 >= s.maxSize {
		return false
	}

	s.head = &node{
		frame: frame,
		next:  s.head,
	}
	s.size++
	return true
}

func (s *callStack) Pop() (stackFrame, bool) {
	if s.head == nil {
		return stackFrame{}, false
	}
	frame := s.head.frame
	s.head = s.head.next
	s.size--
	return frame, true
}

func (s *callStack) Slice() []stackFrame {
	frames := make([]stackFrame, s.size)
	current := s.head
	i := 0
	for current != nil {
		frames[i] = current.frame
		current = current.next
		i++
	}
	return frames
}

func (f stackFrame) String() string {
	return f.name
}
