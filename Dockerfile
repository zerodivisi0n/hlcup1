FROM ubuntu:16.04

ENV MONGO_MAJOR 3.4
ENV MONGO_VERSION 3.4.7
ENV GO_VERSION 1.8.3

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
        && apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 0C49F3730359A14518585931BC711F9BA15703C6 \
        && echo "deb [ arch=amd64,arm64 ] http://repo.mongodb.org/apt/ubuntu xenial/mongodb-org/$MONGO_MAJOR multiverse" | tee /etc/apt/sources.list.d/mongodb-org-$MONGO_MAJOR.list \
        && apt-get update \
        && apt-get install -y --no-install-recommends \
                mongodb-org-server=$MONGO_VERSION \
        && rm -rf /var/lib/apt/lists/* \
        && rm -rf /var/lib/mongodb /var/log/mongodb \
        && mv /etc/mongod.conf /etc/mongod.conf.orig \
        && mkdir -p /data/db

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
