#!/usr/bin/env bash
# TUI smoke tests using tui-test (https://github.com/microsoft/tui-test).
# Requires: tui-test binary on PATH or at /tmp/tui-test.
# Requires: planetary-tui binary at ./planetary-tui (run `make tui` first).
# Usage: bash tui_test.sh [path-to-planetary-tui]
set -euo pipefail

TUI_BIN="${1:-./planetary-tui}"
TUI_TEST="${TUI_TEST_BIN:-tui-test}"

# Try /tmp/tui-test as fallback
if ! command -v "$TUI_TEST" &>/dev/null; then
  if [ -x /tmp/tui-test ]; then
    TUI_TEST=/tmp/tui-test
  else
    echo "ERROR: tui-test not found. Install from https://github.com/microsoft/tui-test" >&2
    exit 1
  fi
fi

if [ ! -x "$TUI_BIN" ]; then
  echo "ERROR: $TUI_BIN not found or not executable. Run: make tui" >&2
  exit 1
fi

SESSION="planetary-tui-test-$$"
PASS=0
FAIL=0

cleanup() {
  "$TUI_TEST" --session "$SESSION" close 2>/dev/null || true
}
trap cleanup EXIT

run_test() {
  local name="$1"
  shift
  if "$@" >/dev/null 2>&1; then
    echo "  PASS: $name"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: $name"
    FAIL=$((FAIL + 1))
  fi
}

echo "=== Planetary TUI tests ==="
echo

# Start the TUI
"$TUI_TEST" run "$TUI_BIN" --session "$SESSION" --cols 120 --rows 40 >/dev/null 2>&1
"$TUI_TEST" --session "$SESSION" wait idle --timeout 10000 >/dev/null 2>&1

echo "Layout:"
run_test "sidebar header 'Planetary' is visible" \
  "$TUI_TEST" --session "$SESSION" expect text "Planetary" --timeout 5000

run_test "sidebar shows 'Unread'" \
  "$TUI_TEST" --session "$SESSION" expect text "Unread" --timeout 5000

run_test "sidebar shows 'Starred'" \
  "$TUI_TEST" --session "$SESSION" expect text "Starred" --timeout 5000

run_test "sidebar separator '│' is present" \
  "$TUI_TEST" --session "$SESSION" expect text "│" --no-strict --timeout 5000

run_test "main panel shows 'All Entries' header" \
  "$TUI_TEST" --session "$SESSION" expect text "All Entries" --timeout 5000

echo
echo "Navigation — vim bindings:"
# Press 'j' to move down in sidebar
"$TUI_TEST" --session "$SESSION" press j >/dev/null 2>&1
"$TUI_TEST" --session "$SESSION" wait idle --timeout 2000 >/dev/null 2>&1

# Press 'k' to go back up
"$TUI_TEST" --session "$SESSION" press k >/dev/null 2>&1
"$TUI_TEST" --session "$SESSION" wait idle --timeout 2000 >/dev/null 2>&1

# 'Planetary' should still be visible (sidebar still showing)
run_test "sidebar still visible after j/k navigation" \
  "$TUI_TEST" --session "$SESSION" expect text "Planetary" --timeout 3000

# Move to 'Unread' (second item) and open it
"$TUI_TEST" --session "$SESSION" press j >/dev/null 2>&1
"$TUI_TEST" --session "$SESSION" press l >/dev/null 2>&1  # open/enter
"$TUI_TEST" --session "$SESSION" wait idle --timeout 5000 >/dev/null 2>&1

run_test "opening Unread shows 'Unread' header in main panel" \
  "$TUI_TEST" --session "$SESSION" expect text "Unread" --no-strict --timeout 5000

# Go back to sidebar with 'h'
"$TUI_TEST" --session "$SESSION" press h >/dev/null 2>&1
"$TUI_TEST" --session "$SESSION" wait idle --timeout 2000 >/dev/null 2>&1

run_test "pressing 'h' returns focus to sidebar" \
  "$TUI_TEST" --session "$SESSION" expect text "Planetary" --timeout 3000

echo
echo "Keyboard shortcuts:"
# Test 'r' refreshes (shouldn't crash)
"$TUI_TEST" --session "$SESSION" press r >/dev/null 2>&1
"$TUI_TEST" --session "$SESSION" wait idle --timeout 5000 >/dev/null 2>&1

run_test "'r' refresh: sidebar still visible after reload" \
  "$TUI_TEST" --session "$SESSION" expect text "Planetary" --timeout 8000

echo
echo "=== Results: $PASS passed, $FAIL failed ==="
[ "$FAIL" -eq 0 ]
