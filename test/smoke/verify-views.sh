#!/bin/bash
# Automated tmux-based smoke test for w9s TUI.
# Validates every view renders correctly with real or mock data.
#
# Usage:
#   ./test/smoke/verify-views.sh            # real server (uses ~/.config/w9s/config.yaml)
#   W9S_ENABLE_MOCK=1 ./test/smoke/verify-views.sh mock  # mock mode
#
# Requires: tmux, w9s binary at ./build/w9s
set -euo pipefail

cd "$(dirname "$0")/../.."

BINARY="./build/w9s"
SESSION="w9stest"
PASS=0; FAIL=0; TOTAL=0

if [ "${1:-}" = "mock" ]; then
  export W9S_ENABLE_MOCK=1
  CMD="$BINARY --mock"
  echo "Running smoke tests in MOCK mode"
else
  CMD="$BINARY"
  echo "Running smoke tests against real server"
fi

assert() {
  local name="$1" pattern="$2" output="$3"
  TOTAL=$((TOTAL+1))
  if echo "$output" | grep -q "$pattern"; then
    echo "  PASS: $name"; PASS=$((PASS+1))
  else
    echo "  FAIL: $name (expected '$pattern')"; FAIL=$((FAIL+1))
  fi
}

assert_not() {
  local name="$1" pattern="$2" output="$3"
  TOTAL=$((TOTAL+1))
  if echo "$output" | grep -q "$pattern"; then
    echo "  FAIL: $name (unexpected '$pattern')"; FAIL=$((FAIL+1))
  else
    echo "  PASS: $name"; PASS=$((PASS+1))
  fi
}

capture() { tmux capture-pane -t "$SESSION" -p 2>/dev/null; }

# Cleanup any previous session
tmux kill-session -t "$SESSION" 2>/dev/null || true

# Launch w9s
tmux new-session -d -s "$SESSION" -x 120 -y 40 "$CMD"
sleep 3

echo ""
echo "--- Dashboard (Tab 1) ---"
OUT=$(capture)
assert "Header: w9s.sh branding" "w9s.sh" "$OUT"
assert "Dashboard: cluster summary" "Cluster Summary\|Total Nodes\|Nodes" "$OUT"

echo ""
echo "--- Nodes (Tab 2) ---"
tmux send-keys -t "$SESSION" "2"; sleep 2
OUT=$(capture)
assert "Nodes: table has NAME column" "NAME" "$OUT"
assert "Nodes: has IP data" "192.168\|10\.0\." "$OUT"
assert "Nodes: has MAC data" "00:11:22\|44:" "$OUT"
assert "Nodes: has Status" "Built\|Stale\|Ready\|No Image" "$OUT"

echo ""
echo "--- Export ---"
# x should export current table to CSV
rm -f /tmp/w9s-export-*.csv
tmux send-keys -t "$SESSION" "x"; sleep 2
ls /tmp/w9s-export-*.csv >/dev/null 2>&1
if [ $? -eq 0 ]; then
  TOTAL=$((TOTAL+1)); PASS=$((PASS+1)); echo "  PASS: Export: CSV file created"
  EXPORTFILE=$(ls /tmp/w9s-export-*.csv | head -1)
  if head -1 "$EXPORTFILE" | grep -qi "name"; then
    TOTAL=$((TOTAL+1)); PASS=$((PASS+1)); echo "  PASS: Export: CSV has headers"
  else
    TOTAL=$((TOTAL+1)); FAIL=$((FAIL+1)); echo "  FAIL: Export: CSV missing headers"
  fi
  rm -f /tmp/w9s-export-*.csv
else
  TOTAL=$((TOTAL+1)); FAIL=$((FAIL+1)); echo "  FAIL: Export: no CSV file created"
  TOTAL=$((TOTAL+1)); FAIL=$((FAIL+1)); echo "  FAIL: Export: CSV headers (skipped)"
fi
# Flash message should appear immediately (before closing detail pane)
sleep 1
OUT=$(capture)
assert "Flash: success message shown" "✓\|Exported" "$OUT"
# Close detail pane and wait for auto-clear
tmux send-keys -t "$SESSION" Escape; sleep 5
OUT=$(capture)
assert "Flash: hints restored after clear" "Filter\|Sort\|Detail" "$OUT"

echo ""
echo "--- Node Add Form ---"
tmux send-keys -t "$SESSION" "a"; sleep 1
OUT=$(capture)
assert "Node Add: form appears" "Add Node\|Save\|Cancel" "$OUT"
# Cancel with Escape
tmux send-keys -t "$SESSION" Escape; sleep 1
OUT=$(capture)
assert "Node Add: cancelled, back to table" "NAME" "$OUT"
assert_not "Node Add: not on help" "w9s Help" "$OUT"

echo ""
echo "--- Node Edit Form ---"
# Select first node, press e to edit
tmux send-keys -t "$SESSION" "e"; sleep 1
OUT=$(capture)
assert "Node Edit: form appears" "Edit Node\|Save\|Cancel" "$OUT"
# Cancel
tmux send-keys -t "$SESSION" Escape; sleep 1
OUT=$(capture)
assert "Node Edit: cancelled, back to table" "NAME" "$OUT"

echo ""
echo "--- Column Sort ---"
# s should sort by next column, header should show sort indicator
tmux send-keys -t "$SESSION" "s"; sleep 1
OUT=$(capture)
assert "Sort: sort indicator shown" "▲\|▼\|↑\|↓\|▴\|▾" "$OUT"
# S (shift+s) should reverse sort
tmux send-keys -t "$SESSION" "S"; sleep 1
OUT=$(capture)
assert "Sort: reverse indicator" "▼\|↓\|▾" "$OUT"
# Escape resets sort
tmux send-keys -t "$SESSION" Escape; sleep 1
OUT=$(capture)
assert "Sort: back to default (NAME first)" "NAME" "$OUT"

echo ""
echo "--- Selection Persistence ---"
# Move down 3 rows (to 4th node)
tmux send-keys -t "$SESSION" Down Down Down; sleep 1
# Wait for auto-refresh (6s > 5s interval)
sleep 6
# Press Enter — detail should show the 4th node, not the 1st
tmux send-keys -t "$SESSION" Enter; sleep 1
OUT=$(capture)
assert "Selection: survived refresh (not first node)" "compute-04\|compute-03\|gpu" "$OUT"
assert_not "Selection: not reset to compute-01" "Node: compute-01" "$OUT"
tmux send-keys -t "$SESSION" Escape; sleep 1

echo ""
echo "--- Detail Pane ---"
tmux send-keys -t "$SESSION" Enter; sleep 1
OUT=$(capture)
assert "Detail: shows node info" "Node:" "$OUT"
assert "Detail: left-aligned content" "Profiles:" "$OUT"
assert_not "Detail: not a centered modal" "Close" "$OUT"
tmux send-keys -t "$SESSION" Escape; sleep 1
OUT=$(capture)
assert "Detail: Escape returns to nodes" "NAME" "$OUT"
assert_not "Detail: not on help after Escape" "w9s Help" "$OUT"
# Reopen and close via Enter
tmux send-keys -t "$SESSION" Enter; sleep 1
tmux send-keys -t "$SESSION" Enter; sleep 1
OUT=$(capture)
assert "Detail: Enter also closes pane" "NAME" "$OUT"
assert_not "Detail: not on help after Enter close" "w9s Help" "$OUT"

echo ""
echo "--- Filter ---"
tmux send-keys -t "$SESSION" "/"; sleep 1
OUT=$(capture)
assert "Filter: opens on /" "Filter:" "$OUT"
tmux send-keys -t "$SESSION" "gpu"; sleep 1
OUT=$(capture)
assert "Filter: text entered" "gpu" "$OUT"
tmux send-keys -t "$SESSION" Escape; sleep 1
OUT=$(capture)
assert_not "Filter: hidden after Escape" "Filter:" "$OUT"

echo ""
echo "--- Profiles (Tab 3) ---"
tmux send-keys -t "$SESSION" "3"; sleep 2
OUT=$(capture)
assert "Profiles: default visible" "default" "$OUT"

echo ""
echo "--- Profile Detail ---"
tmux send-keys -t "$SESSION" Enter; sleep 1
OUT=$(capture)
assert "Profile Detail: detail pane opens" "Profile:\|Init:\|Root:" "$OUT"
tmux send-keys -t "$SESSION" Escape; sleep 1
OUT=$(capture)
assert "Profile Detail: returns to profiles after Escape" "SYSTEM OVERLAYS\|NAME" "$OUT"
assert_not "Profile Detail: not on help after Escape" "w9s Help" "$OUT"

echo ""
echo "--- Profile Add ---"
tmux send-keys -t "$SESSION" "a"; sleep 1
OUT=$(capture)
assert "Profile Add: form appears" "Add Profile\|Save\|Cancel" "$OUT"
tmux send-keys -t "$SESSION" Escape; sleep 1
OUT=$(capture)
assert "Profile Add: cancelled, back to table" "NAME" "$OUT"

echo ""
echo "--- Profile Delete ---"
tmux send-keys -t "$SESSION" "d"; sleep 1
OUT=$(capture)
assert "Profile Delete: confirmation" "Are you sure\|Delete\|Yes\|No" "$OUT"
tmux send-keys -t "$SESSION" Escape; sleep 1
OUT=$(capture)
assert "Profile Delete: cancelled" "NAME\|default" "$OUT"

echo ""
echo "--- Image Import ---"
tmux send-keys -t "$SESSION" "4"; sleep 2
tmux send-keys -t "$SESSION" "i"; sleep 1
OUT=$(capture)
assert "Image Import: form appears" "Import Image\|Save\|Cancel" "$OUT"
tmux send-keys -t "$SESSION" Escape; sleep 1
OUT=$(capture)
assert "Image Import: cancelled, back to table" "NAME\|Images" "$OUT"

echo ""
echo "--- Images (Tab 4) ---"
tmux send-keys -t "$SESSION" "4"; sleep 2
OUT=$(capture)
assert "Images: view rendered" "NAME\|Name\|Images" "$OUT"

echo ""
echo "--- Overlays (Tab 5) ---"
tmux send-keys -t "$SESSION" "5"; sleep 2
OUT=$(capture)
assert "Overlays: wwinit visible" "wwinit" "$OUT"

echo ""
echo "--- Overlay Drill-Down ---"
# Enter → drill into overlay files
tmux send-keys -t "$SESSION" Enter; sleep 1
OUT=$(capture)
assert "Overlay: file list shown" "FILE" "$OUT"
# Enter → view file content (modal)
tmux send-keys -t "$SESSION" Enter; sleep 2
OUT=$(capture)
assert "Overlay: file content shown" "content\|#\|ww4\|conf\|warewulf\|network\|etc" "$OUT"
# Escape → close file view, return to file list
tmux send-keys -t "$SESSION" Escape; sleep 1
OUT=$(capture)
assert "Overlay: back to file list after Escape" "FILE" "$OUT"
assert_not "Overlay: not on help after file modal" "w9s Help" "$OUT"

# Template render: t key on a .ww file should prompt for node name
# First find a .ww file — navigate down to find one
tmux send-keys -t "$SESSION" Escape; sleep 1  # back to overlay list
# Go to a known overlay with .ww files (wwinit has templates)
tmux send-keys -t "$SESSION" Down Down Down Down Down Down Down Down Down Down; sleep 1  # scroll to find one
tmux send-keys -t "$SESSION" Enter; sleep 1  # drill into files
tmux send-keys -t "$SESSION" "t"; sleep 1  # render template
OUT=$(capture)
assert "Template: render form appears" "Render Template\|Save\|Cancel" "$OUT"
tmux send-keys -t "$SESSION" Escape; sleep 1

# Escape → back to overlay list
tmux send-keys -t "$SESSION" Escape; sleep 1
tmux send-keys -t "$SESSION" Escape; sleep 1
OUT=$(capture)
assert "Overlay: back to overlay list" "wwinit\|SITE\|NAME" "$OUT"
assert_not "Overlay: not on help after drill-down" "w9s Help" "$OUT"

echo ""
echo "--- Help (Tab 6) ---"
tmux send-keys -t "$SESSION" "6"; sleep 1
OUT=$(capture)
assert "Help: keyboard info" "Quit\|Filter\|Tab\|Help" "$OUT"

echo ""
echo "--- Quit ---"
tmux send-keys -t "$SESSION" "q"; sleep 1
TOTAL=$((TOTAL+1))
if ! tmux has-session -t "$SESSION" 2>/dev/null; then
  echo "  PASS: q quits app"; PASS=$((PASS+1))
else
  echo "  FAIL: q did not quit app"; FAIL=$((FAIL+1))
  tmux kill-session -t "$SESSION" 2>/dev/null
fi

echo ""
echo "================================"
echo "Results: $PASS/$TOTAL passed, $FAIL failed"
echo "================================"
[ "$FAIL" -eq 0 ] && exit 0 || exit 1
