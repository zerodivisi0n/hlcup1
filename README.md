# HighloadCup Contest Round One

Task for the first round of competition

## Getting Started

### Build

Build docker image:

```
docker build -t highloadcup .
```

### Run

Start server with test data:

```
docker run -v ./data.zip:/tmp/data/data.zip -p 8080:80 highloadcup
```

### Develop

Start container with interactive shell:

```
docker run -it -v ./data.zip:/tmp/data/data.zip -v $(pwd):/go/src/server -p 8080:80 highloadcup bash
```
