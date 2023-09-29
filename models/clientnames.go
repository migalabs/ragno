package models

import (
	"strings"
)

type ClientName string
type ClientOS string
type ClientArch string
type ClientLanguage string

func CustomToString[T ClientName | ClientOS | ClientArch | ClientLanguage](s T) string {
	return string(s)
}

var (
	// GoblaUnknown
	Unknown string = "unknown"

	// avail Clients
	Geth         ClientName = "geth"
	Erigon       ClientName = "erigon"
	Reth         ClientName = "reth"
	Nethermind   ClientName = "nethermind"
	Besu         ClientName = "besu"
	OpenEthereum ClientName = "open-ethereum"
	Parity       ClientName = "parity"
	Reosc        ClientName = "reosc"
	EthereumJS   ClientName = "ethereum-js"
	NimbusEth1   ClientName = "nimbus-eth1"

	// avail OSs
	LinuxOS   ClientOS = "linux"
	WindowsOS ClientOS = "windows"
	MaxOS     ClientOS = "mac"
	FreeBsdOS ClientOS = "free-bsd"

	// Client Archs
	Amd64Arch ClientArch = "amd64"
	X86Arch   ClientArch = "x86"
	ArmArch   ClientArch = "arm"

	// Client Languages
	GoLanguage         ClientLanguage = "go"
	RustLanguage       ClientLanguage = "rust"
	JavaLanguage       ClientLanguage = "java"
	JavaScriptLanguage ClientLanguage = "js"
	DotnetLanguage     ClientLanguage = "dotnet"
	NimLanguage        ClientLanguage = "nim"
)

// Valid strings
var ValidClientNames = map[ClientName][]string{
	Geth:         {"geth", "go-ethereum"},
	Reth:         {"reth"},
	Erigon:       {"erigon"},
	Nethermind:   {"nethermind"},
	Besu:         {"besu"},
	OpenEthereum: {"openethereum"},
	Parity:       {"parity"},
	EthereumJS:   {"ethereum-js"},
	NimbusEth1:   {"nimbus", "nim"},
}

var ValidOSs = map[ClientOS][]string{
	LinuxOS:   {"linux", "ubuntu"},
	WindowsOS: {"win", "windows"},
	MaxOS:     {"macos", "osx"},
	FreeBsdOS: {"free", "bsd"},
}

var ValidArchs = map[ClientArch][]string{
	Amd64Arch: {"amd64", "x64"},
	X86Arch:   {"x86_64", "86"},
	ArmArch:   {"arm", "arm64"},
}

var ValidLanguages = map[ClientLanguage][]string{
	GoLanguage:         {"go"},
	RustLanguage:       {"rust", "reth"},
	JavaLanguage:       {"java", "arm64"},
	JavaScriptLanguage: {"js", "nodejs"},
	DotnetLanguage:     {"arm", "arm64"},
	NimLanguage:        {"nim", "nimvm"},
}

type RemoteNodeClient struct {
	RawClientName      string
	ClientName         ClientName
	ClientVersion      string
	ClientCleanVersion string
	ClientOS           ClientOS
	ClientArch         ClientArch
	ClientLanguage     ClientLanguage
}

func ParseUserAgent(rawString string) RemoteNodeClient {
	details := RemoteNodeClient{
		RawClientName:      rawString,
		ClientName:         classifyItem(rawString, ClientName(splitRawBy(strings.ToLower(rawString), "/", 0)), ValidClientNames), // don't use the default Unknown for the ClientName
		ClientVersion:      getRawVersion(rawString),
		ClientCleanVersion: getCleanVersion(rawString),
		ClientOS:           classifyItem(rawString, ClientOS(Unknown), ValidOSs),
		ClientArch:         classifyItem(rawString, ClientArch(Unknown), ValidArchs),
		ClientLanguage:     classifyItem(rawString, ClientLanguage(Unknown), ValidLanguages),
	}
	return details
}

func classifyItem[T ClientName | ClientOS | ClientArch | ClientLanguage](rawString string, output T, validList map[T][]string) T {
	lowerString := strings.ToLower(rawString)
	for classItem, validItems := range validList {
		for _, vItem := range validItems {
			if strings.Contains(lowerString, vItem) {
				output = classItem
			}
		}
	}
	if output == T("") {
		output = T(Unknown)
	}
	return output
}

func getRawVersion(rawString string) string {
	version := Unknown
	client := classifyItem(rawString, ClientName(Unknown), ValidClientNames)
	switch client {
	// except for nimbus: "nimbus-eth1 v0.1.0 [linux: amd64, rocksdb, nimvm, 6d1328]"
	case NimbusEth1:
		version = splitRawBy(rawString, " ", 1) // second item
	// generally, nodes follow this format "client_name/clean_version-dirty_version/os_arch/language
	default:
		version = splitRawBy(rawString, "/", 1) // second item
	}
	return version
}

func splitRawBy(raw, spliter string, splittedItem int) string {
	version := Unknown
	nameParts := strings.Split(raw, spliter)
	if splittedItem >= len(nameParts) {
		return version
	}
	return nameParts[splittedItem]
}

func getCleanVersion(rawString string) string {
	version := Unknown
	rawVersion := getRawVersion(rawString)
	version = splitRawBy(rawVersion, "-", 0)
	return version
}
