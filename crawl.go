package main

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

type StopWords map[string]struct{}

// returns the expected path to the robots.txt file
func getRobotsTxtURL(seed string) (string, error) {
	parsedSeed, err := url.Parse(seed)
	if err != nil {
		return "", err
	}

	p := parsedSeed.Path
	ext := path.Ext(p)

	if ext == "" {
		if !strings.HasSuffix(p, "/") {
			p += "/"
		}
		p += "robots.txt"
	} else {
		p = path.Join(path.Dir(p), "robots.txt")
	}

	parsedSeed.Path = p

	return parsedSeed.String(), nil
}

/*
    Removes the hostname and the prefix "/" from a string URL or a file path
*/
func removeHostname(fullURL string) (string, error) {
	// Parse the URL
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return "", err
	}

	// Get the path and trim the leading slash
	pathAndQuery := strings.TrimPrefix(parsedURL.EscapedPath(), "/")
	if parsedURL.RawQuery != "" {
		pathAndQuery += "?" + parsedURL.RawQuery
	}

	return pathAndQuery, nil
}


/*
    This function adds new, unvisited, and valid URL or file paths to the queue
*/
func addNewURLsToQueue(hrefs []string, seed string, visited map[string]struct{}, queue *[]string) {
    cleanedHrefs := Clean(seed, hrefs)
    for _, href := range cleanedHrefs {
        if href == "INVALID HREF" {
            continue
        }
		cleanedHref, _ := removeHostname(href)
        // Check if the href is already visited. If not, add to queue and mark as visited
        if _, seen := visited[cleanedHref]; !seen {
            *queue = append(*queue, href)
            visited[cleanedHref] = struct{}{} 
        }
    }
}


/*
	Crawl(): Given a seed URL, download the webpage, extract the words and URLs,
	add all cleaned URLs found to a download queue, and continue to crawl those URLs.

	@params: seed is the seed URL string that the method will crawl
	@returns: error for error handling
*/
func Crawl(idx Index, seed string) error {
	// read the robots.txt file that is stored at (seed-seed.path)+robots.txt
	robotsTxtURL, err := getRobotsTxtURL(seed)
	if err != nil{
		fmt.Errorf("ERROR retriving robots.txt: %s\n", err)
	}

	userRestrictions, _ := ParseRobotsTxt(robotsTxtURL)
	if len(userRestrictions.Records) == 0{
		
		r := Record{
			userAgents: []string{"*"},
			allow:      nil,             
			disallow:   nil,
			crawlDelay: 100 * time.Millisecond,
		}
		userRestrictions.Records = append(userRestrictions.Records, r)
	}

	stopWords, err := GenerateStopWords()
	if err != nil {
		return err
	}

	userAgentName := "ddsamu3l"
	queue := []string{seed}
	visited := make(map[string]struct{})
	
	mostApplicableRecord := findApplicableRecord(userRestrictions, userAgentName)

	for len(queue) > 0 {
		var disallowWg sync.WaitGroup
		disallowCh := make(chan string, len(queue))
		for i := 0; i<len(queue); i++{
			disallowWg.Add(1)
			go Disallow(mostApplicableRecord, queue[i], disallowCh, &disallowWg)
		}
		disallowWg.Wait()
		close(disallowCh)

        disallowedSet := make(map[string]struct{})
        for c := range disallowCh {
            fmt.Println(c, "is disallowed")
            disallowedSet[c] = struct{}{}
            visited[c] = struct{}{}
        }

        filteredQueue := make([]string, 0, len(queue))
        for _, url := range queue {
            if _, isDisallowed := disallowedSet[url]; !isDisallowed {
                filteredQueue = append(filteredQueue, url)
            }
        }
        queue = filteredQueue

		var downloadWg sync.WaitGroup
		downloadCh := make(chan DownloadContent, len(queue))
		for _, url := range queue{
			downloadWg.Add(1)
			go Download(url, downloadCh, &downloadWg)
			time.Sleep(mostApplicableRecord.crawlDelay)
		}
		downloadWg.Wait()
		close(downloadCh)

		// CODE REVIEW CHANGES #1
		var extractWg sync.WaitGroup
		extractChannel := make(chan ExtractResult)
		currentQueueLen := len(queue)
		for extracted := range downloadCh{
			currentURL := extracted.url
			if !isSameHostname(seed, currentURL){
				continue
			}
			extractWg.Add(1)
			go Extract(extracted.body, currentURL, &extractWg, extractChannel)
		}
		go func() {
			extractWg.Wait()
			close(extractChannel)
		}()
		queue = queue[currentQueueLen:]

		for result := range extractChannel {
			if result.err != nil {
				fmt.Printf("Error extracting from %s: %v\n", result.url, result.err)
				continue
			}
			addNewURLsToQueue(result.hrefs, seed, visited, &queue)
			fmt.Println(result.url)
			result.url, err = removeHostname(result.url)
			if err != nil {
				return err
			}
			idx.AddToIndex(stopWords, result.words, result.url, result.title)
		}		
	}
	return nil
}
