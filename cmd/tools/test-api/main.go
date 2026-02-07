package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	// Login
	loginData := map[string]string{
		"email":    "mahasiswa1@campus.edu",
		"password": "Password123!",
	}
	body, _ := json.Marshal(loginData)

	resp, err := http.Post("http://localhost:8102/api/v1/auth/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Login connection failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Login failed with status %d: %s\n", resp.StatusCode, string(respBody))
		return
	}

	var result struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	json.Unmarshal(respBody, &result)

	token := result.Data.Token
	if token == "" {
		// Fallback for different response structure
		var result2 struct {
			Token string `json:"token"`
		}
		json.Unmarshal(respBody, &result2)
		token = result2.Token
	}

	if token == "" {
		fmt.Println("Failed to parse token from response")
		fmt.Println(string(respBody))
		return
	}

	fmt.Printf("Token found: %s...\n", token[:10])

	// Try hit dosen missions
	req, _ := http.NewRequest("GET", "http://localhost:8102/api/v1/dosen/missions", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp2, err := client.Do(req)
	if err != nil {
		fmt.Printf("Missions request failed: %v\n", err)
		return
	}
	defer resp2.Body.Close()

	fmt.Printf("Missions status: %d\n", resp2.StatusCode)
	respBody2, _ := ioutil.ReadAll(resp2.Body)
	fmt.Printf("Missions response: %s\n", string(respBody2))
}
