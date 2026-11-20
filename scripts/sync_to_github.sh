#!/bin/bash
# Complete sync of factorio_optimizer_go to GitHub
set -euo pipefail
REPO="VIRUSGAMING64/factorio_optimizer_go"
DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$DIR"
# Ensure recipes note
test -f data/recipes.json || curl -L -o data/recipes.json \
  https://raw.githubusercontent.com/SeezoCode/factorio_optimizer_v2/main/data/recipes.json || true
git init -b main 2>/dev/null || true
git remote remove origin 2>/dev/null || true
git remote add origin "https://github.com/${REPO}.git"
git add -A
git status -sb
git commit -m "chore: complete sync of Factorio optimizer Go port" || true
git branch -M main
git push -u origin main --force
echo "Done. See https://github.com/${REPO}"
