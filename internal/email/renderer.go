package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"
)

type Source struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	Source string `json:"source"`
}

type BriefingItem struct {
	Tier         string   `json:"tier"`
	Title        string   `json:"title"`
	Summary      string   `json:"summary"`
	SingleSource bool     `json:"single_source"`
	Sources      []Source `json:"sources"`
}

type ContentRadarItem struct {
	Title        string `json:"title"`
	Angle        string `json:"angle"`
	ItalianHook  string `json:"italian_hook"`
	OpeningImage string `json:"opening_image"`
	Missing      string `json:"missing"`
	Readiness    int    `json:"readiness"`
}

type Report struct {
	Date          string             `json:"date"`
	Briefing      []BriefingItem     `json:"briefing"`
	WatchTomorrow []string           `json:"watch_tomorrow"`
	ContentRadar  []ContentRadarItem `json:"content_radar"`
}

func ParseReport(raw string) (*Report, error) {
	// strip markdown code fences if model added them anyway
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var r Report
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		return nil, fmt.Errorf("failed to parse report JSON: %w", err)
	}
	return &r, nil
}

var tierConfig = map[string]struct {
	Color string
	Label string
	Emoji string
}{
	"critical":    {"#c0392b", "CRITICAL", "🔴"},
	"significant": {"#e67e22", "SIGNIFICANT", "🟡"},
	"context":     {"#27ae60", "CONTEXT", "🟢"},
}

var readinessLabels = map[int]string{
	5: "Ready to film",
	4: "Nearly ready",
	3: "Needs one data point",
	2: "Needs research",
	1: "Direction only",
}

func RenderHTML(r *Report) (string, error) {
	funcMap := template.FuncMap{
		"tierColor": func(tier string) string {
			if cfg, ok := tierConfig[tier]; ok {
				return cfg.Color
			}
			return "#888"
		},
		"tierLabel": func(tier string) string {
			if cfg, ok := tierConfig[tier]; ok {
				return cfg.Emoji + " " + cfg.Label
			}
			return tier
		},
		"readinessLabel": func(r int) string {
			if label, ok := readinessLabels[r]; ok {
				return label
			}
			return ""
		},
		"readinessColor": func(r int) string {
			switch {
			case r >= 4:
				return "#27ae60"
			case r == 3:
				return "#e67e22"
			default:
				return "#c0392b"
			}
		},
		"repeat": func(s string, n int) string {
			return strings.Repeat(s, n)
		},
		"formatDate": func(d string) string {
			t, err := time.Parse("2006-01-02", d)
			if err != nil {
				return d
			}
			return t.Format("Monday, January 2, 2006")
		},
	}

	tmpl, err := template.New("report").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, r); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func SaveHTML(r *Report, path string) error {
	html, err := RenderHTML(r)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(html), 0644)
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Daily Briefing — {{formatDate .Date}}</title>
</head>
<body style="margin:0;padding:0;background:#f4f4f4;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif;">

<table width="100%" cellpadding="0" cellspacing="0" style="background:#f4f4f4;padding:24px 0;">
<tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%;">

  <!-- HEADER -->
  <tr><td style="background:#1a1a2e;border-radius:8px 8px 0 0;padding:32px 32px 24px;">
    <div style="font-size:11px;font-weight:600;letter-spacing:0.15em;color:#8b8fa8;text-transform:uppercase;margin-bottom:8px;">Intelligence Briefing</div>
    <div style="font-size:28px;font-weight:700;color:#ffffff;line-height:1.2;">{{formatDate .Date}}</div>
  </td></tr>

  <!-- BRIEFING SECTION HEADER -->
  <tr><td style="background:#16213e;padding:16px 32px;">
    <div style="font-size:11px;font-weight:700;letter-spacing:0.12em;color:#8b8fa8;text-transform:uppercase;">§ Daily Briefing</div>
  </td></tr>

  <!-- BRIEFING ITEMS -->
  {{range .Briefing}}
  <tr><td style="background:#ffffff;padding:0 32px;">
    <div style="border-left:4px solid {{tierColor .Tier}};margin:20px 0 0;padding:16px 20px;background:#fafafa;border-radius:0 6px 6px 0;">

      <!-- tier badge + title -->
      <div style="margin-bottom:10px;">
        <span style="display:inline-block;background:{{tierColor .Tier}};color:#fff;font-size:10px;font-weight:700;letter-spacing:0.1em;padding:3px 8px;border-radius:3px;text-transform:uppercase;">{{tierLabel .Tier}}</span>
        {{if .SingleSource}}<span style="display:inline-block;background:#f39c12;color:#fff;font-size:10px;font-weight:700;padding:3px 8px;border-radius:3px;margin-left:6px;text-transform:uppercase;">SINGLE SOURCE</span>{{end}}
      </div>

      <div style="font-size:17px;font-weight:700;color:#1a1a2e;line-height:1.3;margin-bottom:10px;">{{.Title}}</div>
      <div style="font-size:14px;color:#444;line-height:1.7;margin-bottom:14px;">{{.Summary}}</div>

      <!-- sources -->
      <div style="font-size:11px;color:#888;border-top:1px solid #eee;padding-top:10px;">
        <span style="font-weight:600;text-transform:uppercase;letter-spacing:0.05em;">Sources: </span>
        {{range $i, $s := .Sources}}{{if $i}} · {{end}}<a href="{{$s.URL}}" style="color:#3498db;text-decoration:none;">{{$s.Source}}</a>{{end}}
      </div>

    </div>
  </td></tr>
  {{end}}

  <tr><td style="height:8px;background:#ffffff;"></td></tr>

  <!-- WATCH TOMORROW -->
  <tr><td style="background:#1a1a2e;padding:20px 32px;margin-top:8px;">
    <div style="font-size:11px;font-weight:700;letter-spacing:0.12em;color:#8b8fa8;text-transform:uppercase;margin-bottom:14px;">👁 Watch Tomorrow</div>
    {{range .WatchTomorrow}}
    <div style="font-size:13px;color:#e0e0e0;line-height:1.6;padding:6px 0;border-bottom:1px solid #2a2a4e;">→ {{.}}</div>
    {{end}}
  </td></tr>

  <!-- CONTENT RADAR HEADER -->
  <tr><td style="background:#0f3460;padding:16px 32px;">
    <div style="font-size:11px;font-weight:700;letter-spacing:0.12em;color:#8b8fa8;text-transform:uppercase;">§ Content Radar</div>
    <div style="font-size:12px;color:#aaa;margin-top:4px;">Stories worth turning into a Short</div>
  </td></tr>

  <!-- CONTENT RADAR ITEMS -->
  {{range .ContentRadar}}
  <tr><td style="background:#ffffff;padding:0 32px;">
    <div style="border:1px solid #e8e0f0;border-left:4px solid #8e44ad;margin:20px 0 0;padding:20px;border-radius:0 6px 6px 0;background:#fdf9ff;">

      <!-- readiness badge -->
      <div style="margin-bottom:12px;">
        <span style="display:inline-block;background:{{readinessColor .Readiness}};color:#fff;font-size:10px;font-weight:700;padding:3px 10px;border-radius:3px;">
          {{repeat "●" .Readiness}}{{repeat "○" (5 | subtract .Readiness)}} {{readinessLabel .Readiness}}
        </span>
      </div>

      <div style="font-size:16px;font-weight:700;color:#1a1a2e;margin-bottom:14px;">{{.Title}}</div>

      <table width="100%" cellpadding="0" cellspacing="0">
        <tr>
          <td style="padding:8px 0;border-bottom:1px solid #ede8f5;vertical-align:top;width:120px;">
            <div style="font-size:10px;font-weight:700;color:#8e44ad;text-transform:uppercase;letter-spacing:0.08em;">Angle</div>
          </td>
          <td style="padding:8px 0 8px 12px;border-bottom:1px solid #ede8f5;">
            <div style="font-size:13px;color:#333;line-height:1.6;">{{.Angle}}</div>
          </td>
        </tr>
        <tr>
          <td style="padding:8px 0;border-bottom:1px solid #ede8f5;vertical-align:top;">
            <div style="font-size:10px;font-weight:700;color:#8e44ad;text-transform:uppercase;letter-spacing:0.08em;">🇮🇹 Italian hook</div>
          </td>
          <td style="padding:8px 0 8px 12px;border-bottom:1px solid #ede8f5;">
            <div style="font-size:13px;color:#333;line-height:1.6;">{{.ItalianHook}}</div>
          </td>
        </tr>
        <tr>
          <td style="padding:8px 0;border-bottom:1px solid #ede8f5;vertical-align:top;">
            <div style="font-size:10px;font-weight:700;color:#8e44ad;text-transform:uppercase;letter-spacing:0.08em;">Opening image</div>
          </td>
          <td style="padding:8px 0 8px 12px;border-bottom:1px solid #ede8f5;">
            <div style="font-size:13px;color:#333;line-height:1.6;">{{.OpeningImage}}</div>
          </td>
        </tr>
        <tr>
          <td style="padding:8px 0;vertical-align:top;">
            <div style="font-size:10px;font-weight:700;color:#8e44ad;text-transform:uppercase;letter-spacing:0.08em;">Missing</div>
          </td>
          <td style="padding:8px 0 8px 12px;">
            <div style="font-size:13px;color:#666;line-height:1.6;font-style:italic;">{{.Missing}}</div>
          </td>
        </tr>
      </table>

    </div>
  </td></tr>
  {{end}}

  <!-- FOOTER -->
  <tr><td style="height:8px;background:#ffffff;"></td></tr>
  <tr><td style="background:#1a1a2e;border-radius:0 0 8px 8px;padding:20px 32px;text-align:center;">
    <div style="font-size:11px;color:#555;line-height:1.6;">Generated by news-briefing · Sources linked above</div>
  </td></tr>

</table>
</td></tr>
</table>

</body>
</html>`
