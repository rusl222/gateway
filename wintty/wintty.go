package wintty

import (
	"bufio"
	"errors"
	"net/netip"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	// ErrInvalidFormat is returned when the input does not match the expected format.
	ErrInvalidFormat = "input does not match the expected format"
)

type Wintty struct {
	TTY    []interface{}
	Params map[uint8]string
}

type ComDirection struct {
	Com     string
	Comment string
}

type IpDirection struct {
	Network string
	Role    string
	Self    netip.AddrPort
	Remote  netip.AddrPort
	Comment string
}

// Read - reading a configuration text file with many
// strings like " WINTTY=net:tcp,m:4001,192.168.0.2,4001,192.168.0.1; Moxa Port 1 mbm"
func (wt *Wintty) Read(confpath string) error {
	file, err := os.Open(confpath)
	if err != nil {
		return err
	}
	defer file.Close()

	wt.Params = make(map[uint8]string)

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		bytes1 := CleanForRegex(scanner.Text())

		dir := IpDirection{}
		if err := ParseIpDirection(bytes1, &dir); err == nil {
			wt.TTY = append(wt.TTY, dir)
		}

		com := ComDirection{}
		if err := ParseComDirection(bytes1, &com); err == nil {
			wt.TTY = append(wt.TTY, com)
		}

		param := ChannelParam{}
		if err := ParseChannelParam(bytes1, &param); err == nil {
			wt.Params[param.Channel] = param.Settings
		}
	}
	return nil
}

func CleanForRegex(input string) []byte {
	substr, comment, _ := strings.Cut(input, ";")
	// Удаляем все пробелы и табуляции
	cleaned := strings.ReplaceAll(substr, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "\t", "")
	// Приводим все буквы к нижнему регистру
	return []byte(strings.ToLower(cleaned) + ";" + comment)
}

// Parse parses the input byte slice and fills the IpDirection structure
// 'WINTTY=net:tcp,m:4001,192.168.0.2,4001,192.168.0.1; Moxa Port 1 mbm'.
func ParseIpDirection(input []byte, ipDir *IpDirection) error {
	var err error

	// Define the regular expression pattern
	pattern := `wintty=net:(tcp|udp),(m|s|p)?:?(\d+),([\d\.]+),(\d+),([\d\.]+);?\s*(.*)?`
	re := regexp.MustCompile(pattern)

	// Find matches
	matches := re.FindSubmatch(input)
	if matches == nil {
		return errors.New(ErrInvalidFormat)
	}

	// Fill the IpDirection structure
	ipDir.Network = string(matches[1])
	ipDir.Role = string(matches[2])

	ipDir.Self, err = netip.ParseAddrPort(string(matches[4]) + ":" + string(matches[3]))
	if err != nil {
		return err
	}

	ipDir.Remote, err = netip.ParseAddrPort(string(matches[6]) + ":" + string(matches[5]))
	if err != nil {
		return err
	}

	ipDir.Comment = string(matches[7])

	return nil
}

func ParseComDirection(input []byte, ComDir *ComDirection) error {
	// Define the regular expression pattern
	pattern := `wintty=(com)(\d+);?\s*(.*)?`
	re := regexp.MustCompile(pattern)

	// Find matches
	matches := re.FindSubmatch(input)
	if matches == nil {
		return errors.New(ErrInvalidFormat)
	}

	// Fill the Direction structure
	ComDir.Com = string(matches[1]) + string(matches[2])
	ComDir.Comment = string(matches[3])
	return nil
}

type ChannelParam struct {
	Channel  uint8
	Settings string
	Comment  string
}

// channel_param=04,modem_cnf="gsm.cnf"  ,ttylog
func ParseChannelParam(input []byte, channelParam *ChannelParam) error {
	// Define the regular expression pattern
	pattern := `channel_param=(\d+),([\w=,".]*);?\s*(.*)?`
	re := regexp.MustCompile(pattern)

	// Find matches
	matches := re.FindSubmatch(input)
	if matches == nil {
		return errors.New(ErrInvalidFormat)
	}

	// Fill the Direction structure
	channel, err := strconv.Atoi(string(matches[1]))
	if err != nil {
		return err
	}
	channelParam.Channel = uint8(channel)
	channelParam.Settings = string(matches[2])
	channelParam.Comment = string(matches[3])
	return nil
}

//; line 6

func ParseComment(input []byte, comment *string) error {
	// Define the regular expression pattern
	pattern := `\s*;(.*)`
	re := regexp.MustCompile(pattern)

	// Find matches
	matches := re.FindSubmatch(input)
	if matches == nil {
		return errors.New(ErrInvalidFormat)
	}

	// Fill the Direction structure
	*comment = string(matches[1])
	return nil
}
