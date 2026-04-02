#!/bin/bash
# Scripted w9s demo for asciinema recording
# Uses a FIFO to feed keystrokes to tview with delays
set -e

cd "$(dirname "$0")/../.."

export W9S_ENABLE_MOCK=1
export TERM=xterm-256color
export COLUMNS=120
export LINES=35

FIFO=$(mktemp -u /tmp/w9s-demo-XXXXXX)
mkfifo "$FIFO"

# Feed keystrokes in background
(
    sleep 3          # Dashboard loads

    echo -n "2"      # Nodes view
    sleep 2

    echo -n "/"      # Open filter
    sleep 0.5
    echo -n "compute"
    sleep 1.5
    printf '\x1b'    # Escape - close filter
    sleep 0.5

    echo -n "s"      # Sort
    sleep 1
    echo -n "s"      # Sort again
    sleep 1

    printf '\r'       # Enter - detail
    sleep 2
    printf '\x1b'    # Escape - back
    sleep 0.5

    echo -n "3"      # Profiles
    sleep 2.5

    echo -n "4"      # Images
    sleep 2.5

    echo -n "5"      # Overlays
    sleep 2
    printf '\r'       # Enter - drill in
    sleep 2
    printf '\x1b'    # Escape - back
    sleep 0.5

    echo -n "6"      # Power
    sleep 2

    echo -n "1"      # Dashboard
    sleep 2

    echo -n "q"      # Quit
    sleep 0.5
) > "$FIFO" &
KEYS_PID=$!

# Run w9s reading from the FIFO
./build/w9s --mock < "$FIFO"

wait $KEYS_PID 2>/dev/null
rm -f "$FIFO"
