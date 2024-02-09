package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cortexproject/cortex/pkg/alertmanager/alertspb"
	"github.com/gogo/protobuf/proto"
	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	"github.com/prometheus/alertmanager/nflog/nflogpb"
	"github.com/prometheus/alertmanager/silence/silencepb"
	"io"
	"log"
	"os"
	"strings"
)

var (
	out bytes.Buffer
	outputFile string
	intputFile string
)

func main() {
	flag.CommandLine.StringVar(&intputFile, "file", "", "path of fullstate file to deserialize")
	flag.CommandLine.StringVar(&outputFile, "output", "", "file for deserialized output")

	flag.Parse()

	if intputFile == "" {
		log.Fatalf("No full state file specified.")
	}

	decodeFullState(intputFile)

	outputToFile(outputFile)
}

func decodeFullState(path string) {
	in, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	fs := alertspb.FullStateDesc{}
	err = proto.Unmarshal(in, &fs)
	if err != nil {
		log.Fatalln("Error unmarshalling full state:", err)
	}

	for _, part := range fs.GetState().Parts {
		out.WriteString("\n----\n")
		if isNfLog(part.Key) {
			parseNotifications(part.Data)
		} else if isSilence(part.Key) {
			parseSilences(part.Data)
		} else {
			out.WriteString(fmt.Sprintf("Unknown part type: %s", part.Key))
		}

	}
}

func parseNotifications(data []byte) {
	out.WriteString("Alerts:\n")
	r := bytes.NewReader(data)
	for {
		nf := nflogpb.MeshEntry{}
		n, err := pbutil.ReadDelimited(r, &nf)
		if err != nil && err != io.EOF {
			log.Fatalf("unable to read alert notifications, %v", err)
		}
		if n == 0 || err == io.EOF {
			break
		}
		result, err := json.Marshal(nf)
		if err != nil {
			log.Fatalf("unable to marshal to json, %v", err)
		}
		_, err = out.WriteString(string(result) + "\n")
		if err != nil {
			log.Fatalf("unable to write output, %v", err)
		}
	}
}

func parseSilences(data []byte) {
	out.WriteString("Silences:\n")
	r := bytes.NewReader(data)
	for {
		silence := silencepb.MeshSilence{}
		n, err := pbutil.ReadDelimited(r, &silence)
		if err != nil && err != io.EOF {
			log.Fatalf("unable to read silences, %v", err)
		}
		if n == 0 || err == io.EOF {
			break
		}

		result, err := json.Marshal(silence)
		if err != nil {
			log.Fatalf("unable to marshal to json, %v", err)
		}

		_, err = out.WriteString(string(result) + "\n")
		if err != nil {
			log.Fatalf("unable to write output, %v", err)
		}
	}
}

func outputToFile(file string) {
	if outputFile == "" {
		// write to stdout if output file not specified
		fmt.Print(out.String())
		return
	} 
	fo, err := os.Create(file)
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	_, err = fo.Write(out.Bytes())
	if err != nil {
		log.Fatalf("Failed writing output to file: %v", err)
	}
}

func isNfLog(key string) bool {
	return strings.HasPrefix(key, "nfl")
}

func isSilence(key string) bool {
	return strings.HasPrefix(key, "sil")
}
