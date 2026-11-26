package main

import (
	"bufio"
	"fmt"
	"os"
	"proxy_checker/pkg/logger"
	"proxy_checker/pkg/proxy_checker"
	"strings"
)

const SettingsFile = "settings.yml"

var settings = map[string]string{}

func loadSettings() {
	f, err := os.OpenFile(SettingsFile, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		keyVal := strings.Split(line, ": ")
		settings[keyVal[0]] = keyVal[1]
	}
}

func main() {
	__ := transFn()
	settingsPM := proxy_checker.ProxyManagerSettings{TransFn: __, Logger: logger.NewConsoleLogger()}
	pm := proxy_checker.NewProxyManager(settingsPM)
	loadSettings()

	fmt.Println("*******************************************************************")
	fmt.Println("************************** PROXY_CHECKER **************************")
	fmt.Println("*******************************************************************")

	//fmt.Println("Lang ru/en?")

	for readln := bufio.NewScanner(os.Stdin); readln.Text() != "0"; {
		// Source - https://stackoverflow.com/a/22892171
		// Posted by Kavu, modified by community. See post 'Timeline' for change history
		// Retrieved 2026-11-21, License - CC BY-SA 4.0
		fmt.Print("\033[H\033[2J")

		fmt.Println(__("Что хотите сделать?"))
		fmt.Println(__("0. Выход"))
		fmt.Println(__("1. Скачать все прокси из файла источников (sourcess.txt -> raw.txt)"))
		fmt.Println(__("2. Проверить raw.txt (все прокси должны быть руками записаны)"))
		fmt.Println(__("3. Скачать все прокси из файла источников и автоматически проверить"))
		//fmt.Println("*******************************************************************")
		fmt.Print(">>")
		readln.Scan()

		switch readln.Text() {
		case "0":
			break
		case "1":
			fmt.Println(__("Скачиваем прокси из всех источников в raw.txt"))
			err := pm.DownloadProxiesFromSourcesToRawFile(true)
			if err != nil {
				fmt.Println(err)
			}
		case "2":
			fmt.Println(__("Проверяем raw.txt"))
			pm.ProcessProxiesFile(proxy_checker.FilePathRaw)
		case "3":
			fmt.Println(__("Качаем данные из всех источников и проверяем их"))
			pm.DownloadFromSourcesAndProcess()
		default:
			fmt.Println(__("Неправильный ввод"))
		}
	}

	fmt.Println("Пока")
}
