#!/usr/bin/env bash
# Automated w9s demo capture — records mock mode with scripted navigation.
# Showcases: Dashboard, Nodes (sort, filter, detail, export), Profiles,
# Images, Overlays (drill-down), Cluster Switcher, Help.
#
# Usage: ./demos/auto-capture.sh
# Output: demos/w9s-auto.cast → docs/screenshots/w9s-demo.gif
#
# Requires: asciinema, agg, expect

set -euo pipefail
cd "$(dirname "$0")/.."

CAST_FILE="demos/w9s-auto.cast"
GIF_FILE="docs/screenshots/w9s-demo.gif"

mkdir -p demos docs/screenshots

echo "Recording automated w9s demo..."

cat > /tmp/w9s-demo-expect.exp <<'EXPECT'
set timeout 30
spawn env W9S_ENABLE_MOCK=1 ./build/w9s --mock
expect {
    "w9s.sh" { }
    timeout { puts "Timeout waiting for TUI"; exit 1 }
}

# Dashboard — pause to show cluster overview
sleep 3

# Nodes view
send "2"
sleep 2

# Sort by column
send "s"
sleep 1.5
send "s"
sleep 1.5

# Filter
send "/"
sleep 0.5
send "gpu"
sleep 1.5
send "\033"
sleep 1

# Detail view
send "\r"
sleep 2
send "\033"
sleep 1

# Export CSV
send "x"
sleep 1.5
send "\033"
sleep 1.5

# Profiles view
send "3"
sleep 2

# Profile detail
send "\r"
sleep 2
send "\033"
sleep 1

# Images view
send "4"
sleep 2

# Overlays view
send "5"
sleep 2

# Drill into overlay files
send "\r"
sleep 2
send "\033"
sleep 1

# Cluster switcher
send "C"
sleep 2
send "\033"
sleep 1

# Help view
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

if command -v agg &>/dev/null; then
  echo "Converting to GIF..."
  agg --cols 120 --rows 35 --speed 2 "$CAST_FILE" "$GIF_FILE"
  echo "GIF saved to $GIF_FILE"
  ls -lh "$GIF_FILE"
fi

echo ""
echo "Done! Files:"
echo "  Cast: $CAST_FILE"
echo "  GIF:  $GIF_FILE"
