package client

import (
	"time"
	"fmt"
)

type streamMetadata struct {
	maxCount       *int
	maxAge         *time.Duration
	truncateBefore *int
	cacheControl   *time.Duration
	acl            *streamAcl
	customMetadata map[string]interface{}
}

func newStreamMetadata(
	maxCount *int,
	maxAge *time.Duration,
	truncateBefore *int,
	cacheControl *time.Duration,
	acl *streamAcl,
	customMetadata map[string]interface{},
) *streamMetadata {
	if maxCount != nil && *maxCount <= 0 {
		panic(fmt.Sprintf("maxCount should be positive value"))
	}
	if maxAge != nil && *maxAge <= 0 {
		panic(fmt.Sprintf("maxAge should be positive value"))
	}
	if truncateBefore != nil && *truncateBefore < 0 {
		panic(fmt.Sprintf("maxCount should be non-negative value"))
	}
	if cacheControl != nil && *cacheControl <= 0 {
		panic(fmt.Sprintf("maxAge should be positive value"))
	}
	return &streamMetadata{
		maxCount:       maxCount,
		maxAge:         maxAge,
		truncateBefore: truncateBefore,
		cacheControl:   cacheControl,
		acl:            acl,
		customMetadata: customMetadata,
	}
}

func CreateStreamMetadata(
	maxCount *int,
	maxAge *time.Duration,
	truncateBefore *int,
	cacheControl *time.Duration,
	acl *streamAcl,
) *streamMetadata {
	return newStreamMetadata(maxCount, maxAge, truncateBefore, cacheControl, acl, nil)
}
