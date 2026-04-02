#!/bin/bash
# Record w9s demo via tmux inside asciinema
# Usage: asciinema rec --cols 120 --rows 35 -c './docs/screenshots/record-demo.sh' demo.cast
set -e

cd "$(dirname "$0")/../.."
export W9S_ENABLE_MOCK=1
SESSION="w9s-demo-$$"

# Start tmux session running w9s
tmux new-session -d -s "$SESSION" -x 120 -y 35 "./build/w9s --mock"
trap "tmux kill-session -t '$SESSION' 2>/dev/null" EXIT

# Attach to show the output in asciinema's terminal
tmux attach-session -t "$SESSION" &
ATTACH_PID=$!
sleep 0.5

# Helper to send keys with visible pauses
key() {
    tmux send-keys -t "$SESSION" "$1"
    sleep "${2:-2}"
}

# Dashboard loads
sleep 3

# 2 - Nodes
key "2" 2

# Sort
key "s" 1.5

# Detail view
key "Enter" 2
key "Escape" 1

# 3 - Profiles
key "3" 3

# 4 - Images
key "4" 3

# 5 - Overlays
key "5" 2

# Drill into overlay
key "Enter" 2
key "Escape" 1

# 6 - Power
key "6" 2

# 1 - Dashboard
key "1" 2

# Quit
key "q" 0.5

wait $ATTACH_PID 2>/dev/null || true
