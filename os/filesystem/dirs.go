// Package filesystem provides functionality for interacting with directories and files in a cross-platform manner.
package filesystem

import (
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/libonomy/wallet-cli/os/app/config"
	"github.com/libonomy/wallet-cli/os/log"
)

// Using a function pointer to get the current user so we can more easily mock in tests
var currentUser = user.Current

// Directory and paths funcs

// OwnerReadWriteExec is a standard owner read / write / exec file permission.
const OwnerReadWriteExec = 0700

// OwnerReadWrite is a standard owner read / write file permission.
const OwnerReadWrite = 0600

// PathExists returns true iff file exists in local store and is accessible.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

// GetLibonomyDataDirectoryPath gets the full os-specific path to the libonomy top-level data directory.
func GetLibonomyDataDirectoryPath() (string, error) {
	return GetFullDirectoryPath(config.ConfigValues.DataFilePath)
}

// GetLibonomyTempDirectoryPath gets the libonomy temp files dir so we don't have to work with convoluted os specific temp folders.
func GetLibonomyTempDirectoryPath() (string, error) {
	return ensureDataSubDirectory("temp")
}

// DeleteAllTempFiles deletes all temp files from the temp dir and creates a new temp dir.
func DeleteAllTempFiles() error {
	tempDir, err := GetLibonomyTempDirectoryPath()
	if err != nil {
		return err
	}

	err = os.RemoveAll(tempDir)
	if err != nil {
		return err
	}

	// create temp dir again
	_, err = GetLibonomyTempDirectoryPath()
	return err
}

// EnsureLibonomyDataDirectories return the os-specific path to the libonomy data directory.
// It creates the directory and all predefined sub directories on demand.
func EnsureLibonomyDataDirectories() (string, error) {
	dataPath, err := GetLibonomyDataDirectoryPath()
	if err != nil {
		log.Error("Can't get or create libonomy data folder")
		return "", err
	}

	log.Debug("Data directory: %s", dataPath)

	// ensure sub folders exist - create them on demand
	_, err = GetAccountsDataDirectoryPath()
	if err != nil {
		return "", err
	}

	_, err = GetLogsDataDirectoryPath()
	if err != nil {
		return "", err
	}

	return dataPath, nil
}

// ensureDataSubDirectory ensure a sub-directory exists.
func ensureDataSubDirectory(dirName string) (string, error) {
	dataPath, err := GetLibonomyDataDirectoryPath()
	if err != nil {
		log.Error("Failed to ensure data dir", err)
		return "", err
	}

	pathName := filepath.Join(dataPath, dirName)
	aPath, err := GetFullDirectoryPath(pathName)
	if err != nil {
		log.Error("Can't access libonomy folder", pathName, "Erorr:", err)
		return "", err
	}
	return aPath, nil
}

// GetAccountsDataDirectoryPath returns the path to the accounts data directory.
// It will create the directory if it doesn't already exist.
func GetAccountsDataDirectoryPath() (string, error) {
	return ensureDataSubDirectory(config.AccountsDirectoryName)
}

// GetLogsDataDirectoryPath returns the path to the app logs data directory.
// It will create the directory if it doesn't already exist.
func GetLogsDataDirectoryPath() (string, error) {
	return ensureDataSubDirectory(config.LogDirectoryName)
}

// GetUserHomeDirectory returns the current user's home directory if one is set by the system.
func GetUserHomeDirectory() string {

	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := currentUser(); err == nil {
		return usr.HomeDir
	}
	return ""
}

// GetCanonicalPath returns an os-specific full path following these rules:
// - replace ~ with user's home dir path
// - expand any ${vars} or $vars
// - resolve relative paths /.../
// p: source path name
func GetCanonicalPath(p string) string {

	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		if home := GetUserHomeDirectory(); home != "" {
			p = home + p[1:]
		}
	}
	return path.Clean(os.ExpandEnv(p))
}

// GetFullDirectoryPath gets the OS specific full path for a named directory.
// The directory is created if it doesn't exist.
func GetFullDirectoryPath(name string) (string, error) {

	aPath := GetCanonicalPath(name)

	// create dir if it doesn't exist
	err := os.MkdirAll(aPath, OwnerReadWriteExec)

	return aPath, err
}

// EnsureNodesDataDirectory Gets the os-specific full path to the nodes master data directory.
// Attempts to create the directory on-demand.
func EnsureNodesDataDirectory(nodesDirectoryName string) (string, error) {
	dataPath, err := GetLibonomyDataDirectoryPath()
	if err != nil {
		return "", err
	}

	nodesDir := filepath.Join(dataPath, nodesDirectoryName)
	return GetFullDirectoryPath(nodesDir)
}

// EnsureNodeDataDirectory Gets the path to the node's data directory, e.g. /nodes/[node-id]/
// Directory will be created on demand if it doesn't exist.
func EnsureNodeDataDirectory(nodesDataDir string, nodeID string) (string, error) {
	return GetFullDirectoryPath(filepath.Join(nodesDataDir, nodeID))
}

// NodeDataFile Returns the os-specific full path to the node's data file.
func NodeDataFile(nodesDataDir, NodeDataFileName, nodeID string) string {
	return filepath.Join(nodesDataDir, nodeID, NodeDataFileName)
}
