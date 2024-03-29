FROM golang:1.20.8-bullseye AS builder

WORKDIR /
RUN apt-get install git
COPY ./ /ragno

#RUN make dependencies
WORKDIR /ragno
RUN make dependencies
RUN make build

# FINAL STAGE -> copy the binary and few config files
FROM debian:bullseye

RUN mkdir /ragno
WORKDIR /ragno
COPY --from=builder /ragno/build/ ./
COPY --from=builder /ragno/db/migrations ./db/migrations

# Crawler exposed Port
EXPOSE 5001
EXPOSE 5080

ENTRYPOINT ["./ragno"]