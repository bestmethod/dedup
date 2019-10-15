package main

import (
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type mainStruct struct {
	paths        []string
	detail       bool
	rm           bool
	rmall        bool
	quiet        bool
	dry          bool
	multithread  bool
	sort         bool
	goroutines   *sync.WaitGroup
	fileListLock *sync.Mutex
	fileList     map[string][]*fileStruct
}

type fileStruct struct {
	path  string
	name  string
	inode string
}

func (m *mainStruct) main(args []string) (retVal int) {
	m.fileList = make(map[string][]*fileStruct)
	err := m.processFlags()
	if err != nil {
		log.Fatalf("ERROR processing flags: %s", err)
	}

	if m.multithread {
		m.fileListLock = new(sync.Mutex)
		m.goroutines = new(sync.WaitGroup)
	}

	fmt.Fprintf(os.Stderr, "Generating sums, dryrun=%t\n", m.dry)
	err = m.walk()
	if err != nil {
		log.Fatalf("ERROR walking: %s", err)
	}

	fmt.Fprintf(os.Stderr, "Enumerating table and printing duplicates\n")
	m.doDedup()
	return
}

func (m *mainStruct) doDedup() {
	for sha, f := range m.fileList {
		// sort array, fixes issue when multithreading with remove ordering
		if m.multithread || m.sort {
			sort.Slice(f, func(i, j int) bool {
				locI := -1
				locJ := -1
				for ni, mpath := range m.paths {
					if strings.HasPrefix(f[i].path, mpath) {
						locI = ni
					}
					if strings.HasPrefix(f[j].path, mpath) {
						locJ = ni
					}
					if locI > -1 && locJ > -1 {
						break
					}
				}
				if locI == -1 || locJ == -1 {
					return f[i].path < f[j].path
				}
				if locI < locJ {
					return true
				} else if locI > locJ {
					return false
				}
				// locI == locJ, alphabetical
				return f[i].path < f[j].path
			})
		}
		// find duplicate 'inode+name+path' and delete from list
		var fa []*fileStruct
		for _, nf := range f {
			found := false
			for _, fal := range fa {
				if fal.inode == nf.inode && fal.name == nf.name && fal.path == nf.path {
					found = true
					break
				}
			}
			if !found {
				fa = append(fa, nf)
			}
		}
		// go through everything
		if len(fa) > 1 {
			if !m.quiet {
				fmt.Printf("DUPLICATE: size+sha=%s\n", sha)
			}
			for i, nf := range fa {
				if !m.quiet {
					if m.rmall {
						fmt.Printf("\tremove\tinode=%s\tname=%s\tpath=%s\n", nf.inode, nf.name, nf.path)
					} else if m.rm && i != 0 {
						fmt.Printf("\tremove\tinode=%s\tname=%s\tpath=%s\n", nf.inode, nf.name, nf.path)
					} else {
						fmt.Printf("\t      \tinode=%s\tname=%s\tpath=%s\n", nf.inode, nf.name, nf.path)
					}
				}
				if m.rmall {
					if !m.dry {
						os.Remove(nf.path)
					}
				} else if m.rm && i != 0 {
					if !m.dry {
						os.Remove(nf.path)
					}
				}
			}
		}
	}
}

func main() {
	m := new(mainStruct)
	os.Exit(m.main(os.Args))
}

func (m *mainStruct) processFlags() (err error) {
	flag.BoolVar(&m.detail, "detail", false, "if set, will print each file it's processing")
	flag.BoolVar(&m.rm, "rm", false, "if set, will remove all duplicates except the first one for each file")
	flag.BoolVar(&m.rmall, "rm-all", false, "if set, will remove all files that have duplicates (all copies)")
	flag.BoolVar(&m.quiet, "quiet", false, "if set, will not print duplicate information")
	flag.BoolVar(&m.dry, "dryrun", false, "if set, doesn't perform actual remove actions")
	flag.BoolVar(&m.sort, "sort", false, "employs a file sorting mechanism like multithreading: by specified directory order, then alphabetically; useful when with -rm option (first file remains)")
	flag.BoolVar(&m.multithread, "multithread", false, "if set, enable multithreading - one thread per listed directory; will sort by specified directory order, then alphabetically")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  %s [-detail] [-rm|-rm-all] [-quiet] [-dryrun] directory\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s [-detail] [-rm|-rm-all] [-quiet] [-dryrun] directory1 directory2 directory...\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptional Arguments:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	m.paths = flag.Args()
	if len(m.paths) == 0 {
		err = errors.New("specify a path. Run with '-help' for usage instructions")
		return
	}
	for _, mpath := range m.paths {
		n, err := os.Stat(mpath)
		if os.IsNotExist(err) {
			return fmt.Errorf("Does not exist: %s, %s", mpath, err)
		}
		if err != nil {
			return fmt.Errorf("Cannot stat directory %s: %s", mpath, err)
		}
		if n.IsDir() == false {
			return fmt.Errorf("Is not a directory: %s", mpath)
		}
	}
	return
}

func (m *mainStruct) walkEach(mpath string) (err error) {
	err = filepath.Walk(mpath, m.doWalk)
	if err != nil {
		err = fmt.Errorf("Error executing filepath.Walk on %s: %s", mpath, err)
	}
	if m.multithread {
		m.goroutines.Done()
	}
	return
}

func (m *mainStruct) walk() (err error) {
	for _, mpath := range m.paths {
		fmt.Fprintf(os.Stderr, "Generating sums on %s\n", mpath)
		if m.multithread {
			m.goroutines.Add(1)
			go m.walkEach(mpath)
		} else {
			m.walkEach(mpath)
		}
	}
	if m.multithread {
		m.goroutines.Wait()
	}
	return
}

func (m *mainStruct) doWalk(path string, info os.FileInfo, err error) error {
	if m.detail == true {
		fmt.Fprintf(os.Stderr, "Processing %s\n", path)
	}
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	if !info.Mode().IsRegular() {
		return nil
	}
	nFile := new(fileStruct)
	nFile.name = info.Name()
	nFile.path = path
	nFile.inode = getSys(info)
	fileSize := info.Size()
	fileSha, err := getSum(path)
	if err != nil {
		return err
	}
	fileSum := fmt.Sprintf("%d+%x", fileSize, fileSha)
	if m.multithread {
		m.fileListLock.Lock()
	}
	m.fileList[fileSum] = append(m.fileList[fileSum], nFile)
	if m.multithread {
		m.fileListLock.Unlock()
	}
	return nil
}

func getSum(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open file '%s': %s", path, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, fmt.Errorf("could not read file '%s': %s", path, err)
	}
	return h.Sum(nil), nil
}
