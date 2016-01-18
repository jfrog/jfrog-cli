package utils

import (
    "os"
    "fmt"
    "bytes"
    "strings"
    "encoding/json"
)

func CheckError(err error) {
    if err != nil {
        panic(err)
    }
}

func Exit(msg string) {
    fmt.Println(msg)
    os.Exit(1)
}

func AddTrailingSlashIfNeeded(url string) string {
    if url != "" && !strings.HasSuffix(url, "/") {
        url += "/"
    }
    return url
}

func GetFileNameFromUrl(url string) string {
    parts := strings.Split(url, "/")
    size := len(parts)
    if size == 0 {
        return url
    }
    return parts[size-1]
}

func IndentJson(jsonStr []byte) string {
    var content bytes.Buffer
    err := json.Indent(&content, jsonStr, "", "  ")
    if err == nil {
        return content.String()
    }
    return string(jsonStr)
}

// Creates a string in the form of ["item-1","item-2","item-3"...] from an input
// in the form of item-1,item-1,item-1...
func BuildListString(listStr string) string {
    if listStr == "" {
        return ""
    }
    split := strings.Split(listStr, ",")
    size := len(split)
    str := "[\""
    for i := 0; i < size; i++ {
        str += split[i]
        if i+1 < size {
            str += "\",\""
        }
    }
    str += "\"]"
    return str
}

func MapToJson(m map[string]string) string {
    first := true
    json := "{"

    for key := range m {
        val := m[key]
        if val != "" {
            if !first {
                json += ","
            }
            first = false
            if !strings.HasPrefix(val, "[") || !strings.HasSuffix(val, "]") {
                val = "\"" + val + "\""
            }
            json += "\"" + key + "\": " + val
        }
    }
    if first {
        return ""
    }
    json += "}"
    return json
}

func ConfirmAnswer(answer string) bool {
    answer = strings.ToLower(answer)
    return answer == "y" || answer == "yes"
}