#!/usr/bin/env bash
# Automated w9s demo capture — records mock mode with scripted navigation.
# No manual interaction needed.
#
# Usage: ./demos/auto-capture.sh
# Output: demos/w9s-auto.cast → docs/screenshots/w9s-demo.gif
#
# Requires: asciinema, agg, expect (or python3)

set -euo pipefail
cd "$(dirname "$0")/.."

CAST_FILE="demos/w9s-auto.cast"
GIF_FILE="docs/screenshots/w9s-demo.gif"

mkdir -p demos docs/screenshots

# Check for expect
if ! command -v expect &>/dev/null; then
  echo "Installing expect..."
  sudo dnf install -y expect 2>/dev/null || sudo apt-get install -y expect 2>/dev/null || true
fi

echo "Recording automated w9s demo..."

# Use expect to drive the TUI with scripted keystrokes
# Create the expect script as a file (avoids quoting issues)
cat > /tmp/w9s-demo-expect.exp <<'EXPECT'
set timeout 30
spawn env W9S_ENABLE_MOCK=1 ./build/w9s --mock
# Wait for TUI to render
expect {
    "w9s.sh" { }
    timeout { puts "Timeout waiting for TUI"; exit 1 }
}
sleep 2
# Dashboard view is default — pause to show it
sleep 2
# Switch to Nodes view (Tab)
send "\t"
sleep 2
# Switch to Profiles view (Tab)
send "\t"
sleep 2
# Switch to Images view (Tab)
send "\t"
sleep 2
# Switch to Overlays view (Tab)
send "\t"
sleep 2
# Switch to Power view (Tab)
send "\t"
sleep 2
# Switch to Help view (?)
send "?"
sleep 2
# Quit
send "q"
expect eof
EXPECT

asciinema rec \
  --cols 120 \
  --rows 35 \
  --overwrite \
  --title "w9s — Terminal UI for Warewulf" \
  --command "expect /tmp/w9s-demo-expect.exp" \
  "$CAST_FILE"

echo "Recording saved to $CAST_FILE"

# Convert to GIF
if command -v agg &>/dev/null; then
  echo "Converting to GIF..."
  agg --cols 120 --rows 35 --speed 1.5 "$CAST_FILE" "$GIF_FILE"
  echo "GIF saved to $GIF_FILE"
  ls -lh "$GIF_FILE"
fi

echo ""
echo "Done! Files:"
echo "  Cast: $CAST_FILE"
echo "  GIF:  $GIF_FILE"
