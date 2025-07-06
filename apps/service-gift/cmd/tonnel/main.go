package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/andybalholm/brotli"
)

const (
	url  = "https://gifts3.tonnel.network/api/pageGifts"
	body = `{"page":1,"limit":30,"sort":"{\"message_post_time\":-1,\"gift_id\":-1}","filter":"{\"price\":{\"$exists\":true},\"buyer\":{\"$exists\":false},\"asset\":\"TON\"}","ref":0,"price_range":null,"user_auth":"user=%7B%22id%22%3A7350079261%2C%22first_name%22%3A%22pp%22%2C%22last_name%22%3A%22%22%2C%22username%22%3A%22peterparkish%22%2C%22language_code%22%3A%22en%22%2C%22is_premium%22%3Atrue%2C%22allows_write_to_pm%22%3Atrue%2C%22photo_url%22%3A%22https%3A%5C%2F%5C%2Ft.me%5C%2Fi%5C%2Fuserpic%5C%2F320%5C%2FjMwTE1p_IMe6se6v6t6X8uaS1ymy2hHPJ1Oqt3b13hES-84zfc1MJCUrxxLDLgap.svg%22%7D&chat_instance=-863087820077995196&chat_type=private&auth_date=1751566839&signature=XaRYGoQBTy1ReF6o8IDxgOX4y9-FOyPZIapKM-yHBv-4l7ZvYLDAb-CEjZRMBXFxyelFN7hwV1T8ouX1X482Ag&hash=84fc858e74235897bce7c94da28cacc5988c8bc5ae0602407db028417365445d"}`
)

func main() {
	// HTTP-клиент c таймаутом и выключенным авто-gzip
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			DisableCompression: true, // иначе Go сам добавит Accept-Encoding: gzip
		},
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewBufferString(body))
	if err != nil {
		log.Fatalf("create request: %v", err)
	}

	// точные заголовки, которые ты просил
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://marketplace.tonnel.network")
	req.Header.Set("Referer", "https://marketplace.tonnel.network/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko)")

	// Отправляем
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("http call: %v", err)
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.Status)

	// Декодируем, если сервер сжал ответ
	var reader io.Reader = resp.Body
	switch resp.Header.Get("Content-Encoding") {
	case "br":
		reader = brotli.NewReader(resp.Body)
	case "gzip":
		gr, err := gzip.NewReader(resp.Body)
		if err != nil {
			log.Fatalf("init gzip reader: %v", err)
		}
		defer gr.Close()
		reader = gr
	}

	raw, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("read body: %v", err)
	}

	// Попробуем красиво отформатировать, если это JSON
	var pretty bytes.Buffer
	if json.Valid(raw) && json.Indent(&pretty, raw, "", "  ") == nil {
		fmt.Println(pretty.String())
	} else {
		fmt.Println(string(raw))
	}
}
