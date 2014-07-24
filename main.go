package main

// Glen Newton
// glen.newton@gmail.com

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

type PathInfo struct {
	path      string
	isLibrary bool
}

var re = regexp.MustCompile(" +")
var filesRpmExec = []string{"rpm", "-qlvp"}

func usage(name string) {
	fmt.Println("\n\t" + name + ": Find executables and libraries in RPMs to populate PATH and LD_LIBRARY_PATH")
	fmt.Println("Usage: " + name + " <rpmfile0> ... <rpmfileN>")
	fmt.Println("\t Returns 2 lines with each of the following followed by the paths found for each:   \"PATH: \"   \"LD PATH: \"  \n")
	flag.Usage()
}

var filesFromStdin = false

func init() {
	flag.BoolVar(&filesFromStdin, "c", filesFromStdin, "Files one per line in stdin")
	flag.Parse()
}

func main() {
	flag.Parse()

	files, err := getFileNames(filesFromStdin)
	if err != nil {
		os.Exit(42)
		return
	}

	err = checkFileNamesExist(files)
	if err != nil {
		os.Exit(42)
		return
	}

	fileLineChannel := make(chan string, 200)

	go findRpmFiles(fileLineChannel, filesRpmExec, files)

	pathChannel := make(chan *PathInfo, 200)

	go makePaths(fileLineChannel, pathChannel)

	outPath := ""
	outLdLibraryPath := ""

	for path := range pathChannel {
		if path.isLibrary {
			if outLdLibraryPath != "" {
				outLdLibraryPath += ":"
			}
			outLdLibraryPath += path.path
		} else {
			if outPath != "" {
				outPath += ":"
			}
			outPath += path.path
		}
	}
	fmt.Println("PATH=" + outPath)
	fmt.Println("LD_LIBRARY_PATH=" + outLdLibraryPath)
}

func checkFileNamesExist(fileNames []string) error {
	for _, fileName := range fileNames {
		file, err := os.Open(fileName) // For read access.
		if err != nil {
			log.Print("Problem opening file: [" + fileName + "]")
			return err
		}
		fi, err := file.Stat()
		if err != nil {
			log.Print("Problem stat'ing file: [" + fileName + "]")
			return err
		}

		fm := fi.Mode()
		if !fm.IsRegular() {
			error := new(InternalError)
			error.errorString = "Is directory, needs to be file: " + fileName
			log.Print(error.errorString)
			return error
		}

		if fm.Perm().String()[7] != 'r' {
			error := new(InternalError)
			error.errorString = "Exists but unable to read: " + fileName
			log.Print(error.errorString)
			return error

		}
	}
	return nil
}

type InternalError struct {
	errorString string
}

func (ie *InternalError) Error() string {
	return "Error: " + ie.errorString
}

func findRpmFiles(fileLineChannel chan string, filesRpmExec []string, args []string) {
	var count = 0
	doneChannel := make(chan bool, 20)
	for _, rpmFile := range args {
		count += 1
		prog := append(filesRpmExec, rpmFile)
		runExec(prog, fileLineChannel, doneChannel)
	}

	for i := 0; i < count; i++ {
		_ = <-doneChannel
	}

	close(fileLineChannel)
}

func makePaths(fileLineChan chan string, pathChannel chan *PathInfo) {
	pathsMap := make(map[string]bool)

	for v := range fileLineChan {
		v = re.ReplaceAllString(v, " ")
		isExecutable, isLibrary, pathString := extract(v)

		_, ok := pathsMap[pathString]
		var path *PathInfo
		if (isExecutable || isLibrary) && !ok {
			pathsMap[pathString] = true
			path = new(PathInfo)
			if isLibrary {
				path.isLibrary = true

			} else {
				path.isLibrary = false
			}
			path.path = pathString
			pathChannel <- path
		}
	}
	close(pathChannel)
}

func extract(p string) (bool, bool, string) {
	parts := strings.Split(p, " ")

	isExecutableFile := findExecutableFile(parts[0])

	var pathPart string
	if len(parts) == 9 {
		pathPart = parts[8]
	} else {
		// bug in rpm...
		if len(parts) == 8 {
			pathPart = parts[7]
			fmt.Println("#  " + p)
		} else {
			return false, false, ""
		}
	}

	isLibrary := findLibrary(pathPart)
	path := ""
	if isLibrary || isExecutableFile {
		path = makePath(pathPart)
	}

	if isLibrary {
		isExecutableFile = false
	}

	return isExecutableFile, isLibrary, path
}

func getFileNames(filesFromStdin bool) ([]string, error) {
	var files []string
	if !filesFromStdin {
		// read files line by line from stdin
		files = flag.Args()[0:]
	} else {
		// or command line
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadString('\n')

			if err != nil {
				if err == io.EOF {
					break
				} else {
					log.Fatal(err)
					return nil, err
				}
			}
			files = append(files, line[0:len(line)-1])
		}
	}
	return files, nil
}

func makePath(fullpath string) string {
	index := strings.LastIndex(fullpath, "/")
	if index < 0 {
		return fullpath
	}
	return fullpath[0:index]
}

func findExecutableFile(perms string) bool {
	return (perms[3] == 'x') && (perms[0] != 'd')
}

func findLibrary(path string) bool {
	return strings.HasSuffix(path, ".o") ||
		strings.HasSuffix(path, ".a") ||
		strings.HasSuffix(path, ".so") ||
		strings.Contains(path, ".so.") ||
		strings.Contains(path, ".a.")
}

func separatePaths(p []string) string {
	s := ""

	ps := string(os.PathListSeparator)
	for i, v := range p {
		if i != 0 {
			s += ps
		}
		s += v
	}
	return s
}
