# Fix: r.lanes.cap undefined

## Error
```
router.go:430: r.lanes.cap undefined (type *LaneTracker has no field or method cap)
```

## Cause
`capPerLane` is an **unexported** field. Use the public method:

```go
// BAD
if rateShare > r.lanes.cap {
    rateShare = r.lanes.cap * 0.5
}

// GOOD
if rateShare > r.lanes.CapPerLane() {
    rateShare = r.lanes.CapPerLane() * 0.5
}
```

In `internal/route/router.go` around the belt placement loop (pathToEntities).
