package chatbuilder

import "encoding/json"

var matcher = []string{"command", "path"}

func bashInputToStr(jsonString string) string {
	var args map[string]any
	err := json.Unmarshal([]byte(jsonString), &args)
	if err != nil {
		return jsonString
	}
	for _, key := range matcher {
		if value, ok := args[key].(string); ok && value != "" {
			return value
		}
	}
	return jsonString
}
