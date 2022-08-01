package storage

type S3FileOption struct {
	Region     string
	Endpoint   string
	Bucket     string
	LocalFile  string
	RemoteFile string
	Direction string
}
