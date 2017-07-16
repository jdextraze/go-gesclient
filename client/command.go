package client

type Command byte

const (
	_ Command = 0x00

	Command_HeartbeatRequestCommand  = 0x01
	Command_HeartbeatResponseCommand = 0x02

	Command_Ping = 0x03
	Command_Pong = 0x04

	Command_PrepareAck = 0x05
	Command_CommitAck  = 0x06

	Command_SlaveAssignment = 0x07
	Command_CloneAssignment = 0x08

	Command_SubscribeReplica         = 0x10
	Command_ReplicaLogPositionAck    = 0x11
	Command_CreateChunk              = 0x12
	Command_RawChunkBulk             = 0x13
	Command_DataChunkBulk            = 0x14
	Command_ReplicaSubscriptionRetry = 0x15
	Command_ReplicaSubscribed        = 0x16

	// CLIENT COMMANDS
	//    Command_CreateStream = 0x80
	//    Command_CreateStreamCompleted = 0x81

	Command_WriteEvents          = 0x82
	Command_WriteEventsCompleted = 0x83

	Command_TransactionStart           = 0x84
	Command_TransactionStartCompleted  = 0x85
	Command_TransactionWrite           = 0x86
	Command_TransactionWriteCompleted  = 0x87
	Command_TransactionCommit          = 0x88
	Command_TransactionCommitCompleted = 0x89

	Command_DeleteStream          = 0x8A
	Command_DeleteStreamCompleted = 0x8B

	Command_ReadEvent                         = 0xB0
	Command_ReadEventCompleted                = 0xB1
	Command_ReadStreamEventsForward           = 0xB2
	Command_ReadStreamEventsForwardCompleted  = 0xB3
	Command_ReadStreamEventsBackward          = 0xB4
	Command_ReadStreamEventsBackwardCompleted = 0xB5
	Command_ReadAllEventsForward              = 0xB6
	Command_ReadAllEventsForwardCompleted     = 0xB7
	Command_ReadAllEventsBackward             = 0xB8
	Command_ReadAllEventsBackwardCompleted    = 0xB9

	Command_SubscribeToStream                         = 0xC0
	Command_SubscriptionConfirmation                  = 0xC1
	Command_StreamEventAppeared                       = 0xC2
	Command_UnsubscribeFromStream                     = 0xC3
	Command_SubscriptionDropped                       = 0xC4
	Command_ConnectToPersistentSubscription           = 0xC5
	Command_PersistentSubscriptionConfirmation        = 0xC6
	Command_PersistentSubscriptionStreamEventAppeared = 0xC7
	Command_CreatePersistentSubscription              = 0xC8
	Command_CreatePersistentSubscriptionCompleted     = 0xC9
	Command_DeletePersistentSubscription              = 0xCA
	Command_DeletePersistentSubscriptionCompleted     = 0xCB
	Command_PersistentSubscriptionAckEvents           = 0xCC
	Command_PersistentSubscriptionNakEvents           = 0xCD
	Command_UpdatePersistentSubscription              = 0xCE
	Command_UpdatePersistentSubscriptionCompleted     = 0xCF

	Command_ScavengeDatabase          = 0xD0
	Command_ScavengeDatabaseCompleted = 0xD1

	Command_BadRequest       = 0xF0
	Command_NotHandled       = 0xF1
	Command_Authenticate     = 0xF2
	Command_Authenticated    = 0xF3
	Command_NotAuthenticated = 0xF4
)

var commands [255]string

func init() {
	commands[0x01] = "HeartbeatRequestCommand"
	commands[0x02] = "HeartbeatResponseCommand"

	commands[0x03] = "Ping"
	commands[0x04] = "Pong"

	commands[0x05] = "PrepareAck"
	commands[0x06] = "CommitAck"

	commands[0x07] = "SlaveAssignment"
	commands[0x08] = "CloneAssignment"

	commands[0x10] = "SubscribeReplica"
	commands[0x11] = "ReplicaLogPositionAck"
	commands[0x12] = "CreateChunk"
	commands[0x13] = "RawChunkBulk"
	commands[0x14] = "DataChunkBulk"
	commands[0x15] = "ReplicaSubscriptionRetry"
	commands[0x16] = "ReplicaSubscribed"

	commands[0x80] = "CreateStream"
	commands[0x81] = "CreateStreamCompleted"

	commands[0x82] = "WriteEvents"
	commands[0x83] = "WriteEventsCompleted"

	commands[0x84] = "TransactionStart"
	commands[0x85] = "TransactionStartCompleted"
	commands[0x86] = "TransactionWrite"
	commands[0x87] = "TransactionWriteCompleted"
	commands[0x88] = "TransactionCommit"
	commands[0x89] = "TransactionCommitCompleted"

	commands[0x8A] = "DeleteStream"
	commands[0x8B] = "DeleteStreamCompleted"

	commands[0xB0] = "ReadEvent"
	commands[0xB1] = "ReadEventCompleted"
	commands[0xB2] = "ReadStreamEventsForward"
	commands[0xB3] = "ReadStreamEventsForwardCompleted"
	commands[0xB4] = "ReadStreamEventsBackward"
	commands[0xB5] = "ReadStreamEventsBackwardCompleted"
	commands[0xB6] = "ReadAllEventsForward"
	commands[0xB7] = "ReadAllEventsForwardCompleted"
	commands[0xB8] = "ReadAllEventsBackward"
	commands[0xB9] = "ReadAllEventsBackwardCompleted"

	commands[0xC0] = "SubscribeToStream"
	commands[0xC1] = "SubscriptionConfirmation"
	commands[0xC2] = "StreamEventAppeared"
	commands[0xC3] = "UnsubscribeFromStream"
	commands[0xC4] = "SubscriptionDropped"
	commands[0xC5] = "ConnectToPersistentSubscription"
	commands[0xC6] = "PersistentSubscriptionConfirmation"
	commands[0xC7] = "PersistentSubscriptionStreamEventAppeared"
	commands[0xC8] = "CreatePersistentSubscription"
	commands[0xC9] = "CreatePersistentSubscriptionCompleted"
	commands[0xCA] = "DeletePersistentSubscription"
	commands[0xCB] = "DeletePersistentSubscriptionCompleted"
	commands[0xCC] = "PersistentSubscriptionAckEvents"
	commands[0xCD] = "PersistentSubscriptionNakEvents"
	commands[0xCE] = "UpdatePersistentSubscription"
	commands[0xCF] = "UpdatePersistentSubscriptionCompleted"

	commands[0xD0] = "ScavengeDatabase"
	commands[0xD1] = "ScavengeDatabaseCompleted"

	commands[0xF0] = "BadRequest"
	commands[0xF1] = "NotHandled"
	commands[0xF2] = "Authenticate"
	commands[0xF3] = "Authenticated"
	commands[0xF4] = "NotAuthenticated"
}

func (c Command) String() string {
	return commands[c]
}
