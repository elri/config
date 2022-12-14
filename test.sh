#!/bin/sh -e
go test -buildvcs=false -coverprofile=coverage.txt -covermode=atomic -coverpkg=./... $* ./...
go run github.com/boumenot/gocover-cobertura@latest < coverage.txt > coverage.xml
