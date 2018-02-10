#!/bin/bash
# Use to make sure that everything compiles fine.
# Does not check proj1-tester right now.

go test ./*.go
go test ./blockartlib/*.go
go test ./libminer/*.go
go test ./miner/*.go
go test ./minerserver/*.go
go test ./shapelib/*.go
go test ./utils/*.go
# go test ./misc/*.go
