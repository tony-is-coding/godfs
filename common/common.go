package common

import (
	"errors"
	"regexp"
	"time"
)

const (
	VERSION = "2.0.0"
	//
	BOOT_CLIENT  BootMode = 0
	BOOT_STORAGE BootMode = 1
	BOOT_TRACKER BootMode = 2
	//
	GROUP_PATTERN       = "^[0-9a-zA-Z-_]{1,30}$"
	SECRET_PATTERN      = "^[^@]{1,30}$"
	SERVER_PATTERN      = "^(([^@^,]{1,30})@)?([^@]+):([1-9][0-9]{0,5})$"
	HTTP_AUTH_PATTERN   = "^([^:]+):([^:]+)$"
	INSTANCE_ID_PATTERN = "^[0-9a-z-]{8}$"
	FILE_META_PATTERN   = "^([0-9a-zA-Z-_]{1,30})/([0-9A-F]{2})/([0-9A-F]{2})/([0-9a-f]{32})$"
	//
	DEFAULT_STORAGE_TCP_PORT  = 9012
	DEFAULT_STORAGE_HTTP_PORT = 8001
	DEFAULT_TRACKER_TCP_PORT  = 9022
	DEFAULT_TRACKER_HTTP_PORT = 8011
	BUFFER_SIZE               = 1 << 15 // 32k
	DEFAULT_GROUP             = "G01"
	//
	OPERATION_RESPONSE       Operation = 0
	OPERATION_CONNECT        Operation = 1
	OPERATION_UPLOAD         Operation = 2
	OPERATION_DOWNLOAD       Operation = 3
	OPERATION_QUERY          Operation = 4
	OPERATION_SYNC_INSTANCES Operation = 5
	OPERATION_PUSH_BINLOGS   Operation = 6
	//
	SUCCESS           OperationResult = 0
	ERROR             OperationResult = 1
	UNAUTHORIZED      OperationResult = 2
	NOT_FOUND         OperationResult = 3
	UNKNOWN_OPERATION OperationResult = 4
	//
	CMD_SHOW_HELP     Command = 0
	CMD_SHOW_VERSION  Command = 1
	CMD_UPDATE_CONFIG Command = 2
	CMD_SHOW_CONFIG   Command = 3
	CMD_UPLOAD_FILE   Command = 4
	CMD_DOWNLOAD_FILE Command = 5
	CMD_INSPECT_FILE  Command = 6
	CMD_BOOT_TRACKER  Command = 7
	CMD_BOOT_STORAGE  Command = 8
	CMD_TEST_UPLOAD   Command = 9
	//
	ROLE_TRACKER Role = 1
	ROLE_STORAGE Role = 2
	ROLE_PROXY   Role = 3
	ROLE_CLIENT  Role = 4
	//
	REGISTER_HOLD RegisterState = 1
	REGISTER_FREE RegisterState = 2
	//
	REGISTER_INTERVAL    = time.Second * 30
	SYNCHRONIZE_INTERVAL = time.Second * 45

	STORAGE_CONFIG_MAP_KEY = "STORAGE_CONFIG_MAP_KEY"
	TRACKER_CONFIG_MAP_KEY = "TRACKER_CONFIG_MAP_KEY"
	PROXY_CONFIG_MAP_KEY   = "PROXY_CONFIG_MAP_KEY"
)

var (
	NotFoundErr                     = errors.New("file not found")
	ServerErr                       = errors.New("server internal error")
	InitializedTrackerConfiguration *TrackerConfig
	InitializedStorageConfiguration *StorageConfig
	InitializedClientConfiguration  *ClientConfig
	FileMetaPatternRegexp           = regexp.MustCompile(FILE_META_PATTERN)
	ServerPatternRegexp             = regexp.MustCompile(SERVER_PATTERN)
	BootAs                          BootMode
	configMaps                      = make(map[string]*ConfigMap)
)

func SetConfigMap(configName string, config *ConfigMap) {
	configMaps[configName] = config
}

func GetConfigMap(configName string) *ConfigMap {
	return configMaps[configName]
}
