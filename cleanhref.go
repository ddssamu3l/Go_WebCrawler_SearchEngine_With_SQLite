package main

import (
	"net/url"
	//"path/filepath"
	"strings"
	"unicode"
)

func IsValidHTTPSURL(hostname string) bool {
	// parse the URL, and check for errors with parsing
	u, err := url.Parse(hostname)
	if err != nil {
		return false
	}
	// if the hostname does not start with "https" or has no host, it's not a valid URL and return false
	if u.Scheme != "https" || u.Host == "" {
		return false
	}
	// check for a valid Top-Level-Domain
	// if there isn't a .something behind the URL, it's not valid
	parts := strings.Split(u.Host, ".")
	if len(parts) < 2 {
		return false
	}

	return true
}

// containsInvalidURLChars checks if a URL contains any invalid characters
func containsInvalidURLChars(url string) bool {
    for _, currentChar := range url {
        if !unicode.IsPrint(currentChar) {
            return true
        }
    }
    return false
}

/*
	isSameHostname will parse its two input URL strings and check if they have the same hostname.
*/
func isSameHostname (url1, url2 string) (bool) {
	parsedURL1, err := url.Parse(url1)
	if err != nil {
		return false
	}

	parsedURL2, err := url.Parse(url2)
	if err != nil {
		return false
	}

	// if the second url is just a partial url, then just assume the two urls have the same hostnames
	if(!parsedURL2.IsAbs()){
		return true
	}

	// Compare the Host fields (ignores schemes and paths)
	return parsedURL1.Hostname() == parsedURL2.Hostname()
}

/*
	Clean(): Processes a list of hrefs (URLs) based on the given host.
	@params: host is the base URL
	@params: hrefs is the list of relative or absolute URLs
	@returns: []string is the list of cleaned, absolute URLs             
*/
func Clean(host string, hrefs []string) []string {
	//TODO: Could possibly add in validation for parsing hostnames and check if they are the same
	var parsedUrls []string

	// Parse the base URL and check if the host is valid
	base, err := url.Parse(host)
	if err != nil {
		return []string{"INVALID HOSTNAME"}
	}

	for _, href := range hrefs {
		if len(href) == 0{
			continue
		}
		if href == "INVALID HREF" || href[0] == '#'{
			parsedUrls = append(parsedUrls, "INVALID HREF")
			continue
		}
		// check if the URL contains invalid characters
		if containsInvalidURLChars(href){
			parsedUrls = append(parsedUrls, "INVALID HREF")
			continue
		}

		baseHostOnly := &url.URL{
			Scheme: base.Scheme,
			Host:   base.Host,
		}

		// Parse href
		hrefURL, err := url.Parse(href)
		if err != nil {
			parsedUrls = append(parsedUrls, "INVALID HREF")
			continue
		}
		// If the href is already a complete URL, it's valid, so add it as is
		// also check if the href has the same hostname as the host
		if IsValidHTTPSURL(href){
			if isSameHostname(host, href){
				parsedUrls = append(parsedUrls, href)
			}else{
				parsedUrls = append(parsedUrls, "INVALID HREF")
			}
			continue
		}

		// if the href is a partial url, match it with the host
		mergedURL := baseHostOnly.ResolveReference(hrefURL)
		// add the resolved, cleaned URL to the list
		parsedUrls = append(parsedUrls, mergedURL.String())
	}

	return parsedUrls
}
