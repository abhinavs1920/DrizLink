# 🔗 DrizLink - P2P File Sharing Application 🔗

A peer-to-peer file sharing application with integrated chat functionality, allowing users to connect, communicate, and share files directly with each other.

## ✨ Features

- **👤 User Authentication**: Connect with a username and maintain persistent sessions
- **💬 Real-time Chat**: Send and receive messages with all connected users
- **📁 File Sharing**: Transfer files directly between users
- **📂 Folder Sharing**: Share entire folders with other users
- **🔍 File Discovery**: Look up and browse other users' shared directories
- **🔄 Automatic Reconnection**: Seamlessly reconnect with your existing session
- **👥 Status Tracking**: Monitor which users are currently online
- **🎨 Colorful UI**: Enhanced CLI interface with colors and emojis
- **📊 Progress Bars**: Visual feedback for file and folder transfers
- **🔒 Data Integrity**: MD5 checksum verification for files and folders

## 🚀 Installation

### Prerequisites
- Go (1.16 or later) 🔧

### Steps
1. Clone the repository ⬇️
```bash
git clone https://github.com/Harsh2563/DrizLink_Cli.git
cd DrizLink_Cli
```

2. Build the application 🛠️
```bash
go build -o DrizLink_Cli
```

## 🎮 Usage

### Starting the Server 🖥️
```bash
# Start server on default port 8080
go run ./server/cmd --port 8080

# Start server on custom port
go run ./server/cmd --port 3000

```

### Connecting as a Client 📱
```bash
# Connect to local server with default port
go run ./client/cmd --server localhost:8080

# Connect to remote server
go run ./client/cmd --server 192.168.0.203:4000

```

The application will validate:
- Server availability before client connection attempts
- Port availability before starting a server
- Existence of shared folder paths

## 🏗️ Architecture

The application follows a hybrid P2P architecture:
- 🌐 A central server handles user registration, discovery, and connection brokering
- ↔️ File and folder transfers occur directly between peers
- 💓 Server maintains connection status through regular heartbeat checks

## 📝 Commands

### Chat Commands 💬
| Command | Description |
|---------|-------------|
| `/help` | Show all available commands |
| `/status` | Show online users |
| `exit` | Disconnect and exit the application |

### File Operations 📂
| Command | Description |
|---------|-------------|
| `/lookup <userId>` | Browse user's shared files |
| `/sendfile <userId> <filePath>` | Send a file to another user |
| `/sendfolder <userId> <folderPath>` | Send a folder to another user |
| `/download <userId> <filename>` | Download a file from another user |

## Terminal UI Features 🎨

- 🌈 **Color-coded messages**:
  - Commands appear in blue
  - Success messages appear in green
  - Error messages appear in red
  - User status notifications in yellow
  
- 📊 **Progress bars for file transfers**:
  ```
  [===================================>------] 75% (1.2 MB/1.7 MB)
  ```

- 📁 **Improved file listings**:
  ```
  === FOLDERS ===
  📁 [FOLDER] documents (Size: 0 bytes)
  📁 [FOLDER] images (Size: 0 bytes)
  
  === FILES ===
  📄 [FILE] document.pdf (Size: 1024 bytes)
  📄 [FILE] image.jpg (Size: 2048 bytes)
  ```

## 🔒 Security

The application implements basic reconnection security by tracking IP addresses and user sessions.

- **📁 Folder Path Validation**: The application verifies that shared folder paths exist before establishing a connection. If an invalid path is provided, the user will be prompted to enter a valid folder path.
- **🔌 Server Availability Check**: Client automatically verifies server availability before attempting connection, preventing connection errors.
- **🚫 Port Conflict Prevention**: Server detects if a port is already in use and alerts the user to choose another port.
- **🔐 Checksum Verification**: All file and folder transfers include MD5 checksum calculation to verify data integrity:
  - When sending, a unique MD5 hash is calculated for the file/folder contents
  - During transfer, the hash is securely transmitted alongside the data
  - Upon receiving, a new hash is calculated from the received data
  - The application compares both hashes to confirm the transfer was successful and uncorrupted
  - Users receive visual confirmation of integrity checks with clear success/failure messages

This checksum process ensures that files and folders arrive exactly as they were sent, protecting against data corruption during transfer.

Made with ❤️ by the DrizLink Team
