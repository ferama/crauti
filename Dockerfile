# react frontend builder
FROM node:18-alpine as uibuilder
WORKDIR /src
COPY pkg/admin/ui .
RUN npm install && npm run build

FROM golang:1.20 as gobuilder
WORKDIR /go/src/app
COPY . .
COPY --from=uibuilder /src/dist pkg/admin/ui/dist
RUN go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /crauti .

# Final docker image
FROM ubuntu:latest
RUN set -eux; \
    apt update && \
    apt install -y \
        ca-certificates \
        curl \
        vim \
        byobu \
        psmisc \
        iputils-ping \
        netcat \
        dnsutils \
        bash-completion \
    && \
    apt clean

COPY --from=gobuilder /crauti /usr/local/bin/crauti

ENTRYPOINT ["/usr/local/bin/crauti"]