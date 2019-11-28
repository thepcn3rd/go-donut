package main

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Binject/go-donut/donut"
	"github.com/akamensky/argparse"
)

func main() {

	parser := argparse.NewParser("go-donut", "Convert a VBS/JS or PE/.NET EXE/DLL to shellcode.\n\t\t"+
		"Only the finest artisanal donuts are made of shells.")

	// -MODULE OPTIONS-
	moduleName := parser.String("n", "module", &argparse.Options{Required: false,
		Help: "Module name. Randomly generated by default with entropy enabled."})
	url := parser.String("u", "url", &argparse.Options{Required: false,
		Help: "HTTP server that will host the donut module."})
	entropy := parser.Int("e", "entropy", &argparse.Options{Required: false,
		Help: "Entropy. 1=disable, 2=use random names, 3=random names + symmetric encryption (default)"})

	//  -PIC/SHELLCODE OPTIONS-
	archStr := parser.String("a", "arch", &argparse.Options{Required: false,
		Default: "x84", Help: "Target Architecture: x32, x64, or x84"})
	bypass := parser.Int("b", "bypass", &argparse.Options{Required: false,
		Default: 3, Help: "Bypass AMSI/WLDP : 1=skip, 2=abort on fail, 3=continue on fail."})
	dstFile := parser.String("o", "out", &argparse.Options{Required: false,
		Default: "loader.bin", Help: "Output file."})
	format := parser.Int("f", "format", &argparse.Options{Required: false,
		Default: 1, Help: "Output format. 1=raw, 2=base64, 3=c, 4=ruby, 5=python, 6=powershell, 7=C#, 8=hex"})
	oepString := parser.String("y", "oep", &argparse.Options{Required: false,
		Help: "Create a new thread for loader. Optionally execute original entrypoint of host process."})
	action := parser.Int("x", "exit", &argparse.Options{Required: false,
		Default: 1, Help: "Exiting. 1=exit thread, 2=exit process"})

	//  -FILE OPTIONS-
	className := parser.String("c", "class", &argparse.Options{Required: false,
		Help: "Optional class name.  (required for .NET DLL)"})
	appDomain := parser.String("d", "domain", &argparse.Options{Required: false,
		Help: "AppDomain name to create for .NET.  Randomly generated by default with entropy enabled."})
	method := parser.String("m", "method", &argparse.Options{Required: false,
		Help: "Optional method or API name for DLL. (a method is required for .NET DLL)"})
	params := parser.String("p", "params", &argparse.Options{Required: false,
		Help: "Optional parameters/command line inside quotations for DLL method/function or EXE."})
	wFlag := parser.Flag("w", "unicode", &argparse.Options{Required: false,
		Help: "Command line is passed to unmanaged DLL function in UNICODE format. (default is ANSI)"})
	runtime := parser.String("r", "runtime", &argparse.Options{Required: false,
		Help: "CLR runtime version."})
	tFlag := parser.Flag("t", "thread", &argparse.Options{Required: false,
		Help: "Create new thread for entrypoint of unmanaged EXE."})
	zFlag := parser.Int("z", "compress", &argparse.Options{Required: false, Default: 1,
		Help: "Pack/Compress file. 1=disable, 2=LZNT1, 3=Xpress, 4=Xpress Huffman"})

	// go-donut only flags
	dotNet := parser.Flag("", "dotnet", &argparse.Options{Required: false,
		Help: ".NET Mode, set true for .NET exe and DLL files (autodetect not implemented)"})
	srcFile := parser.String("i", "in", &argparse.Options{Required: true,
		Help: ".NET assembly, EXE, DLL, VBS, JS or XSL file to execute in-memory."})

	if err := parser.Parse(os.Args); err != nil || *srcFile == "" {
		log.Println(parser.Usage(err))
		return
	}

	var err error
	oep := uint64(0)
	if *oepString != "" {
		oep, err = strconv.ParseUint(*oepString, 16, 64)
		if err != nil {
			log.Println("Invalid OEP: " + err.Error())
			return
		}
	}

	var donutArch donut.DonutArch
	switch strings.ToLower(*archStr) {
	case "x32":
		donutArch = donut.X32
	case "x64":
		donutArch = donut.X64
	case "x84":
		donutArch = donut.X84
	default:
		log.Fatal("Unknown architecture provided")
	}

	config := new(donut.DonutConfig)
	config.Arch = donutArch
	config.Entropy = uint32(*entropy)
	config.OEP = oep

	if *url == "" {
		config.InstType = donut.DONUT_INSTANCE_PIC
	} else {
		config.InstType = donut.DONUT_INSTANCE_URL
	}

	config.DotNetMode = *dotNet
	config.Parameters = *params
	config.Runtime = *runtime
	config.URL = *url
	config.Class = *className
	config.Method = *method
	config.Domain = *appDomain
	config.Bypass = *bypass
	config.ModuleName = *moduleName
	config.Compress = uint32(*zFlag)
	config.Format = uint32(*format)

	if *tFlag {
		config.Thread = 1
	}
	if *wFlag { // convert command line to unicode? only applies to unmanaged DLL function
		config.Unicode = 1
	}
	config.ExitOpt = uint32(*action)

	if *srcFile == "" {
		if *url == "" {
			log.Fatal("No source URL or file provided")
		}
		payload, err := donut.ShellcodeFromURL(*url, config)
		if err == nil {
			err = ioutil.WriteFile(*dstFile, payload.Bytes(), 0644)
		}
	} else {
		payload, err := donut.ShellcodeFromFile(*srcFile, config)
		if err == nil {
			f, err := os.Create(*dstFile)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			if _, err = payload.WriteTo(f); err != nil {
				log.Fatal(err)
			}
		}
	}
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Done!")
	}
}
