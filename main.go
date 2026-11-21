package main

import (
	"bufio"
	"fmt"
	"os"
	_ "proxy_checker/pkg/proxy_checker"
)

func main() {
	//pm := proxy_checker.NewManager()

	for readln := bufio.NewScanner(os.Stdin); readln.Text() != "0"; {
		// Source - https://stackoverflow.com/a/22892171
		// Posted by Kavu, modified by community. See post 'Timeline' for change history
		// Retrieved 2026-11-21, License - CC BY-SA 4.0
		fmt.Print("\033[H\033[2J")

		fmt.Println("*******************************************************************")
		fmt.Println("************************** PROXY_CHECKER **************************")
		fmt.Println("*******************************************************************")
		fmt.Println("Что хотите сделать?")
		fmt.Println("0. Выход")
		fmt.Println("1. Проверить raw.txt")
		fmt.Println("2. Скачать все прокси из файла источников (sourcess.txt -> raw.txt)")
		fmt.Println("3. Скачать все прокси из файла источников и автоматически проверить")
		fmt.Println("*******************************************************************")
		fmt.Print(">>")
		readln.Scan()

		switch readln.Text() {
		case "0":
			break
		case "1":
			fmt.Println("Проверяем raw.txt")
		case "2":
			fmt.Println("Качаем данные из sourcess.txt")
		case "3":
			fmt.Println("Делаем всё и сразу")
		default:
			fmt.Println("Неправильный ввод")
		}
		fmt.Print(">>")
		readln.Scan()
	}

	fmt.Println("Пока")
}
