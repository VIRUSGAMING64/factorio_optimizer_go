#!/usr/bin/env bash
# Full sync of local tree to GitHub (requires gh auth login)
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
git init 2>/dev/null || true
git remote remove origin 2>/dev/null || true
git remote add origin https://github.com/VIRUSGAMING64/factorio_optimizer_go.git
git add -A
git status -sb
git commit -m "sync: complete Factorio optimizer Go port" || true
git branch -M main
git pull origin main --rebase --allow-unrelated-histories || true
git push -u origin main
