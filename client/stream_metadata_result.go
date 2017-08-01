package client

type streamMetadataResult struct {
	stream string
	isStreamDeleted bool
	metastreamVersion int
	streamMetadata streamMetadata
}