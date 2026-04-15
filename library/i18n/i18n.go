package i18n

import (
	"encoding/json"
	"os"
	"sync"
)

var (
	messages map[string]map[string]string
	once     sync.Once
)

// Init hanya dipanggil sekali
func Init() {
	once.Do(func() {
		messages = make(map[string]map[string]string)

		loadFile("en", "i18n/en.json")
		loadFile("id", "i18n/id.json")
	})
}

func loadFile(lang, path string) {
	file, err := os.ReadFile(path)
	if err != nil {
		panic("failed to load i18n file: " + path)
	}

	var data map[string]string
	if err := json.Unmarshal(file, &data); err != nil {
		panic("invalid json in: " + path)
	}

	messages[lang] = data
}

// Translate
func T(lang, key string) string {
	// fallback default
	if lang == "" {
		lang = "en"
	}

	// ambil bahasa utama (id-ID → id)
	if len(lang) > 2 {
		lang = lang[:2]
	}

	if m, ok := messages[lang]; ok {
		if val, ok := m[key]; ok {
			return val
		}
	}

	// fallback ke en
	if val, ok := messages["en"][key]; ok {
		return val
	}

	return key
}
