package algorithm

type Algorithm interface {
	Next(numBackends int) int
}
