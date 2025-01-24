package wintty

import (
	"errors"
	"net/netip"
	"regexp"
	"strconv"
)

const (
	// ErrInvalidFormat is returned when the input does not match the expected format.
	ErrInvalidFormat = "input does not match the expected format"
)

type Wintty struct {
	Record []interface{}
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

// Parse parses the input byte slice and fills the IpDirection structure
// 'WINTTY=net:tcp,m:4001,192.168.0.2,4001,192.168.0.1; Moxa Port 1 mbm'.
func ParseIpDirection(input []byte, ipDir *IpDirection) error {
	var err error

	// Define the regular expression pattern
	pattern := `\s*WINTTY.*=.*net.*:.*(tcp|udp),\s*(m|s|p)?:?\s*(\d+),([\d\.]+),(\d+),([\d\.]+);*\s*(.*)?`
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
	pattern := `\s*WINTTY\s*=\s*(COM|com|Com)(\d+)\s*;*\s*(.*)?`
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
}

// channel_param=04,modem_cnf="gsm.cnf"  ,ttylog
func ParseChannelParam(input []byte, channelParam *ChannelParam) error {
	// Define the regular expression pattern
	pattern := `\s*channel_param\s*=\s*(\d+)\s*,(.*)`
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
