package main

import (
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/urfave/cli"
)

const baseurl = "https://dds.cr.usgs.gov/srtm/version2_1"

type downloadTask struct {
	uri        string
	outputPath string
}

func main() {
	var resolution string
	var subdir string
	var outputPath string

	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name:  "download",
			Usage: "Download SRTM data",
			Action: func(c *cli.Context) error {
				downloadAsync(baseurl, resolution, subdir, outputPath)
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "resolution, res",
					Value:       "SRTM3",
					Usage:       "SRTM resolution to download",
					Destination: &resolution,
				},
				cli.StringFlag{
					Name:        "subdir",
					Value:       "Australia",
					Usage:       "Subdirectory to download",
					Destination: &subdir,
				},
				cli.StringFlag{
					Name:        "output, o",
					Value:       "/tmp",
					Usage:       "Directory to download HGT files to",
					Destination: &outputPath,
				},
			},
		},
	}

	app.Run(os.Args)
}

func downloadAsync(baseurl string, resolution string, subdir string, outputPath string) error {
	var wg sync.WaitGroup

	concurrency := 20
	list := getList(baseurl, resolution, subdir)
	downloadTo := path.Join(outputPath, "_raw")

	fmt.Printf("Files to download: %d\n", len(list))

	var tasks []downloadTask
	for _, uri := range list {
		tasks = append(tasks, downloadTask{
			uri:        uri,
			outputPath: downloadTo,
		})
	}

	// create the output Directory
	err := os.MkdirAll(downloadTo, os.FileMode(0755))
	if err != nil {
		panic(err)
	}

	wg.Add(concurrency)

	go pool(&wg, concurrency, tasks, outputPath)

	wg.Wait()

	return nil
}

func pool(wg *sync.WaitGroup, workers int, tasks []downloadTask, outputPath string) {

	downloadTasksCh := make(chan downloadTask)
	unzipTasksCh := make(chan string)

	for i := 0; i < workers; i++ {
		go downloadWorker(downloadTasksCh, unzipTasksCh)
		go unzipWorker(unzipTasksCh, wg, outputPath)
	}

	for _, task := range tasks {
		downloadTasksCh <- task
	}

	// indicates that no more jobs will be sent
	close(downloadTasksCh)
	close(unzipTasksCh)
}

func downloadWorker(downloadTasksCh <-chan downloadTask, unzipTasksCh chan<- string) {
	// defer wg.Done()
	for {
		task, ok := <-downloadTasksCh

		if !ok {
			return
		}
		// d := time.Duration(task) * time.Millisecond
		// time.Sleep(100)
		// fmt.Printf("Downloading: %s -> %s\n", task.uri, task.outputPath)
		filepath, err := downloadHGT(task.uri, task.outputPath)

		if err != nil {
			fmt.Println("Error while downloading", task.uri, "-", err)
			panic(err)
		}

		fmt.Println("Successfully downloaded", filepath)
		unzipTasksCh <- filepath
	}
}

func unzipWorker(unzipTasksCh <-chan string, wg *sync.WaitGroup, outputPath string) {
	defer wg.Done()

	for {
		src, ok := <-unzipTasksCh

		if !ok {
			return
		}

		fmt.Printf("Unzipping %s\n", src)
		unzip(src, outputPath)
	}
}
