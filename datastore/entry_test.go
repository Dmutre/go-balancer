package datastore

import (
	"bufio"
	"bytes"
	"testing"
)

func TestEntry_Encode(t *testing.T) {
	// Test StringType
	encoder := NewEntry("tK", StringType, "tV")
	data := encoder.Encode()
	decoder := &Entry{}
	err := decoder.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if decoder.GetLength() != encoder.GetLength() {
		t.Error("Incorrect length")
	}
	if decoder.key != "tK" {
		t.Error("Incorrect key")
	}
	if decoder.value != "tV" {
		t.Error("Incorrect value")
	}

	// Test Int64Type
	encoder = NewEntry("tK", Int64Type, int64(12345))
	data = encoder.Encode()
	decoder = &Entry{}
	err = decoder.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if decoder.GetLength() != encoder.GetLength() {
		t.Error("Incorrect length")
	}
	if decoder.key != "tK" {
		t.Error("Incorrect key")
	}
	if decoder.value != int64(12345) {
		t.Error("Incorrect value")
	}
}

func TestReadValue(t *testing.T) {
	// Test StringType
	encoder := NewEntry("tK", StringType, "tV")
	data := encoder.Encode()
	readData := bytes.NewReader(data)
	bReadData := bufio.NewReader(readData)
	valueType, value, err := readValue(bReadData)
	if err != nil {
		t.Fatal(err)
	}
	if valueType != StringType {
		t.Errorf("Wrong value type: [%d]", valueType)
	}
	if value != "tV" {
		t.Errorf("Wrong value: [%s]", value)
	}

	// Test Int64Type
	encoder = NewEntry("tK", Int64Type, int64(12345))
	data = encoder.Encode()
	readData = bytes.NewReader(data)
	bReadData = bufio.NewReader(readData)
	valueType, value, err = readValue(bReadData)
	if err != nil {
		t.Fatal(err)
	}
	if valueType != Int64Type {
		t.Errorf("Wrong value type: [%d]", valueType)
	}
	if value != int64(12345) {
		t.Errorf("Wrong value: [%d]", value)
	}
}