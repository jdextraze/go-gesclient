package client

type Command byte

const (
	_ Command = 0x00

	Command_HeartbeatRequestCommand  Command = 0x01
	Command_HeartbeatResponseCommand Command = 0x02

	Command_Ping Command = 0x03
	Command_Pong Command = 0x04

	Command_PrepareAck Command = 0x05
	Command_CommitAck  Command = 0x06

	Command_SlaveAssignment Command = 0x07
	Command_CloneAssignment Command = 0x08

	Command_SubscribeReplica         Command = 0x10
	Command_ReplicaLogPositionAck    Command = 0x11
	Command_CreateChunk              Command = 0x12
	Command_RawChunkBulk             Command = 0x13
	Command_DataChunkBulk            Command = 0x14
	Command_ReplicaSubscriptionRetry Command = 0x15
	Command_ReplicaSubscribed        Command = 0x16

	// CLIENT COMMANDS
	//    Command_CreateStream Command = 0x80
	//    Command_CreateStreamCompleted Command = 0x81

	Command_WriteEvents          Command = 0x82
	Command_WriteEventsCompleted Command = 0x83

	Command_TransactionStart           Command = 0x84
	Command_TransactionStartCompleted  Command = 0x85
	Command_TransactionWrite           Command = 0x86
	Command_TransactionWriteCompleted  Command = 0x87
	Command_TransactionCommit          Command = 0x88
	Command_TransactionCommitCompleted Command = 0x89

	Command_DeleteStream          Command = 0x8A
	Command_DeleteStreamCompleted Command = 0x8B

	Command_ReadEvent                         Command = 0xB0
	Command_ReadEventCompleted                Command = 0xB1
	Command_ReadStreamEventsForward           Command = 0xB2
	Command_ReadStreamEventsForwardCompleted  Command = 0xB3
	Command_ReadStreamEventsBackward          Command = 0xB4
	Command_ReadStreamEventsBackwardCompleted Command = 0xB5
	Command_ReadAllEventsForward              Command = 0xB6
	Command_ReadAllEventsForwardCompleted     Command = 0xB7
	Command_ReadAllEventsBackward             Command = 0xB8
	Command_ReadAllEventsBackwardCompleted    Command = 0xB9

	Command_SubscribeToStream                         Command = 0xC0
	Command_SubscriptionConfirmation                  Command = 0xC1
	Command_StreamEventAppeared                       Command = 0xC2
	Command_UnsubscribeFromStream                     Command = 0xC3
	Command_SubscriptionDropped                       Command = 0xC4
	Command_ConnectToPersistentSubscription           Command = 0xC5
	Command_PersistentSubscriptionConfirmation        Command = 0xC6
	Command_PersistentSubscriptionStreamEventAppeared Command = 0xC7
	Command_CreatePersistentSubscription              Command = 0xC8
	Command_CreatePersistentSubscriptionCompleted     Command = 0xC9
	Command_DeletePersistentSubscription              Command = 0xCA
	Command_DeletePersistentSubscriptionCompleted     Command = 0xCB
	Command_PersistentSubscriptionAckEvents           Command = 0xCC
	Command_PersistentSubscriptionNakEvents           Command = 0xCD
	Command_UpdatePersistentSubscription              Command = 0xCE
	Command_UpdatePersistentSubscriptionCompleted     Command = 0xCF

	Command_ScavengeDatabase          Command = 0xD0
	Command_ScavengeDatabaseCompleted Command = 0xD1

	Command_BadRequest       Command = 0xF0
	Command_NotHandled       Command = 0xF1
	Command_Authenticate     Command = 0xF2
	Command_Authenticated    Command = 0xF3
	Command_NotAuthenticated Command = 0xF4
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
