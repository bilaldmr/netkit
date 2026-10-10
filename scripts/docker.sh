#!/usr/bin/env bash

set -e

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "$REPO_ROOT"

COMPOSE_FILE=docker-compose.yml

STACK=(-f "$COMPOSE_FILE")

usage() {
    echo "usage: $0 <build|up|down|logs|ps|clean>"
    exit 1
}

case "${1:-}" in
    build)
        docker compose "${STACK[@]}" build
        ;;
    up)
        docker compose "${STACK[@]}" up -d --build
        ;;
    down)
        docker compose "${STACK[@]}" down
        ;;
    logs)
        docker compose "${STACK[@]}" logs -f
        ;;
    ps)
        docker compose "${STACK[@]}" ps
        ;;
    clean)
        docker compose "${STACK[@]}" down -v --remove-orphans
        ;;
    *)
        usage
        ;;
esac
