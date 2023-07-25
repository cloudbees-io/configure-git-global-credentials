FROM alpine:3.18 AS certs

RUN apk add -U --no-cache ca-certificates

FROM golang:1.20.5-alpine3.18 AS build

WORKDIR /work

COPY go.mod* go.sum* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /build-out/ .

FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=build /build-out/* /usr/bin/

WORKDIR /cloudbees/home

ENV HOME=/cloudbees/home
ENV PATH=/usr/bin

ENTRYPOINT ["configure-git-global-credentials"]

## Add a secondary target that allows testing locally under similar conditions without modifying your local gitconfig

FROM golang:1.20 AS testing

COPY --from=build /build-out/* /usr/local/bin/

RUN mkdir -p /cloudbees/home

WORKDIR /cloudbees/workspace

ENV HOME=/cloudbees/home

ENTRYPOINT ["bash"]

