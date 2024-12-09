# QuickSecShare

## Overview
QuickSecShare is a secure and user-friendly application designed for sharing files across a network. It features a graphical user interface (GUI) powered by the Fyne framework and leverages WebSocket technology for real-time communication. This project prioritizes security and usability, integrating encryption mechanisms and network discovery to ensure smooth and secure file transfers.

## Features
- **Secure File Sharing**: AES encryption is used to protect data.
- **Real-Time Communication**: WebSocket enables efficient and real-time interaction.
- **Service Discovery**: Networking capabilities allow devices to discover each other seamlessly.
- **User-Friendly Interface**: Built using Fyne, a cross-platform GUI framework.

## Project Structure

### Root Files
- **`main.go`**: The entry point of the application. It initializes the GUI, establishes WebSocket connections, and integrates with the cryptography and networking modules.
- **`go.mod` & `go.sum`**: Manage dependencies and ensure reproducible builds.
- **`Icon.png`**: The application's icon.

### Directories

#### `cryptography`
Contains cryptographic utilities to secure data:
- **`aes.go`**: Implements AES encryption for file security.

#### `networking`
Handles networking and service discovery:
- **`advertise.go`**: Facilitates network advertisement and discovery of peers for file sharing.

#### `utils`
Likely contains general-purpose helper functions used across the application.

## Technical Decisions

1. **GUI Framework**:
   - The Fyne framework was chosen for its simplicity, cross-platform support, and modern design capabilities.
   - Provides features like dialog boxes, containers, and widgets for building a polished user experience.

2. **Real-Time Communication**:
   - WebSocket was selected for its lightweight protocol and ability to establish persistent connections, ideal for real-time updates during file sharing.

3. **Security**:
   - AES-GCM encryption ensures data privacy, protecting files during transit.
   - Cryptographic operations are isolated in the `cryptography` package for modularity and ease of maintenance.

4. **Networking**:
   - The `advertise.go` module supports service discovery, enabling devices to find each other on the same network without manual configuration.

## How It Works
- The application launches a GUI where users can select files for sharing.
- Files are encrypted using AES before transmission.
- The networking module discovers peers on the network using a service advertisement mechanism.
- WebSocket connections are established between devices for real-time file transfer.

## Build and Run Instructions

### Desktop
1. Ensure you have Go installed.
2. Clone the repository:
   ```bash
   git clone https://github.com/mohamedjawady/quicksecshare
   ```
3. Navigate to the project directory:
   ```bash
   cd file-sharing-app
   ```
4. Build and run the application:
   ```bash
   go run .
   ```

### Mobile
The app can also be packaged for mobile platforms using Fyne:
1. Package the application for Android:
   ```bash
   fyne package -os android/arm --appID com.quicksecshare.net
   ```
1. Ensure you have Go installed.
2. Clone the repository:
   ```bash
   git clone https://github.com/mohamedjawady/quicksecshare
   ```
3. Navigate to the project directory:
   ```bash
   cd file-sharing-app
   ```
4. Package the application:
   ```bash
   fyne package -os android/arm --appID com.quicksecshare.net
   ```

## Future Improvements
- **Support for Additional Encryption Algorithms**: Expand beyond AES to offer more options.
- **Improved Peer Discovery**: Use advanced protocols for faster and more reliable service discovery.
