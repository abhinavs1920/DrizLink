package connection

import (

)

type TransferType int

const (
	FileTransfer TransferType = iota
	FolderTransfer
)

func (t TransferType) String() string {
	switch t {
	case FileTransfer:
		return "File"
	case FolderTransfer:
		return "Folder"
	default:
		return "Unknown"
	}
}

type TransferStatus int

const (
	Active TransferStatus = iota
	Paused
	Completed
	Failed
)

func (s TransferStatus) String() string {
	switch s {
	case Active:
		return "Active"
	case Paused:
		return "Paused"
	case Completed:
		return "Completed"
	case Failed:
		return "Failed"
	default:
		return "Unknown"
	}
}

