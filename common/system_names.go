package common

import "strings"

type SystemConsumerStrategies string

const (
	SystemConsumerStrategies_DispatchToSingle SystemConsumerStrategies = "DispatchToSingle"
	SystemConsumerStrategies_RoundRobin       SystemConsumerStrategies = "RoundRobin"
	SystemConsumerStrategies_Pinned           SystemConsumerStrategies = "Pinned"
)

func (s SystemConsumerStrategies) IsRoundRobin() bool {
	return s == SystemConsumerStrategies_RoundRobin
}

func (s SystemConsumerStrategies) ToString() string {
	return string(s)
}

const (
	SystemMetadata_UserStreamAcl   = "$userStreamAcl"
	SystemMetadata_SystemStreamAcl = "$systemStreamAcl"

	SystemStreams_StreamsStream     = "$streams"
	SystemStreams_SettingsStream    = "$settings"
	SystemStreams_StatsStreamPrefix = "$stats"

	SystemEventTypes_StreamDeleted   = "$streamDeleted"
	SystemEventTypes_StatsCollection = "$statsCollected"
	SystemEventTypes_LinkTo          = "$>"
	SystemEventTypes_StreamMetadata  = "$metadata"
	SystemEventTypes_Settings        = "$settings"
)

func SystemStreams_MetastreamOf(stream string) string {
	return "$$" + stream
}

func SystemStreams_IsMetastream(stream string) bool {
	return strings.HasPrefix(stream, "$$")
}

func SystemStreams_OriginalStreamOf(stream string) string {
	return stream[2:]
}
