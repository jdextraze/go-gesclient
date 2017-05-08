package internal

type OperationManager struct{}

func (m *OperationManager) TotalOperationCount() int { return 0 }

func (m *OperationManager) CheckTimeoutsAndRetry(c *PackageConnection) {}

func (m *OperationManager) CleanUp() {}
