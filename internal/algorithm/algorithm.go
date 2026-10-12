package algorithm

// TODO: Refactor to not have to pass in number each time

type Algorithm interface {
	Next(numBackends int) int
}
