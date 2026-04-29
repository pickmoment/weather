package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const skillName = "wt"

const skillContent = `---
name: wt
description: 날씨 정보를 조회하는 스킬. 현재 날씨, 시간별·일별 예보, 미세먼지/대기질 정보를 제공한다. 사용자가 날씨 정보를 묻거나 /wt를 입력했을 때 사용.
---

# wt

Open-Meteo와 Nominatim을 이용하는 날씨 CLI 도구 ` + "`weather`" + `을 실행해 날씨 데이터를 가져오는 스킬.

## CLI 위치

` + "`weather`" + ` 바이너리가 PATH에 설치되어 있어야 합니다.

모든 명령은 ` + "`weather <subcommand>`" + ` 형태로 실행합니다.

## 서브커맨드

### now — 현재 날씨 및 공기질

` + "```bash" + `
weather now <도시> [-f json|telegram]
` + "```" + `

- 현재 기온/체감온도, 습도, 바람, 강수량, UV 지수
- 미세먼지(PM2.5, PM10), 미국 AQI
- ` + "`-f`" + `: 출력 형식 (기본: telegram)

` + "```bash" + `
weather now 서울
weather now 부산
weather now Tokyo
` + "```" + `

반환 필드: city, country, time, temperature, apparent_temperature, humidity,
precipitation, weather_code, weather_desc, wind_speed, wind_direction,
wind_direction_str, uv_index, pm2_5, pm10, us_aqi, aqi_desc

### hourly — 시간별 예보

` + "```bash" + `
weather hourly <도시> [-n 시간수] [-f json|telegram]
` + "```" + `

- 현재 시각부터 N시간 예보
- ` + "`-n`" + `: 조회 시간 수 (기본: 24)
- ` + "`-f`" + `: 출력 형식 (기본: telegram)

` + "```bash" + `
weather hourly 서울
weather hourly 서울 -n 12
weather hourly 제주 -n 48
` + "```" + `

반환: city, country, hours[]{time, temperature, apparent_temperature,
humidity, precip_probability, precipitation, weather_code, weather_desc, wind_speed}

### daily — 일별 예보

` + "```bash" + `
weather daily <도시> [-n 일수] [-f json|telegram]
` + "```" + `

- 오늘부터 N일 예보 (최대 16일)
- ` + "`-n`" + `: 조회 일 수 (기본: 7)
- ` + "`-f`" + `: 출력 형식 (기본: telegram)

` + "```bash" + `
weather daily 서울
weather daily 서울 -n 14
weather daily 뉴욕 -n 5
` + "```" + `

반환: city, country, days[]{date, temp_max, temp_min, precip_sum,
precip_prob_max, weather_code, weather_desc, wind_speed_max, uv_index_max}

## 출력 형식

- ` + "`-f telegram`" + ` (기본): 코드블록(` + "``` ```" + `) 형태 — 사람이 읽기 좋음
- ` + "`-f json`" + `: JSON 출력 — 데이터 가공 시 사용

사용자에게 날씨를 보여줄 때는 기본 telegram 형식을 그대로 출력한다.

## 사용 패턴

**"서울 날씨 어때?" / "지금 날씨"**
` + "```bash" + `
weather now 서울
` + "```" + `

**"오늘 오후 날씨" / "시간별 예보"**
` + "```bash" + `
weather hourly 서울
` + "```" + `

**"이번 주 날씨" / "주간 예보"**
` + "```bash" + `
weather daily 서울
` + "```" + `

**"내일 비 와?" / "주말 날씨"**
` + "```bash" + `
weather daily 서울 -n 7
` + "```" + `

**"미세먼지 어때?"**
` + "```bash" + `
weather now 서울
# aqi_desc 필드 확인: 좋음/보통/민감군 위험/나쁨/매우 나쁨/위험
` + "```" + `

**"부산 다음주까지 날씨 알려줘"**
` + "```bash" + `
weather daily 부산 -n 14
` + "```" + `

## 데이터 소스

- 지오코딩: Nominatim (OpenStreetMap) — 한국어 도시명 지원
- 날씨 예보: Open-Meteo (무료, 무인증)
- 공기질: Open-Meteo Air Quality API

## 날씨 코드(weather_code) 주요 값

| 코드 | 설명 |
|------|------|
| 0 | 맑음 |
| 1-3 | 대체로 맑음 ~ 흐림 |
| 45,48 | 안개 |
| 51-55 | 이슬비 |
| 61-65 | 비 |
| 71-75 | 눈 |
| 80-82 | 소나기 |
| 95 | 뇌우 |

## 주의사항

- 도시명은 한국어/영어 모두 지원 (Nominatim 지오코딩)
- UV 지수는 야간에 0으로 표시됨
- 공기질 데이터는 위치에 따라 unavailable일 수 있음 (AQI 0)
- forecast_days 최대 16일 (Open-Meteo 제한)
`

var stdin = bufio.NewReader(os.Stdin)

func ask(question string, choices []string) string {
	for {
		fmt.Print(question)
		line, _ := stdin.ReadString('\n')
		ans := strings.TrimSpace(line)
		for _, c := range choices {
			if ans == c {
				return ans
			}
		}
		fmt.Printf("  %s 중 하나를 입력하세요.\n", strings.Join(choices, "/"))
	}
}

func confirm(question string) bool {
	fmt.Print(question)
	line, _ := stdin.ReadString('\n')
	ans := strings.TrimSpace(strings.ToLower(line))
	return ans == "" || ans == "y" || ans == "yes"
}

type installTarget struct {
	agent string
	scope string
	dir   string
}

var targets = []installTarget{
	{"Claude Code", "global", filepath.Join(os.Getenv("HOME"), ".claude", "skills")},
	{"Claude Code", "project", ".claude/skills"},
	{"Codex", "global", filepath.Join(os.Getenv("HOME"), ".agents", "skills")},
	{"Codex", "project", ".agents/skills"},
}

func runInstall(_ []string) {
	fmt.Println("에이전트를 선택하세요:")
	fmt.Println("  1) Claude Code")
	fmt.Println("  2) Codex")
	agent := ask("선택 [1/2]: ", []string{"1", "2"})

	fmt.Println("\n설치 범위를 선택하세요:")
	fmt.Println("  1) global  — 모든 프로젝트에서 사용")
	fmt.Println("  2) project — 현재 프로젝트에서만 사용")
	scope := ask("선택 [1/2]: ", []string{"1", "2"})

	idx := (atoi(agent)-1)*2 + (atoi(scope) - 1)
	t := targets[idx]

	dest := filepath.Join(t.dir, skillName, "SKILL.md")
	fmt.Printf("\n설치 위치: %s\n", dest)
	if !confirm("설치할까요? [Y/n]: ") {
		fmt.Println("취소했습니다.")
		return
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		errExit("디렉토리 생성 실패: " + err.Error())
	}
	if err := os.WriteFile(dest, []byte(skillContent), 0644); err != nil {
		errExit("파일 쓰기 실패: " + err.Error())
	}
	fmt.Printf("\n스킬 설치 완료: %s\n", dest)
	fmt.Printf("%s에서 /%s 으로 호출할 수 있습니다.\n", t.agent, skillName)
}

func atoi(s string) int {
	if s == "2" {
		return 2
	}
	return 1
}
