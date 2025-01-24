package wintty

import (
	"net/netip"
	"os"
	"testing"
)

func TestParseComDirection(t *testing.T) {
	tests := []struct {
		input    []byte
		expected ComDirection
		hasError bool
	}{
		{
			input: []byte("wintty=com1; Serial Port 1"),
			expected: ComDirection{
				Com:     "com1",
				Comment: "Serial Port 1",
			},
			hasError: false,
		},
		{
			input: []byte("wintty=com12; Another Serial Port"),
			expected: ComDirection{
				Com:     "com12",
				Comment: "Another Serial Port",
			},
			hasError: false,
		},
		{
			input: []byte("wintty=com3"),
			expected: ComDirection{
				Com:     "com3",
				Comment: "",
			},
			hasError: false,
		},
		{
			input:    []byte("INVALID INPUT"),
			expected: ComDirection{},
			hasError: true,
		},
	}

	for _, test := range tests {
		var comDir ComDirection
		err := ParseComDirection(test.input, &comDir)
		if test.hasError {
			if err == nil {
				t.Errorf("expected error for input %s, but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for input %s: %v", test.input, err)
			}
			if comDir != test.expected {
				t.Errorf("expected %+v, but got %+v", test.expected, comDir)
			}
		}
	}
}
func TestParseChannelParam(t *testing.T) {
	tests := []struct {
		input    []byte
		expected ChannelParam
		hasError bool
	}{
		{
			input: []byte("channel_param=04,modem_cnf=\"gsm.cnf\",ttylog"),
			expected: ChannelParam{
				Channel:  4,
				Settings: "modem_cnf=\"gsm.cnf\",ttylog",
			},
			hasError: false,
		},
		{
			input: []byte("channel_param=1,setting1,setting2; comment"),
			expected: ChannelParam{
				Channel:  1,
				Settings: "setting1,setting2",
				Comment:  "comment",
			},
			hasError: false,
		},
		{
			input: []byte("channel_param=255,full_settings"),
			expected: ChannelParam{
				Channel:  255,
				Settings: "full_settings",
			},
			hasError: false,
		},
		{
			input:    []byte("channel_param=invalid,settings"),
			expected: ChannelParam{},
			hasError: true,
		}, {
			input:    []byte("channel _param=invalid,settings"),
			expected: ChannelParam{},
			hasError: true,
		},
		{
			input:    []byte("INVALID INPUT"),
			expected: ChannelParam{},
			hasError: true,
		},
	}

	for _, test := range tests {
		var channelParam ChannelParam
		err := ParseChannelParam(test.input, &channelParam)
		if test.hasError {
			if err == nil {
				t.Errorf("expected error for input %s, but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for input %s: %v", test.input, err)
			}
			if channelParam != test.expected {
				t.Errorf("expected %+v, but got %+v", test.expected, channelParam)
			}
		}
	}
}
func TestParseIpDirection(t *testing.T) {
	tests := []struct {
		input    []byte
		expected IpDirection
		hasError bool
	}{
		{
			input: []byte("wintty=net:tcp,m:4001,192.168.0.2,4001,192.168.0.1; Moxa Port 1 mbm"),
			expected: IpDirection{
				Network: "tcp",
				Role:    "m",
				Self:    netip.MustParseAddrPort("192.168.0.2:4001"),
				Remote:  netip.MustParseAddrPort("192.168.0.1:4001"),
				Comment: "Moxa Port 1 mbm",
			},
			hasError: false,
		},
		{
			input: []byte("wintty=net:udp,1234,10.0.0.1,5678,10.0.0.2"),
			expected: IpDirection{
				Network: "udp",
				Role:    "",
				Self:    netip.MustParseAddrPort("10.0.0.1:1234"),
				Remote:  netip.MustParseAddrPort("10.0.0.2:5678"),
				Comment: "",
			},
			hasError: false,
		},
		{
			input:    []byte("INVALID INPUT"),
			expected: IpDirection{},
			hasError: true,
		},
	}

	for _, test := range tests {
		var ipDir IpDirection
		err := ParseIpDirection(test.input, &ipDir)
		if test.hasError {
			if err == nil {
				t.Errorf("expected error for input %s, but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for input %s: %v", test.input, err)
			}
			if ipDir != test.expected {
				t.Errorf("expected %+v, but got %+v", test.expected, ipDir)
			}
		}
	}
}

func TestWintty_Read(t *testing.T) {
	// Create a temporary file for the test
	tempFile, err := os.CreateTemp("", "config_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write test data to the file
	testData := `
	WINTTY=net:tcp,m:4001,192.168.0.2,4001,192.168.0.1; Moxa Port 1 mbm
	WINTTY= net:tcp,s:4002,192.168.0.2,4001,192.168.0.2; Moxa Port 2 mbm
	WINTTY=net:udp, 4003,192.168.0.2,4001,192.168.0.3; Moxa Port 3 mbm
	WINTTY=COM1; Port 1 serial
	channel_param=03, modem_cnf="gsm.cnf"
	`
	if _, err := tempFile.WriteString(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tempFile.Close()

	// Create a Wintty instance and call the Read method
	wt := Wintty{}
	err = wt.Read(tempFile.Name())
	if err != nil {
		t.Fatalf("Read method returned an error: %v", err)
	}

	// Verify the results
	if len(wt.TTY) != 4 {
		t.Errorf("Expected 4 TTY entries, got %d", len(wt.TTY))
	}

	if len(wt.Params) != 1 {
		t.Errorf("Expected 1 Param entry, got %d", len(wt.Params))
	}

	// Optionally, verify specific entries in wt.TTY and wt.Params
	if wt.Params[3] != `modem_cnf="gsm.cnf"` {
		t.Errorf(`Expected Params[3] to be 'modem_cnf="gsm.cnf"', got '%s'`, wt.Params[3])
	}
}
