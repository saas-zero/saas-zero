package id

import (
	"strconv"
	"strings"
)

// ToString renders a snowflake int64 ID as a decimal string so it can be
// safely transported to JavaScript, which loses precision for int64 (> 2^53).
// Every ID returned to the frontend must go through this (or ToStrings).
func ToString(id int64) string {
	return strconv.FormatInt(id, 10)
}

// Parse converts a string ID (as sent by the frontend via `idStr`) back to the
// int64 expected by the gRPC/storage layer. Invalid or empty input yields 0.
func Parse(s string) int64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return v
}

// ToStrings converts a slice of int64 IDs to their lossless string form.
func ToStrings(ids []int64) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, ToString(id))
	}
	return out
}

// ParseStrings converts a slice of string IDs back to int64, skipping empties.
func ParseStrings(ss []string) []int64 {
	out := make([]int64, 0, len(ss))
	for _, s := range ss {
		if v := Parse(s); v != 0 {
			out = append(out, v)
		}
	}
	return out
}
