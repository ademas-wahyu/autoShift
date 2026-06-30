package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ademaswahyu/autoshift-backend/models"
)

type HolidayFetcher struct {
	apiURL  string
	client  *http.Client
}

type NagerHoliday struct {
	Date        string `json:"date"`
	LocalName   string `json:"localName"`
	Name        string `json:"name"`
	CountryCode string `json:"countryCode"`
	Global      bool   `json:"global"`
}

func NewHolidayFetcher(apiURL string) *HolidayFetcher {
	return &HolidayFetcher{
		apiURL: apiURL,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (h *HolidayFetcher) FetchAndStore(year int, countryCode string) error {
	url := fmt.Sprintf("%s/PublicHolidays/%d/%s", h.apiURL, year, countryCode)
	log.Printf("fetching holidays from: %s", url)

	resp, err := h.client.Get(url)
	if err != nil {
		return fmt.Errorf("holiday api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("holiday api returned status %d", resp.StatusCode)
	}

	var nagerHolidays []NagerHoliday
	if err := json.NewDecoder(resp.Body).Decode(&nagerHolidays); err != nil {
		return fmt.Errorf("failed to decode holidays: %w", err)
	}

	// Store in DB (upsert)
	for _, nh := range nagerHolidays {
		holiday := models.Holiday{
			Date:       nh.Date,
			Name:       nh.LocalName,
			IsNational: nh.Global,
		}
		// Upsert by date
		var existing models.Holiday
		result := models.DB.Where("date = ?", nh.Date).First(&existing)
		if result.Error != nil {
			models.DB.Create(&holiday)
		} else {
			models.DB.Model(&existing).Updates(map[string]interface{}{
				"name":        nh.LocalName,
				"is_national": nh.Global,
			})
		}
	}

	log.Printf("stored %d holidays for %s %d", len(nagerHolidays), countryCode, year)
	return nil
}

func (h *HolidayFetcher) GetHolidays(year int) []models.Holiday {
	var holidays []models.Holiday
	models.DB.Where("date >= ? AND date < ?",
		fmt.Sprintf("%d-01-01", year),
		fmt.Sprintf("%d-01-01", year+1),
	).Find(&holidays)
	return holidays
}
