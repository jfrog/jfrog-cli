package client
import "strconv"

const defaultApiUrl = "https://bintray.com/api/v1/"

type Bintray struct {
    Username string
    ApiKey   string
    ApiUrl   string
    Flags    map[string]string
}

func New(username string, apiKey string, apiUrl string, flags map[string]string) *Bintray {
    if apiUrl == "" { apiUrl = apiUrl }
    if apiUrl == "" {
        apiUrl = defaultApiUrl
    }
    return &Bintray{username, apiKey, apiUrl, flags}
}

func (bt *Bintray) FlagStr(name string) string {
    return bt.Flags[name]
}

func (bt *Bintray) FlagInt(name string) int {
    if v, ok := bt.Flags[name]; ok {
        f, _ := strconv.Atoi(v)
        return f
    }
    return 0
}

func (bt *Bintray) FlagBool(name string) bool {
    if v, ok := bt.Flags[name]; ok {
        f, _ := strconv.ParseBool(v)
        return f
    }
    return false
}

func (bt *Bintray) Subject() string {
    //Get the provided subject, otherwise default to the user
    if subject, ok:= bt.Flags["subject"];ok {
        return subject
    }
    return bt.Username
}
