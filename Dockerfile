FROM golang:1.24-bullseye AS builder
RUN apt-get update && apt-get install -y libseccomp-dev gcc pkg-config
WORKDIR /app
COPY main.go go.mod go.sum ./
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o seccomp-agent main.go

FROM ubuntu:22.04
RUN apt-get update && apt-get install -y libseccomp2 net-tools && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/seccomp-agent /usr/bin/seccomp-agent
ENTRYPOINT ["seccomp-agent"]
