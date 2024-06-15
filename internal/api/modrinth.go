package api

import (
	"fmt"
	"slices"
	"time"

	util "github.com/telecter/cmd-launcher/internal"
)

type Project struct {
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
type ProjectVersion struct {
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

func GetModrinthProject(id string) (Project, error) {
	data := Project{}
	err := util.GetJSON("https://api.modrinth.com/v2/project/"+id, &data)
	if err != nil {
		return data, fmt.Errorf("failed to get project: %s", err)
	}
	return data, nil
}
func DownloadModrinthProject(path string, id string, gameVersion string, loader string) error {
	data := []ProjectVersion{}
	err := util.GetJSON(fmt.Sprintf("https://api.modrinth.com/v2/project/%s/version", id), &data)
	if err != nil {
		return fmt.Errorf("failed to get version info: %v", err)
	}
	for _, version := range data {
		if slices.Contains(version.Loaders, loader) && slices.Contains(version.GameVersions, gameVersion) {
			err := util.DownloadFile(version.Files[0].URL, path+"/"+version.Files[0].Filename)
			if err != nil {
				return fmt.Errorf("failed to download project file: %s", err)
			}
			return nil
		}
	}
	return fmt.Errorf("no version found")
}
