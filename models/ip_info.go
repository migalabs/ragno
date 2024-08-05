package models

import "time"

const (
	IPInfoTTL = 30 * 24 * time.Hour // 30 days
)

// IP-API message structure
type IPInfoMsg struct {
	IP            string  `json:"query"`
	Status        string  `json:"status"`
	Continent     string  `json:"continent"`
	ContinentCode string  `json:"continentCode"`
	Country       string  `json:"country"`
	CountryCode   string  `json:"countryCode"`
	Region        string  `json:"region"`
	RegionName    string  `json:"regionName"`
	City          string  `json:"city"`
	Zip           string  `json:"zip"`
	Lat           float64 `json:"lat"`
	Lon           float64 `json:"lon"`
	Isp           string  `json:"isp"`
	Org           string  `json:"org"`
	As            string  `json:"as"`
	AsName        string  `json:"asname"`
	Mobile        bool    `json:"mobile"`
	Proxy         bool    `json:"proxy"`
	Hosting       bool    `json:"hosting"`
}

// Returns if the struct reply from the IP API is empty
func (m *IPInfoMsg) IsEmpty() bool {
	return m.Country == "" && m.City == ""
}

type IPInfoResponse struct {
	IPInfo       IPInfo
	DelayTime    time.Duration
	AttemptsLeft int
	Err          error
}

type IPInfo struct {
	IPInfoMsg
	ExpirationTime time.Time
}
