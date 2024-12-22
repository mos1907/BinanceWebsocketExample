package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	API_KEY    = "Your_Apikey"
	API_SECRET = "Your_Secretkey"
)

type Position struct {
	Symbol           string  `json:"symbol"`
	PositionAmt      string  `json:"positionAmt"`
	EntryPrice       string  `json:"entryPrice"`
	UnRealizedProfit string  `json:"unRealizedProfit"`
	Leverage         string  `json:"leverage"`
	MarkPrice        string  `json:"markPrice"`
	ROE              float64 // ROE değerini tutacak alan
}

type Order struct {
	OrderID   int64  `json:"orderId"`
	Symbol    string `json:"symbol"`
	Type      string `json:"type"`
	Side      string `json:"side"`
	Price     string `json:"price"`
	OrigQty   string `json:"origQty"`
	StopPrice string `json:"stopPrice"`
	Status    string `json:"status"`
}

type Account struct {
	TotalWalletBalance    string `json:"totalWalletBalance"`
	AvailableBalance      string `json:"availableBalance"`
	TotalUnrealizedProfit string `json:"totalUnrealizedProfit"`
}

func clearConsole() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func getSignature(queryString string) string {
	h := hmac.New(sha256.New, []byte(API_SECRET))
	h.Write([]byte(queryString))
	return hex.EncodeToString(h.Sum(nil))
}

func getInitialData() ([]Position, []Order, Account) {
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	queryString := fmt.Sprintf("timestamp=%s", timestamp)
	signature := getSignature(queryString)

	client := &http.Client{}

	// Pozisyonları al
	positionsURL := fmt.Sprintf("https://fapi.binance.com/fapi/v2/positionRisk?%s&signature=%s", queryString, signature)
	req, _ := http.NewRequest("GET", positionsURL, nil)
	req.Header.Add("X-MBX-APIKEY", API_KEY)
	resp, _ := client.Do(req)

	var positions []Position
	json.NewDecoder(resp.Body).Decode(&positions)

	// Aktif pozisyonları filtrele
	activePositions := []Position{}
	for _, pos := range positions {
		positionAmt, _ := strconv.ParseFloat(pos.PositionAmt, 64)
		if positionAmt != 0.0 {
			activePositions = append(activePositions, pos)
		}
	}

	// Emirleri al
	ordersURL := fmt.Sprintf("https://fapi.binance.com/fapi/v1/openOrders?%s&signature=%s", queryString, signature)
	req, _ = http.NewRequest("GET", ordersURL, nil)
	req.Header.Add("X-MBX-APIKEY", API_KEY)
	resp, _ = client.Do(req)

	var orders []Order
	json.NewDecoder(resp.Body).Decode(&orders)

	// Hesap bilgilerini al
	accountURL := fmt.Sprintf("https://fapi.binance.com/fapi/v2/account?%s&signature=%s", queryString, signature)
	req, _ = http.NewRequest("GET", accountURL, nil)
	req.Header.Add("X-MBX-APIKEY", API_KEY)
	resp, _ = client.Do(req)

	var account Account
	json.NewDecoder(resp.Body).Decode(&account)

	return activePositions, orders, account
}

func getListenKey() string {
	client := &http.Client{}
	url := "https://fapi.binance.com/fapi/v1/listenKey"
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Add("X-MBX-APIKEY", API_KEY)
	resp, _ := client.Do(req)

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)
	return result["listenKey"]
}

func printCurrentStatus(positions []Position, orders []Order, account Account) {
	clearConsole()
	fmt.Println("\n=== MEVCUT HESAP DURUMU ===")
	fmt.Printf("Toplam Bakiye: %s USDT\n", account.TotalWalletBalance)
	fmt.Printf("Kullanılabilir Bakiye: %s USDT\n", account.AvailableBalance)
	fmt.Printf("Unrealized PNL: %s USDT\n", account.TotalUnrealizedProfit)
	fmt.Println(strings.Repeat("-", 50))

	if len(positions) > 0 {
		fmt.Println("\n=== AÇIK POZİSYONLAR ===")
		for _, pos := range positions {
			// ROE hesaplama
			posAmt, _ := strconv.ParseFloat(pos.PositionAmt, 64)
			entryPrice, _ := strconv.ParseFloat(pos.EntryPrice, 64)
			unRealizedProfit, _ := strconv.ParseFloat(pos.UnRealizedProfit, 64)
			leverage, _ := strconv.ParseFloat(pos.Leverage, 64)

			investment := math.Abs(posAmt) * entryPrice / leverage
			if investment > 0 {
				pos.ROE = (unRealizedProfit / investment) * 100
			}

			fmt.Printf("\nSembol: %s\n", pos.Symbol)
			fmt.Printf("Miktar: %s\n", pos.PositionAmt)
			fmt.Printf("Kaldıraç: %sx\n", pos.Leverage)
			fmt.Printf("Giriş Fiyatı: %s\n", pos.EntryPrice)
			fmt.Printf("Güncel Fiyat: %s\n", pos.MarkPrice)
			fmt.Printf("PNL: %s\n", pos.UnRealizedProfit)
			fmt.Printf("ROE: %.2f%%\n", pos.ROE)
			fmt.Println(strings.Repeat("-", 50))
		}

	} else {
		fmt.Println("\n=== AÇIK POZİSYON YOK ===")
		fmt.Println(strings.Repeat("-", 50))
	}

	if len(orders) > 0 {
		fmt.Println("\n=== AKTİF EMİRLER ===")
		for _, order := range orders {
			fmt.Printf("\nEmir ID: %d\n", order.OrderID)
			fmt.Printf("Sembol: %s\n", order.Symbol)
			fmt.Printf("Tip: %s\n", order.Type)
			fmt.Printf("Taraf: %s\n", order.Side)
			fmt.Printf("Fiyat: %s\n", order.Price)
			fmt.Printf("Miktar: %s\n", order.OrigQty)
			if order.StopPrice != "0" {
				fmt.Printf("Stop Fiyat: %s\n", order.StopPrice)
			}
			fmt.Printf("Durum: %s\n", order.Status)
			fmt.Println(strings.Repeat("-", 50))
		}
	} else {
		fmt.Println("\n=== AKTİF EMİR YOK ===")
		fmt.Println(strings.Repeat("-", 50))
	}

	fmt.Println("\nGüncellemeler takip ediliyor...")
}

func handleUserData(positions []Position, orders []Order, account Account) {
	listenKey := getListenKey()
	wsURL := fmt.Sprintf("wss://fstream.binance.com/ws/%s", listenKey)

	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Fatal("Bağlantı hatası:", err)
	}
	defer c.Close()

	// Market verisi akışını başlat
	go func() {
		for {
			symbols := make([]string, 0)
			for _, pos := range positions {
				symbols = append(symbols, strings.ToLower(pos.Symbol))
			}

			if len(symbols) > 0 {
				streams := make([]string, 0)
				for _, symbol := range symbols {
					streams = append(streams, fmt.Sprintf("%s@markPrice@1s", symbol))
				}

				streamURL := fmt.Sprintf("wss://fstream.binance.com/stream?streams=%s", strings.Join(streams, "/"))
				ws, _, err := websocket.DefaultDialer.Dial(streamURL, nil)
				if err != nil {
					log.Println("Market akışı bağlantı hatası:", err)
					time.Sleep(time.Second)
					continue
				}

				for {
					_, message, err := ws.ReadMessage()
					if err != nil {
						log.Println("Market verisi okuma hatası:", err)
						break
					}

					var data map[string]interface{}
					if err := json.Unmarshal(message, &data); err != nil {
						continue
					}

					if markData, ok := data["data"].(map[string]interface{}); ok {
						symbol := markData["s"].(string)
						markPrice := markData["p"].(string)

						for i := range positions {
							if positions[i].Symbol == symbol {
								positions[i].MarkPrice = markPrice
								printCurrentStatus(positions, orders, account)
								break
							}
						}
					}
				}
			}
			time.Sleep(time.Second)
		}
	}()

	// Kullanıcı verisi akışını işle
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("Okuma hatası:", err)
			return
		}

		var data map[string]interface{}
		if err := json.Unmarshal(message, &data); err != nil {
			continue
		}

		if eventType, ok := data["e"].(string); ok {
			switch eventType {
			case "ACCOUNT_UPDATE":
				if accountData, ok := data["a"].(map[string]interface{}); ok {
					if balances, ok := accountData["B"].([]interface{}); ok {
						for _, balance := range balances {
							b := balance.(map[string]interface{})
							if b["a"].(string) == "USDT" {
								account.TotalWalletBalance = b["wb"].(string)
							}
						}
					}
				}
				printCurrentStatus(positions, orders, account)

			case "ORDER_TRADE_UPDATE":
				orderUpdate := data["o"].(map[string]interface{})
				orderID := int64(orderUpdate["i"].(float64))
				orderStatus := orderUpdate["X"].(string)

				// Emir kapandıysa listeden kaldır
				if orderStatus == "FILLED" || orderStatus == "CANCELED" || orderStatus == "EXPIRED" || orderStatus == "REJECTED" {
					for i, order := range orders {
						if order.OrderID == orderID {
							orders = append(orders[:i], orders[i+1:]...)
							break
						}
					}
				}
				printCurrentStatus(positions, orders, account)
			}
		}
	}
}

func main() {
	for {
		positions, orders, account := getInitialData()
		handleUserData(positions, orders, account)
		time.Sleep(3 * time.Second)
	}
}
