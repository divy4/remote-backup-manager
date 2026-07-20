package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// The global config value. All code reads from here!
var Config = loadConfig()

// The path of the config file.
const configPath string = "/etc/upload_backup.json"

// The raw format of the config file
type rawConfig struct {
	ChecksumMethod 			string 		`json:"checksum_method"`
	MaxTarSize 					string 		`json:"max_tar_size"`
	DesiredTarSize 			string 		`json:"desired_tar_size"`
	LocalBackupPath 		string 		`json:"local_backup_path"`
	LocalChecksumsPath 	string 		`json:"local_checksums_path"`
	GpgEncryptionKey 		string 		`json:"gpg_encryption_key"`
	S3Bucket 						string 		`json:"s3_bucket"`
	S3Subpath 					string 		`json:"s3_subpath"`
	S3LockDays 					int64 		`json:"s3_lock_days"`
	S3UploadSpeed 			string 		`json:"s3_upload_speed"`
	AwsAccessKeyId 			string 		`json:"aws_access_key_id"`
	AwsSecretAccessKey 	string 		`json:"aws_secret_access_key"`
	StaticSections 			[]string 	`json:"static_sections"`
}

// The formatted config the app code reads
type ConfigType struct {
	ChecksumMethod 			ChecksumMethod
	MaxTarSize 					int64
	DesiredTarSize 			int64
	LocalBackupPath 		string
	LocalChecksumsPath 	string
	GpgEncryptionKey 		string
	S3Bucket 						string
	S3Subpath 					string
	S3LockDays 					int64
	S3UploadSpeed 			int64
	AwsAccessKeyId 			string
	AwsSecretAccessKey 	string
	StaticSections 			[]string
}

// Loads the config for the app
func loadConfig() *ConfigType {
	log.Printf("Loading config from %s...", configPath)

	// Load the config as text
	configText, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal(err)
	}
	
	// Then convert it into a raw config object
	var rawConfig rawConfig
	err = json.Unmarshal(configText, &rawConfig)
	if err != nil {
		log.Fatal(err)
	}

	config := ConfigType{
		ChecksumMethod: loadChecksumMethod("checksum_method", rawConfig.ChecksumMethod),
		MaxTarSize: loadBytes("max_tar_size", rawConfig.MaxTarSize, byte, false),
		DesiredTarSize: loadBytes("desired_tar_size", rawConfig.DesiredTarSize, byte, false),
		LocalBackupPath: loadPath("local_backup_path", rawConfig.LocalBackupPath, dir, exists),
		LocalChecksumsPath: loadPath("local_checksums_path", rawConfig.LocalChecksumsPath, file, existsOrNotExists),
		GpgEncryptionKey: loadString("gpg_encryption_key", rawConfig.GpgEncryptionKey, "^[A-Za-z0-9]{40}$"),
		S3Bucket: loadString("s3_bucket", rawConfig.S3Bucket, "^[a-z0-9-]+$"),
		S3Subpath: loadString("s3_subpath", rawConfig.S3Subpath, "."),
		S3LockDays: loadInt("s3_lock_days", rawConfig.S3LockDays, 1, 730),
		S3UploadSpeed: loadBytes("s3_upload_speed", rawConfig.S3UploadSpeed, byte, true),
		AwsAccessKeyId: loadString("aws_access_key_id", rawConfig.AwsAccessKeyId, "."),
		AwsSecretAccessKey: loadString("aws_secret_access_key", rawConfig.AwsSecretAccessKey, "."),
		StaticSections: make([]string, len(rawConfig.StaticSections)),
	}
	
	for i, section := range rawConfig.StaticSections {
		key := fmt.Sprintf("static_sections.%d", i)
		config.StaticSections = append(config.StaticSections, loadPath(key, section, dirOrFile, existsOrNotExists))
	}

	log.Println("Config loaded.")
	return &config
}

// Basic values

func loadString(key string, s string, pattern string) string {
	regex := regexp.MustCompile(pattern)
	if !regex.MatchString(s) {
		configError(key, "'%s' does not match expression '%s'.", s, pattern)
	}
	return s
}

func loadInt(key string, i int64, minVal int64, maxVal int64) int64 {
	if i < minVal || i > maxVal {
		configError(key, "Value '%d' is not within [%d, %d].", i, minVal, maxVal)
	}
	return i
}

// Checksum methods

type ChecksumMethod int64
const (
	ChecksumByMetadata ChecksumMethod = iota
	ChecksumByContent
)

// Converts a string to a checksum method
func loadChecksumMethod(key string, s string) ChecksumMethod {
	switch s {
	case "content":
		return ChecksumByContent
	case "metadata":
		return ChecksumByMetadata
	}
	configError(key, "Unknown checksum method '%s'.", s)
	return ChecksumByContent
}

// Bytes

var bytePattern = regexp.MustCompile("^([1-9][0-9]*)(K|Ki|M|Mi|G|Gi|T|Ti)(b|B)((ps|/s)?)$")

type bitOrByteType int64
const (
	bit  bitOrByteType = iota
	byte
)

var siSuffixes = map[string]int64{
	"": 1,
	"K": 1000,
	"M": 1000000,
	"G": 1000000000,
	"T": 1000000000000,
	"Ki": 1024,
	"Mi": 1024 * 1024,
	"Gi": 1024 * 1024 * 1024,
	"Ti": 1024 * 1024 * 1024 * 1024,
}

// Converts a string to an integer number of bits/bytes.
//
// bitType defines if the result should be in bits or bytes. If s describes
// bits and this value is set to byte, the number will be divided by 8 and
// vice versa.
//
// isRate defines if s should be a constant value, e.g. 1MB, or a rate value,
// e.g. 1MB/s.
func loadBytes(key string, s string, bitOrByte bitOrByteType, isRate bool) int64 {
	// Match string
	match := bytePattern.FindStringSubmatch(s)
	if match == nil {
		configError(key, "'%s' is not a valid byte string.", s)
	}

	// Split it up
	num, err := strconv.ParseInt(match[1], 10, 64)
	multiplier := match[2]
	matchedType := match[3]
	rate := match[4]
	// Shouldn't happen because the regex above should prevent invalid values.
	// But just in case...
	if err != nil {
		log.Fatal(err)
	}

	// Fail if the rate doesn't line up
	if isRate && rate == "" {
		configError(key, "Expected string %s to be a rate, e.g. end with '/s'.", s)
	} else if ! isRate && rate != "" {
		configError(key, "Expected string %s to NOT be a rate, e.g. not end with '/s'.", s)
	}

	// Figure out the number
	i := num * siSuffixes[multiplier]
	
	// Convert bytes -> bits
	if matchedType == "B" && bitOrByte == bit {
		i *= 8
	// and bits -> bytes
	} else if matchedType == "b" && bitOrByte == byte {
		i /= 8
	}

	return i
}

// Paths

type pathType int64
const (
	dir pathType = iota
	file
	dirOrFile
)

type existsType int64
const (
	exists existsType = iota
	notExists
	existsOrNotExists
)

// Verifies a string path and returns the same string.
func loadPath(key string, s string, t pathType, e existsType) string {
	// Convert to absolute path
	absPath, err := filepath.Abs(s)
	if err != nil {
		configError(key, "Error while converting '%s' to an absolute path: %s", s, err)
	}

	// Ensure the file exists
	info, err := os.Lstat(absPath)
	if err != nil {
		// If we don't expect the path to exist, that's fine, return early
		if e != exists && strings.Contains(err.Error(), "no such file or directory") {
			return s
		}
		// Fail on any other error
		log.Fatal(err)
	// But if the path did exist but we didn't expect it to, then that's a problem
	} else if e == notExists {
		configError(key, "Path '%s' already exists.", s)
	}

	// Throw an error if it isn't the correct type
	if t == dir && !info.IsDir() {
		configError(key, "'%s' is not a directory.", s)
	} else if t == file && info.IsDir() {
		configError(key, "'%s' is not a file.", s)
	}

	return absPath
}

// Helpers

// Executes log.Fatalf with a predictable format for configuration errors.
func configError(key string, message string, args ...any) {
	message = fmt.Sprintf("Config error in %s - %s: %s", configPath, key, message)
	log.Fatalf(message, args...)
}
