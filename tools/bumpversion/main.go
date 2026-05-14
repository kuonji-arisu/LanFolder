package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

type replacement struct {
	path    string
	pattern string
	replace string
}

func main() {
	args := os.Args[1:]
	if len(args) > 1 {
		fatalf("usage: go run ./tools/bumpversion [version]")
	}

	version := ""
	printOnly := len(args) == 1 && args[0] == "--print"
	if printOnly {
		data, err := os.ReadFile("VERSION")
		if err != nil {
			fatalf("%v", err)
		}
		version = strings.TrimSpace(string(data))
	} else if len(args) == 1 {
		version = args[0]
	} else {
		data, err := os.ReadFile("VERSION")
		if err != nil {
			fatalf("%v", err)
		}
		version = strings.TrimSpace(string(data))
	}
	if !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(version) {
		fatalf("version must use numeric semver format like 0.0.2")
	}
	if printOnly {
		fmt.Println(version)
		return
	}
	if len(args) == 1 {
		if err := os.WriteFile("VERSION", []byte(version+"\n"), 0644); err != nil {
			fatalf("%v", err)
		}
	}

	version4 := version + ".0"
	replacements := []replacement{
		{"build/config.yml", `(version: ")\d+\.\d+\.\d+(" # The application version)`, "${1}" + version + "${2}"},
		{"build/windows/wails.exe.manifest", `(version=")\d+\.\d+\.\d+(")`, "${1}" + version + "${2}"},
		{"build/windows/info.json", `("file_version": ")\d+\.\d+\.\d+(")`, "${1}" + version + "${2}"},
		{"build/windows/info.json", `("ProductVersion": ")\d+\.\d+\.\d+(")`, "${1}" + version + "${2}"},
		{"build/windows/nsis/wails_tools.nsh", `(!define INFO_PRODUCTVERSION ")\d+\.\d+\.\d+(")`, "${1}" + version + "${2}"},
		{"build/windows/nsis/project.nsi", `(# Default ")\d+\.\d+\.\d+(")`, "${1}" + version + "${2}"},
		{"build/windows/msix/template.xml", `(?m)(^\s+Version=")\d+\.\d+\.\d+\.0"\s*$`, "${1}" + version4 + "\""},
		{"build/windows/msix/app_manifest.xml", `(?m)(^\s+Version=")\d+\.\d+\.\d+\.0"\s*$`, "${1}" + version4 + "\""},
		{"build/linux/nfpm/nfpm.yaml", `(version: ")\d+\.\d+\.\d+(")`, "${1}" + version + "${2}"},
		{"build/darwin/Info.plist", `(<key>CFBundleVersion</key>\s*<string>)\d+\.\d+\.\d+(</string>)`, "${1}" + version + "${2}"},
		{"build/darwin/Info.plist", `(<key>CFBundleShortVersionString</key>\s*<string>)\d+\.\d+\.\d+(</string>)`, "${1}" + version + "${2}"},
		{"build/darwin/Info.dev.plist", `(<key>CFBundleVersion</key>\s*<string>)\d+\.\d+\.\d+(</string>)`, "${1}" + version + "${2}"},
		{"build/darwin/Info.dev.plist", `(<key>CFBundleShortVersionString</key>\s*<string>)\d+\.\d+\.\d+(</string>)`, "${1}" + version + "${2}"},
	}

	for _, item := range replacements {
		if err := replaceInFile(item); err != nil {
			fatalf("%v", err)
		}
	}
	fmt.Printf("Synced release metadata to %s\n", version)
}

func replaceInFile(item replacement) error {
	data, err := os.ReadFile(item.path)
	if err != nil {
		return err
	}
	re, err := regexp.Compile(item.pattern)
	if err != nil {
		return err
	}
	if !re.Match(data) {
		return fmt.Errorf("version pattern not found in %s", item.path)
	}
	next := re.ReplaceAll(data, []byte(item.replace))
	return os.WriteFile(item.path, next, 0644)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
