// package algorithm_test Unit tests for algorithms
package algorithm_test

import (
	"testing"

	"github.com/anyangateny1/Load-Balancer/internal/algorithm"
)

func TestRoundRobinIsInOrder(t *testing.T) {
	t.Parallel()

	rr := algorithm.RoundRobin{}

	const serverNumbers int = 1_000_000

	for serverNum := range serverNumbers {
		expected := rr.Next(serverNumbers)
		if expected != serverNum {
			t.Fatalf("expected does not match serverNum: expected, serverNum %d, %d", expected, serverNum)
		}
	}
}

func TestRoundRobinOneElement(t *testing.T) {
	t.Parallel()

	rr := algorithm.RoundRobin{}

	const serverNumber = 1
	const iterations = 1_000_000

	for i := range iterations {
		index := rr.Next(serverNumber)

		if index != 0 {
			t.Fatalf("iteration %d: expected index 0, got %d", i, index)
		}
	}
}

func TestRoundRobinEachNumberIsNTimes(t *testing.T) {
	t.Parallel()

	rr := algorithm.RoundRobin{}

	const serverNumbers int = 1_000_000
	const iterations int = 10

	numbers_seen := map[int]int{}

	for range iterations {
		for range serverNumbers {
			numbers_seen[rr.Next(serverNumbers)]++
		}
	}

	for serverNum := range serverNumbers {
		if numbers_seen[serverNum] != iterations {
			t.Fatalf("number not seem amount as iteration: number, number of iterations, expected %d, %d, %d",
				serverNum, numbers_seen[serverNum], iterations)
		}
	}
}
