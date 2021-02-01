package uchess

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/artdarek/go-unzip"
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
		return err
	}

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	// Create our bytes counter and pass it to be used alongside our writer
	counter := &WriteCounter{}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
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
		return err
	}

	return nil
}

// MatchStockfishWin returns a boolean indicating whether a given file
// looks like the stockfish executable (Windows). Fuzzy matching is applied
func MatchStockfishWin(file string) bool {
	stockfishExe := regexp.MustCompile(`(?i)^stockfish.+\.exe$`)
	return stockfishExe.MatchString(file)
}

// FindStockfish scans a directory and looks for a stockfish binary. Upon
// matching, a string with the binary name is returned. If no match is found,
// an empty string is returned
func FindStockfish(dir string) string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return ""
	}

	for _, f := range files {
		fullPath := fmt.Sprintf("%v%v%v", dir, string(os.PathSeparator), f.Name())
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
		return ""
	}

	uchessDir := fmt.Sprintf("%v%v%v", home, string(os.PathSeparator), "uchess")
	if !FileExists(uchessDir) {
		err = os.Mkdir(uchessDir, os.ModeDir)
		if err != nil {
			return ""
		}
	}
	return uchessDir
}

var stockfishWin = "https://stockfishchess.org/files/stockfish_12_win_x64.zip"
var stockfishLin = "https://stockfishchess.org/files/stockfish_12_linux_x64.zip"

func downloadStockfish() bool {
	var stockfish string

	if runtime.GOOS == "windows" && runtime.GOARCH == "amd64" {
		stockfish = stockfishWin
	} else if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		stockfish = stockfishLin
	}

	if uchessDir := ensureUchessDir(); uchessDir != "" {
		file := fmt.Sprintf("%v%v%v", uchessDir, string(os.PathSeparator), "stockfish.zip")
		err := DownloadFile(stockfish, file)

		if err != nil {
			fmt.Println(err.Error())
			return false
		}

		uz := unzip.New(file, uchessDir)
		err = uz.Extract()

		if err != nil {
			fmt.Println(err)
			return false
		}
	}
	return true
}

// FindOrFetchStockfish attempts to load Stockfish via the path and attempts
// to download the binary for a limited number of platforms
func FindOrFetchStockfish() string {
	var pathDelim string

	if runtime.GOOS == "windows" {
		pathDelim = ";"
	} else {
		pathDelim = ":"
	}

	path := os.Getenv("PATH")
	paths := strings.Split(path, pathDelim)
	homeDir, err := homedir.Dir()
	uchessDir := fmt.Sprintf("%v%v%v", homeDir, string(os.PathSeparator), "uchess")

	if err != nil {
		fmt.Println(err.Error())
	} else {
		paths = append(paths, uchessDir)
	}

	foundStockfish := false
	for _, p := range paths {
		fmt.Printf("Search: %v\n", p)
		searchResult := FindStockfish(p)
		if searchResult != "" {
			fmt.Printf("Found: %v%v%v\n", p, string(os.PathSeparator), searchResult)
			return fmt.Sprintf("%v%v%v", p, string(os.PathSeparator), searchResult)
		}
	}

	if !foundStockfish {
		fmt.Printf("\nStockfish could not be found in your path.\n")

		if (runtime.GOOS == "windows" || runtime.GOOS == "linux") && runtime.GOARCH == "amd64" {
			var resp string
			for resp != "yes" && resp != "no" && resp != "y" && resp != "n" {
				fmt.Println("Attempt an automated install? [y/n] ")
				reader := bufio.NewReader(os.Stdin)
				resp, _ = reader.ReadString('\n')
				resp = strings.ToLower(resp)
			}

			if resp == "yes" || resp == "y" {
				success := downloadStockfish()

				if success {
					fmt.Println("Install was successful. Restart uchess.")
					os.Exit(0)
				}
			}
		} else {
			fmt.Println("Automated install is not supported on your platform.")
			fmt.Println("Please install Stockfish using your package manager ")
			fmt.Println("or download a binary from the official website and ")
			fmt.Printf("make sure the executable is in your path.\n\n")
			fmt.Println("https://stockfishchess.org/download/")
			os.Exit(0)
		}
	}
	// unreachable
	return ""
}
