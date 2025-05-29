package connection

import (
	"drizlink/utils"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
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

type TransferReader interface {
	Read(p []byte) (n int, err error)
	GetBytesProcessed() int64
}

type TransferWriter interface {
	Write(p []byte) (n int, err error)
	GetBytesProcessed() int64
}

type Transfer struct {
	ID            string
	Type          TransferType
	Name          string
	Size          int64
	BytesComplete int64
	Status        TransferStatus
	Direction     string 
	Recipient     string
	Path          string
	Checksum      string
	StartTime     time.Time
	File          *os.File
	Connection    net.Conn
	ProgressBar   *utils.ProgressBar
	pauseMutex    sync.Mutex
	isPaused      bool
}

var (
	ActiveTransfers   = make(map[string]*Transfer)
	TransfersMutex    sync.RWMutex
	transferIDCounter = 1
	DefaultManager    *TransferManager
)

func init() {
	DefaultManager = NewTransferManager()
}

type TransferManager struct {
	transfers      map[string]*Transfer
	mutex          sync.RWMutex
	nextID         int
}

func NewTransferManager() *TransferManager {
	return &TransferManager{
		transfers: make(map[string]*Transfer),
		nextID:    1,
	}
}

func (tm *TransferManager) GenerateID() string {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	id := strconv.Itoa(tm.nextID)
	tm.nextID++
	return id
}

// GenerateTransferID creates a unique ID for a transfer (legacy function for compatibility)
func GenerateTransferID() string {
	return DefaultManager.GenerateID()
}

func (tm *TransferManager) Register(transfer *Transfer) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	tm.transfers[transfer.ID] = transfer
}

// RegisterTransfer adds a new transfer to the tracking system (legacy function for compatibility)
func RegisterTransfer(transfer *Transfer) {
	DefaultManager.Register(transfer)
	// Also maintain old behavior for backward compatibility
	TransfersMutex.Lock()
	defer TransfersMutex.Unlock()
	ActiveTransfers[transfer.ID] = transfer
}

func (tm *TransferManager) Get(id string) (*Transfer, bool) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	transfer, exists := tm.transfers[id]
	return transfer, exists
}
