FROM golang:1.6

RUN apt-get update \ 
 && apt-get install --yes netcat redis-tools

ARG user=relax
ARG group=relax
ARG uid=1000
ARG gid=1000
ENV GOPATH=/opt/relax

WORKDIR ${GOPATH}

COPY . .



RUN go get github.com/zerobotlabs/relax/slack \
 && go get github.com/zerobotlabs/relax/healthcheck \
 && go build \ 
 && mv relax bin

RUN chown ${uid}:${gid} $GOPATH \
 && groupadd -g ${gid} ${group} \
 && useradd -d "$GOPATH" -u ${uid} -g ${gid} -m -s /bin/bash ${user} \
 && chown -R ${uid}:${gid} /usr/local /bin

USER ${user}

CMD ["bin/relax"]

