package compressor

import (
	"archive/zip"
	"io"
	"os"
	"path"
)

const (
	ZipExtension      = ".zip"
	ZipExtensionRegex = "\\.zip"
)

func CompressFile(fileNametoCompress string) error {
	compressedName := fileNametoCompress + ZipExtension
	compressFile, err := os.Create(compressedName)
	if err != nil {
		return err
	}
	defer compressFile.Close()
	zipWriter := zip.NewWriter(compressFile)
	defer zipWriter.Close()
	baseCompressedName := path.Base(fileNametoCompress)
	fileZipWriter, err := zipWriter.Create(baseCompressedName)
	if err != nil {
		os.Remove(compressedName)
		return err
	}
	fileToZip, err := os.Open(fileNametoCompress)
	if err != nil {
		os.Remove(compressedName)
		return err
	}
	defer fileToZip.Close()
	_, err = io.Copy(fileZipWriter, fileToZip)
	return err
}
