package process

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Chunk struct {
	Paths []string
}

type GitInfo struct {
	Author    string
	CreatedAt time.Time
}

type ProcessorResult struct {
	FullPath string
	Name     string
	Info     *GitInfo
}

type Processor struct {
	ChunkSize int
	RootDir   string
	Dig       bool

	chunks  map[int]*Chunk
	results []*ProcessorResult
}

func NewProcessor(chunkSize int, rootDir string, dig bool) *Processor {
	return &Processor{
		ChunkSize: chunkSize,
		RootDir:   rootDir,
		Dig:       dig,
		chunks:    make(map[int]*Chunk),
	}
}

func (p *Processor) Begin() ([]*ProcessorResult, error) {
	dirs, err := ioutil.ReadDir(p.RootDir)
	if err != nil {
		return nil, err
	}

	correctedIndex := 0
	currentChunk := 0

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		if p.chunks[currentChunk] == nil {
			p.chunks[currentChunk] = new(Chunk)
		}

		if correctedIndex < p.ChunkSize-1 {
			p.chunks[currentChunk].Paths = append(p.chunks[currentChunk].Paths, dir.Name())
			correctedIndex += 1
		} else {
			if p.chunks[currentChunk+1] == nil {
				p.chunks[currentChunk+1] = &Chunk{
					Paths: []string{dir.Name()},
				}
			}

			currentChunk += 1
			correctedIndex = 0
		}
	}

	// totalPaths := 0
	// for idx, ch := range p.chunks {
	// 	totalPaths += len(ch.Paths)
	// 	fmt.Printf("%d - %+v\n", idx, ch.Paths)
	// }

	// fmt.Printf("TOTAL PATHS: %d\n", totalPaths)

	return p.processChunks(), nil
}

func (p *Processor) processChunks() []*ProcessorResult {
	var waiter sync.WaitGroup
	waiter.Add(len(p.chunks))

	for _, chunk := range p.chunks {
		go p.processChunk(chunk, &waiter)
	}

	waiter.Wait()
	return p.results
}

func (p *Processor) processChunk(chunk *Chunk, waiter *sync.WaitGroup) {
	defer waiter.Done()

	for _, path := range chunk.Paths {
		fullPath := filepath.Join(p.RootDir, path)
		if err := filepath.Walk(fullPath, func(walkPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() && info.Name() == ".git" {
				walkPath = walkPath[:len(walkPath)-5]

				result := ProcessorResult{
					FullPath: walkPath,
					Name:     path,
				}

				if p.Dig {
					result.Info = p.dig(walkPath)
				}

				p.results = append(p.results, &result)
			}

			return nil
		}); err != nil {
			panic(err)
		}
	}
}

func (p *Processor) dig(fullPath string) *GitInfo {
	rawCommand := "git log --reverse | head -3"

	cmd := exec.Command("bash", "-c", rawCommand)
	cmd.Dir = fullPath

	output, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	data := string(output)

	author := strings.Split(data, "Author:")
	createdAt := strings.Split(data, "Date:")

	var info GitInfo

	if len(author) > 1 {
		info.Author = strings.TrimSpace(strings.Split(author[1], "\n")[0])
	}

	if len(createdAt) > 1 {
		createdTime := strings.TrimSpace(createdAt[1])

		parsedTime, err := time.Parse(time.RFC3339, createdTime)
		if err == nil {
			info.CreatedAt = parsedTime
		}
	}

	return &info
}
