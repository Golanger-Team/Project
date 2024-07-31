package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const REPORTER_PORT = 9999

type AdvertiserPublisherEventCount struct {
	advertiser_id	string
	publisher_id	string
	event_type		string
	total			int
}

type AdPublisherEventCount struct {
	ad_id			string
	publisher_id	string
	event_type		string
	total			int
}

// A struct reflecting the collaboration of an ad with a publisher.
type AdPublisherCollaboration struct {
	AdID        int
	PublisherID int
}

type AdvertiserPublisherCollaboration struct {
	AdvertiserID int
	PublisherID  int
}


// Stores the statistics of a collaboration. Namely, impression count, click count and ctr.
type Statistics struct {
	Impressions int
	Clicks      int
	CTR         float64
}

// Maps advertiser-publisher collaborations to their emprical success statistics.
var advertiserEvaluation map[AdvertiserPublisherCollaboration]Statistics

// Map 
var adEvaluation map[AdPublisherCollaboration]Statistics

/* Sends the mean ctr of each advertiser's ads, per publisher. */
func sendAdvertisersMeanCTR(c *gin.Context) {
	var timeCondition = "time > now() - INTERVAL '1 hour'"
	var eventCounts []AdvertiserPublisherEventCount
	db.Table("events").Select("advertiser_id, publisher_id, event_type, count(1) AS total").Where(timeCondition).Group("advertiser_id, publisher_id, event_type").Scan(&eventCounts)

	var collaboration AdvertiserPublisherCollaboration
	var statistics Statistics
	for _, eventCount := range eventCounts {
		collaboration.AdvertiserID, _ = strconv.Atoi(eventCount.advertiser_id)
		collaboration.PublisherID,  _ = strconv.Atoi(eventCount.publisher_id)
		
		statistics = advertiserEvaluation[collaboration]
		if eventCount.event_type == "impression" {
			statistics.Impressions = eventCount.total
		} else {
			statistics.Clicks = eventCount.total
		}
		advertiserEvaluation[collaboration] = statistics
	}

	/* Fix possible inconsistencies in data. These inconsistencies
	 can happen, for example by latency in arrival of click and impression
	 events. */
	for apc := range advertiserEvaluation {
		var statistics = advertiserEvaluation[apc]
		if statistics.Impressions < statistics.Clicks {
			statistics.Impressions = statistics.Clicks
		}
		if statistics.Impressions > 0 {
			statistics.CTR = float64(statistics.Clicks) / float64(statistics.Impressions)
		}
	}
	
	/* Our statistics map is ready to be sent. */
	c.JSON(http.StatusOK, advertiserEvaluation)
}

/* Sends the per-publisher success statistics of each Ad. */
func sendAdStatistics(c *gin.Context) {
	var timeCondition = "time > now() - INTERVAL '1 hour'"
	
	var eventCounts []AdPublisherEventCount
	db.Table("events").Select("ad_id, publisher_id, event_type, count(1) AS total").Where(timeCondition).Group("ad_id, publisher_id, event_type").Scan(&eventCounts)

	var collaboration AdPublisherCollaboration
	for _, eventCount := range eventCounts {
		collaboration.AdID, _ = strconv.Atoi(eventCount.ad_id)
		collaboration.PublisherID, _ = strconv.Atoi(eventCount.publisher_id)

		var statistics = adEvaluation[collaboration]
		switch eventCount.event_type {
		case "impression":
			statistics.Impressions = eventCount.total
		case "click":
			statistics.Clicks = eventCount.total
		default:
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		adEvaluation[collaboration] = statistics
	}
	
	/* Fix possible inconsistencies in data. These inconsistencies
	 can happen, for example by latency in arrival of click and impression
	 events. */
	 for apc := range adEvaluation {
		var statistics = adEvaluation[apc]
		if statistics.Impressions < statistics.Clicks {
			statistics.Impressions = statistics.Clicks
		}
		if statistics.Impressions > 0 {
			statistics.CTR = float64(statistics.Clicks) / float64(statistics.Impressions)
		}
	}
}

/* Runs the router that will route api calls from ad server to
 handlers. Note that this function block the calling goroutine
 indefinitely. */
func setupAndRunAPIRouter() {
	router := gin.Default()
	
	router.Run(":" + strconv.Itoa(REPORTER_PORT))
}