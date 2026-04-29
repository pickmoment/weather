package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	nominatimURL  = "https://nominatim.openstreetmap.org/search"
	openMeteoURL  = "https://api.open-meteo.com/v1/forecast"
	airQualityURL = "https://air-quality-api.open-meteo.com/v1/air-quality"
	tz            = "Asia/Seoul"
)

// ── Nominatim ──────────────────────────────────────────────────────────────

type geoResult struct {
	Name    string `json:"name"`
	Lat     string `json:"lat"`
	Lon     string `json:"lon"`
	Address struct {
		Country     string `json:"country"`
		CountryCode string `json:"country_code"`
	} `json:"address"`
}

func geocode(city string) (*geoResult, error) {
	u := nominatimURL + "?q=" + url.QueryEscape(city) +
		"&format=json&limit=1&accept-language=ko&addressdetails=1"
	data, err := getJSON(u)
	if err != nil {
		return nil, err
	}
	var results []geoResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("도시를 찾을 수 없습니다: %s", city)
	}
	return &results[0], nil
}

// ── Output types ───────────────────────────────────────────────────────────

type CurrentWeather struct {
	City          string  `json:"city"`
	Country       string  `json:"country"`
	Time          string  `json:"time"`
	Temperature   float64 `json:"temperature"`
	ApparentTemp  float64 `json:"apparent_temperature"`
	Humidity      int     `json:"humidity"`
	Precipitation float64 `json:"precipitation"`
	WeatherCode   int     `json:"weather_code"`
	WeatherDesc   string  `json:"weather_desc"`
	WindSpeed     float64 `json:"wind_speed"`
	WindDir       int     `json:"wind_direction"`
	WindDirStr    string  `json:"wind_direction_str"`
	UVIndex       float64 `json:"uv_index"`
	PM25          float64 `json:"pm2_5"`
	PM10          float64 `json:"pm10"`
	AQI           int     `json:"us_aqi"`
	AQIDesc       string  `json:"aqi_desc"`
}

type HourlyEntry struct {
	Time          string  `json:"time"`
	Temperature   float64 `json:"temperature"`
	ApparentTemp  float64 `json:"apparent_temperature"`
	Humidity      int     `json:"humidity"`
	PrecipProb    int     `json:"precip_probability"`
	Precipitation float64 `json:"precipitation"`
	WeatherCode   int     `json:"weather_code"`
	WeatherDesc   string  `json:"weather_desc"`
	WindSpeed     float64 `json:"wind_speed"`
}

type HourlyForecast struct {
	City    string        `json:"city"`
	Country string        `json:"country"`
	Hours   []HourlyEntry `json:"hours"`
}

type DailyEntry struct {
	Date          string  `json:"date"`
	TempMax       float64 `json:"temp_max"`
	TempMin       float64 `json:"temp_min"`
	PrecipSum     float64 `json:"precip_sum"`
	PrecipProbMax int     `json:"precip_prob_max"`
	WeatherCode   int     `json:"weather_code"`
	WeatherDesc   string  `json:"weather_desc"`
	WindSpeedMax  float64 `json:"wind_speed_max"`
	UVIndexMax    float64 `json:"uv_index_max"`
}

type DailyForecast struct {
	City    string       `json:"city"`
	Country string       `json:"country"`
	Days    []DailyEntry `json:"days"`
}

// ── Internal API response structs ──────────────────────────────────────────

type apiErrResp struct {
	Error  bool   `json:"error"`
	Reason string `json:"reason"`
}

type forecastCurrentResp struct {
	Current struct {
		Time             string  `json:"time"`
		Temperature2m    float64 `json:"temperature_2m"`
		RelativeHumidity int     `json:"relative_humidity_2m"`
		ApparentTemp     float64 `json:"apparent_temperature"`
		Precipitation    float64 `json:"precipitation"`
		WeatherCode      int     `json:"weather_code"`
		Windspeed10m     float64 `json:"windspeed_10m"`
		WindDir10m       int     `json:"winddirection_10m"`
		UVIndex          float64 `json:"uv_index"`
	} `json:"current"`
}

type airQualityResp struct {
	Current struct {
		PM25  float64 `json:"pm2_5"`
		PM10  float64 `json:"pm10"`
		USAQI int     `json:"us_aqi"`
	} `json:"current"`
}

type forecastHourlyResp struct {
	Hourly struct {
		Time         []string  `json:"time"`
		Temperature  []float64 `json:"temperature_2m"`
		Humidity     []int     `json:"relative_humidity_2m"`
		ApparentTemp []float64 `json:"apparent_temperature"`
		PrecipProb   []int     `json:"precipitation_probability"`
		Precip       []float64 `json:"precipitation"`
		WeatherCode  []int     `json:"weather_code"`
		Windspeed    []float64 `json:"windspeed_10m"`
	} `json:"hourly"`
}

type forecastDailyResp struct {
	Daily struct {
		Time          []string  `json:"time"`
		TempMax       []float64 `json:"temperature_2m_max"`
		TempMin       []float64 `json:"temperature_2m_min"`
		PrecipSum     []float64 `json:"precipitation_sum"`
		PrecipProbMax []int     `json:"precipitation_probability_max"`
		WeatherCode   []int     `json:"weather_code"`
		WindspeedMax  []float64 `json:"windspeed_10m_max"`
		UVIndexMax    []float64 `json:"uv_index_max"`
	} `json:"daily"`
}

// ── Helpers ────────────────────────────────────────────────────────────────

func safeF(s []float64, i int) float64 {
	if i < len(s) {
		return s[i]
	}
	return 0
}

func safeI(s []int, i int) int {
	if i < len(s) {
		return s[i]
	}
	return 0
}

func checkAPIErr(data []byte) error {
	var e apiErrResp
	if json.Unmarshal(data, &e) == nil && e.Error {
		return fmt.Errorf("API 오류: %s", e.Reason)
	}
	return nil
}

// ── Fetch functions ────────────────────────────────────────────────────────

func fetchNow(city string) (*CurrentWeather, error) {
	geo, err := geocode(city)
	if err != nil {
		return nil, err
	}

	fURL := fmt.Sprintf(
		"%s?latitude=%s&longitude=%s&current=temperature_2m,relative_humidity_2m,apparent_temperature,precipitation,weather_code,windspeed_10m,winddirection_10m,uv_index&timezone=%s",
		openMeteoURL, geo.Lat, geo.Lon, url.QueryEscape(tz),
	)
	fData, err := getJSON(fURL)
	if err != nil {
		return nil, err
	}
	if err := checkAPIErr(fData); err != nil {
		return nil, err
	}
	var fr forecastCurrentResp
	if err := json.Unmarshal(fData, &fr); err != nil {
		return nil, err
	}

	aURL := fmt.Sprintf(
		"%s?latitude=%s&longitude=%s&current=pm2_5,pm10,us_aqi&timezone=%s",
		airQualityURL, geo.Lat, geo.Lon, url.QueryEscape(tz),
	)
	aData, err := getJSON(aURL)
	if err != nil {
		return nil, err
	}
	var ar airQualityResp
	_ = json.Unmarshal(aData, &ar)

	c := fr.Current
	return &CurrentWeather{
		City:          geo.Name,
		Country:       strings.ToUpper(geo.Address.CountryCode),
		Time:          c.Time,
		Temperature:   c.Temperature2m,
		ApparentTemp:  c.ApparentTemp,
		Humidity:      c.RelativeHumidity,
		Precipitation: c.Precipitation,
		WeatherCode:   c.WeatherCode,
		WeatherDesc:   wmoDesc(c.WeatherCode),
		WindSpeed:     c.Windspeed10m,
		WindDir:       c.WindDir10m,
		WindDirStr:    windDirStr(c.WindDir10m),
		UVIndex:       c.UVIndex,
		PM25:          ar.Current.PM25,
		PM10:          ar.Current.PM10,
		AQI:           ar.Current.USAQI,
		AQIDesc:       aqiDesc(ar.Current.USAQI),
	}, nil
}

func fetchHourly(city string, n int) (*HourlyForecast, error) {
	geo, err := geocode(city)
	if err != nil {
		return nil, err
	}

	forecastDays := (n+23)/24 + 1
	if forecastDays < 2 {
		forecastDays = 2
	}

	u := fmt.Sprintf(
		"%s?latitude=%s&longitude=%s&hourly=temperature_2m,relative_humidity_2m,apparent_temperature,precipitation_probability,precipitation,weather_code,windspeed_10m&timezone=%s&forecast_days=%d",
		openMeteoURL, geo.Lat, geo.Lon, url.QueryEscape(tz), forecastDays,
	)
	data, err := getJSON(u)
	if err != nil {
		return nil, err
	}
	if err := checkAPIErr(data); err != nil {
		return nil, err
	}
	var resp forecastHourlyResp
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	seoulLoc, _ := time.LoadLocation("Asia/Seoul")
	nowStr := time.Now().In(seoulLoc).Format("2006-01-02T15:00")

	startIdx := 0
	for i, t := range resp.Hourly.Time {
		if t >= nowStr {
			startIdx = i
			break
		}
	}

	hours := make([]HourlyEntry, 0, n)
	h := resp.Hourly
	for i := startIdx; i < len(h.Time) && len(hours) < n; i++ {
		hours = append(hours, HourlyEntry{
			Time:          h.Time[i],
			Temperature:   safeF(h.Temperature, i),
			ApparentTemp:  safeF(h.ApparentTemp, i),
			Humidity:      safeI(h.Humidity, i),
			PrecipProb:    safeI(h.PrecipProb, i),
			Precipitation: safeF(h.Precip, i),
			WeatherCode:   safeI(h.WeatherCode, i),
			WeatherDesc:   wmoDesc(safeI(h.WeatherCode, i)),
			WindSpeed:     safeF(h.Windspeed, i),
		})
	}

	return &HourlyForecast{
		City:    geo.Name,
		Country: strings.ToUpper(geo.Address.CountryCode),
		Hours:   hours,
	}, nil
}

func fetchDaily(city string, n int) (*DailyForecast, error) {
	geo, err := geocode(city)
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf(
		"%s?latitude=%s&longitude=%s&daily=temperature_2m_max,temperature_2m_min,precipitation_sum,precipitation_probability_max,weather_code,windspeed_10m_max,uv_index_max&timezone=%s&forecast_days=%d",
		openMeteoURL, geo.Lat, geo.Lon, url.QueryEscape(tz), n,
	)
	data, err := getJSON(u)
	if err != nil {
		return nil, err
	}
	if err := checkAPIErr(data); err != nil {
		return nil, err
	}
	var resp forecastDailyResp
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	d := resp.Daily
	days := make([]DailyEntry, len(d.Time))
	for i := range d.Time {
		days[i] = DailyEntry{
			Date:          d.Time[i],
			TempMax:       safeF(d.TempMax, i),
			TempMin:       safeF(d.TempMin, i),
			PrecipSum:     safeF(d.PrecipSum, i),
			PrecipProbMax: safeI(d.PrecipProbMax, i),
			WeatherCode:   safeI(d.WeatherCode, i),
			WeatherDesc:   wmoDesc(safeI(d.WeatherCode, i)),
			WindSpeedMax:  safeF(d.WindspeedMax, i),
			UVIndexMax:    safeF(d.UVIndexMax, i),
		}
	}

	return &DailyForecast{
		City:    geo.Name,
		Country: strings.ToUpper(geo.Address.CountryCode),
		Days:    days,
	}, nil
}
