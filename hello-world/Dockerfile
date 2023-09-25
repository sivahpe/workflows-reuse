# Copyright 2022 Hewlett Packard Enterprise Development LP

FROM alpine:3.13 AS build

RUN set -eux; \
	apk add -U --no-cache \
		curl \
		git  \
		make \
	;

# add new user
ARG USER=default
RUN addgroup -g 1000 ${USER} \
        && adduser -h /build -D -u 1000 -G ${USER} ${USER} \
    ;

USER ${USER}

WORKDIR /build

COPY --chown=${USER}:${USER} go.mod ./
COPY --chown=${USER}:${USER} ./hello-world ./hello-world
COPY --chown=${USER}:${USER} ./Makefile ./Makefile

ARG VERSION
RUN set -eux; \
    CGO_ENABLED=0;

FROM alpine:3.13 AS base-runtime

ARG HTTPS_PROXY

# add new user
ARG USER=default
RUN set -eux; \
    addgroup ${USER}; \
    adduser -D -G ${USER} ${USER};

USER ${USER}
WORKDIR /home/${USER}

FROM base-runtime AS hello-world
