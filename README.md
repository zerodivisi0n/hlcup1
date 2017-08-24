# HighloadCup Contest Round One

Task for the first round of competition

## Getting Started

### Build

Build docker image:

```
./build.sh
```

### Run

Start server with test data:

```
docker run -v ./data.zip:/tmp/data/data.zip -p 8080:80 hlcup1
```

### Develop

Start container with interactive shell:

```
docker run -it -v ./data.zip:/tmp/data/data.zip -v $(pwd):/go/src/server -w /go/src/server -p 8080:80 golang:1.8 bash
```
