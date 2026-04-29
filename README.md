# weather

Open-Meteo와 Nominatim을 사용하는 날씨 CLI 도구.

현재 날씨·대기질, 시간별·일별 예보를 터미널에서 조회하고, Claude Code / Codex용 스킬로 설치할 수 있습니다.

## 빌드

```bash
go build -o weather .
```

## 사용법

```
weather <명령어> [옵션]
```

### 명령어

| 명령어 | 설명 |
|--------|------|
| `now` | 현재 날씨 및 공기질 조회 |
| `hourly` | 시간별 예보 조회 |
| `daily` | 일별 예보 조회 |
| `install` | AI 에이전트용 스킬 파일 설치 |

각 명령어에 `-h` 플래그로 도움말을 볼 수 있습니다.

---

### `now` — 현재 날씨

```bash
weather now <도시> [-f json|telegram]
```

- 현재 기온/체감온도, 습도, 바람, 강수량, UV 지수
- 미세먼지(PM2.5, PM10), 미국 AQI 포함

```bash
weather now 서울
weather now 부산
weather now Tokyo
```

### `hourly` — 시간별 예보

```bash
weather hourly <도시> [-n 시간수] [-f json|telegram]
```

- 현재 시각부터 N시간 예보 (기본: 24시간)

```bash
weather hourly 서울
weather hourly 서울 -n 12
weather hourly 제주 -n 48
```

### `daily` — 일별 예보

```bash
weather daily <도시> [-n 일수] [-f json|telegram]
```

- 오늘부터 N일 예보 (기본: 7일, 최대 16일)

```bash
weather daily 서울
weather daily 서울 -n 14
weather daily 뉴욕 -n 5
```

### 출력 형식 (`-f`)

| 값 | 설명 |
|----|------|
| `telegram` (기본) | 코드블록 형태, 읽기 좋음 |
| `json` | JSON 출력, 데이터 가공용 |

---

## AI 에이전트 스킬 설치

```bash
weather install
```

대화형 프롬프트로 에이전트(Claude Code / Codex)와 설치 범위(global / project)를 선택하면 스킬 파일을 설치합니다.

설치 후 `/wt` 명령으로 에이전트에서 날씨를 조회할 수 있습니다.

## 데이터 소스

- **지오코딩**: [Nominatim](https://nominatim.openstreetmap.org/) (OpenStreetMap) — 한국어 도시명 지원
- **날씨 예보**: [Open-Meteo](https://open-meteo.com/) — 무료, 인증 불필요
- **공기질**: Open-Meteo Air Quality API
