package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Record struct {
	userAgents []string
	allow      []string
	disallow   []string
	crawlDelay time.Duration
}

type Records struct {
	Records []Record
}

/*
	parseRobotsTxt builds the Records data structure that allows the Crawl() method to determine which files it can and cannot crawl
	@params url is the file path to the robots.txt file we want to read
	@returns the completed Records data structure
	@returns error for error checking
*/
func ParseRobotsTxt(url string) (Records, error) {
	var robots Records

	// get and read the robots.txt file
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return robots, fmt.Errorf("ERROR failed to fetch robots.txt file: %v\n", err)
	}
	defer resp.Body.Close()

	// read the robots.txt file line by line and build 'robots'
	reader := bufio.NewReader(resp.Body)
	var currentRecord Record
	makeNewRecord := false // parsingDirectives tracks whether a new Record object should be made. Whenever the parser encounters a "User-agent" flag again after encountering other flags, the parser will know to make a new Record struct.

	// regexp to match robots.txt properties. Using regexp.MustCompile skips the error checking code
	userAgent := regexp.MustCompile(`(?i)^\s*(User-agent)\s*:\s*(.+)$`)    // (?i) makes userAgent case insensitive, ^\s* matches the first non space char at the from of the line, User-agent\s*: matches the "User-agent" text followed by white space and followed by the ':', and \s*(.+)$ extracts any non space text after the ':' and stores it as userAgent's value.
	disallow := regexp.MustCompile(`(?i)^\s*(Disallow)\s*:\s*(.*)$`)      // (.*) extracts ZERO OR MORE characters.
	allow := regexp.MustCompile(`(?i)^\s*(Allow)\s*:\s*(.*)$`)
	crawlDelay := regexp.MustCompile(`(?i)^\s*(Crawl-delay)\s*:\s*(.+)$`)
	sitemap := regexp.MustCompile(`(?i)^\s*Sitemap\s*:`) // Saving Sitemap for future use

	for {
		// Read the file one line at a time, and read the entire line at one time.
		currentLine, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return robots, fmt.Errorf("ERROR reading line in robots.txt: %v\n", err)
		}

		currentLine = strings.TrimSpace(currentLine)

		// if the line is empty or is a comment (starts with '#'), we skip the line
		if currentLine == "" || currentLine[0] == '#' || sitemap.MatchString(currentLine) {
			if err == io.EOF {
				break
			}
			continue
		}

		// matching regular expressions
		// matches' will be the output of the regular expression string matching process
		// e.g if the line looks like: "User-agent: GoogleBot", then matches = ["User-agent: GoogleBot", "User-agent", "GoogleBot"]
		if matches := userAgent.FindStringSubmatch(currentLine); matches != nil {
			if makeNewRecord || len(currentRecord.userAgents) == 0 {
				if len(currentRecord.userAgents) > 0 {
					robots.Records = append(robots.Records, currentRecord)
					currentRecord = Record{}
					currentRecord.crawlDelay = 100 * time.Millisecond
				}
				currentRecord.userAgents = append(currentRecord.userAgents, matches[2])
				makeNewRecord = false
			} else {
				currentRecord.userAgents = append(currentRecord.userAgents, matches[2])
			}
		} else {
			if len(currentRecord.userAgents) > 0 {
				makeNewRecord = true // Next User-agent will start a new Record

				// match "Disallow" rules
				if matches := disallow.FindStringSubmatch(currentLine); matches != nil {
					currentRecord.disallow = append(currentRecord.disallow, matches[2])
					continue
				}
				// match "Allow" rules
				if matches := allow.FindStringSubmatch(currentLine); matches != nil {
					currentRecord.allow = append(currentRecord.allow, matches[2])
					continue
				}
				// match Crawl-delay
				if matches := crawlDelay.FindStringSubmatch(currentLine); matches != nil {
					delayTime, err := strconv.Atoi(matches[2])
					if err != nil {
						fmt.Printf("ERROR invalid Crawl-delay value: %s\n", matches[2])
					} else {
						currentRecord.crawlDelay = time.Duration(delayTime) * time.Second
					}
				}
			}
		}

		if err == io.EOF {
			if len(currentRecord.userAgents) > 0 {
				robots.Records = append(robots.Records, currentRecord)
			}
			break
		}
	}

	return robots, nil
}
