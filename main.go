package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type eventType string

const (
	CreateEvent      eventType = "CreateEvent"
	DeleteEvent      eventType = "DeleteEvent"
	PushEvent        eventType = "PushEvent"
	PullRequestEvent eventType = "PullRequestEvent"
	WatchEvent       eventType = "WatchEvent"
	ForkEvent        eventType = "ForkEvent"
	ReleaseEvent     eventType = "ReleaseEvent"
	IssueEvent       eventType = "IssueEvent"
)

type event struct {
	Type    eventType      `json:"type"`
	Payload map[string]any `json:"payload"`
	Repo    struct {
		Name string `json:"name"`
	} `json:"repo"`
}

func (e event) Format() string {
	switch e.Type {
	case PushEvent:
		return fmt.Sprintf("Pushed %v commit(s) to %s", e.Payload["distinct_size"], e.Repo.Name)
	case WatchEvent:
		return fmt.Sprintf("Starred %s", e.Repo.Name)
	case ForkEvent:
		return fmt.Sprintf("Forked %s", e.Repo.Name)
	case CreateEvent:
		refType := e.Payload["ref_type"]
		ref := e.Payload["ref"]
		switch refType {
		case "repository":
			return fmt.Sprintf("Created %v %s", refType, e.Repo.Name)
		default:
			return fmt.Sprintf("Created %v %v in %s", refType, ref, e.Repo.Name)
		}
	case DeleteEvent:
		refType := e.Payload["ref_type"]
		ref := e.Payload["ref"]
		switch refType {
		case "repository":
			return fmt.Sprintf("Deleted %v %s", refType, e.Repo.Name)
		default:
			return fmt.Sprintf("Deleted %v %v in %s", refType, ref, e.Repo.Name)
		}
	case PullRequestEvent:
		pr := (e.Payload["pull_request"].(map[string]any))["number"]
		return fmt.Sprintf(
			"%s pull request #%v in %s",
			strings.Title(e.Payload["action"].(string)),
			pr,
			e.Repo.Name,
		)
	case ReleaseEvent:
		return fmt.Sprintf(
			"%s release %s in %s",
			strings.Title(e.Payload["action"].(string)),
			(e.Payload["release"].(map[string]any))["tag_name"],
			e.Repo.Name,
		)
	case IssueEvent:
		return fmt.Sprintf(
			"%s issue in %s",
			strings.Title(e.Payload["action"].(string)),
			e.Repo.Name,
		)
	default:
		return fmt.Sprintf("%s in %s", e.Type, e.Repo.Name)
	}
}

func fetchEvents(username string) ([]event, error) {
	uri := fmt.Sprintf("https://api.github.com/users/%s/events", username)
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, nil
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, errors.New("username not found")
	}
	if string(res.StatusCode)[0] == '4' {
		return nil, errors.New("client error")
	}
	if string(res.StatusCode)[0] == '5' {
		return nil, errors.New("server error")
	}

	defer res.Body.Close()

	var events []event
	err = json.NewDecoder(res.Body).Decode(&events)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func run() error {
	args := make([]string, 1)
	copy(args, os.Args[1:])

	username := args[0]
	if username == "" {
		return errors.New("please provide username")
	}
	events, err := fetchEvents(args[0])
	if err != nil {
		return err
	}
	if len(events) == 0 {
		fmt.Println("no recent activity")
	}
	for _, event := range events {
		fmt.Println(event.Format())
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println("ERROR:", err)
	}
}
