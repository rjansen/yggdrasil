FROM golang:1.11

ARG GID=1000
ARG UID=1000
RUN groupadd -g $GID -o yggdrasil && useradd -m -u $UID -g $GID -o -s /bin/bash yggdrasil

RUN curl -O -L https://github.com/gotestyourself/gotestsum/releases/download/v0.3.2/gotestsum_0.3.2_linux_amd64.tar.gz && \
    tar xf gotestsum_0.3.2_linux_amd64.tar.gz && \
    mv gotestsum /usr/local/bin && \
    rm gotestsum_0.3.2_linux_amd64.tar.gz
RUN curl -L -o codecov https://codecov.io/bash && \
    chmod a+x codecov && \
    mv codecov /usr/local/bin
RUN go get -u github.com/derekparker/delve/cmd/dlv

WORKDIR /go/src/github.com/rjansen/yggdrasil
ENTRYPOINT ["make"]
