//go:build !amd64.v3

package main

// fusedMultiplyAdd says the tests were built at a level of amd64 at which Go
// fuses a multiplication and an addition into one step that rounds once.
const fusedMultiplyAdd = false
