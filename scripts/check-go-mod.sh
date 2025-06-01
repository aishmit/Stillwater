#!/bin/bash
set -euo pipefail

echo "Checking for go.mod/go.sum tidy..."
cp go.mod go.mod.bak
cp go.sum go.sum.bak

go mod tidy

if ! diff -q go.mod go.mod.bak >/dev/null || ! diff -q go.sum go.sum.bak >/dev/null; then
    echo "❌ go.mod or go.sum is not tidy. Run 'go mod tidy'"
    diff -u go.mod.bak go.mod || true
    diff -u go.sum.bak go.sum || true
    rm -f go.mod.bak go.sum.bak
    exit 1
fi

echo "✅ go.mod and go.sum are tidy"
rm -f go.mod.bak go.sum.bak