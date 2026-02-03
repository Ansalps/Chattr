package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func VerifyRazorpaySignature(paymentID, subscriptionID, signature string, keySecret string) bool {
	pID := paymentID
	sID := subscriptionID
	secret := keySecret
	fmt.Println(paymentID, subscriptionID, keySecret)
	res1 := GenerateHmacSHA256(pID+"|"+sID, secret)
	fmt.Println("Result 1 (Pay|Sub):", res1)
	return true
}

func GenerateHmacSHA256Payment(data, secret string) string {
	// Ensure no accidental whitespace is breaking the signature
	key := []byte(strings.TrimSpace(secret))
	h := hmac.New(sha256.New, key)
	h.Write([]byte(strings.TrimSpace(data)))
	return hex.EncodeToString(h.Sum(nil))
}

func GenerateHmacSHA256(data, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func VerifyRazorpayWebhookSignature(body []byte, webhookSecret string, signature string) bool {
	h := hmac.New(sha256.New, []byte(webhookSecret))
	h.Write(body)
	computedSignature := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(computedSignature), []byte(signature))
}
