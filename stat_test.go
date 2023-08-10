package main

import (
	_ "embed"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func Test_scan(t *testing.T) {
	t.Run("two method result same", func(t *testing.T) {
		for _, input := range sampleLines {
			oneStat, oneErr := scan(input)
			twoStat, twoErr := scanThroughReflection(input)
			if oneErr != nil || twoErr != nil {
				if oneErr == nil || twoErr == nil {
					t.Errorf("on input %s not err both one %v two %v", input, oneErr, twoErr)
				}
			}
			if !reflect.DeepEqual(oneStat, twoStat) {
				t.Errorf("on input %s not same output \none %+v\ntwo %+v", input, oneStat, twoStat)
			}
		}
	})
}

// The file comes from a sample of my arch-linux surface real /proc/pid/stat
//
//go:embed stat_line_samples.txt
var sampleText string
var sampleLines = strings.Split(sampleText, "\n")

func BenchmarkScanFast(b *testing.B) {
	for i := 0; i < b.N; i++ {
		input := sampleLines[i%len(sampleLines)]
		if _, err := scan(input); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkScanSlow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		input := sampleLines[i%len(sampleLines)]
		if _, err := scanThroughReflection(input); err != nil {
			b.Fatal(err)
		}
	}
}

func scanThroughReflection(s string) (*Stat, error) {
	var ret Stat
	value := reflect.ValueOf(&ret).Elem()
	tp := value.Type()
	rest := s + " X" // add one extra suffix so every part could be a before, not leaving the tail as rest.
	var part string
	var found bool
	for i := 0; i < value.NumField(); i++ {
		part, rest, found = strings.Cut(rest, " ")
		if !found {
			return nil, fmt.Errorf("fail to cut %d when scan on rest %s", i, rest)
		}
		field := value.Field(i)
		structField := tp.Field(i)
		specifier := fieldToSpecifier(structField)
		ptr := reflect.New(field.Type())
		if _, err := fmt.Sscanf(part, specifier, ptr.Interface()); err != nil {
			return nil, fmt.Errorf("parse %s as %s on [%d]%s: %v", part, specifier, i, structField.Name, err)
		}
		field.Set(ptr.Elem())
	}
	return &ret, nil
}
