package client

import (
	"encoding/json"
	"fmt"
	"time"
)

type StreamMetadata map[string]interface{}

func newStreamMetadata(
	maxCount *int,
	maxAge *time.Duration,
	truncateBefore *int,
	cacheControl *time.Duration,
	acl *StreamAcl,
	customMetadata map[string]interface{},
) StreamMetadata {
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
	m := StreamMetadata{}
	if maxCount != nil {
		m["$maxCount"] = maxCount
	}
	if maxAge != nil {
		m["$maxAge"] = maxAge
	}
	if truncateBefore != nil {
		m["$tb"] = truncateBefore
	}
	if cacheControl != nil {
		m["$cacheControl"] = cacheControl
	}
	if acl != nil {
		m["$acl"] = acl
	}
	for k, v := range customMetadata {
		m[k] = v
	}
	return m
}

func CreateStreamMetadata(
	maxCount *int,
	maxAge *time.Duration,
	truncateBefore *int,
	cacheControl *time.Duration,
	acl *StreamAcl,
) StreamMetadata {
	return newStreamMetadata(maxCount, maxAge, truncateBefore, cacheControl, acl, nil)
}

func StreamMetadataFromJsonBytes(data []byte) (StreamMetadata, error) {
	metadata := StreamMetadata{}
	return metadata, json.Unmarshal(data, &metadata)
}
