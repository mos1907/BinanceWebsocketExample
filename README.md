---
title: "Monitoring Binance Futures Trading Using Websocket API with GOLANG"
date: 2024-12-22
lastmod: 2024-12-22
draft: false
tags: ["Golang", "Binance", "BTC", "Cyrpto","Trading"]
authors: ["Murat"]
categories: ["Software Development"]
description: "Monitoring Binance Futures Trading Using Websocket API with GOLANG"
lightgallery: true
featuredImage: "gobinance.png"
---
# Monitoring Binance Futures Trading Using Websocket API with GOLANG

## Introduction

This article presents a technical analysis of a real-time position and order monitoring application developed for the Binance futures platform. The application is developed using the Go programming language and provides the capability to monitor positions, orders, and account status in real-time through the Binance API.

## Technical Infrastructure

### Technologies Used

- **Go Programming Language**: Main development language
- **Gorilla WebSocket**: Library used for WebSocket connections
- **Binance Futures API**: Communication with the futures platform
- **HMAC-SHA256**: Hash algorithm used for API security signing

### Core Data Structures

The application is built on three fundamental data structures:

```go
type Position struct {
    Symbol           string  `json:"symbol"`
    PositionAmt      string  `json:"positionAmt"`
    EntryPrice       string  `json:"entryPrice"`
    UnRealizedProfit string  `json:"unRealizedProfit"`
    Leverage         string  `json:"leverage"`
    MarkPrice        string  `json:"markPrice"`
    ROE              float64
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
    AvailableBalance     string `json:"availableBalance"`
    TotalUnrealizedProfit string `json:"totalUnrealizedProfit"`
}
```

## Application Architecture

### 1. Security and Authentication

The application uses HMAC-SHA256 signing mechanism for secure communication with the Binance API:

```go
func getSignature(queryString string) string {
    h := hmac.New(sha256.New, []byte(API_SECRET))
    h.Write([]byte(queryString))
    return hex.EncodeToString(h.Sum(nil))
}
```

### 2. Data Collection and Processing

The application manages two main data streams:
1. **Initial Data via REST API**: Account status, positions, and orders are retrieved using the `getInitialData()` function.
2. **Real-Time Updates via WebSocket**: Live data flow is managed through the `handleUserData()` function.

### 3. Real-Time Data Flow

WebSocket connections are used for two distinct purposes:

1. **User Data Stream**: Account updates and order changes
2. **Market Data Stream**: Price updates for open positions

```go
func handleUserData(positions []Position, orders []Order, account Account) {
    listenKey := getListenKey()
    wsURL := fmt.Sprintf("wss://fstream.binance.com/ws/%s", listenKey)
    // ... WebSocket operations
}
```

### 4. Console Interface

The application provides a user-friendly interface through the console. The `printCurrentStatus()` function regularly displays:

- Account balance
- Open positions and profit/loss status
- Active orders
- ROE (Return on Equity) calculations

## Key Features

### 1. ROE Calculation

The profitability ratio of positions is calculated according to this formula:

```go
investment := math.Abs(posAmt) * entryPrice / leverage
if investment > 0 {
    pos.ROE = (unRealizedProfit / investment) * 100
}
```

### 2. Multi-Symbol Support

The application can monitor multiple open positions simultaneously and track real-time price updates for each one.

### 3. Error Handling

Features include automatic reconnection in case of connection loss and ensuring data flow continuity.

## Implementation Details

### WebSocket Stream Management

The application maintains two types of WebSocket connections:

1. **User Data Stream**: 
   - Monitors account updates
   - Tracks order status changes
   - Updates position modifications

2. **Market Data Stream**:
   - Subscribes to price updates for active positions
   - Processes mark price updates in real-time
   - Updates ROE calculations automatically

### Console Output Management

The application implements a clean console interface that:
- Refreshes automatically with new data
- Organizes information in clearly defined sections
- Formats numbers and percentages for readability
- Supports cross-platform operation (Windows/Unix)

## Performance Considerations

The application is designed with several performance optimizations:

1. **Efficient Data Structures**:
   - Uses appropriate data types for numerical values
   - Implements efficient string handling
   - Maintains minimal memory footprint

2. **Resource Management**:
   - Properly manages WebSocket connections
   - Implements appropriate error handling
   - Includes automatic cleanup procedures

## Conclusion

This application serves as a powerful monitoring tool for users trading on the Binance futures platform. It enhances the trading experience through real-time data streaming, automatic ROE calculation, and a user-friendly console interface.

Future Development Suggestions:
- Addition of a graphical user interface
- Implementation of an alarm system
- Integration of automated trading strategies
- Performance analysis and reporting features

## Technical Requirements

- Go 1.15 or higher
- Binance API keys
- Internet connection
- Terminal access

## Security Considerations

- API keys must be kept secure
- Implementation follows Binance API best practices
- Includes proper error handling for security-related operations
- Implements secure WebSocket connection management
