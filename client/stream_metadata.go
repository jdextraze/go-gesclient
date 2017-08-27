package client

import (
	"encoding/json"
	"fmt"
	"github.com/jdextraze/go-gesclient/common"
	"time"
)

type StreamMetadata struct {
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
) *StreamMetadata {
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
	return &StreamMetadata{
		maxCount:       maxCount,
		maxAge:         maxAge,
		truncateBefore: truncateBefore,
		cacheControl:   cacheControl,
		acl:            acl,
		customMetadata: customMetadata,
	}
}

func (d *StreamMetadata) UnmarshalJSON(data []byte) error {
	var metadataMap map[string]interface{}

	if err := json.Unmarshal(data, &metadataMap); err != nil {
		return fmt.Errorf("Failed to unmarshal to stream metadata: %v", err)
	}

	if v, ok := metadataMap[common.SystemMetadata_MaxCount]; ok {
		if vc, ok := v.(float64); ok {
			maxCount := int(vc)
			d.maxCount = &maxCount
		} else {
			return fmt.Errorf("Invalid metadata format for $maxCount: %v", v)
		}
		delete(metadataMap, common.SystemMetadata_MaxCount)
	}
	if v, ok := metadataMap[common.SystemMetadata_MaxAge]; ok {
		if vc, ok := v.(float64); ok {
			maxAge := time.Duration(vc)
			d.maxAge = &maxAge
		} else {
			return fmt.Errorf("Invalid metadata format for $maxAge: %v", v)
		}
		delete(metadataMap, common.SystemMetadata_MaxAge)
	}
	if v, ok := metadataMap[common.SystemMetadata_TruncateBefore]; ok {
		if vc, ok := v.(float64); ok {
			truncateBefore := int(vc)
			d.truncateBefore = &truncateBefore
		} else {
			return fmt.Errorf("Invalid metadata format for $tb: %v", v)
		}
		delete(metadataMap, common.SystemMetadata_TruncateBefore)
	}
	if v, ok := metadataMap[common.SystemMetadata_CacheControl]; ok {
		if vc, ok := v.(float64); ok {
			cacheControl := time.Duration(vc)
			d.cacheControl = &cacheControl
		} else {
			return fmt.Errorf("Invalid metadata format for $cacheControl: %v", v)
		}
		delete(metadataMap, common.SystemMetadata_CacheControl)
	}
	if v, ok := metadataMap[common.SystemMetadata_Acl]; ok {
		if vc, ok := v.(map[string][]string); ok {
			d.acl = NewStreamAcl(
				vc[common.SystemMetadata_AclRead],
				vc[common.SystemMetadata_AclWrite],
				vc[common.SystemMetadata_AclDelete],
				vc[common.SystemMetadata_AclMetaRead],
				vc[common.SystemMetadata_AclMetaWrite],
			)
		} else {
			return fmt.Errorf("Invalid metadata format for $acl: %v", v)
		}
		delete(metadataMap, common.SystemMetadata_Acl)
	}
	d.customMetadata = metadataMap

	return nil
}

func (d *StreamMetadata) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{}
	if d.maxCount != nil {
		data[common.SystemMetadata_MaxCount] = *d.maxCount
	}
	if d.maxAge != nil {
		data[common.SystemMetadata_MaxAge] = *d.maxAge
	}
	if d.truncateBefore != nil {
		data[common.SystemMetadata_TruncateBefore] = *d.truncateBefore
	}
	if d.cacheControl != nil {
		data[common.SystemMetadata_CacheControl] = *d.cacheControl
	}
	if d.acl != nil {
		data[common.SystemMetadata_Acl] = map[string][]string{
			common.SystemMetadata_AclRead:      d.acl.readRoles,
			common.SystemMetadata_AclWrite:     d.acl.writeRoles,
			common.SystemMetadata_AclDelete:    d.acl.deleteRoles,
			common.SystemMetadata_AclMetaRead:  d.acl.metaReadRoles,
			common.SystemMetadata_AclMetaWrite: d.acl.metaWriteRoles,
		}
	}
	for key, value := range d.customMetadata {
		data[key] = value
	}
	return json.Marshal(data)
}

func CreateStreamMetadata(
	maxCount *int,
	maxAge *time.Duration,
	truncateBefore *int,
	cacheControl *time.Duration,
	acl *streamAcl,
) *StreamMetadata {
	return newStreamMetadata(maxCount, maxAge, truncateBefore, cacheControl, acl, nil)
}

func StreamMetadataFromJsonBytes(data []byte) (*StreamMetadata, error) {
	metadata := &StreamMetadata{}
	return metadata, json.Unmarshal(data, metadata)
}
