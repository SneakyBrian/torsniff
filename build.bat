@echo off

pushd .\client
CALL gulp
popd

go generate
go build
