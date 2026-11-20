package proxy_checker

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const FromStringReg = "^(.{4,6})\\:\\/\\/?(.{1,}):(.{1,5})$"

// IpToIntFunc экспорт для теста. https://stackoverflow.com/a/60813569
var IpToIntFunc = ipToInt

type Data struct {
	IP       string
	Protocol string
	Port     int
	Username string
	Password string
}

type ValidationError struct {
	What string
	When time.Time
}

func (t *Data) String() string {
	var authData string

	if t.Username == "" {
		authData = ""
	} else {
		if t.Password != "" {
			authData = fmt.Sprintf("%s:%s@", t.Username, t.Password)
		} else {
			authData = fmt.Sprintf("%s@", t.Username)
		}
	}
	// костыльно как-то, но хз как лучше сделать

	ipWithCredentials := fmt.Sprintf("%s%s", authData, t.IP)

	return fmt.Sprintf("%s://%s:%d", t.Protocol, ipWithCredentials, t.Port)
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("Ошибка валидации: %s:%s", e.When, e.What)
}

func (t *Data) ValidateData() error {
	if t.IP == "" {
		return &ValidationError{"IP адрес не может быть пустым", time.Now()}
	}

	if t.Protocol == "" {
		return &ValidationError{"Протокол не может быть пустым", time.Now()}
	}

	if t.Port == 0 {
		return &ValidationError{"Порт не может быть пустым", time.Now()}
	}

	isIpValid, err := ipCheck(t.IP)
	if !isIpValid {
		return err
	}

	return nil
}

var regex = regexp.MustCompile("^[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}$")

func ipCheck(ip string) (bool, error) {
	const MaxIntIp = 4294967295 // 255.255.255.255 просто потому что так хочу

	if !regex.MatchString(ip) {
		return false, &ValidationError{What: "Неверный формат IP адреса", When: time.Now()}
	}

	intIp := ipToInt(ip)

	if intIp > MaxIntIp {
		return false, &ValidationError{What: "IP адрес вылезает за все допустимые рамки", When: time.Now()}
	}

	return true, nil
}

// IpToInt По-хорошему его бы оставить приватным, но тогда его тест не увидит. А я чо зря его писал?
// Ну и по-хорошему бы вынести в какой-либо пакет типа IP, но плевать
func ipToInt(ip string) int {
	ipParts := strings.Split(ip, ".")
	firstOct, _ := strconv.ParseInt(ipParts[0], 10, 32) // плевать на ошибку ее не допустить должна регулярка
	secondOct, _ := strconv.ParseInt(ipParts[1], 10, 32)
	thirdOct, _ := strconv.ParseInt(ipParts[2], 10, 32)
	forthOct, _ := strconv.ParseInt(ipParts[3], 10, 32)

	intIp := firstOct<<24 + secondOct<<16 + thirdOct<<8 + forthOct
	return int(intIp)
}

var reg = regexp.MustCompile(FromStringReg)

// TODO: по хорошему нужно логин пароль распарсить. но пока что плевать.
// socks5://username:password@127.0.0.1:1080 - в итоге получится не подходящим
func FromString(str string) (*Data, error) {
	//reg := regexp.MustCompile(FromStringReg)

	// слайс слайсов
	substr := reg.FindAllStringSubmatch(str, 1)
	d := &Data{}
	if len(substr) == 0 {
		return d, fmt.Errorf("прокси %s имеет неверный формат", str)
	}

	proxyData := substr[0]
	port, err := strconv.Atoi(proxyData[3])
	if err != nil {
		return d, err
	}

	d.Protocol = proxyData[1]
	d.IP = proxyData[2]
	d.Port = port

	err = d.ValidateData()
	return d, err
}
