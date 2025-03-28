package segments

import (
	"fmt"
	"math"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
)

// StravaAPI is a wrapper around http.Oauth
type StravaAPI interface {
	GetActivities() ([]*StravaData, error)
}

type stravaAPI struct {
	http.OAuthRequest
}

func (s *stravaAPI) GetActivities() ([]*StravaData, error) {
	url := "https://www.strava.com/api/v3/athlete/activities?page=1&per_page=1"
	return http.OauthResult[[]*StravaData](&s.OAuthRequest, url, nil)
}

// segment struct, makes templating easier
type Strava struct {
	base

	api   StravaAPI
	Icon  string
	Ago   string
	Error string
	URL   string
	StravaData
	Hours        int
	Authenticate bool
}

const (
	RideIcon            properties.Property = "ride_icon"
	RunIcon             properties.Property = "run_icon"
	SkiingIcon          properties.Property = "skiing_icon"
	WorkOutIcon         properties.Property = "workout_icon"
	UnknownActivityIcon properties.Property = "unknown_activity_icon"

	StravaAccessTokenKey  = "strava_access_token"
	StravaRefreshTokenKey = "strava_refresh_token"

	noActivitiesFound = "No activities found"
)

// StravaData struct contains the API data
type StravaData struct {
	StartDate            time.Time `json:"start_date"`
	Type                 string    `json:"type"`
	Name                 string    `json:"name"`
	ID                   int       `json:"id"`
	Distance             float64   `json:"distance"`
	Duration             float64   `json:"moving_time"`
	AverageWatts         float64   `json:"average_watts"`
	WeightedAverageWatts float64   `json:"weighted_average_watts"`
	AverageHeartRate     float64   `json:"average_heartrate"`
	MaxHeartRate         float64   `json:"max_heartrate"`
	KudosCount           int       `json:"kudos_count"`
	DeviceWatts          bool      `json:"device_watts"`
}

func (s *Strava) Template() string {
	return " {{ if .Error }}{{ .Error }}{{ else }}{{ .Ago }}{{ end }} "
}

func (s *Strava) Enabled() bool {
	s.initAPI()

	data, err := s.api.GetActivities()
	if err == nil && len(data) > 0 {
		s.StravaData = *data[0]
		s.Icon = s.getActivityIcon()
		s.Hours = s.getHours()
		s.Ago = s.getAgo()
		s.URL = fmt.Sprintf("https://www.strava.com/activities/%d", s.ID)
		return true
	}
	if err == nil && len(data) == 0 {
		s.Error = noActivitiesFound
		return true
	}
	if _, s.Authenticate = err.(*http.OAuthError); s.Authenticate {
		s.Error = err.(*http.OAuthError).Error()
		return true
	}
	return false
}

func (s *Strava) initAPI() {
	if s.api != nil {
		return
	}

	oauth := &http.OAuthRequest{
		AccessTokenKey:  StravaAccessTokenKey,
		RefreshTokenKey: StravaRefreshTokenKey,
		SegmentName:     "strava",
		AccessToken:     s.props.GetString(properties.AccessToken, ""),
		RefreshToken:    s.props.GetString(properties.RefreshToken, ""),
		Request: http.Request{
			Env:         s.env,
			HTTPTimeout: s.props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout),
		},
	}

	s.api = &stravaAPI{
		OAuthRequest: *oauth,
	}
}

func (s *Strava) getHours() int {
	hours := time.Since(s.StartDate).Hours()
	return int(math.Floor(hours))
}

func (s *Strava) getAgo() string {
	if s.Hours > 24 {
		days := int32(math.Floor(float64(s.Hours) / float64(24)))
		return fmt.Sprintf("%d", days) + string('d')
	}
	return fmt.Sprintf("%d", s.Hours) + string("h")
}

func (s *Strava) getActivityIcon() string {
	switch s.Type {
	case "VirtualRide":
		fallthrough
	case "Ride":
		return s.props.GetString(RideIcon, "\uf206")
	case "Run":
		return s.props.GetString(RunIcon, "\ue213")
	case "NordicSki":
	case "AlpineSki":
	case "BackcountrySki":
		return s.props.GetString(SkiingIcon, "\ue213")
	case "WorkOut":
		return s.props.GetString(WorkOutIcon, "\ue213")
	default:
		return s.props.GetString(UnknownActivityIcon, "\ue213")
	}
	return s.props.GetString(UnknownActivityIcon, "\ue213")
}
