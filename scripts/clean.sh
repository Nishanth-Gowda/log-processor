#!/bin/bash

# Clean script for log processor
# Usage: ./scripts/clean.sh [logs|offsets|all]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

LOGS_DIR="$PROJECT_DIR/logs"
OFFSETS_DIR="$PROJECT_DIR/offsets"

clean_logs() {
    echo "🧹 Cleaning log files..."
    if [ -d "$LOGS_DIR" ]; then
        rm -rf "$LOGS_DIR"/*
        echo "   Removed contents of $LOGS_DIR"
    else
        echo "   Logs directory does not exist"
    fi
}

clean_offsets() {
    echo "🧹 Cleaning offset files..."
    if [ -d "$OFFSETS_DIR" ]; then
        rm -rf "$OFFSETS_DIR"/*
        echo "   Removed contents of $OFFSETS_DIR"
    else
        echo "   Offsets directory does not exist"
    fi
}

clean_all() {
    clean_logs
    clean_offsets
    echo "✅ All cleaned!"
}

# Show help
show_help() {
    echo "Usage: $0 [logs|offsets|all]"
    echo ""
    echo "Commands:"
    echo "  logs     - Remove all log files from logs/"
    echo "  offsets  - Remove all offset files from offsets/"
    echo "  all      - Remove both logs and offsets"
    echo ""
    echo "Without arguments, shows this help."
}

# Main
case "${1:-help}" in
    logs)
        clean_logs
        ;;
    offsets)
        clean_offsets
        ;;
    all)
        clean_all
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo "Unknown option: $1"
        show_help
        exit 1
        ;;
esac
