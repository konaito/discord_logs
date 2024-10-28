package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// WebhookData は各codeのデータを格納する構造体
type WebhookData struct {
	WebhooksURL string `json:"webhooks_url"`
}

// loadWebhookMap はJSONファイルからWebhookのマッピングを読み込む
func loadWebhookMap(filePath string) (map[string]WebhookData, error) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read webhook map file: %w", err)
	}

	var webhookMap map[string]WebhookData
	err = json.Unmarshal(file, &webhookMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook map: %w", err)
	}

	return webhookMap, nil
}

// DiscordMessage はDiscordに送信するメッセージの構造体
type DiscordMessage struct {
	Content string `json:"content"`
}

// sendToDiscord は指定されたWebhook URLにメッセージを送信する
func sendToDiscord(webhookURL string, message string) error {
	msg := DiscordMessage{
		Content: message,
	}

	// メッセージをJSON形式にエンコード
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Discord WebhookにPOSTリクエストを送信
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonMsg))
	if err != nil {
		return fmt.Errorf("failed to send message to Discord: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
	}

	return nil
}

// handleRequest はPOSTリクエストを処理し、codeに基づいてメッセージを送信し、GETリクエストでJSONを表示する
func handleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// POSTリクエストの場合
		var reqData struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}

		// JSONリクエストをデコード
		err := json.NewDecoder(r.Body).Decode(&reqData)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// codeに対応するWebhook Dataを取得
		webhookMap, err := loadWebhookMap("webhook_map.json") // リクエストのたびにJSONファイルを読み込む
		if err != nil {
			http.Error(w, "Failed to load webhook map", http.StatusInternalServerError)
			log.Println("Error loading webhook map:", err)
			return
		}

		webhookData, exists := webhookMap[reqData.Code]
		if !exists {
			http.Error(w, "Invalid code", http.StatusBadRequest)
			return
		}

		// Discordにメッセージを送信
		err = sendToDiscord(webhookData.WebhooksURL, reqData.Message)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Message sent successfully")

	case http.MethodGet:
		// GETリクエストの場合、webhook_map.jsonを読み込んでその内容を表示する
		webhookMap, err := loadWebhookMap("webhook_map.json")
		if err != nil {
			http.Error(w, "Failed to load webhook map", http.StatusInternalServerError)
			log.Println("Error loading webhook map:", err)
			return
		}

		// webhookMapをJSON形式でレスポンスに書き込む
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(webhookMap)
		if err != nil {
			http.Error(w, "Failed to encode webhook map", http.StatusInternalServerError)
			log.Println("Error encoding webhook map:", err)
			return
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/dl", handleRequest)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
