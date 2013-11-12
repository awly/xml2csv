package main

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sync"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("usage:", os.Args[0], "csv-filename xml-filenames")
		return
	}

	ch := make(chan record, len(os.Args)-2)
	wg := &sync.WaitGroup{}

	for i, fname := range os.Args[2:] {
		fin, err := os.Open(fname)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer fin.Close()

		go unmarshal(i, fin, ch, wg)
	}

	resch := make(chan map[string][]string)
	go func() {
		resch<- build(ch, len(ch))
	}
	wg.Wait()
	close(ch)
	data := <-resch

	fout, err := os.Create(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fout.Close()
	out := csv.NewWriter(fout)

	data := data{}
	err = xml.NewDecoder(fin).Decode(&data)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err = out.Write(append([]string{""}, os.Args[1:len(os.Args)-1]...)); err != nil {
		fmt.Println(err)
		return
	}
	for _, r := range data.Records {
		if err = out.Write([]string{r.Name, r.Val}); err != nil {
			fmt.Println(err)
			return
		}
	}
	out.Flush()
}

func unmarshalAll(files []string) (map[string][]string)

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
