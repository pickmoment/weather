package main

import (
	"flag"
	"fmt"
	"os"
)

func usage() {
	fmt.Fprintln(os.Stderr, `날씨 CLI

사용법: weather <명령어> [옵션]

명령어:
  now      현재 날씨 및 공기질 조회
  hourly   시간별 예보 조회
  daily    일별 예보 조회
  install  AI 에이전트용 스킬 파일 설치

각 명령어에 -h 플래그로 도움말을 볼 수 있습니다.`)
}

func errExit(msg string) {
	fmt.Fprintf(os.Stderr, "오류: %s\n", msg)
	os.Exit(1)
}

// sortArgs moves flags (and their values) before positional args so that
// `weather hourly 서울 -n 6` works the same as `weather hourly -n 6 서울`.
func sortArgs(args []string) []string {
	var flags, positional []string
	for i := 0; i < len(args); i++ {
		if len(args[i]) > 1 && args[i][0] == '-' {
			flags = append(flags, args[i])
			if i+1 < len(args) && (len(args[i+1]) == 0 || args[i+1][0] != '-') {
				flags = append(flags, args[i+1])
				i++
			}
		} else {
			positional = append(positional, args[i])
		}
	}
	return append(flags, positional...)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "now":
		runNow(os.Args[2:])
	case "hourly":
		runHourly(os.Args[2:])
	case "daily":
		runDaily(os.Args[2:])
	case "install":
		runInstall(os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "알 수 없는 명령어: %s\n\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func runNow(args []string) {
	fs := flag.NewFlagSet("now", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "사용법: weather now <도시> [-f json|telegram]\n")
		fs.PrintDefaults()
	}
	format := fs.String("f", "telegram", "출력 형식 (json|telegram)")
	_ = fs.Parse(sortArgs(args))

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "오류: 도시명이 필요합니다")
		fs.Usage()
		os.Exit(1)
	}
	data, err := fetchNow(fs.Arg(0))
	if err != nil {
		errExit(err.Error())
	}
	fmt.Println(fmtNow(data, *format))
}

func runHourly(args []string) {
	fs := flag.NewFlagSet("hourly", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "사용법: weather hourly <도시> [-n 시간수] [-f json|telegram]\n")
		fs.PrintDefaults()
	}
	n := fs.Int("n", 24, "조회할 시간 수")
	format := fs.String("f", "telegram", "출력 형식 (json|telegram)")
	_ = fs.Parse(sortArgs(args))

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "오류: 도시명이 필요합니다")
		fs.Usage()
		os.Exit(1)
	}
	data, err := fetchHourly(fs.Arg(0), *n)
	if err != nil {
		errExit(err.Error())
	}
	fmt.Println(fmtHourly(data, *format))
}

func runDaily(args []string) {
	fs := flag.NewFlagSet("daily", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "사용법: weather daily <도시> [-n 일수] [-f json|telegram]\n")
		fs.PrintDefaults()
	}
	n := fs.Int("n", 7, "조회할 일 수")
	format := fs.String("f", "telegram", "출력 형식 (json|telegram)")
	_ = fs.Parse(sortArgs(args))

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "오류: 도시명이 필요합니다")
		fs.Usage()
		os.Exit(1)
	}
	data, err := fetchDaily(fs.Arg(0), *n)
	if err != nil {
		errExit(err.Error())
	}
	fmt.Println(fmtDaily(data, *format))
}
