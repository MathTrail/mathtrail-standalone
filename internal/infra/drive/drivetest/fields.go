package drivetest

import (
	"fmt"
	"strings"
)

// The fields Drive answers with when a call names none.
var defaultFields = []string{"kind", "id", "name", "mimeType"}

// fileFields are the fields of a file the stand-in knows how to answer with.
var fileFields = map[string]bool{
	"kind": true, "id": true, "name": true, "mimeType": true, "parents": true,
	"appProperties": true, "modifiedTime": true, "webViewLink": true, "trashed": true,
}

// fileSelection reads the fields a call asks of one file: a list such as
// id,name. A field the stand-in does not know is refused, as Drive refuses a
// field it does not have.
func fileSelection(fields string) ([]string, error) {
	if fields == "" {
		return defaultFields, nil
	}
	selected := strings.Split(fields, ",")
	for _, field := range selected {
		if !fileFields[field] {
			return nil, fmt.Errorf("%q is no field of a file the stand-in answers with", field)
		}
	}
	return selected, nil
}

// listSelection reads the fields a search asks of each file it finds, in the
// form files(id,name), which may be followed by nextPageToken.
func listSelection(fields string) ([]string, error) {
	if fields == "" {
		return defaultFields, nil
	}
	inner, rest, found := strings.Cut(strings.TrimPrefix(fields, "files("), ")")
	if !found || !strings.HasPrefix(fields, "files(") || (rest != "" && rest != ",nextPageToken") {
		return nil, fmt.Errorf("%q does not ask for files(…)", fields)
	}
	return fileSelection(inner)
}

// pick is a file as Drive answers with it: the fields asked for, and the ones
// the file has nothing in left out.
func (f *File) pick(fields []string) map[string]any {
	answer := map[string]any{}
	for _, field := range fields {
		switch field {
		case "kind":
			answer["kind"] = "drive#file"
		case "id":
			answer["id"] = f.ID
		case "name":
			answer["name"] = f.Name
		case "mimeType":
			answer["mimeType"] = f.MimeType
		case "parents":
			if len(f.Parents) > 0 {
				answer["parents"] = f.Parents
			}
		case "appProperties":
			if len(f.AppProperties) > 0 {
				answer["appProperties"] = f.AppProperties
			}
		case "modifiedTime":
			answer["modifiedTime"] = f.ModifiedTime.UTC().Format("2006-01-02T15:04:05.000Z")
		case "webViewLink":
			answer["webViewLink"] = WebViewLink(f.ID)
		case "trashed":
			answer["trashed"] = f.Trashed
		}
	}
	return answer
}

// WebViewLink is where a file of the stand-in's opens, as Drive links a file.
func WebViewLink(id string) string {
	return "https://drive.google.com/file/d/" + id + "/view?usp=drivesdk"
}
