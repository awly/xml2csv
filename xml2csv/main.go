package main

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage:", os.Args[0], "csv-filename xml-filenames")
		return
	}

	data, err := unmarshalAll(os.Args[2:])
	if err != nil {
		fmt.Println(err)
		return
	}

	fout, err := os.Create(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fout.Close()
	out := csv.NewWriter(fout)

	if err = out.Write(append([]string{""}, os.Args[2:len(os.Args)]...)); err != nil {
		fmt.Println(err)
		return
	}
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if err = out.Write(append([]string{k}, data[k]...)); err != nil {
			fmt.Println(err)
			return
		}
	}
	out.Flush()
}

func unmarshalAll(files []string) (map[string][]string, error) {
	ch := make(chan record, len(files))
	wg := &sync.WaitGroup{}

	for i, fname := range files {
		fin, err := os.Open(fname)
		if err != nil {
			return nil, err
		}
		defer fin.Close()

		wg.Add(1)
		go unmarshal(i, fin, ch, wg)
	}

	resch := make(chan map[string][]string)
	go func() {
		resch <- build(ch, len(files))
	}()
	wg.Wait()
	close(ch)
	return <-resch, nil
}

func unmarshal(pos int, in io.Reader, out chan<- record, wg *sync.WaitGroup) {
	defer wg.Done()
	dec := xml.NewDecoder(in)
	d := &data{}
	err := dec.Decode(d)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, r := range d.Records {
		r.pos = pos
		out <- r
	}
}

func build(in <-chan record, l int) map[string][]string {
	res := make(map[string][]string)
	for r := range in {
		if _, ok := res[r.Name]; !ok {
			res[r.Name] = make([]string, l)
		}
		res[r.Name][r.pos] = r.Val
	}
	return res
}

type data struct {
	Records []record `xml:"string"`
}
type record struct {
	Name string `xml:"name,attr"`
	Val  string `xml:",chardata"`
	pos  int
}
