package utils

func GetString(data map[string]interface{}, key string) string {
    if value, ok := data[key]; ok && value != nil {
        return value.(string)
    }
    return ""
}
