package common

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
	SystemMetadata_MaxAge         = "$maxAge"
	SystemMetadata_MaxCount       = "$maxCount"
	SystemMetadata_TruncateBefore = "$tb"
	SystemMetadata_CacheControl   = "$cacheControl"
	SystemMetadata_Acl            = "$acl"
	SystemMetadata_AclRead        = "$r"
	SystemMetadata_AclWrite       = "$w"
	SystemMetadata_AclDelete      = "$d"
	SystemMetadata_AclMetaRead    = "$mr"
	SystemMetadata_AclMetaWrite   = "$mw"
)
