package translator

import (
	"bufio"
	"dbf-column-filler/internal/models"
	"fmt"
	"os"
	"strings"
)

func LoadTranslations() (models.Translations, error) {
	file, err := os.Open(fmt.Sprintf("lang.txt"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	translations := make(models.Translations)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		translations[parts[0]] = parts[1]
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return translations, nil
}
