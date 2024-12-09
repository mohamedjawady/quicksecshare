package main

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/gorilla/websocket"
	"github.com/mohamedhabib/file-sharing-app/cryptography"
	"github.com/mohamedhabib/file-sharing-app/networking"
	"github.com/mohamedhabib/file-sharing-app/utils"
)

var (
	conn *websocket.Conn

	statusLabel        *widget.Label
	isServer           bool
	clients            []*websocket.Conn
	upgrader           websocket.Upgrader
	discoveryPort      = 9999
	serviceList        *widget.List
	discoveredServices = make(map[string]struct{})
	window             fyne.Window

	key []byte
)

func main() {
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	myApp := app.New()
	myWindow := myApp.NewWindow("QuickSecShare")

	window = myWindow
	myWindow.Resize(fyne.NewSize(600, 400))

	statusLabel = widget.NewLabel("Status: Disconnected")
	statusLabel.Alignment = fyne.TextAlignCenter

	input := widget.NewEntry()
	input.SetPlaceHolder("Enter username (must be kept secret)")

	serverButton := widget.NewButton("Start as Server", func() {
		if !setAESKey(input.Text) {
			showError(myWindow, "Invalid AES Key")
			return
		}
		isServer = true
		showMainScreen(myApp, myWindow)
	})

	clientButton := widget.NewButton("Start as Client", func() {
		if !setAESKey(input.Text) {
			showError(myWindow, "Invalid AES Key.")
			return
		}
		isServer = false
		showMainScreen(myApp, myWindow)
	})

	myWindow.SetContent(
		container.NewVBox(
			widget.NewLabelWithStyle("File Exchange App", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabel("Enter Username:"),
			input,
			widget.NewLabel("Select Mode:"),
			container.NewGridWithColumns(2, serverButton, clientButton),
		),
	)
	myWindow.ShowAndRun()
}

func setAESKey(userKey string) bool {
	key = []byte(cryptography.GenerateAESKey(userKey))
	return true
}

func updateStatus(message string) {
	statusLabel.SetText("Status: " + message)
}

func showError(window fyne.Window, message string) {
	dialog.ShowError(fmt.Errorf(message), window)
}

func startServer(address string) {
	go networking.AdvertiseService(strings.Split(address, ":")[1])
	http.HandleFunc("/ws", handleClientConnection)

	err := http.ListenAndServe(address, nil)
	if err != nil {
		utils.LogError("Error starting server: " + err.Error())
		updateStatus("Error starting server.")
		return
	}
	updateStatus("Server started at " + address)
}

func connectToServer(address string) {
	var err error
	conn, _, err = websocket.DefaultDialer.Dial("ws://"+address+"/ws", nil)
	if err != nil {
		utils.LogError("Error connecting: " + err.Error())
		updateStatus("Error connecting.")
		return
	}
	updateStatus("Connected to " + address)
	go listenForFiles()
}

func handleClientConnection(w http.ResponseWriter, r *http.Request) {
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.LogError("Error upgrading connection: " + err.Error())
		return
	}
	clients = append(clients, clientConn)
	updateStatus("Client connected.")
	go handleClient(clientConn)
}

func handleClient(client *websocket.Conn) {
	defer client.Close()
	for {
		_, fileName, err := client.ReadMessage()
		if err != nil {
			updateStatus("Disconnected from client.")
			return
		}

		_, encryptedData, err := client.ReadMessage()
		if err != nil {
			updateStatus("Error receiving file.")
			return
		}

		fileData, err := cryptography.Decrypt(string(encryptedData), key)
		if err != nil {
			updateStatus("Decryption failed.")
			return
		}
		saveFile(string(fileName), fileData)
	}
}

func listenForFiles() {
	for {
		_, fileName, err := conn.ReadMessage()
		if err != nil {
			updateStatus("Disconnected from server.")
			return
		}

		_, encryptedData, err := conn.ReadMessage()
		if err != nil {
			updateStatus("Error receiving file.")
			return
		}

		fileData, err := cryptography.Decrypt(string(encryptedData), key)
		if err != nil {
			updateStatus("Decryption failed.")
			return
		}
		saveFile(string(fileName), fileData)
	}
}

func sendWithFileName(client *websocket.Conn, data []byte, fileName string) {
	fileName = sanitizeFileName(fileName)

	err := client.WriteMessage(websocket.TextMessage, []byte(fileName))
	if err != nil {
		utils.LogError("Error sending file name: " + err.Error())
		return
	}

	err = client.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		utils.LogError("Error sending file content: " + err.Error())
		return
	}
	utils.LogDebug("File " + fileName + " sent successfully")
}

func saveFile(fileName string, fileData []byte) {

	absolutePath, err := filepath.Abs(fileName)
	if err != nil {
		utils.LogError("Error getting absolute path: " + err.Error())
		return
	}

	err = ioutil.WriteFile(absolutePath, fileData, 0644)
	if err != nil {
		utils.LogError("Error saving file: " + err.Error())
		return
	}

	utils.LogDebug("File saved: " + absolutePath)
	showSuccess(window, "File saved at: "+absolutePath)
}

func sanitizeFileName(fileName string) string {
	return strings.ReplaceAll(fileName, "..", "")
}

func showSuccess(window fyne.Window, message string) {

	content := canvas.NewText(message, color.Black)
	var popUp *widget.PopUp

	popUpContent := container.NewVBox(
		content,
	)

	popUp = widget.NewPopUp(popUpContent, window.Canvas())

	popUp.Show()
}

func discoverServices(myWindow fyne.Window) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", discoveryPort))
	if err != nil {
		utils.LogError("Error resolving UDP address: " + err.Error())
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		utils.LogError("Error listening on UDP: " + err.Error())
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	utils.LogDebug("Listening for UDP broadcasts on port " + fmt.Sprintf("%d", discoveryPort))

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			utils.LogError("Error reading from UDP: " + err.Error())
			continue
		}

		serviceInfo := string(buffer[:n])
		utils.LogDebug(fmt.Sprintf("Discovered service from %s: %s", remoteAddr, serviceInfo))

		discoveredServices[serviceInfo] = struct{}{}

		myWindow.Content().Refresh()
	}
}

func showMainScreen(app fyne.App, window fyne.Window) {
	addressEntry := widget.NewEntry()
	addressEntry.SetPlaceHolder("Enter address:port (e.g., localhost:8080)")
	addressEntry.SetText("0.0.0.0:8080")

	filePathLabel := widget.NewLabel("No file selected")

	connectButton := widget.NewButton("Connect", func() {
		address := addressEntry.Text
		if isServer {
			startServer(address)
		} else {
			connectToServer(address)
		}
	})

	sendFileButton := widget.NewButton("Send File", func() {
		dialog.ShowFileOpen(func(r fyne.URIReadCloser, err error) {
			if r == nil {
				updateStatus("File selection canceled.")
				return
			}

			data, err := ioutil.ReadAll(r)
			if err != nil {
				showError(window, "Error reading file: "+err.Error())
				return
			}

			filePathLabel.SetText("Selected: " + filepath.Base(r.URI().Name()))
			sendFile(data, filepath.Base(r.URI().Name()), window)
		}, window)
	})

	serviceList = widget.NewList(
		func() int {
			return len(discoveredServices)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Service")
		},
		func(i widget.ListItemID, item fyne.CanvasObject) {
			services := make([]string, 0, len(discoveredServices))
			for k := range discoveredServices {
				services = append(services, k)
			}
			item.(*widget.Label).SetText(services[i])
		},
	)

	window.SetContent(
		container.NewVBox(
			widget.NewLabelWithStyle("File Exchange App", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			statusLabel,
			addressEntry, connectButton,
			filePathLabel, sendFileButton,
			widget.NewLabel("Discovered Services:"),
			serviceList,
		),
	)

	if !isServer {
		go discoverServices(window)
	}
}

func sendFile(data []byte, fileName string, window fyne.Window) {
	encryptedData, err := cryptography.Encrypt(data, key)
	if err != nil {
		showError(window, "Encryption failed: "+err.Error())
		return
	}

	if isServer {
		for _, client := range clients {
			sendWithFileName(client, []byte(encryptedData), fileName)
		}
		showSuccess(window, "File sent securely.")
	} else if conn != nil {
		sendWithFileName(conn, []byte(encryptedData), fileName)
		showSuccess(window, "File sent securely.")
	} else {
		showError(window, "No connection established.")
	}
}
