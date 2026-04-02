# w9s demo recording
# Install VHS: https://github.com/charmbracelet/vhs
# Usage: cd w9s && make build && vhs docs/screenshots/w9s-demo.tape

Output docs/screenshots/w9s-demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 700
Set Framerate 15
Set Theme "Dracula"
Set TypingSpeed 80ms

# Launch w9s in mock mode
Type "W9S_ENABLE_MOCK=1 ./build/w9s --mock"
Enter
Sleep 2s

# Dashboard view (loads first)
Sleep 3s

# Nodes view
Type "2"
Sleep 2s

# Filter nodes
Type "/"
Sleep 500ms
Type "compute"
Sleep 1.5s
Escape
Sleep 500ms

# Sort by column
Type "s"
Sleep 1s
Type "s"
Sleep 1s

# View node detail
Enter
Sleep 2s
Escape
Sleep 500ms

# Profiles view
Type "3"
Sleep 2.5s

# Images view
Type "4"
Sleep 2.5s

# Overlays view
Type "5"
Sleep 2s

# Drill into an overlay
Enter
Sleep 2s
Escape
Sleep 500ms

# Power view
Type "6"
Sleep 2s

# Back to dashboard
Type "1"
Sleep 2s

# Quit
Type "q"
Sleep 500ms
