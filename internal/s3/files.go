package s3

import (
	"fmt"
	"os"
)

// getFilesInDirectory retrieves the list of files in the specified directory.
func getFilesInDirectory(dir string) ([]string, error) {
	const op = "s3.getFilesInDirectory"

	dis, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read directory %s: %w", op, dir, err)
	}

	var files []string
	for _, di := range dis {
		if !di.IsDir() {
			files = append(files, di.Name())
		}
	}

	return files, nil
}

// readFileContent reads the content of the specified file.
func readFileContent(filePath string) ([]byte, error) {
	const op = "s3.readFileContent"

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read file %s: %w", op, filePath, err)
	}
	return data, nil
}
