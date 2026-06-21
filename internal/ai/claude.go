package ai

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/alittlelate/news-briefing/internal/rss"
	"github.com/anthropics/anthropic-sdk-go"
)

func LoadPrompt(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	prompt := string(data)

	return prompt, nil
}

func BuildArticlesList(articles []rss.Article) string {
	var sb strings.Builder

	for i, a := range articles {
		fmt.Fprintf(&sb, "[%d] TITLE: %s\nSOURCE: %s\nURL: %s\nTEXT: %s\n\n",
			i+1, a.Title, a.Source, a.URL, a.Description)
	}

	return sb.String()
}

func GenerateReport(systemPrompt string, articles []rss.Article) (string, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	articlesList := BuildArticlesList(articles)

	userMessage := fmt.Sprintf("Today is %s.\n\nHere are today's articles\n\n%s", time.Now().Format("Monday, January 2, 2006"), articlesList)

	client := anthropic.NewClient()

	response, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 8192,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)),
		},
	})
	if err != nil {
		log.Printf("Failed to recieve valid answer from claude's api: %v", err)
		return "", err
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude")
	}
	return response.Content[0].Text, nil
}
