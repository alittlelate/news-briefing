package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/alittlelate/news-briefing/internal/ai"
	"github.com/alittlelate/news-briefing/internal/email"
	"github.com/alittlelate/news-briefing/internal/report"
	"github.com/alittlelate/news-briefing/internal/rss"
)

func main() {
	cfg, err := rss.LoadConfig("feeds.yml")
	if err != nil {
		log.Fatalf("failed to load feeds: %v", err)
	}

	systemPrompt, err := ai.LoadPrompt("prompt.txt")
	if err != nil {
		log.Fatalf("failed to load feeds: %v", err)
	}

	log.Println("fetching RSS feeds...")
	articles := rss.FetchAll(cfg)
	log.Printf("fetched %d articles", len(articles))

	log.Println("generating report...")
	reportText, err := ai.GenerateReport(systemPrompt, articles)
	if err != nil {
		log.Fatalf("failed to generate report %v", err)
	}

	path, err := report.Write(reportText)
	if err != nil {
		log.Fatalf("failed to write report %v", err)
	} else {
		log.Printf("report written to %s: ", path)
	}

	log.Println("parsing report...")
	parsedReport, err := email.ParseReport(reportText)
	if err != nil {
		log.Fatalf("failed to parse report: %v", err)
	}

	log.Println("rendering html report for email...")
	renderedReport, err := email.RenderHTML(parsedReport)
	if err != nil {
		log.Fatalf("failed to render report: %v", err)
	}

	log.Println("sending email...")
	subject := fmt.Sprintf("Briefing — %s", time.Now().Format("Monday, January 2, 2006"))

	if err := email.Send(subject, renderedReport); err != nil {
		log.Fatalf("failed to send email %v", err)
	}
}
