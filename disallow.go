package main

import (
	"net/url"
	"regexp"
	"strings"
	"sync"
)

/*
	return true if the user agent is allowed to crawl 'url', false otherwise.
	@params userRestrictions is the "Records" data structure that keeps track of which user agents are allowed/disallowed on which URLs
	@params userAgentName is the name of the user's crawler. This will be used to match the rules inside of userRestrictions
	@params url is the URL that the crawler is currently trying to crawl. Disallow() is trying to determine whether the user can access this URL.
*/
func Disallow(applicableRecord *Record, link string, ch chan<- string, wg *sync.WaitGroup){
	defer wg.Done()

	parsedLink, err := url.Parse(link)
	if err != nil{
		ch <- link
		return
	}
	// Find the most specific applicable Record struct for the user agent
	
	if applicableRecord == nil {
		return
	}

	// Determine if the URL is allowed or disallowed based on the rules
	allowed := isAllowedByRules(applicableRecord, parsedLink.Path)

	if !allowed{
		ch <- link
	}
	
	return
}

func findApplicableRecord(userRestrictions Records, userAgentName string) *Record {
	// make a wildcardRecord ("User-agent: *") in case no exact matches are found.
	var wildcardRecord *Record

	uaNameLower := strings.ToLower(userAgentName)
	for _, record := range userRestrictions.Records {
		for _, ua := range record.userAgents {
			uaLower := strings.ToLower(strings.TrimSpace(ua))
			if uaLower == "*" {
				wildcardRecord = &record
			}else if uaLower == uaNameLower {
				return &record
			}
		}
	}

	// if no exact match, use the wildcard record
	if wildcardRecord != nil {
		return wildcardRecord
	}

	// if even the wild card record is nil, then there are no rules that apply to the user's crawler, and Disallow() can assuem the user is allowed to crawl
	return nil
}

func isAllowedByRules(applicableRecord *Record, url string)bool{

	for _, pattern := range applicableRecord.allow {
		if pattern == "" {
			continue
		}
		escapedPattern := regexp.QuoteMeta(pattern)
		escapedPattern = strings.ReplaceAll(escapedPattern, `\*`, `.*`)
		if strings.HasSuffix(pattern, "$") {
			escapedPattern = "^" + escapedPattern
		} else {
			escapedPattern = ".*" + escapedPattern
		}
		matched, err := regexp.MatchString(escapedPattern, url)
		if err == nil && matched {
			return true;
		}
	}

	for _, pattern := range applicableRecord.disallow {
		if pattern == "" {
			continue
		}
		escapedPattern := regexp.QuoteMeta(pattern)
		escapedPattern = strings.ReplaceAll(escapedPattern, `\*`, `.*`)
		if strings.HasSuffix(pattern, "$") {
			escapedPattern = "^" + escapedPattern
		} else {
			escapedPattern = ".*" + escapedPattern
		}
		matched, err := regexp.MatchString(escapedPattern, url)
		if err == nil && matched {
			return false
		}
	}

	return true;
}