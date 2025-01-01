package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"golang.org/x/net/proxy"
)

type Config struct {
	WsToken        string `json:"ws_token"`
	MarketplaceMsg string `json:"message"`
}

var (
	proxyAddr     = ""
	proxyUser     = ""
	proxyPassword = ""
)

func main() {
	color.Cyan("Cracked.io - Automation")
	color.White("Auto Marketplace Posting (30 min)\n========================================\n\n")
	color.Red("Do not close this program.")

	data, err := os.ReadFile("config.json")
	if err != nil {
		color.Red("Cannot read config.json")
		return
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		color.Red("Cannot decode config.json")
		return
	}

	WSS_URL := "wss://yelling.cc/socket.io/?token=" + config.WsToken + "&EIO=3&transport=websocket"
	go updateMarketplace(WSS_URL, config.MarketplaceMsg)
	for range time.Tick(10 * time.Minute) {
		go updateMarketplace(WSS_URL, config.MarketplaceMsg)
	}
}

func updateMarketplace(WSS_URL string, MarketplaceMsg string) {
	color.Yellow("Checking..")

	header := make(http.Header)
	header.Add("Origin", "https://cracked.io")

	var socketConn *websocket.Conn

	if proxyAddr != "" {
		auth := &proxy.Auth{
			User:     proxyUser,
			Password: proxyPassword,
		}
		proxyDialer, err := proxy.SOCKS5("tcp", proxyAddr, auth, proxy.Direct)

		if err != nil {
			color.Red("Failed to create SOCKS5 dialer: %v, trying again..", err)
			updateMarketplace(WSS_URL, MarketplaceMsg)
			return
		}

		websocketDialer := &websocket.Dialer{
			Proxy:   nil,
			NetDial: proxyDialer.Dial,
		}

		conn, _, err := websocketDialer.Dial(WSS_URL, header)
		if err != nil {
			color.Red("Failed websocket connection: %v", err)
			updateMarketplace(WSS_URL, MarketplaceMsg)
			return
		} else {
			socketConn = conn
		}
	} else {
		conn, _, err := websocket.DefaultDialer.Dial(WSS_URL, header)
		if err != nil {
			color.Red("Failed websocket connection: %v", err)
			updateMarketplace(WSS_URL, MarketplaceMsg)
			return
		} else {
			socketConn = conn
		}
	}

	if socketConn == nil {
		color.Red("Websocket connection is nil")
		return
	}

	defer socketConn.Close()

	_, _, err := socketConn.ReadMessage()
	if err != nil {
		return
	}

	message := map[string]interface{}{
		"room":    "market",
		"message": MarketplaceMsg,
	}

	payload := []interface{}{2, message}
	messageBytes, err := json.Marshal(payload)
	if err != nil {
		color.Red("Failed to marshal message: %v", err)
	}

	err = socketConn.WriteMessage(websocket.TextMessage, append([]byte("42"), messageBytes...))
	if err != nil {
		checkln("Failed")
	} else {
		checkln("Done")
	}
}

func checkln(format string, a ...interface{}) {
	file, err := os.OpenFile("status.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	log.SetOutput(file)
	currentTime := time.Now().Format("02-01-2006 - 15:04:05")
	log.Printf("%s - %s", currentTime, fmt.Sprintf(format, a...))
}
