package gesclient

// EventStore.ClientApi.SystemData

type tcpCommand byte

const (
	_ tcpCommand = 0x00

	tcpCommand_HeartbeatRequestCommand  = 0x01
	tcpCommand_HeartbeatResponseCommand = 0x02

	tcpCommand_Ping = 0x03
	tcpCommand_Pong = 0x04

	tcpCommand_PrepareAck = 0x05
	tcpCommand_CommitAck  = 0x06

	tcpCommand_SlaveAssignment = 0x07
	tcpCommand_CloneAssignment = 0x08

	tcpCommand_SubscribeReplica         = 0x10
	tcpCommand_ReplicaLogPositionAck    = 0x11
	tcpCommand_CreateChunk              = 0x12
	tcpCommand_RawChunkBulk             = 0x13
	tcpCommand_DataChunkBulk            = 0x14
	tcpCommand_ReplicaSubscriptionRetry = 0x15
	tcpCommand_ReplicaSubscribed        = 0x16

	// CLIENT COMMANDS
	//    tcpCommand_CreateStream = 0x80
	//    tcpCommand_CreateStreamCompleted = 0x81

	tcpCommand_WriteEvents          = 0x82
	tcpCommand_WriteEventsCompleted = 0x83

	tcpCommand_TransactionStart           = 0x84
	tcpCommand_TransactionStartCompleted  = 0x85
	tcpCommand_TransactionWrite           = 0x86
	tcpCommand_TransactionWriteCompleted  = 0x87
	tcpCommand_TransactionCommit          = 0x88
	tcpCommand_TransactionCommitCompleted = 0x89

	tcpCommand_DeleteStream          = 0x8A
	tcpCommand_DeleteStreamCompleted = 0x8B

	tcpCommand_ReadEvent                         = 0xB0
	tcpCommand_ReadEventCompleted                = 0xB1
	tcpCommand_ReadStreamEventsForward           = 0xB2
	tcpCommand_ReadStreamEventsForwardCompleted  = 0xB3
	tcpCommand_ReadStreamEventsBackward          = 0xB4
	tcpCommand_ReadStreamEventsBackwardCompleted = 0xB5
	tcpCommand_ReadAllEventsForward              = 0xB6
	tcpCommand_ReadAllEventsForwardCompleted     = 0xB7
	tcpCommand_ReadAllEventsBackward             = 0xB8
	tcpCommand_ReadAllEventsBackwardCompleted    = 0xB9

	tcpCommand_SubscribeToStream                         = 0xC0
	tcpCommand_SubscriptionConfirmation                  = 0xC1
	tcpCommand_StreamEventAppeared                       = 0xC2
	tcpCommand_UnsubscribeFromStream                     = 0xC3
	tcpCommand_SubscriptionDropped                       = 0xC4
	tcpCommand_ConnectToPersistentSubscription           = 0xC5
	tcpCommand_PersistentSubscriptionConfirmation        = 0xC6
	tcpCommand_PersistentSubscriptionStreamEventAppeared = 0xC7
	tcpCommand_CreatePersistentSubscription              = 0xC8
	tcpCommand_CreatePersistentSubscriptionCompleted     = 0xC9
	tcpCommand_DeletePersistentSubscription              = 0xCA
	tcpCommand_DeletePersistentSubscriptionCompleted     = 0xCB
	tcpCommand_PersistentSubscriptionAckEvents           = 0xCC
	tcpCommand_PersistentSubscriptionNakEvents           = 0xCD
	tcpCommand_UpdatePersistentSubscription              = 0xCE
	tcpCommand_UpdatePersistentSubscriptionCompleted     = 0xCF

	tcpCommand_ScavengeDatabase          = 0xD0
	tcpCommand_ScavengeDatabaseCompleted = 0xD1

	tcpCommand_BadRequest       = 0xF0
	tcpCommand_NotHandled       = 0xF1
	tcpCommand_Authenticate     = 0xF2
	tcpCommand_Authenticated    = 0xF3
	tcpCommand_NotAuthenticated = 0xF4
)

var tcpCommands [255]string

func init() {
	tcpCommands[0x01] = "HeartbeatRequestCommand"
	tcpCommands[0x02] = "HeartbeatResponseCommand"

	tcpCommands[0x03] = "Ping"
	tcpCommands[0x04] = "Pong"

	tcpCommands[0x05] = "PrepareAck"
	tcpCommands[0x06] = "CommitAck"

	tcpCommands[0x07] = "SlaveAssignment"
	tcpCommands[0x08] = "CloneAssignment"

	tcpCommands[0x10] = "SubscribeReplica"
	tcpCommands[0x11] = "ReplicaLogPositionAck"
	tcpCommands[0x12] = "CreateChunk"
	tcpCommands[0x13] = "RawChunkBulk"
	tcpCommands[0x14] = "DataChunkBulk"
	tcpCommands[0x15] = "ReplicaSubscriptionRetry"
	tcpCommands[0x16] = "ReplicaSubscribed"

	tcpCommands[0x80] = "CreateStream"
	tcpCommands[0x81] = "CreateStreamCompleted"

	tcpCommands[0x82] = "WriteEvents"
	tcpCommands[0x83] = "WriteEventsCompleted"

	tcpCommands[0x84] = "TransactionStart"
	tcpCommands[0x85] = "TransactionStartCompleted"
	tcpCommands[0x86] = "TransactionWrite"
	tcpCommands[0x87] = "TransactionWriteCompleted"
	tcpCommands[0x88] = "TransactionCommit"
	tcpCommands[0x89] = "TransactionCommitCompleted"

	tcpCommands[0x8A] = "DeleteStream"
	tcpCommands[0x8B] = "DeleteStreamCompleted"

	tcpCommands[0xB0] = "ReadEvent"
	tcpCommands[0xB1] = "ReadEventCompleted"
	tcpCommands[0xB2] = "ReadStreamEventsForward"
	tcpCommands[0xB3] = "ReadStreamEventsForwardCompleted"
	tcpCommands[0xB4] = "ReadStreamEventsBackward"
	tcpCommands[0xB5] = "ReadStreamEventsBackwardCompleted"
	tcpCommands[0xB6] = "ReadAllEventsForward"
	tcpCommands[0xB7] = "ReadAllEventsForwardCompleted"
	tcpCommands[0xB8] = "ReadAllEventsBackward"
	tcpCommands[0xB9] = "ReadAllEventsBackwardCompleted"

	tcpCommands[0xC0] = "SubscribeToStream"
	tcpCommands[0xC1] = "SubscriptionConfirmation"
	tcpCommands[0xC2] = "StreamEventAppeared"
	tcpCommands[0xC3] = "UnsubscribeFromStream"
	tcpCommands[0xC4] = "SubscriptionDropped"
	tcpCommands[0xC5] = "ConnectToPersistentSubscription"
	tcpCommands[0xC6] = "PersistentSubscriptionConfirmation"
	tcpCommands[0xC7] = "PersistentSubscriptionStreamEventAppeared"
	tcpCommands[0xC8] = "CreatePersistentSubscription"
	tcpCommands[0xC9] = "CreatePersistentSubscriptionCompleted"
	tcpCommands[0xCA] = "DeletePersistentSubscription"
	tcpCommands[0xCB] = "DeletePersistentSubscriptionCompleted"
	tcpCommands[0xCC] = "PersistentSubscriptionAckEvents"
	tcpCommands[0xCD] = "PersistentSubscriptionNakEvents"
	tcpCommands[0xCE] = "UpdatePersistentSubscription"
	tcpCommands[0xCF] = "UpdatePersistentSubscriptionCompleted"

	tcpCommands[0xD0] = "ScavengeDatabase"
	tcpCommands[0xD1] = "ScavengeDatabaseCompleted"

	tcpCommands[0xF0] = "BadRequest"
	tcpCommands[0xF1] = "NotHandled"
	tcpCommands[0xF2] = "Authenticate"
	tcpCommands[0xF3] = "Authenticated"
	tcpCommands[0xF4] = "NotAuthenticated"
}

func (c tcpCommand) String() string {
	return tcpCommands[c]
}
