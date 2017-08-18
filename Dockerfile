FROM ubuntu:16.04

# Install dependencies
RUN set -x\
        && apt-get update \
        && apt-get install -y --no-install-recommends \
                ca-certificates \
                wget \
                git \
        && rm -rf /var/lib/apt/lists/*

# Install MongoDB
RUN set -x\
        && wget https://repo.percona.com/apt/percona-release_0.1-4.xenial_all.deb \
        && dpkg -i percona-release_0.1-4.xenial_all.deb \
        && apt-get update \
        && apt-get install -y --no-install-recommends \
                percona-server-mongodb-34-server \
        && rm -rf /var/lib/apt/lists/* \
        && rm -rf /var/lib/mongodb /var/log/mongodb \
        && rm -f percona-release_0.1-4.xenial_all.deb \
        && mv /etc/mongod.conf /etc/mongod.conf.orig \
        && mkdir -p /data/db

ENV GO_VERSION 1.8.3

# Install Go
RUN set -x \
        && wget https://storage.googleapis.com/golang/go$GO_VERSION.linux-amd64.tar.gz \
        && tar -C /usr/local -xzf go$GO_VERSION.linux-amd64.tar.gz \
        && rm -f go$GO_VERSION.linux-amd64.tar.gz \
        && apt-get purge -y --auto-remove wget \
        && export PATH="/usr/local/go/bin:$PATH" \
        && go version

# Update env path for Go
ENV GOPATH /go
ENV GOROOT /usr/local/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin"

COPY . /go/src/server
WORKDIR /go/src/server

# Build app
RUN go get -v . && go build . && go install .

CMD ["./wrapper.sh"]
