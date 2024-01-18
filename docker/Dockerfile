FROM alpine:3.19.0

USER root

RUN apk add tcpdump
RUN apk add bash
RUN apk add curl
RUN apk add jq

WORKDIR /opt/dumpy

COPY * ./
RUN echo "cat dumpy_ASCII.txt" > ~/.bashrc

ENTRYPOINT ["/bin/bash"]

                          
