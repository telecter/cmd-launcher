package meta

import (
	"fmt"
	"time"

	"github.com/telecter/cmd-launcher/internal/network"
)

type ModrinthProject struct {
	ClientSide       string      `json:"client_side"`
	ServerSide       string      `json:"server_side"`
	GameVersions     []string    `json:"game_versions"`
	ID               string      `json:"id"`
	Slug             string      `json:"slug"`
	ProjectType      string      `json:"project_type"`
	Team             string      `json:"team"`
	Organization     string      `json:"organization"`
	Title            string      `json:"title"`
	Description      string      `json:"description"`
	Body             string      `json:"body"`
	BodyURL          interface{} `json:"body_url"`
	Published        time.Time   `json:"published"`
	Updated          time.Time   `json:"updated"`
	Approved         time.Time   `json:"approved"`
	Queued           interface{} `json:"queued"`
	Status           string      `json:"status"`
	RequestedStatus  interface{} `json:"requested_status"`
	ModeratorMessage interface{} `json:"moderator_message"`
	License          struct {
		ID   string      `json:"id"`
		Name string      `json:"name"`
		URL  interface{} `json:"url"`
	} `json:"license"`
	Downloads            int           `json:"downloads"`
	Followers            int           `json:"followers"`
	Categories           []string      `json:"categories"`
	AdditionalCategories []interface{} `json:"additional_categories"`
	Loaders              []string      `json:"loaders"`
	Versions             []string      `json:"versions"`
	IconURL              string        `json:"icon_url"`
	IssuesURL            string        `json:"issues_url"`
	SourceURL            string        `json:"source_url"`
	WikiURL              string        `json:"wiki_url"`
	DiscordURL           string        `json:"discord_url"`
	DonationUrls         []interface{} `json:"donation_urls"`
	Gallery              []struct {
		URL         string      `json:"url"`
		Featured    bool        `json:"featured"`
		Title       string      `json:"title"`
		Description interface{} `json:"description"`
		Created     time.Time   `json:"created"`
		Ordering    int         `json:"ordering"`
	} `json:"gallery"`
	Color              int    `json:"color"`
	ThreadID           string `json:"thread_id"`
	MonetizationStatus string `json:"monetization_status"`
}
type ModrinthProjectVersion struct {
	GameVersions    []string    `json:"game_versions"`
	Loaders         []string    `json:"loaders"`
	ID              string      `json:"id"`
	ProjectID       string      `json:"project_id"`
	AuthorID        string      `json:"author_id"`
	Featured        bool        `json:"featured"`
	Name            string      `json:"name"`
	VersionNumber   string      `json:"version_number"`
	Changelog       string      `json:"changelog"`
	ChangelogURL    interface{} `json:"changelog_url"`
	DatePublished   time.Time   `json:"date_published"`
	Downloads       int         `json:"downloads"`
	VersionType     string      `json:"version_type"`
	Status          string      `json:"status"`
	RequestedStatus interface{} `json:"requested_status"`
	Files           []struct {
		Hashes struct {
			Sha1   string `json:"sha1"`
			Sha512 string `json:"sha512"`
		} `json:"hashes"`
		URL      string      `json:"url"`
		Filename string      `json:"filename"`
		Primary  bool        `json:"primary"`
		Size     int         `json:"size"`
		FileType interface{} `json:"file_type"`
	} `json:"files"`
	Dependencies []interface{} `json:"dependencies"`
}
type ModrinthSearchResults struct {
	Hits []struct {
		ProjectID         string    `json:"project_id"`
		ProjectType       string    `json:"project_type"`
		Slug              string    `json:"slug"`
		Author            string    `json:"author"`
		Title             string    `json:"title"`
		Description       string    `json:"description"`
		Categories        []string  `json:"categories"`
		DisplayCategories []string  `json:"display_categories"`
		Versions          []string  `json:"versions"`
		Downloads         int       `json:"downloads"`
		Follows           int       `json:"follows"`
		IconURL           string    `json:"icon_url"`
		DateCreated       time.Time `json:"date_created"`
		DateModified      time.Time `json:"date_modified"`
		LatestVersion     string    `json:"latest_version"`
		License           string    `json:"license"`
		ClientSide        string    `json:"client_side"`
		ServerSide        string    `json:"server_side"`
		Gallery           []string  `json:"gallery"`
		FeaturedGallery   any       `json:"featured_gallery"`
		Color             int       `json:"color"`
	} `json:"hits"`
	Offset    int `json:"offset"`
	Limit     int `json:"limit"`
	TotalHits int `json:"total_hits"`
}

func SearchModrinthProjects(query string, page int) (ModrinthSearchResults, error) {
	if page < 1 {
		return ModrinthSearchResults{}, fmt.Errorf("negative page not allowed")
	}
	var data ModrinthSearchResults
	url := fmt.Sprintf("https://api.modrinth.com/v2/search?query=%s&offset=%d", query, 10*(page-1))
	if err := network.FetchJSON(url, &data); err != nil {
		return ModrinthSearchResults{}, fmt.Errorf("failed to get search results: %w", err)
	}
	return data, nil
}
func GetModrinthProject(id string) (ModrinthProject, error) {
	var data ModrinthProject
	if err := network.FetchJSON("https://api.modrinth.com/v2/project/"+id, &data); err != nil {
		return ModrinthProject{}, fmt.Errorf("failed to get project data: %w", err)
	}
	return data, nil
}

func GetModrinthProjectVersions(id string) ([]ModrinthProjectVersion, error) {
	var data []ModrinthProjectVersion
	if err := network.FetchJSON(fmt.Sprintf("https://api.modrinth.com/v2/project/%s/version", id), &data); err != nil {
		return []ModrinthProjectVersion{}, fmt.Errorf("failed to get version info: %w", err)
	}
	return data, nil
}
