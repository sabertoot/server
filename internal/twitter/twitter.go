package twitter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sabertoot/server/internal/plog"
)

const (
	v2BaseURL = "https://api.twitter.com/2"
)

func GetTweetsByUser(
	ctx context.Context,
	userHandle string,
	bearerToken string,
	startDate time.Time,
	sinceId string,
	nextToken string,
) (
	map[string]any,
	error,
) {

	// Build the request URL.
	query := fmt.Sprintf("from:%s -is:retweet -is:reply -is:quote", userHandle)
	requestURL :=
		fmt.Sprintf(
			"%s/tweets/search/recent?query=%s&max_results=100&sort_order=recency&tweet.fields=created_at,entities&expansions=author_id,attachments.media_keys&user.fields=profile_image_url&media.fields=type,preview_image_url,url,width,height",
			v2BaseURL,
			url.QueryEscape(query))

	// If we have a sinceId, use that. Otherwise, use the start date.
	// The Twitter API doesn't allow both parameters at the same time.
	if sinceId != "" {
		requestURL = fmt.Sprintf("%s&since_id=%s", requestURL, sinceId)
	} else {
		// Recent search in Twitter API v2 only allows to go back as far as 7 days.
		earliestDate := time.Now().UTC().AddDate(0, 0, -7).Add(time.Hour)
		if startDate.Before(earliestDate) {
			startDate = earliestDate
		}
		requestURL = fmt.Sprintf("%s&start_time=%s", requestURL, startDate.Format(time.RFC3339))
	}

	if nextToken != "" {
		requestURL = fmt.Sprintf("%s&next_token=%s", requestURL, nextToken)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
	}
	req.Header.Set("User-Agent", "Sabertoot/1.0")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearerToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing HTTP request: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad status code from Twitter API: %d", resp.StatusCode)
	}

	// ToDo: Check for rate limit exceeded.
	rateLimitValue := resp.Header.Get("x-rate-limit-remaining")
	rateLimit, err := strconv.Atoi(rateLimitValue)
	plog.Debugf("Twitter API rate limit remaining: %d", rateLimit)
	if err != nil {
		return nil, fmt.Errorf("bad rate limit value '%s': %w", rateLimitValue, err)
	}

	defer resp.Body.Close()
	result := map[string]any{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("error deserializing HTTP response: %w", err)
	}

	return result, nil
}
