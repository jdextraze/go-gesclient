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
