package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/qedus/osmpbf"
)

func generateWayid2nodeidsMapping(input, output string) {
	infile, err := os.Open(input)
	defer infile.Close()
	if err != nil {
		log.Fatal(err)
		fmt.Printf("Open pbf file of %v failed.\n", input)
		return
	}
	fmt.Printf("Open pbf file of %s succeed.\n", input)

	outfile, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE, 0755)
	defer outfile.Close()
	defer outfile.Sync()
	if err != nil {
		log.Fatal(err)
		fmt.Printf("Open output file of %s failed.\n", output)
		return
	}
	fmt.Printf("Open output file of %s succeed.\n", output)

	wayid2nodeids(infile, outfile)
}

func wayid2nodeids(infile io.Reader, outfile io.Writer) {
	// Init extractor
	extractor := osmpbf.NewDecoder(infile)
	extractor.SetBufferSize(osmpbf.MaxBlobSize)
	err := extractor.Start(runtime.GOMAXPROCS(-1))
	if err != nil {
		log.Fatal(err)
		return
	}

	// Init loader
	loader := bufio.NewWriter(outfile)
	defer loader.Flush()

	var wc, nc uint32
	for {
		if v, err := extractor.Decode(); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		} else {
			switch v := v.(type) {
			case *osmpbf.Node:
			case *osmpbf.Way:
				// Transform
				str := covertWayObj2IdMappingString(v)
				//str := covertWayObj2MockSpeed(v)

				_, err := loader.WriteString(str)
				if err != nil {
					log.Fatal(err)
					return
				}

				wc++
				nc += (uint32)(len(v.NodeIDs))
			case *osmpbf.Relation:
			default:
				log.Fatalf("unknown type %T\n", v)
			}
		}
	}

	fmt.Printf("Total ways: %d, total nodes: %d\n", wc, nc)
}

func covertWayObj2IdMappingString(v *osmpbf.Way) string {
	// format: wayid,nodeid1,nodeid2, ...
	return strconv.FormatUint((uint64)(v.ID), 10) + "," +
		strings.Trim(strings.Join(strings.Fields(fmt.Sprint(v.NodeIDs)), ","), "[]") +
		"\n"
}

func covertWayObj2MockSpeed(v *osmpbf.Way) string {
	// format: wayid,random speed
	return strconv.FormatUint((uint64)(v.ID), 10) +
		"," +
		strconv.Itoa(rand.Intn(100)) +
		"\n"
}

var flags struct {
	input  string
	output string
}

func init() {
	flag.StringVar(&flags.input, "i", "", "Input pbf file.")
	flag.StringVar(&flags.output, "o", "", "Output csv file")
}

func main() {
	flag.Parse()

	if len(flags.input) == 0 || len(flags.output) == 0 {
		fmt.Printf("[ERROR]Input or Output file path is empty.\n")
		return
	}

	startTime := time.Now()
	generateWayid2nodeidsMapping(flags.input, flags.output)
	endTime := time.Now()
	fmt.Printf("Total processing time for wayid2nodeids-extract takes %f seconds\n", endTime.Sub(startTime).Seconds())
}
