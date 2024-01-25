package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Load environment variables from .env file
func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

const (
	dateFormat = "2006-01-02"
	eventsAPI  = "https://api.github.com/users/%s/events"
)

func main() {
	// Load environment variables
	loadEnv()

	// Get GitHub token and username from environment variables
	githubToken := os.Getenv("GITHUB_TOKEN")
	username := os.Getenv("GITHUB_USERNAME")

	// Get today's date in the format used in GitHub events
	today := time.Now().Format(dateFormat)

	// Get daily events from GitHub profile
	dailyEvents, err := getDailyEvents(today, username, githubToken)
	if err != nil {
		log.Fatal(err)
	}

	// Format events for the report
	report := formatEvents(dailyEvents)

	// Print or save the report as needed
	fmt.Println(report)

	// Save the report to a file
	reportFile := os.Getenv("REPORT_FILE")
	err = ioutil.WriteFile(reportFile, []byte(report), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func getDailyEvents(date, username, token string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf(eventsAPI, username)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var events []map[string]interface{}
	err = parseJSON(body, &events)
	if err != nil {
		return nil, err
	}

	var dailyEvents []map[string]interface{}
	for _, event := range events {
		createdAt, ok := event["created_at"].(string)
		if !ok {
			continue
		}

		if startTime, err := time.Parse(time.RFC3339, createdAt); err == nil {
			if startTime.Format(dateFormat) == date {
				dailyEvents = append(dailyEvents, event)
			}
		}
	}

	return dailyEvents, nil
}

func formatEvents(events []map[string]interface{}) string {
	report := fmt.Sprintf("%s:\n", time.Now().Format("Jan 02, 2006"))

	// Default lines in every report
	report += "• Done | Attended frail-check meeting\n"
	report += "• Done | Attended frail-check followup meeting\n"

	// Keep track of seen pull request titles
	seenTitles := make(map[string]bool)

	for _, event := range events {
		eventType, ok := event["type"].(string)
		if !ok {
			continue
		}

		var prTitle, status, action, author string
		merged := false

		switch eventType {
		case "PullRequestEvent":
			action, _ = event["payload"].(map[string]interface{})["action"].(string)
			merged, _ = event["payload"].(map[string]interface{})["pull_request"].(map[string]interface{})["merged"].(bool)
			prTitle, _ = event["payload"].(map[string]interface{})["pull_request"].(map[string]interface{})["title"].(string)
			author, _ = event["payload"].(map[string]interface{})["pull_request"].(map[string]interface{})["user"].(map[string]interface{})["login"].(string)

		case "PullRequestReviewEvent":
			action, _ = event["payload"].(map[string]interface{})["action"].(string)
			merged, _ = event["payload"].(map[string]interface{})["pull_request"].(map[string]interface{})["merged"].(bool)
			prTitle, _ = event["payload"].(map[string]interface{})["pull_request"].(map[string]interface{})["title"].(string)
			author, _ = event["payload"].(map[string]interface{})["review"].(map[string]interface{})["user"].(map[string]interface{})["login"].(string)
		}

		// Check if the title has been seen before
		if !seenTitles[prTitle] {
			seenTitles[prTitle] = true

			switch {
			//myside
			case author == os.Getenv("GITHUB_USERNAME") && action == "opened" && merged:
				status = "Done"
			case author == os.Getenv("GITHUB_USERNAME") && action == "opened":
				status = "In Review"
			case author == os.Getenv("GITHUB_USERNAME") && action == "closed" && merged:
				status = "Done"
			case author == os.Getenv("GITHUB_USERNAME") && merged:
				status = "Done"

			//other side
			case author != os.Getenv("GITHUB_USERNAME") && action == "closed" && merged:
				status = "Reviewed and merged"
			case author != os.Getenv("GITHUB_USERNAME") && action == "closed":
				status = "Reviewed"
			}

			// Append to the report only if status and prTitle are not empty
			if status != "" && prTitle != "" {
				report += fmt.Sprintf("• %s | %s\n", status, prTitle)
			}
		}
	}

	report += "Next:\n• Continue with assigned task and R&D\n"

	return report
}


func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
