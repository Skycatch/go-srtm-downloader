package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// GetList Retrieve the list of srtm data to download
func getList(baseurl string, resolution string, subdir string) []string {
	var list []string

	// fetch the list of srtm files
	url := baseurl + "/" + resolution + "/" + subdir

	resp, err := http.Get(url)

	if err != nil {
		panic(err)
	}
	root, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

	files := scrape.FindAll(root, scrape.ByTag(atom.A))

	for _, file := range files {
		fileurl := url + "/" + scrape.Attr(file, "href")

		if strings.Contains(fileurl, ".zip") {
			// fmt.Printf("%2d %s (%s)\n", i, scrape.Text(file), fileurl)
			list = append(list, fileurl)
		}
	}

	return list
}

func downloadHGT(uri string, outputDir string) (string, error) {
	parts := strings.Split(uri, "/")
	fileName := parts[len(parts)-1]
	filePath := path.Join(outputDir, fileName)

	fmt.Printf("Downloading %s -> %s\n", uri, filePath)

	output, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error while creating", filePath, "-", err)
		return "", err
	}
	defer output.Close()

	response, err := http.Get(uri)
	if err != nil {
		fmt.Println("Error while downloading", uri, "-", err)
		return "", err
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", uri, "-", err)
		return "", err
	}

	fmt.Println(n, "bytes downloaded.")

	return filePath, nil
}
