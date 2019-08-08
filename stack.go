package main

import "os"

type Stack struct {
	stack []os.FileInfo
}

func (s *Stack) push(elem os.FileInfo) {
	s.stack = append(s.stack, elem)
}

func (s *Stack) pop() os.FileInfo {
	n := len(s.stack) - 1
	elem := s.stack[n]

	s.stack = s.stack[:n]

	return elem
}
