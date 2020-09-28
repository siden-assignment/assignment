package store

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"
)

var (
	testDir    = "./test"
	testPrefix = "testing"

	basic = `hello
woot
woot
nope
`

	basicResult = `hello
nope
woot
`
)

func TestBadgerBasic(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll(testDir)
	})
	badger, err := New(BADGER, testDir)
	if err != nil {
		t.Fatal(err)
	}
	if badger == nil {
		t.Fatal("New returned a nil instace of the badger based store")
	}

	ctx := context.Background()
	input := bytes.NewBufferString(basic)
	if err := badger.Write(ctx, testPrefix, input); err != nil {
		t.Fatal(err, "failed to write basic input to store")
	}

	buff := make([]byte, 0, len([]byte(basicResult)))
	output := bytes.NewBuffer(buff)
	if err := badger.Read(ctx, testPrefix, output); err != nil {
		t.Fatal(err, "failed to read basic input out of store")
	}

	out := output.String()
	if out != basicResult {
		t.Fatal("failed to read data properly out of store, got:", out)
	}
}

func TestBadgerDifferentPrefix(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll(testDir)
	})
	badger, err := New(BADGER, testDir)
	if err != nil {
		t.Fatal(err)
	}
	if badger == nil {
		t.Fatal("New returned a nil instace of the badger based store")
	}

	ctx := context.Background()
	input := bytes.NewBufferString(basic)
	if err := badger.Write(ctx, testPrefix, input); err != nil {
		t.Fatal(err, "failed to write basic input to store")
	}

	buff := make([]byte, 0, len([]byte("")))
	output := bytes.NewBuffer(buff)
	if err := badger.Read(ctx, "differentPrefix", output); err != nil {
		t.Fatal(err, "failed to read basic input out of store")
	}

	out := output.String()
	if out != "" {
		t.Fatal("failed to read data properly out of store, got:", out)
	}
}

func TestFull(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll(testDir)
	})

	badger, err := New(BADGER, testDir)
	if err != nil {
		t.Fatal(err)
	}
	if badger == nil {
		t.Fatal("New returned a nil instace of the badger based store")
	}

	ctx := context.Background()
	input, err := os.Open("../../dist/test/input-test-file.txt")
	if err != nil {
		t.Fatal("failed to open test file", err)
	}

	expectedFile, err := os.Open("../../dist/test/output-test-file.txt")
	if err != nil {
		t.Fatal("Failed to open test file", err)
	}

	stats, err := expectedFile.Stat()
	if err != nil {
		t.Fatal("failed to retrieve file stats on output file", err)
	}

	if err := badger.Write(ctx, testPrefix, input); err != nil {
		t.Fatal("Failed to write file to badeger store", err)
	}

	buff := make([]byte, 0, stats.Size())
	output := bytes.NewBuffer(buff)
	if err := badger.Read(ctx, testPrefix, output); err != nil {
		t.Fatal("Failed to read data properly out of store", err)
	}

	expected, err := ioutil.ReadAll(expectedFile)
	if err != nil {
		t.Fatal("failed to properly read in expected output", err)
	}

	out := output.String()
	if out != string(expected) {
		t.Fatal("Failed to read data properly out of store", len(out), len(string(expected)))
	}
}
