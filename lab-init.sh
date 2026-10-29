#!/usr/bin/env bash
#
# Каркас практики 01 (01-url-shortener).
# Сгенерирован tools/pack-lab.sh — вручную не править.
#
# Развернуть каркас:      ./lab-init.sh
# Сверить с текущей:      ./lab-init.sh --check
#
set -euo pipefail

LAB_DIR="${LAB_DIR:-lab01}"
STARTER_VERSION="9a1b25843cfd"
PACKED_AT="2026-09-04"

FILES=(
  ".golangci.yml"
  "Dockerfile"
  "Makefile"
  "cmd/shortener/main.go"
  "docker-compose.yml"
  "go.mod"
  "internal/config/config.go"
  "internal/httpapi/api.go"
  "internal/storage/storage.go"
  "internal/storage/storage_test.go"
)

INFRA=(
  "go.mod"
  "Makefile"
  "Dockerfile"
  "docker-compose.yml"
  ".golangci.yml"
)

emit() {
  case "$1" in
  ".golangci.yml")
    cat <<'__LAB_FILE_EOF__'
version: "2"

linters:
  enable:
    - errcheck
    - govet
    - ineffassign
    - staticcheck
    - unused
    - bodyclose
    - errorlint
    - noctx

  exclusions:
    rules:
      # httptest.NewRequest не уходит в сеть, и требовать для него контекст
      # незачем: правило noctx нацелено на запросы без дедлайна.
      - path: _test\.go
        linters:
          - noctx

formatters:
  enable:
    - gofmt
    - goimports
__LAB_FILE_EOF__
    ;;
  "Dockerfile")
    cat <<'__LAB_FILE_EOF__'
FROM golang:1.27-alpine AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/shortener ./cmd/shortener

FROM alpine:3.21

RUN adduser -D -u 10001 app
USER app
COPY --from=build /out/shortener /usr/local/bin/shortener

EXPOSE 8080

# Именно exec-форма. В shell-форме бинарник запускается через оболочку,
# сигнал до него не доходит, и корректная остановка молча не работает.
CMD ["/usr/local/bin/shortener"]
__LAB_FILE_EOF__
    ;;
  "Makefile")
    cat <<'__LAB_FILE_EOF__'
.PHONY: build test race lint up down clean

build:
	go build ./...

test:
	go test ./...

race:
	go test -race ./...

lint:
	golangci-lint run

up:
	docker compose up --build -d

down:
	docker compose down -v

clean:
	go clean -testcache
__LAB_FILE_EOF__
    ;;
  "cmd/shortener/main.go")
    cat <<'__LAB_FILE_EOF__'
// Команда shortener — сервис сокращения ссылок.
package main

// TODO: прочитать конфигурацию, собрать зависимости, поднять HTTP-сервер.
func main() {
	panic("не реализовано")
}
__LAB_FILE_EOF__
    ;;
  "docker-compose.yml")
    cat <<'__LAB_FILE_EOF__'
services:
  shortener:
    build: .
    ports:
      - "8080:8080"
    # Значения берутся из окружения, если заданы, иначе используются умолчания.
    # Так конфигурацию можно менять, не правя файл.
    environment:
      HTTP_ADDR: "${HTTP_ADDR:-:8080}"
      BASE_URL: "${BASE_URL:-http://localhost:8080}"
      CODE_LENGTH: "${CODE_LENGTH:-7}"
      LOG_LEVEL: "${LOG_LEVEL:-info}"
__LAB_FILE_EOF__
    ;;
  "go.mod")
    cat <<'__LAB_FILE_EOF__'
module urlshortener

go 1.27
__LAB_FILE_EOF__
    ;;
  "internal/config/config.go")
    cat <<'__LAB_FILE_EOF__'
// Package config читает конфигурацию сервиса из переменных окружения.
package config

// TODO: разобрать переменные окружения из задания (шаг «Конфигурация»),
// проверить значения и вернуть ошибку, если они бессмысленны.
// Сервис с некорректной конфигурацией стартовать не должен.
__LAB_FILE_EOF__
    ;;
  "internal/httpapi/api.go")
    cat <<'__LAB_FILE_EOF__'
// Package httpapi содержит HTTP-слой сервиса: маршруты, обработчики и middleware.
package httpapi

// TODO: собрать два маршрута из задания (шаг «HTTP»), реализовать обработчики
// и запись в логи.
__LAB_FILE_EOF__
    ;;
  "internal/storage/storage.go")
    cat <<'__LAB_FILE_EOF__'
// Package storage хранит соответствие короткого кода и исходной ссылки.
//
// Интерфейс, ошибка и сигнатура New() — часть контракта практики.
// Менять их нельзя: против них компилируется тест, который приходит
// из пайплайна курса.
package storage

import "errors"

// ErrCodeTaken возвращается, если код уже занят другой ссылкой.
var ErrCodeTaken = errors.New("code already taken")

// Storage — хранилище ссылок.
type Storage interface {
	// Save сохраняет ссылку под кодом.
	// Проверить, что код свободен, и записать нужно одной неделимой операцией.
	Save(code, url string) error

	// Lookup возвращает ссылку по коду.
	Lookup(code string) (url string, ok bool)
}

// New создаёт хранилище в памяти.
//
// TODO: вернуть реализацию Storage. Она должна выдерживать параллельные
// вызовы: net/http обрабатывает каждый запрос в отдельной горутине.
func New() Storage {
	panic("storage.New: не реализовано")
}
__LAB_FILE_EOF__
    ;;
  "internal/storage/storage_test.go")
    cat <<'__LAB_FILE_EOF__'
package storage

import (
	"errors"
	"testing"
)

func TestSaveAndLookup(t *testing.T) {
	s := New()

	if err := s.Save("abc", "https://example.com"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, ok := s.Lookup("abc")
	if !ok {
		t.Fatal("Lookup: код не найден")
	}
	if got != "https://example.com" {
		t.Fatalf("Lookup = %q, ожидалось %q", got, "https://example.com")
	}
}

func TestSaveRejectsTakenCode(t *testing.T) {
	s := New()

	if err := s.Save("abc", "https://first.example"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := s.Save("abc", "https://second.example"); !errors.Is(err, ErrCodeTaken) {
		t.Fatalf("Save на занятый код вернул %v, ожидалась ErrCodeTaken", err)
	}

	if got, _ := s.Lookup("abc"); got != "https://first.example" {
		t.Fatalf("занятый код перезаписан: %q", got)
	}
}

func TestLookupMissing(t *testing.T) {
	if _, ok := New().Lookup("nope"); ok {
		t.Fatal("Lookup: несуществующий код найден")
	}
}
__LAB_FILE_EOF__
    ;;
  *) echo "неизвестный файл: $1" >&2; return 1 ;;
  esac
}

red()   { printf '\033[31m%s\033[0m\n' "$*"; }
green() { printf '\033[32m%s\033[0m\n' "$*"; }

install_all() {
  if [ -e "$LAB_DIR" ] && [ -n "$(ls -A "$LAB_DIR" 2>/dev/null)" ]; then
    red "каталог $LAB_DIR существует и не пуст"
    echo "Каркас разворачивается только в пустой каталог, чтобы ничего не затереть."
    echo "Сверить содержимое с текущей версией каркаса: $0 --check"
    exit 1
  fi

  for f in "${FILES[@]}"; do
    mkdir -p "$LAB_DIR/$(dirname "$f")"
    emit "$f" > "$LAB_DIR/$f"
  done
  printf '%s\n' "$STARTER_VERSION" > "$LAB_DIR/.starter-version"

  green "каркас развёрнут в $LAB_DIR (версия $STARTER_VERSION, собран $PACKED_AT)"
  echo
  echo "Дальше:"
  echo "  cd $LAB_DIR && make test     # тесты падают, пока не написано хранилище"
  echo "  cd $LAB_DIR && make up       # поднять сервис"
}

check_all() {
  [ -d "$LAB_DIR" ] || { red "каталога $LAB_DIR нет"; exit 1; }

  local have differs=0 missing=0
  have=$(cat "$LAB_DIR/.starter-version" 2>/dev/null || echo "неизвестна")
  echo "версия каркаса в $LAB_DIR: $have"
  echo "версия в этом скрипте:     $STARTER_VERSION"
  echo

  for f in "${FILES[@]}"; do
    if [ ! -f "$LAB_DIR/$f" ]; then
      echo "  нет      $f"
      missing=$((missing + 1))
    elif ! emit "$f" | diff -q - "$LAB_DIR/$f" >/dev/null 2>&1; then
      local mark="изменён"
      case " ${INFRA[*]:-} " in *" $f "*) mark="изменён (служебный)" ;; esac
      echo "  $mark  $f"
      differs=$((differs + 1))
    fi
  done

  if [ $differs -eq 0 ] && [ $missing -eq 0 ]; then
    green "расхождений нет"
    return 0
  fi
  echo
  echo "Изменённые файлы с кодом — это нормально, работа в них и ведётся."
}

# Режим намеренно не описан ни в задании, ни в справке скрипта: неизвестно,
# понадобится ли он вообще. Если каркас придётся править после выдачи —
# рассказать о нём можно будет отдельно.
update_infra() {
  [ -d "$LAB_DIR" ] || { red "каталога $LAB_DIR нет"; exit 1; }
  if [ ${#INFRA[@]} -eq 0 ]; then
    red "для этой практики список сопровождаемых файлов не задан"
    exit 1
  fi

  local changed=0
  for f in "${INFRA[@]}"; do
    if ! emit "$f" | diff -q - "$LAB_DIR/$f" >/dev/null 2>&1; then
      mkdir -p "$LAB_DIR/$(dirname "$f")"
      emit "$f" > "$LAB_DIR/$f"
      echo "  обновлён $f"
      changed=$((changed + 1))
    fi
  done
  printf '%s\n' "$STARTER_VERSION" > "$LAB_DIR/.starter-version"

  if [ $changed -eq 0 ]; then
    green "служебные файлы уже актуальны"
  else
    green "обновлено файлов: $changed, версия $STARTER_VERSION"
  fi
  echo "Файлы с вашим кодом не тронуты."
}

case "${1:-install}" in
  install)  install_all ;;
  --check)  check_all ;;
  --update) update_infra ;;
  -h|--help)
    echo "Использование: $0 [--check]"
    echo "  без аргументов  развернуть каркас в \$LAB_DIR (по умолчанию $LAB_DIR)"
    echo "  --check         показать расхождения с текущей версией каркаса"
    ;;
  *) echo "неизвестный аргумент: $1" >&2; exit 2 ;;
esac
