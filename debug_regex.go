package main

import (
	"fmt"
	"regexp"
)

func main() {
	line := "Jan  1 10:30:15 server1 sshd[1234]: Failed password for root from 192.168.1.100 port 22 ssh2"
	regex := regexp.MustCompile(`^(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+(\S+)\s+(\S+?)(?:\[(\d+)\])?:\s*(.*)$`)
	matches := regex.FindStringSubmatch(line)
	
	fmt.Printf("Line: %s\n", line)
	fmt.Printf("Matches: %d\n", len(matches))
	for i, match := range matches {
		fmt.Printf("  [%d]: '%s'\n", i, match)
	}
}
