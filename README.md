# MessagePack
MessagePack is an efficient binary serialization format. It lets you exchange data among multiple languages like JSON. But it’s faster and smaller. Small integers are encoded into a single byte, and typical short strings require only one extra byte in addition to the strings themselves.

## Prerequisites

- docker

## Install

#### Build docker image
```
./run.sh build
```

#### Run docker container
```
./run.sh start
```

#### Run msgpack
```
./msgpack
```
After running msgpack, user can input the following functions:

| Type | Function  | Comment  | Format |
| ------------ | ------------ |------------ |------------ |
| string  | encode  | encode JSON to MessagePack format  | Bool / Int / Float64 / Map / Slice / String |
| string  | decode  | decode MessagePack format to JSON  | - |
| string  | exit | stop msgpack program  | - |
| signal  | Ctrl+C  | stop msgpack program  | -|

#### Run manual
- User who need to manually interact with the program should mount the repository in a docker container.
- Run with files
```
go run main.go
```
- Unittest
```
go test -v
```

This project test on macOS 10.11.6.

## References
- [MessagePack](https://msgpack.org/index.html)
- [vmihailenco/msgpack](https://github.com/vmihailenco/msgpack)