package generator

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

var existingTransID map[string]bool
var currentTime time.Time

func GenerateRandomSequence() int {
	n, err := rand.Int(rand.Reader, big.NewInt(9000))
	if err != nil {
		return 1000
	}
	return int(n.Int64()) + 1000
}

func GenerateTransactionID() string {
	now := time.Now()
	dateStr := now.Format("20060102150405")
	sequence := GenerateRandomSequence()
	sequenceStr := fmt.Sprintf("%04d", sequence)
	transID := fmt.Sprintf("%s%s", dateStr, sequenceStr)

	if existingTransID == nil {
		currentTime = time.Now()
		existingTransID = make(map[string]bool)
	} else if existingTransID[transID] {
		return GenerateTransactionID()
	}

	if time.Now().After(currentTime.Add(time.Second * 2)) {
		existingTransID = make(map[string]bool)
		currentTime = time.Now()
	}

	existingTransID[transID] = true

	return transID
}
