#!/bin/bash

# Значения по умолчанию
PORT=80
BODY=""
CODE=""

# Функция помощи
show_help() {
  cat << EOF
Usage: $0 [-p port] [-b body] [-c code] [-h]

Options:
  -p  Port to listen on (default: 8080)
  -b  Response body text
  -c  Response HTTP code (default: 200 if -b specified)
  -h  Show this help

Examples:
  # Only listen, no response
  $0 -p 8080

  # Respond with 200 OK and text
  $0 -p 8080 -b "Success"

  # Respond with 404 and no body
  $0 -p 8080 -c 404

  # Respond with 201 and text
  $0 -p 8080 -b "Created" -c 201
EOF
  exit 0
}

# Парсинг аргументов
while getopts "p:b:c:h" opt; do
  case $opt in
    p) PORT="$OPTARG" ;;
    b) BODY="$OPTARG" ;;
    c) CODE="$OPTARG" ;;
    h) show_help ;;
    *) show_help ;;
  esac
done

# Если указан body или code, формируем ответ
if [ -n "$BODY" ] || [ -n "$CODE" ]; then
  # Код по умолчанию 200
  [ -z "$CODE" ] && CODE=200
  
  # Длина тела
  CONTENT_LENGTH=${#BODY}
  
  RESPONSE="HTTP/1.1 $CODE OK\r\nContent-Length: $CONTENT_LENGTH\r\n\r\n$BODY"
  
  echo "=== HTTP Listener on port $PORT (responding with $CODE) ==="
  while true; do
    echo -e "\n--- New request ---"
    echo -e "$RESPONSE" | nc -l -p "$PORT"
  done
else
  echo "=== HTTP Listener on port $PORT (no response mode) ==="
  while true; do
    echo -e "\n--- New request ---"
    nc -l -p "$PORT"
  done
fi