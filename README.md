# SRTM Downloader
Library to assist in downloading SRTM data from https://dds.cr.usgs.gov/srtm/version2_1

> Work in progress

## Usage

Help
```
go run *.go download -help
```

Download SRTM3 data for Australia
```
go run *.go download -resolution SRTM3 -subdir Australia -o ~/Desktop/SRTM3/Australia/raw1
```
