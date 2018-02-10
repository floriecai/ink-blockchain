#!/bin/bash
# Use to make sure that everything compiles fine.
# Does not check proj1-server right now.

# Check top files individually since each have their own func main()
top_files=($(find ./*.go -maxdepth 0 -type f))
for f in ${top_files[@]}; do
	go test $f;
done

go test ./blockartlib/*.go
go test ./libminer/*.go
go test ./miner/*.go
go test ./minerserver/*.go
go test ./shapelib/*.go
go test ./utils/*.go
# go test ./misc/*.go
