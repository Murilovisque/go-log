package compressor

import (
	"archive/zip"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestComprimirArquivoDeveZip(t *testing.T) {
	dir, err := ioutil.TempDir("", "teste-logs")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	nomeArqLog := "teste.log"
	caminhoCompletoArqLog := path.Join(dir, nomeArqLog)
	conteudoArq := []byte("conteúdo")
	err = ioutil.WriteFile(caminhoCompletoArqLog, conteudoArq, 0644)
	if err != nil {
		t.Fatal(err)
	}
	ComprimirArquivo(caminhoCompletoArqLog)

	caminhoCompletoArqLogComprimido := caminhoCompletoArqLog + ".zip"
	zipReader, err := zip.OpenReader(caminhoCompletoArqLogComprimido)
	if err != nil {
		t.Fatal(err)
	}
	if len(zipReader.File) != 1 {
		t.Fatalf("Should be 1 but is %d", len(zipReader.File))
	}
	zippedFile, err := zipReader.Open(nomeArqLog)
	if err != nil {
		t.Fatal(err)
	}
	conteudoZippedFile, err := ioutil.ReadAll(zippedFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(conteudoZippedFile) != string(conteudoArq) {
		t.Fatalf("Should be %s but is %s", string(conteudoArq), string(conteudoZippedFile))
	}
}
