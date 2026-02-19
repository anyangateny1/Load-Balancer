# Load Balancer

A TCP load balancer built from scratch in Go using only the standard library.

This is a learning project, as I am trying to learn Go and gain more practice with Test Driven Development, with the goal of getting hands-on experience with the language while exploring networking and distributed systems concepts.

## Goals

- Practice TDD for reliable, testable code
- Learn Go by building something real
- Understand how load balancers work under the hood
- Get more familiar with TCP networking, concurrency, and connection management
- Explore distributed systems ideas like health checks, failover, and request routing

## Building & Running

Requires **Go 1.25.0+**.

```bash
make build   # compile to bin/lb
make run     # run directly
make test    # run tests with race detection
make all     # fmt, vet, lint, test, build
```

## Project Structure

```
cmd/lb/                  – entry point
internal/ – load balancer implementation + tests
```
