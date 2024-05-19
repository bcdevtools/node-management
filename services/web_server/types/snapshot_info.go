package types

type SnapshotInfo struct {
	FileName         string
	Size             string
	ModTime          string
	DownloadFilePath string
	Error            error
}
