// ref: https://golangcode.com/create-zip-files-in-go/
// ref: https://gist.github.com/maximilien/328c9ac19ab0a158a8df
package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

// func main() {
// 	var outfile string
// 	var files []string
// 	var verbose bool
// 	const suffix = "[.tar.gz|.zip]"

// 	flag.BoolVar(&verbose, "v", false, "Print more stuff!")
// 	flag.StringVar(&outfile, "o", "", "Output file with extension "+suffix)
// 	flag.Parse()
// 	files = flag.Args()

// 	if outfile == "" || (!strings.HasSuffix(outfile, ".tar.gz") && !strings.HasSuffix(outfile, ".zip")) {
// 		fmt.Println("Outfile [-o] required with suffix " + suffix)
// 		os.Exit(1)
// 	}

// 	if len(files) == 0 {
// 		fmt.Println("File list required")
// 		os.Exit(1)
// 	}

// 	os.MkdirAll(path.Dir(outfile), 0755)

// 	if strings.HasSuffix(outfile, ".zip") {
// 		if err := ZipFiles(outfile, files); err != nil {
// 			panic(err)
// 		}
// 		if verbose {
// 			fmt.Println("Zipped File:", outfile)
// 		}
// 	} else if strings.HasSuffix(outfile, ".tar.gz") {
// 		if err := CreateTarball(outfile, files); err != nil {
// 			panic(err)
// 		}
// 		if verbose {
// 			fmt.Println("Tar GZipped File:", outfile)
// 		}
// 	}

// }

// ZipFiles compresses one or many files into a single zip archive file.
// @Param 1: filename is the output zip file's name.
// @Param 2: files is a list of files to add to the zip.
func ZipFiles(filename string, files []string) error {

	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err = AddFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func AddFileToZip(zipWriter *zip.Writer, filename string) error {

	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	// header.Name = filename
	header.SetMode(0755)

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

func CreateTarball(tarballFilePath string, filePaths []string) error {
	file, err := os.Create(tarballFilePath)
	if err != nil {
		return fmt.Errorf("could not create tarball file '%s', got error '%s'", tarballFilePath, err.Error())
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	for _, filePath := range filePaths {
		err := addFileToTarWriter(filePath, tarWriter)
		if err != nil {
			return fmt.Errorf("could not add file '%s', to tarball, got error '%s'", filePath, err.Error())
		}
	}

	return nil
}

func addFileToTarWriter(filePath string, tarWriter *tar.Writer) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("could not open file '%s', got error '%s'", filePath, err.Error())
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("could not get stat for file '%s', got error '%s'", filePath, err.Error())
	}

	header, err := tar.FileInfoHeader(stat, "")
	if err != nil {
		return fmt.Errorf("could not create header for file '%s', got error '%s'", filePath, err.Error())
	}

	header.Mode = 0755
	err = tarWriter.WriteHeader(header)
	if err != nil {
		return fmt.Errorf("could not write header for file '%s', got error '%s'", filePath, err.Error())
	}

	_, err = io.Copy(tarWriter, file)
	if err != nil {
		return fmt.Errorf("could not copy the file '%s' data to the tarball, got error '%s'", filePath, err.Error())
	}

	return nil
}
