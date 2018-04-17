package projections

import (
	cli "github.com/jdextraze/go-gesclient/client"
	"github.com/jdextraze/go-gesclient/tasks"
	"net"
	"time"
)

type Manager struct {
	client       *client
	httpEndpoint *net.TCPAddr
}

func NewManager(
	httpEndpoint *net.TCPAddr,
	operationTimeout time.Duration,
) *Manager {
	if httpEndpoint == nil {
		panic("httpEndpoint is nil")
	}

	return &Manager{
		client:       newClient(operationTimeout),
		httpEndpoint: httpEndpoint,
	}
}

// Task.Result() returns nil
func (m *Manager) EnableAsync(name string, userCredentials *cli.UserCredentials) *tasks.Task {
	if name == "" {
		panic("name must be present")
	}

	return m.client.Enable(m.httpEndpoint, name, userCredentials)
}

// Task.Result() returns nil
func (m *Manager) DisableAsync(name string, userCredentials *cli.UserCredentials) *tasks.Task {
	if name == "" {
		panic("name must be present")
	}

	return m.client.Disable(m.httpEndpoint, name, userCredentials)
}

// Task.Result() returns nil
func (m *Manager) AbortAsync(name string, userCredentials *cli.UserCredentials) *tasks.Task {
	if name == "" {
		panic("name must be present")
	}

	return m.client.Abort(m.httpEndpoint, name, userCredentials)
}

// Task.Result() returns nil
func (m *Manager) CreateOneTimeAsync(query string, userCredentials *cli.UserCredentials) *tasks.Task {
	if query == "" {
		panic("query must be present")
	}

	return m.client.CreateOneTime(m.httpEndpoint, query, userCredentials)
}

// Task.Result() returns nil
func (m *Manager) CreateTransientAsync(name string, query string, userCredentials *cli.UserCredentials) *tasks.Task {
	if name == "" {
		panic("name must be present")
	}
	if query == "" {
		panic("query must be present")
	}

	return m.client.CreateTransient(m.httpEndpoint, name, query, userCredentials)
}

// Task.Result() returns nil
func (m *Manager) CreateContinuousAsync(
	name string,
	query string,
	trackEmittedStreams bool,
	userCredentials *cli.UserCredentials,
) *tasks.Task {
	if name == "" {
		panic("name must be present")
	}
	if query == "" {
		panic("query must be present")
	}

	return m.client.CreateContinuous(m.httpEndpoint, name, query, trackEmittedStreams, userCredentials)
}

// Task.Result() returns []*projections.ProjectionDetails
func (m *Manager) ListAllAsync(userCredentials *cli.UserCredentials) *tasks.Task {
	return m.client.ListAll(m.httpEndpoint, userCredentials)
}

// Task.Result() returns []projections.ProjectionDetails
func (m *Manager) ListOneTimeAsync(userCredentials *cli.UserCredentials) *tasks.Task {
	return m.client.ListOneTime(m.httpEndpoint, userCredentials)
}

// Task.Result() returns []projections.ProjectionDetails
func (m *Manager) ListContinuousAsync(userCredentials *cli.UserCredentials) *tasks.Task {
	return m.client.ListContinuous(m.httpEndpoint, userCredentials)
}

// Task.Result() returns a string
func (m *Manager) GetStatusAsync(name string, userCredentials *cli.UserCredentials) *tasks.Task {
	if name == "" {
		panic("name must be present")
	}

	return m.client.GetStatus(m.httpEndpoint, name, userCredentials)
}

// Task.Result() returns a string
func (m *Manager) GetStateAsync(name string, userCredentials *cli.UserCredentials) *tasks.Task {
	if name == "" {
		panic("name must be present")
	}

	return m.client.GetState(m.httpEndpoint, name, userCredentials)
}

// Task.Result() returns a string
func (m *Manager) GetPartitionStateAsync(
	name string,
	partitionId string,
	userCredentials *cli.UserCredentials,
) *tasks.Task {
	if name == "" {
		panic("name must be present")
	}
	if partitionId == "" {
		panic("partitionId must be present")
	}

	return m.client.GetPartitionStateAsync(m.httpEndpoint, name, partitionId, userCredentials)
}

// Task.Result() returns a string
func (m *Manager) GetResultAsync(name string, userCredentials *cli.UserCredentials) *tasks.Task {
	if name == "" {
		panic("name must be present")
	}

	return m.client.GetResult(m.httpEndpoint, name, userCredentials)
}

// Task.Result() returns a string
func (m *Manager) GetPartitionResultAsync(
	name string,
	partitionId string,
	userCredentials *cli.UserCredentials,
) *tasks.Task {
	if name == "" {
		panic("name must be present")
	}
	if partitionId == "" {
		panic("partitionId must be present")
	}

	return m.client.GetPartitionResultAsync(m.httpEndpoint, name, partitionId, userCredentials)
}

// Task.Result() returns a string
func (m *Manager) GetStatisticsAsync(name string, userCredentials *cli.UserCredentials) *tasks.Task {
	if name == "" {
		panic("name must be present")
	}

	return m.client.GetStatistics(m.httpEndpoint, name, userCredentials)
}

// Task.Result() returns a string
func (m *Manager) GetQueryAsync(name string, userCredentials *cli.UserCredentials) *tasks.Task {
	if name == "" {
		panic("name must be present")
	}

	return m.client.GetQuery(m.httpEndpoint, name, userCredentials)
}

// Task.Result() returns nil
func (m *Manager) UpdateQueryAsync(name string, query string, userCredentials *cli.UserCredentials) *tasks.Task {
	if name == "" {
		panic("name must be present")
	}
	if query == "" {
		panic("query must be present")
	}

	return m.client.GetQuery(m.httpEndpoint, name, userCredentials)
}

// Task.Result() returns nil
func (m *Manager) DeleteQueryAsync(
	name string,
	deleteEmittedStreams bool,
	userCredentials *cli.UserCredentials,
) *tasks.Task {
	if name == "" {
		panic("name must be present")
	}

	return m.client.Delete(m.httpEndpoint, name, deleteEmittedStreams, userCredentials)
}
