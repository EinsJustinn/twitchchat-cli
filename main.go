package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

type User struct {
	Username string
	Color    string
	Message  string
}

const (
	twitchServer = "irc.chat.twitch.tv:6667"
	password     = "SCHMOOPIIE"
	defaultColor = "#FFFFFF"
)

var writer *bufio.Writer

func main() {

	channelName := flag.String("channel", "", "Auto connect to channel")

	flag.Parse()

	setTitle("Twitch Chat")

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	nickname := "justinfan" + strconv.Itoa(r.Intn(90000)+10000)

	channel := *channelName
	if channel == "" {
		fmt.Print("Enter channel: ")
		_, err := fmt.Scanln(&channel)
		if err != nil {
			fmt.Println("Error reading channel")
			return
		}
	}

	fmt.Println("Connecting to Twitch IRC")

	dial, err := net.Dial("tcp", twitchServer)
	if err != nil {
		panic(err)
	}
	defer dial.Close()

	fmt.Println("Connected to Twitch IRC")

	writer = bufio.NewWriter(dial)
	reader := bufio.NewReader(dial)

	fmt.Println("Trying to login")

	writeToChat("CAP REQ :twitch.tv/tags")
	writeToChat("PASS " + password)
	writeToChat("NICK " + nickname)
	_ = writer.Flush()

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		splits := strings.SplitN(line, " ", 3)

		if isInt(splits[1]) {
			if splits[1] == "001" {
				fmt.Println("Logged in")
				writeToChat("JOIN #" + channel)

				fmt.Println("Join channel")
				_ = writer.Flush()
			}
			continue
		}
		if splits[1] == "JOIN" {
			setTitle("Twitch Chat - " + channel)
			clearScreen()
			continue
		}

		if strings.HasPrefix(line, "@") {
			user, err := getUser(line)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Print(colorize(user.Username, user.Color) + ": " + user.Message)
			continue
		}

	}
}

func getUser(s string) (User, error) {
	var user User

	for _, part := range strings.Split(s, ";") {
		splitN := strings.SplitN(part, "=", 2)
		if len(splitN) != 2 {
			continue
		}

		key := splitN[0]
		value := splitN[1]

		switch key {
		case "color":
			if value == "" {
				user.Color = defaultColor
			} else {
				user.Color = value
			}
		case "user-type", "vip":
			if value == "" {
				continue
			}

			message, err := parseUserMessage(value)
			if err != nil {
				return User{}, fmt.Errorf("fehler beim Parsen der Nachricht: %w", err)
			}

			user.Message = message
		case "display-name":
			user.Username = value
		}
	}

	return user, nil
}

func parseUserMessage(s string) (string, error) {
	splits := strings.SplitN(s, ":", 2)
	if len(splits) != 2 {
		return "", fmt.Errorf("failed to parse user message: %s", s)
	}

	message := strings.SplitN(splits[1], " ", 4)[3][1:]

	return message, nil
}

func colorize(text string, hexColor string) string {
	var r, g, b int
	_, err := fmt.Sscanf(hexColor, "#%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("\033[38;2;%d;%d;%dm%s\033[0m", r, g, b, text)
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func setTitle(s string) {
	fmt.Printf("\033]0;%s\007", s)
}

func isInt(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func writeToChat(s string) {
	_, _ = writer.WriteString(s)
	_, _ = writer.WriteString("\r\n")
}
