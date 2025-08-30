package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/apparentlyarhm/app-proxy-go/config"
)

type Client struct {
	config config.GitHubConfig
}

func NewClient(cfg config.GitHubConfig) *Client {

	return &Client{
		config: cfg,
	}
}

func (c *Client) GetGithubData() (any, error) {
	q := `
	query {
        user(login: "apparentlyarhm") {
          repositories(first: 8, orderBy: {field: PUSHED_AT, direction: DESC}) {
            nodes {
              name
              primaryLanguage {
                  name
                  color
              }
              defaultBranchRef {
                target {
                  ... on Commit {
                    history(first: 5) {
                      edges {
                        node {
                          committedDate
                          messageHeadline
                          url
                          author {
                            name
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
	`
	reqBody := GraphQLRequest{Query: q}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request body: %w", err)
	}

	url := "https://" + c.config.Host + "/graphql"
	log.Printf("[Github] gh url :: %v\n", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(reqBytes)))
	if err != nil {
		fmt.Println("Error forming request:", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.GhToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("[Github] Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body) // ignore error here, best effort
		return nil, fmt.Errorf("HTTP request failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// Unmarshal the raw JSON bytes into a generic map.
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding github response: %w", err)
	}

	return result, nil

}
