package messages

import (
	"fmt"
	"github.com/satori/go.uuid"
	"time"
)

type ClusterInfoDto struct {
	Members []*MemberInfoDto
}

type MemberInfoDto struct {
	InstanceId uuid.UUID

	Timestamp time.Time
	State     VNodeState
	IsAlive   bool

	InternalTcpIp         string
	InternalTcpPort       int
	InternalSecureTcpPort int

	ExternalTcpIp         string
	ExternalTcpPort       int
	ExternalSecureTcpPort int

	InternalHttpIp   string
	InternalHttpPort int

	ExternalHttpIp   string
	ExternalHttpPort int

	LastCommitPosition int64
	WriterCheckpoint   int64
	ChaserCheckpoint   int64

	EpochPosition int64
	EpochNumber   int
	EpochId       uuid.UUID

	NodePriority int
}

func (x MemberInfoDto) String() string {
	if x.State == VNodeState_Manager {
		return fmt.Sprintf("MAN %s <%t> [%s, %s:%d, %s:%d | %s", x.InstanceId, x.IsAlive, x.State, x.InternalHttpIp,
			x.InternalHttpPort, x.ExternalHttpIp, x.ExternalHttpPort, x.Timestamp)
	}
	return fmt.Sprintf("VND %s <%t> [%s, %s:%d, %d, %s:%d, %d, %s:%d, %s:%d] %d/%d/%d/E%d@%d:%s | %s", x.InstanceId,
		x.IsAlive, x.State, x.InternalTcpIp, x.InternalTcpPort, x.InternalSecureTcpPort, x.ExternalTcpIp,
		x.ExternalTcpPort, x.ExternalSecureTcpPort, x.InternalHttpIp, x.InternalHttpPort, x.ExternalHttpIp,
		x.ExternalHttpPort, x.LastCommitPosition, x.WriterCheckpoint, x.ChaserCheckpoint, x.EpochNumber,
		x.EpochPosition, x.EpochId, x.Timestamp)
}

type VNodeState int

const (
	VNodeState_Initializing VNodeState = 0
	VNodeState_Unknown      VNodeState = 1
	VNodeState_PreReplica   VNodeState = 2
	VNodeState_CatchingUp   VNodeState = 3
	VNodeState_Clone        VNodeState = 4
	VNodeState_Slave        VNodeState = 5
	VNodeState_PreMaster    VNodeState = 6
	VNodeState_Master       VNodeState = 7
	VNodeState_Manager      VNodeState = 8
	VNodeState_ShuttingDown VNodeState = 9
	VNodeState_Shutdown     VNodeState = 10
)

var VNodeState_name = map[int]string{
	0:  "Initializing",
	1:  "Unknown",
	2:  "PreReplica",
	3:  "CatchingUp",
	4:  "Clone",
	5:  "Slave",
	6:  "PreMaster",
	7:  "Master",
	8:  "Manager",
	9:  "ShuttingDown",
	10: "Shutdown",
}

func (x VNodeState) String() string {
	return VNodeState_name[int(x)]
}
