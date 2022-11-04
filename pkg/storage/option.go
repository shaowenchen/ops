package storage

type S3FileOption struct {
	Region     string
	Endpoint   string
	Bucket     string
	AK         string
	SK         string
	Direction  string
	LocalFile  string
	RemoteFile string
}
