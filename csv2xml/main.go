package main

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage:", os.Args[0], "input.csv")
		return
	}

	inf, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	defer inf.Close()
	in := csv.NewReader(inf)
	recs, err := in.Read()
	if err != nil {
		fmt.Println(err)
		return
	}
	outf := make([]*os.File, len(recs)-1)
	for i := range outf {
		outf[i], err = os.Create(recs[i+1])
		if err != nil {
			fmt.Println(err)
			return
		}
		defer outf[i].Close()
	}

	data := make([]resources, len(outf))
	for i := range data {
		data[i].Records = make([]record, 0)
	}
	for recs, err = in.Read(); err == nil; recs, err = in.Read() {
		for i := range data {
			data[i].Records = append(data[i].Records, record{Name: recs[0], Val: recs[i+1]})
		}
	}
	for i, f := range outf {
		enc := xml.NewEncoder(f)
		enc.Indent("", "\t")
		err = enc.Encode(data[i])
		if err != nil {
			fmt.Println(err)
		}
		enc.Flush()
	}
}

type resources struct {
	Records []record `xml:"string"`
}
type record struct {
	Name string `xml:"name,attr"`
	Val  string `xml:",chardata"`
}
