package usecase

import (
  "io/ioutil"
  "testing"
)

const (
  validDataPath = "../testdata/test.jpeg"
  invalidDataPath = "../testdata/test.txt"
  widthResizeRatio = 0.8
  heightResizeRatio = 0.8
  expectedFormat = "jpeg"
	invalidFormat = "invalid format"
)

func readTestData(t *testing.T, filepath string) []byte {
  t.Helper()
  data, err := ioutil.ReadFile(filepath)
  if err != nil {
    t.Fatalf("テストデータの読み込みに失敗しました: %v", err)
  }
  return data
}

func setupTestData(t *testing.T) (validData, invalidData []byte) {
	t.Helper()
	validData = readTestData(t, validDataPath)
	invalidData = readTestData(t, invalidDataPath)
	return
}

func TestCheckImageFormat(t *testing.T) {
  validData, invalidData := setupTestData(t)
  tests := map[string]struct {
    data    []byte
    want    string
    wantErr bool
  }{
    "valid case":   {validData, expectedFormat, false},
    "invalid case": {invalidData, "", true},
  }
  for testName, tt := range tests {
    t.Run(testName, func(t *testing.T) {
      got, err := CheckImageFormat(tt.data)
      if (err != nil) != tt.wantErr {
        t.Errorf("CheckImageFormat() error = %v, wantErr = %v", err, tt.wantErr)
      } else if got != tt.want {
        t.Errorf("CheckImageFormat() = %v, want = %v", got, tt.want)
      }
    })
  }
}

func TestResizeImage(t *testing.T) {
  validData, _ := setupTestData(t)
  tests := map[string]struct {
    format  string
    wantErr bool
  }{
    "valid case":   {expectedFormat, false},
    "invalid case": {invalidFormat, true},
  }
  for testName, tt := range tests {
    t.Run(testName, func(t *testing.T) {
      _, err := ResizeImage(validData, widthResizeRatio, heightResizeRatio, tt.format)
      if (err != nil) != tt.wantErr {
        t.Errorf("ResizeImage() error = %v, wantErr = %v", err, tt.wantErr)
      }
    })
  }
}
