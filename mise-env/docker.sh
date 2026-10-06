#!/bin/sh

if [ -z "${DOCKER_HOST:-}" ] && command -v docker >/dev/null 2>&1; then
	docker_context_host="$(docker context inspect --format '{{.Endpoints.docker.Host}}' 2>/dev/null)"
	if [ -n "$docker_context_host" ]; then
		export DOCKER_HOST="$docker_context_host"
	fi
fi

case "${DOCKER_HOST:-}" in
	unix://*/.lima/*|unix://*/.colima/*|unix://*/.docker/run/*)
		if [ -z "${TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE:-}" ]; then
			docker_security_options="$(docker info --format '{{json .SecurityOptions}}' 2>/dev/null || true)"
			case "$docker_security_options" in
				*name=rootless*)
					TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE="/run/user/$(id -u)/docker.sock"
					;;
				*)
					TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=/var/run/docker.sock
					;;
			esac
		fi
		export TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE
		;;
esac

unset docker_context_host docker_security_options
