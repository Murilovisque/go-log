package compressor

import (
	"archive/zip"
	"io"
	"os"
	"path"
)

func ComprimirArquivo(nomeArquivoParaComprimir string) error {
	nomeArquivoComprimido := nomeArquivoParaComprimir + ".zip"
	arquivoZip, err := os.Create(nomeArquivoComprimido)
	if err != nil {
		return err
	}
	defer arquivoZip.Close()
	arquivoZipWriter := zip.NewWriter(arquivoZip)
	defer arquivoZipWriter.Close()
	nomeBaseArquivoParaComprimir := path.Base(nomeArquivoParaComprimir)
	arquivoParaZiparWriter, err := arquivoZipWriter.Create(nomeBaseArquivoParaComprimir)
	if err != nil {
		os.Remove(nomeArquivoComprimido)
		return err
	}
	arquivoParaZipar, err := os.Open(nomeArquivoParaComprimir)
	if err != nil {
		os.Remove(nomeArquivoComprimido)
		return err
	}
	defer arquivoParaZipar.Close()
	_, err = io.Copy(arquivoParaZiparWriter, arquivoParaZipar)
	return err
}
