package main

// Glen Newton
// glen.newton@gmail.com

import (
	"fmt"
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
	fmt.Println("\t Returns 2 lines with each of the following followed by the paths found for each:  \"LD PATH: \"   \"PATH: \" \n")
}

func main() {

	if len(os.Args) < 2 {
		usage(os.Args[0])
		return
	}

	err := checkFileNamesExist(os.Args[1:])
	if err != nil {
		return
	}

	fileLineChannel := make(chan string, 200)

	go findRpmFiles(fileLineChannel, filesRpmExec, os.Args)

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
			log.Print("Problem opening file: " + fileName)
			return err
		}
		fi, err := file.Stat()
		if err != nil {
			log.Print("Problem stat'ing file: " + fileName)
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
	// rpmFiles := []string{
	// 	"/home/newtong/work/rpms-biocluster/branches/rocks-6.1/mira/dependencies/gperftools-devel-2.0-3.el6.2.x86_64.rpm",
	// 	"/home/newtong/work/R/R2/R-2.15.1-1.x86_64.rpm",
	// 	"/home/newtong/work/R/R2/RPMS/x86_64/R2-core-2.15.1-2.x86_64.rpm",
	// 	"/home/newtong/work/rpms-biocluster/branches/rocks-6.1/sparsehash/sparsehash-2.0.2-1.noarch.rpm",
	// }

	var count = 0
	doneChannel := make(chan bool, 20)
	for i, rpmFile := range args {
		if i == 0 {
			continue
		}
		count += 1
		prog := append(filesRpmExec, rpmFile)
		runExec(prog, fileLineChannel, doneChannel)
		//fmt.Println(prog)
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
			//fmt.Println(isExecutable, isLibrary, path, v)
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
