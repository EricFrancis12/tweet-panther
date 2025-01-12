#!/bin/bash

SCRIPT_DIR="$(dirname "$(realpath "$0")")"
ROOR_DIR="$(dirname "$SCRIPT_DIR")"

ENV_FILE_PATH="$ROOR_DIR/.env"
ENV_EXAMPLE_FILE_PATH="$ROOR_DIR/.env.example"

if [ ! -e "$ENV_EXAMPLE_FILE_PATH" ]; then
    echo "error: .env.example file is missing"
    exit 1
fi

if [ -e "$ENV_FILE_PATH" ]; then
    echo ".env file already exists. Override it? (Y/N)"
    read CHOICE
    if [ "$CHOICE" != "Y" ] && [ "$CHOICE" != "y" ]; then
        echo "Abborted .env file creation"
        exit 0
    fi
fi

ENV_EXAMPLE_FILE_CONTENT=$(<"$ENV_EXAMPLE_FILE_PATH")
echo -e "$ENV_EXAMPLE_FILE_CONTENT" > "$ENV_FILE_PATH"

echo "Created new .env file at project root"
