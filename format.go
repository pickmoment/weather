package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ── WMO weather codes ──────────────────────────────────────────────────────

var wmoDescMap = map[int]string{
	0: "맑음", 1: "대체로 맑음", 2: "부분 흐림", 3: "흐림",
	45: "안개", 48: "서리 안개",
	51: "가벼운 이슬비", 53: "보통 이슬비", 55: "강한 이슬비",
	56: "빙결 이슬비(약)", 57: "빙결 이슬비(강)",
	61: "가벼운 비", 63: "보통 비", 65: "강한 비",
	66: "빙우(약)", 67: "빙우(강)",
	71: "가벼운 눈", 73: "보통 눈", 75: "강한 눈",
	77: "눈 결정",
	80: "가벼운 소나기", 81: "보통 소나기", 82: "강한 소나기",
	85: "약한 눈 소나기", 86: "강한 눈 소나기",
	95: "뇌우", 96: "약한 우박 뇌우", 99: "강한 우박 뇌우",
}

func wmoDesc(code int) string {
	if d, ok := wmoDescMap[code]; ok {
		return d
	}
	return fmt.Sprintf("코드%d", code)
}

// ── Wind direction ─────────────────────────────────────────────────────────

var windDirs = []string{"N", "NNE", "NE", "ENE", "E", "ESE", "SE", "SSE", "S", "SSW", "SW", "WSW", "W", "WNW", "NW", "NNW"}

func windDirStr(deg int) string {
	idx := int(float64(deg)/22.5+0.5) % 16
	return windDirs[idx]
}

// ── AQI ────────────────────────────────────────────────────────────────────

func aqiDesc(aqi int) string {
	switch {
	case aqi <= 0:
		return "-"
	case aqi <= 50:
		return "좋음"
	case aqi <= 100:
		return "보통"
	case aqi <= 150:
		return "민감군 위험"
	case aqi <= 200:
		return "나쁨"
	case aqi <= 300:
		return "매우 나쁨"
	default:
		return "위험"
	}
}

// ── Date helpers ───────────────────────────────────────────────────────────

var weekdays = []string{"일", "월", "화", "수", "목", "금", "토"}

func formatDate(s string) string {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return s
	}
	return fmt.Sprintf("%02d/%02d(%s)", t.Month(), t.Day(), weekdays[t.Weekday()])
}

func formatHour(s string) string {
	// "2024-01-15T14:00" → "14:00"
	if len(s) >= 16 {
		return s[11:16]
	}
	return s
}

// ── Shared helpers ─────────────────────────────────────────────────────────

const divider = "──────────────────────────────────────────"

func toJSON(v any) string {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
	return strings.TrimRight(buf.String(), "\n")
}

func tg(text string) string {
	return "```\n" + text + "\n```"
}

// ── Formatters ─────────────────────────────────────────────────────────────

func fmtNow(data *CurrentWeather, format string) string {
	if format == "json" {
		return toJSON(data)
	}

	lines := []string{
		fmt.Sprintf("%s, %s  |  %s", data.City, data.Country, strings.Replace(data.Time, "T", " ", 1)),
		divider,
		fmt.Sprintf("%s  %.1f°C (체감 %.1f°C)  습도 %d%%", data.WeatherDesc, data.Temperature, data.ApparentTemp, data.Humidity),
		fmt.Sprintf("바람 %.1fkm/h %s  UV %.1f  강수 %.1fmm", data.WindSpeed, data.WindDirStr, data.UVIndex, data.Precipitation),
	}

	if data.AQI > 0 {
		lines = append(lines, divider)
		lines = append(lines, fmt.Sprintf("공기질  PM2.5 %.1f  PM10 %.1f  AQI %d %s",
			data.PM25, data.PM10, data.AQI, data.AQIDesc))
	}

	return tg(strings.Join(lines, "\n"))
}

func fmtHourly(data *HourlyForecast, format string) string {
	if format == "json" {
		return toJSON(data)
	}

	lines := []string{
		fmt.Sprintf("%s 시간별 예보  (%s)", data.City, data.Country),
		divider,
	}

	for _, h := range data.Hours {
		line := fmt.Sprintf("%-5s  %-10s  %5.1f°C  강수 %3d%%  %4.1fmm  바람 %.1fkm/h",
			formatHour(h.Time),
			h.WeatherDesc,
			h.Temperature,
			h.PrecipProb,
			h.Precipitation,
			h.WindSpeed,
		)
		lines = append(lines, line)
	}

	return tg(strings.Join(lines, "\n"))
}

func fmtDaily(data *DailyForecast, format string) string {
	if format == "json" {
		return toJSON(data)
	}

	lines := []string{
		fmt.Sprintf("%s 일별 예보  (%s)", data.City, data.Country),
		divider,
	}

	for _, d := range data.Days {
		line := fmt.Sprintf("%s  %-10s  최고 %5.1f°C  최저 %5.1f°C  강수 %4.1fmm %3d%%  바람 %.1fkm/h  UV %.1f",
			formatDate(d.Date),
			d.WeatherDesc,
			d.TempMax,
			d.TempMin,
			d.PrecipSum,
			d.PrecipProbMax,
			d.WindSpeedMax,
			d.UVIndexMax,
		)
		lines = append(lines, line)
	}

	return tg(strings.Join(lines, "\n"))
}
