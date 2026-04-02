#!/usr/bin/env bash
# Record w9s TUI demo in mock mode using asciinema.
# Usage: ./demos/record-mock.sh
# Output: demos/w9s-demo.cast (asciinema recording)
#         docs/screenshots/w9s-demo.gif (animated GIF)
#
# This uses a helper script that sends keystrokes to w9s
# to navigate through all views.

set -euo pipefail
cd "$(dirname "$0")/.."

CAST_FILE="demos/w9s-demo.cast"
GIF_FILE="docs/screenshots/w9s-demo.gif"

mkdir -p demos docs/screenshots

echo "Recording w9s demo in mock mode..."
echo "The TUI will launch. Navigate through views with Tab, then press q to quit."
echo ""

# Record with asciinema — interactive session
W9S_ENABLE_MOCK=1 asciinema rec \
  --cols 120 \
  --rows 35 \
  --title "w9s — Terminal UI for Warewulf" \
  --command "./build/w9s --mock" \
  "$CAST_FILE"

echo ""
echo "Recording saved to $CAST_FILE"

# Convert to GIF
if command -v agg &>/dev/null; then
  echo "Converting to GIF..."
  agg --cols 120 --rows 35 "$CAST_FILE" "$GIF_FILE"
  echo "GIF saved to $GIF_FILE"
else
  echo "Install agg to convert to GIF: https://github.com/asciinema/agg"
fi
