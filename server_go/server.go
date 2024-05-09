package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type apiTokenRequest struct {
	Code string `json:"code"`
}

type apiTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func startServer() {
	router := mux.NewRouter()
	router.HandleFunc("/api/token", apiToken).Methods("POST")

	fmt.Println("Listening at :3001")
	err := http.ListenAndServe(":3001", router)
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}

func main() {
	godotenv.Load("../.env")

	go startServer()

	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_TOKEN"))
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.Identify.Intents |= discordgo.IntentMessageContent
	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func apiToken(w http.ResponseWriter, r *http.Request) {
	var body apiTokenRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		fmt.Println("Failed to decode body:", err)
		return
	}

	fmt.Println("Received code:", body.Code)

	data := url.Values{}
	data.Set("client_id", os.Getenv("VITE_DISCORD_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("DISCORD_CLIENT_SECRET"))
	data.Set("grant_type", "authorization_code")
	data.Set("code", body.Code)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://discord.com/api/oauth2/token", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, _ := client.Do(req)
	defer resp.Body.Close()

	var tokenResponse apiTokenResponse
	json.NewDecoder(resp.Body).Decode(&tokenResponse)

	fmt.Println("Received token:", tokenResponse.AccessToken)

	json.NewEncoder(w).Encode(map[string]string{"access_token": tokenResponse.AccessToken})
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	mentions := m.Mentions
	if len(mentions) > 0 && mentions[0].ID == s.State.User.ID {
		s.ChannelMessageSend(m.ChannelID, "https://poker.sogetsu.online")
	}
}
