package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type slackClient struct {
	token string
}

func newSlackClient(token string) *slackClient {
	return &slackClient{token: token}
}

type slackResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

func (c *slackClient) post(method string, params url.Values) ([]byte, error) {
	req, err := http.NewRequest("POST", "https://slack.com/api/"+method,
		strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var buf []byte
	buf = make([]byte, 0, 4096)
	tmp := make([]byte, 512)
	for {
		n, err := resp.Body.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if err != nil {
			break
		}
	}
	return buf, nil
}

// lookupUserID finds a Slack user ID by display name or real name.
func (c *slackClient) lookupUserID(username string) (string, error) {
	// Strip leading @ if present
	username = strings.TrimPrefix(username, "@")

	params := url.Values{"limit": {"200"}}
	for {
		body, err := c.post("users.list", params)
		if err != nil {
			return "", err
		}

		var result struct {
			slackResponse
			Members []struct {
				ID      string `json:"id"`
				Name    string `json:"name"`
				Profile struct {
					DisplayName string `json:"display_name"`
					RealName    string `json:"real_name"`
				} `json:"profile"`
				Deleted bool `json:"deleted"`
			} `json:"members"`
			ResponseMetadata struct {
				NextCursor string `json:"next_cursor"`
			} `json:"response_metadata"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return "", fmt.Errorf("parse error: %w", err)
		}
		if !result.OK {
			return "", fmt.Errorf("slack error: %s", result.Error)
		}

		for _, m := range result.Members {
			if m.Deleted {
				continue
			}
			if m.Name == username ||
				strings.EqualFold(m.Profile.DisplayName, username) ||
				strings.EqualFold(m.Profile.RealName, username) {
				return m.ID, nil
			}
		}

		cursor := result.ResponseMetadata.NextCursor
		if cursor == "" {
			break
		}
		params = url.Values{"limit": {"200"}, "cursor": {cursor}}
	}
	return "", fmt.Errorf("user not found: %s", username)
}

// openDM opens (or reuses) a DM channel with the given user ID.
func (c *slackClient) openDM(userID string) (string, error) {
	body, err := c.post("conversations.open", url.Values{"users": {userID}})
	if err != nil {
		return "", err
	}

	var result struct {
		slackResponse
		Channel struct {
			ID string `json:"id"`
		} `json:"channel"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}
	if !result.OK {
		return "", fmt.Errorf("slack error: %s", result.Error)
	}
	return result.Channel.ID, nil
}

// sendMessage posts a message to a channel or DM.
func (c *slackClient) sendMessage(channelID, text string) error {
	body, err := c.post("chat.postMessage", url.Values{
		"channel": {channelID},
		"text":    {text},
	})
	if err != nil {
		return err
	}

	var result slackResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parse error: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("slack error: %s", result.Error)
	}
	return nil
}

// SendDM looks up a user by name and sends them a DM.
func (c *slackClient) SendDM(username, message string) error {
	userID, err := c.lookupUserID(username)
	if err != nil {
		return err
	}
	channelID, err := c.openDM(userID)
	if err != nil {
		return err
	}
	return c.sendMessage(channelID, message)
}
