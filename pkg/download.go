package uchess

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/codeclysm/extract/v3"
	"github.com/dustin/go-humanize"
	"github.com/mitchellh/go-homedir"
)

// WriteCounter counts the number of bytes written to it. By implementing the Write method,
// it is of the io.Writer interface and we can pass this into io.TeeReader()
// Every write to this writer, will print the progress of the file write
type WriteCounter struct {
	Total uint64
}

// Write updates the WriteCounter total and prints progress
func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

// PrintProgress prints the progress of a file write
func (wc WriteCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	fmt.Printf("\r%s", strings.Repeat(" ", 50))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}

// DownloadFile will download a url and store it in local filepath.
// It writes to the destination file as it downloads it, without
// loading the entire file into memory.
// We pass an io.TeeReader into Copy() to report progress on the download.
func DownloadFile(url string, filepath string) error {
	// Create the file with .tmp extension, so that we won't overwrite a
	// file until it's downloaded fully
	out, err := os.Create(filepath + ".tmp")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	// Create our bytes counter and pass it to be used alongside our writer
	counter := &WriteCounter{}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	// The progress use the same line so print a new line once it's finished downloading
	fmt.Println()
	// Close the file to avoid a resource lock
	out.Close()
	resp.Body.Close()

	// Rename the tmp file back to the original file
	err = os.Rename(filepath+".tmp", filepath)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}

// MatchStockfishWin returns a boolean indicating whether a given file
// looks like the stockfish executable (Windows). Fuzzy matching is applied
func MatchStockfishWin(file string) bool {
	stockfishExe := regexp.MustCompile(`(?i)^stockfish.*\.exe$`)
	return stockfishExe.MatchString(file)
}

// SearchDirForStockfish scans a directory and looks for a stockfish binary. Upon
// matching, a string with the binary name is returned. If no match is found,
// an empty string is returned
func SearchDirForStockfish(dir string) string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return ""
	}

	for _, f := range files {
		fullPath := filepath.Join(dir, f.Name())
		if runtime.GOOS == "windows" && MatchStockfishWin(f.Name()) {
			return f.Name()
		}

		if runtime.GOOS != "windows" && f.Name() == "stockfish" && IsFile(fullPath) {
			return f.Name()
		}
	}
	return ""
}

func ensureUchessDir() string {
	home, err := homedir.Dir()

	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	uchessDir := filepath.Join(home, uchessDirName())
	if !FileExists(uchessDir) {
		err = os.Mkdir(uchessDir, os.ModeDir)
		if err != nil {
			fmt.Println(err.Error())
			return ""
		}
	}
	return uchessDir
}

func isStockfishBin(file string) bool {
	return strings.HasPrefix(file, "stockfish") && !strings.HasSuffix(file, ".zip")
}

var stockfishWin = "https://stockfishchess.org/files/stockfish_12_win_x64.zip"
var stockfishLin = "https://stockfishchess.org/files/stockfish_12_linux_x64.zip"

func mkDownloadDir() (string, error) {
	// Create a temp directory to hold the download
	tmp := os.Getenv("TMP")
	dir, err := ioutil.TempDir(tmp, "uchess")

	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	return dir, nil
}

func extractZip(dlDir, dlPath string) error {
	zipData, err := ioutil.ReadFile(dlPath)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	buffer := bytes.NewBuffer(zipData)
	if err = extract.Zip(context.Background(), buffer, dlDir, nil); err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func findTmpStockfish(dlDir string) (string, error) {
	files, err := ioutil.ReadDir(dlDir)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	for _, file := range files {
		fileName := file.Name()
		if isStockfishBin(fileName) {
			return filepath.Join(dlDir, fileName), nil
		}
	}
	return "", nil
}

// StockfishFilename returns a platform specific filename for the Stockfish binary
func StockfishFilename() string {
	if runtime.GOOS == "windows" {
		return "stockfish.exe"
	}
	return "stockfish"
}

func downloadStockfish() bool {
	var stockfish, dlDir, sfPath string
	var err error

	// Windows/amd64 and Linux/amd64 support automated download/install
	if runtime.GOOS == "windows" && runtime.GOARCH == "amd64" {
		stockfish = stockfishWin
	} else if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		stockfish = stockfishLin
	}

	// Set up a temp download dir
	if dlDir, err = mkDownloadDir(); err != nil {
		fmt.Println(err.Error())
		return false
	}
	// Clean up the temp dir when finished
	defer os.RemoveAll(dlDir)

	// dlPath is the full path that the initial zip was saved to
	dlPath := filepath.Join(dlDir, "stockfish.zip")
	if err = DownloadFile(stockfish, dlPath); err != nil {
		fmt.Println(err.Error())
		return false
	}

	// Extract the zip to the temp location
	if err = extractZip(dlDir, dlPath); err != nil {
		fmt.Println(err.Error())
		return false
	}

	// Locate the extracted stockfish binary
	if sfPath, err = findTmpStockfish(dlDir); err != nil || sfPath == "" {
		fmt.Println(err.Error())
		return false
	}

	// Move and rename the extracted binary
	if uchessDir := ensureUchessDir(); uchessDir != "" {
		finalPath := filepath.Join(uchessDir, StockfishFilename())
		if err = os.Rename(sfPath, finalPath); err != nil {
			fmt.Println(err.Error())
			return false
		}
	}
	return true
}

func uchessDirName() string {
	if runtime.GOOS == "windows" {
		return "uchess"
	}
	return ".uchess"
}

// AppDir returns the uchess directory location
func AppDir() string {
	homeDir, err := homedir.Dir()

	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return filepath.Join(homeDir, uchessDirName())
}

func pathDelim() string {
	if runtime.GOOS == "windows" {
		return ";"
	}
	return ":"
}

// FindStockfish attempts to find Stockfish in the PATH and config dir
// A string with the path is returned when found, an empty string is returned otherwise
func FindStockfish() string {
	pathDelim := pathDelim()
	osPath := os.Getenv("PATH")
	paths := strings.Split(osPath, pathDelim)
	appDir := AppDir()
	if appDir != "" && FileExists(appDir) {
		paths = append(paths, appDir)
	}

	for _, p := range paths {
		searchResult := SearchDirForStockfish(p)
		if searchResult != "" {
			return filepath.Join(p, searchResult)
		}
	}
	return ""
}

// FetchStockfish attempts to load Stockfish via the path and attempts
// to download the binary for a limited number of platforms
func FetchStockfish() string {
	stockfish := FindStockfish()

	if stockfish == "" {
		fmt.Printf("\nStockfish could not be found in your path.\n")

		if (runtime.GOOS == "windows" || runtime.GOOS == "linux") && runtime.GOARCH == "amd64" {
			var resp string
			for resp != "yes" && resp != "no" && resp != "y" && resp != "n" {
				fmt.Printf("Attempt an automated install? [y/n] ")
				reader := bufio.NewReader(os.Stdin)
				resp, _ = reader.ReadString('\n')
				resp = strings.TrimSpace(strings.ToLower(resp))
			}

			if resp == "yes" || resp == "y" {
				success := downloadStockfish()

				if success {
					fmt.Printf("Install was successful. Restart uchess.\n\n")
					os.Exit(0)
				} else {
					fmt.Printf("Install was unsuccessful. Please install Stockfish manually.\n\n")
					os.Exit(0)
				}
			} else {
				fmt.Printf("Please see the docs for manual Stockfish configuration.\n\n")
				os.Exit(0)
			}
		} else {
			fmt.Println("Automated installation is not supported on your platform.")
			fmt.Println("Please install Stockfish using your package manager ")
			fmt.Println("or download a binary from the official website and ")
			fmt.Printf("make sure the executable is in your path.\n\n")
			fmt.Printf("https://stockfishchess.org/download/\n\n")
			os.Exit(0)
		}
	}
	return ""
}
