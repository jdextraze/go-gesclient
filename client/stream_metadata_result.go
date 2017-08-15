package client

type StreamMetadataResult struct {
	stream            string
	isStreamDeleted   bool
	metastreamVersion int
	streamMetadata    *StreamMetadata
}

func NewStreamMetadataResult(
	stream string,
	isStreamDeleted bool,
	metastreamVersion int,
	streamMetadata *StreamMetadata,
) *StreamMetadataResult {
	return &StreamMetadataResult{
		stream:            stream,
		isStreamDeleted:   isStreamDeleted,
		metastreamVersion: metastreamVersion,
		streamMetadata:    streamMetadata,
	}
}

func (r *StreamMetadataResult) Stream() string { return r.stream }

func (r *StreamMetadataResult) IsStreamDeleted() bool { return r.isStreamDeleted }

func (r *StreamMetadataResult) MetastreamVersion() int { return r.metastreamVersion }

func (r *StreamMetadataResult) StreamMetadata() *StreamMetadata { return r.streamMetadata }
