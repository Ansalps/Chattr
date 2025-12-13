package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func VerifyRazorpaySignature(paymentID, subscriptionID, signature string, keySecret string) bool {
	data := paymentID + "|" + subscriptionID
	fmt.Println(paymentID,subscriptionID,keySecret)
	computed := GenerateHmacSHA256(data, keySecret)
	fmt.Println(computed,signature)
	return computed == signature
}

func GenerateHmacSHA256(data, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func VerifyRazorpayWebhookSignature(body []byte,webhookSecret string, signature string) bool {
	h := hmac.New(sha256.New, []byte(webhookSecret))
	h.Write(body)
	computedSignature := hex.EncodeToString(h.Sum(nil))
	return  hmac.Equal([]byte(computedSignature), []byte(signature))
}
