# Sync status

This repository is being synchronized from the local Factorio optimizer Go port.

## Complete local tree

The full source lives in the project workspace. To force-push everything in one step:

```bash
cd factorio_optimizer_go
bash scripts/sync_to_github.sh
# or:
git init -b main
git remote add origin https://github.com/VIRUSGAMING64/factorio_optimizer_go.git
git add -A
git commit -m "chore: complete sync"
git push -u origin main --force
```

## Note

`data/recipes.json` is large (~1.2MB). Fetch it with:

```bash
curl -L -o data/recipes.json \
  https://raw.githubusercontent.com/SeezoCode/factorio_optimizer_v2/main/data/recipes.json
```
