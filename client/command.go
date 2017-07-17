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

var commands = map[byte]string{
	0x01: "HeartbeatRequestCommand",
	0x02: "HeartbeatResponseCommand",

	0x03: "Ping",
	0x04: "Pong",

	0x05: "PrepareAck",
	0x06: "CommitAck",

	0x07: "SlaveAssignment",
	0x08: "CloneAssignment",

	0x10: "SubscribeReplica",
	0x11: "ReplicaLogPositionAck",
	0x12: "CreateChunk",
	0x13: "RawChunkBulk",
	0x14: "DataChunkBulk",
	0x15: "ReplicaSubscriptionRetry",
	0x16: "ReplicaSubscribed",

	0x80: "CreateStream",
	0x81: "CreateStreamCompleted",

	0x82: "WriteEvents",
	0x83: "WriteEventsCompleted",

	0x84: "TransactionStart",
	0x85: "TransactionStartCompleted",
	0x86: "TransactionWrite",
	0x87: "TransactionWriteCompleted",
	0x88: "TransactionCommit",
	0x89: "TransactionCommitCompleted",

	0x8A: "DeleteStream",
	0x8B: "DeleteStreamCompleted",

	0xB0: "ReadEvent",
	0xB1: "ReadEventCompleted",
	0xB2: "ReadStreamEventsForward",
	0xB3: "ReadStreamEventsForwardCompleted",
	0xB4: "ReadStreamEventsBackward",
	0xB5: "ReadStreamEventsBackwardCompleted",
	0xB6: "ReadAllEventsForward",
	0xB7: "ReadAllEventsForwardCompleted",
	0xB8: "ReadAllEventsBackward",
	0xB9: "ReadAllEventsBackwardCompleted",

	0xC0: "SubscribeToStream",
	0xC1: "SubscriptionConfirmation",
	0xC2: "StreamEventAppeared",
	0xC3: "UnsubscribeFromStream",
	0xC4: "SubscriptionDropped",
	0xC5: "ConnectToPersistentSubscription",
	0xC6: "PersistentSubscriptionConfirmation",
	0xC7: "PersistentSubscriptionStreamEventAppeared",
	0xC8: "CreatePersistentSubscription",
	0xC9: "CreatePersistentSubscriptionCompleted",
	0xCA: "DeletePersistentSubscription",
	0xCB: "DeletePersistentSubscriptionCompleted",
	0xCC: "PersistentSubscriptionAckEvents",
	0xCD: "PersistentSubscriptionNakEvents",
	0xCE: "UpdatePersistentSubscription",
	0xCF: "UpdatePersistentSubscriptionCompleted",

	0xD0: "ScavengeDatabase",
	0xD1: "ScavengeDatabaseCompleted",

	0xF0: "BadRequest",
	0xF1: "NotHandled",
	0xF2: "Authenticate",
	0xF3: "Authenticated",
	0xF4: "NotAuthenticated",
}

func (c Command) String() string {
	return commands[byte(c)]
}
