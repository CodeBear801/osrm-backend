package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var flags struct {
	idsmapping  string
	mocktraffic string
	output      string
}

func init() {
	flag.StringVar(&flags.idsmapping, "i", "", "Input id mapping file.")
	flag.StringVar(&flags.mocktraffic, "m", "", "Mock traffic file")
	flag.StringVar(&flags.output, "o", "", "Output csv file")
}

// todo: Integrate with Jay's code
//       Statistic to avoid unmatched element
//       Write data into more compressed format(parquet)
//       Multiple go routine for convert()

func main() {
	flag.Parse()

	if len(flags.idsmapping) == 0 || len(flags.mocktraffic) == 0 || len(flags.output) == 0 {
		fmt.Printf("[ERROR]Input or Output file path is empty.\n")
		return
	}

	wayid2speed := make(map[uint64]int)
	loadMockTraffic(flags.mocktraffic, wayid2speed)

	startTime := time.Now()

	// format is: wayid, nodeid, nodeid, nodeid...
	source := make(chan string)
	// format is fromid, toid, speed
	sink := make(chan string)

	go load(flags.idsmapping, source)
	go convert(source, sink, wayid2speed)
	write(flags.output, sink)

	endTime := time.Now()
	fmt.Printf("Total processing time for wayid2nodeids-extract takes %f seconds\n", endTime.Sub(startTime).Seconds())
}

func load(mappingPath string, source chan<- string) {
	defer close(source)

	f, err := os.Open(mappingPath)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
		fmt.Printf("Open idsmapping file of %v failed.\n", mappingPath)
		return
	}
	fmt.Printf("Open idsmapping file of %s succeed.\n", mappingPath)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		source <- (scanner.Text())
	}
}

func convert(source <-chan string, sink chan<- string, wayid2speed map[uint64]int) {
	var err error
	defer close(sink)

	for str := range source {
		elements := strings.Split(str, ",")
		if len(elements) < 3 {
			fmt.Printf("Invalid string %s in %s\n", str, flags.idsmapping)
			continue
		}

		var wayid uint64
		if wayid, err = strconv.ParseUint(elements[0], 10, 64); err != nil {
			fmt.Printf("#Error during decoding wayid, row = %v\n", elements)
			continue
		}

		if speed, ok := wayid2speed[wayid]; ok {
			var nodes []string = elements[1:]
			for i := 0; (i + 1) < len(nodes); i++ {
				var n1, n2 uint64
				if n1, err = strconv.ParseUint(nodes[i], 10, 64); err != nil {
					fmt.Printf("#Error during decoding nodeid, row = %v\n", elements)
					continue
				}
				if n2, err = strconv.ParseUint(nodes[i+1], 10, 64); err != nil {
					fmt.Printf("#Error during decoding nodeid, row = %v\n", elements)
					continue
				}
				s := fmt.Sprintf("%d,%d,%d\n", n1, n2, speed)
				sink <- s
			}
		} else {
			fmt.Printf("Invalid wayid(%d) from traffic component.\n", wayid)
		}
	}
}

func write(targetPath string, sink chan string) {
	outfile, err := os.OpenFile(targetPath, os.O_RDWR|os.O_CREATE, 0755)
	defer outfile.Close()
	defer outfile.Sync()
	if err != nil {
		log.Fatal(err)
		fmt.Printf("Open output file of %s failed.\n", targetPath)
		return
	}
	fmt.Printf("Open output file of %s succeed.\n", targetPath)

	w := bufio.NewWriter(outfile)
	defer w.Flush()
	for str := range sink {
		_, err := w.WriteString(str)
		if err != nil {
			log.Fatal(err)
			return
		}
	}
}

func loadMockTraffic(trafficPath string, wayid2speed map[uint64]int) {
	// load mock traffic file
	mockfile, err := os.Open(trafficPath)
	defer mockfile.Close()
	if err != nil {
		log.Fatal(err)
		fmt.Printf("Open pbf file of %v failed.\n", trafficPath)
		return
	}
	fmt.Printf("Open pbf file of %s succeed.\n", trafficPath)

	csvr := csv.NewReader(mockfile)
	for {
		row, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			} else {
				fmt.Printf("Error during decoding mock traffic, row = %v\n", err)
				return
			}
		}

		var wayid uint64
		var speed int64
		if wayid, err = strconv.ParseUint(row[0], 10, 64); err != nil {
			fmt.Printf("#Error during decoding wayid, row = %v\n", row)
		}
		if speed, err = strconv.ParseInt(row[1], 10, 32); err != nil {
			fmt.Printf("#Error during decoding speed, row = %v\n", row)
		}

		wayid2speed[wayid] = (int)(speed)
	}
}
