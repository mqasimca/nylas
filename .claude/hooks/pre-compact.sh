#!/bin/bash
# Save session learnings before context compaction
# This hook runs when Claude's context window is about to be compacted

set -euo pipefail

DATE=$(date +%Y-%m-%d)
DIARY_DIR="$HOME/.claude/memory/diary"
mkdir -p "$DIARY_DIR"

# Count existing sessions for today
SESSION_NUM=$(ls "$DIARY_DIR"/$DATE-session-*.md 2>/dev/null | wc -l || echo "0")
SESSION_NUM=$((SESSION_NUM + 1))

DIARY_FILE="$DIARY_DIR/$DATE-session-$SESSION_NUM.md"

# Notify about compaction
echo "=================================================="
echo "CONTEXT COMPACTION IMMINENT"
echo "=================================================="
echo ""
echo "Your context window is being compacted."
echo "Any learnings not saved will be lost."
echo ""
echo "Diary file ready: $DIARY_FILE"
echo ""
echo "Use /diary to capture learnings before they're lost."
echo "=================================================="

exit 0
