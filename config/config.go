package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/imdario/mergo"
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

const BasicAuthentication = "basic"
const OIDCAuthentication = "oidc"

// PasswordCredentials holds password-credentials
type PasswordCredentials struct {
	Username string
	Password string
}

// UsersByRoles maps roles to lists of user holding the role
type UsersByRoles map[string][]string

type basicAuthnData []PasswordCredentials

// Options holds the available command-line options
type Options struct {
	ServerHostname              string         `json:"serverHostname" env:"SERVER_HOSTNAME" long:"server-hostname" short:"h" default:"localhost" description:"Server hostname"`
	ServerPort                  int            `json:"serverPort" env:"SERVER_PORT" long:"server-port" short:"p" default:"8080" description:"Server port"`
	ServerURLContext            string         `json:"serverUrlContext" env:"SERVER_URL_CONTEXT" long:"server-url-context" short:"c" default:"" description:"Server url context"`
	AppDescription              string         `json:"appDescription" env:"APP_DESCRIPTION" long:"app-description" short:"" default:"" description:"Application description"`
	PathToStaticFiles           string         `json:"pathToStaticFiles" env:"PATH_TO_STATIC_FILES" long:"path-to-static-files" short:"f" default:"" description:"Path to static files"`
	IconDataLocationGit         string         `json:"iconDataLocationGit" env:"ICON_DATA_LOCATION_GIT" long:"icon-data-location-git" short:"g" default:"" description:"Icon data location git"`
	IconDataCreateNew           string         `json:"iconDataCreateNew" env:"ICON_DATA_CREATE_NEW" long:"icon-data-create-new" short:"n" default:"never" description:"Icon data create new"`
	AuthenticationType          string         `json:"authenticationType" env:"AUTHENTICATION_TYPE" long:"authentication-type" short:"a" default:"oidc" description:"Authentication type"`
	PasswordCredentials         basicAuthnData `json:"passwordCredentials" env:"PASSWORD_CREDENTIALS" long:"password-credentials"`
	OIDCClientID                string         `json:"oidcClientId" env:"OIDC_CLIENT_ID" long:"oidc-client-id" short:"" default:"" description:"OIDC client id"`
	OIDCClientSecret            string         `json:"oidcClientSecret" env:"OIDC_CLIENT_SECRET" long:"oidc-client-secret" short:"" default:"" description:"OIDC client secret"`
	OIDCAccessTokenURL          string         `json:"oidcAccessTokenUrl" env:"OIDC_ACCESS_TOKEN_URL" long:"oidc-access-token-url" short:"" default:"" description:"OIDC access token url"`
	OIDCUserAuthorizationURL    string         `json:"oidcUserAuthorizationUrl" env:"OIDC_USER_AUTHORIZATION_URL" long:"oidc-user-authorization-url" short:"" default:"" description:"OIDC user authorization url"`
	OIDCClientRedirectBackURL   string         `json:"oidcClientRedirectBackUrl" env:"OIDC_CLIENT_REDIRECT_BACK_URL" long:"oidc-client-redirect-back-url" short:"" default:"" description:"OIDC client redirect back url"`
	OIDCTokenIssuer             string         `json:"oidcTokenIssuer" env:"OIDC_TOKEN_ISSUER" long:"oidc-token-issuer" short:"" default:"" description:"OIDC token issuer"`
	OIDCIpJwtPublicKeyURL       string         `json:"oidcIpJwtPublicKeyUrl" env:"OIDC_IP_JWT_PUBLIC_KEY_URL" long:"oidc-ip-jwt-public-key-url" short:"" default:"" description:"OIDC ip jwt public key url"`
	OIDCIpJwtPublicKeyPemBase64 string         `json:"oidcIpJwtPublicKeyPemBase64" env:"OIDC_IP_JWT_PUBLIC_KEY_PEM_BASE64" long:"oidc-ip-jwt-public-key-pem-base64" short:"" default:"" description:"OIDC ip jwt public key pem base64"`
	OIDCIpLogoutURL             string         `json:"oidcIpLogoutUrl" env:"OIDC_IP_LOGOUT_URL" long:"oidc-ip-logout-url" short:"" default:"" description:"OIDC ip logout url"`
	UsersByRoles                UsersByRoles   `json:"usersByRoles" env:"USERS_BY_ROLES" long:"users-by-roles" short:"" default:"" description:"Users by roles"`
	DBHost                      string         `json:"dbHost" env:"DB_HOST" long:"db-host" short:"" default:"localhost" description:"DB host"`
	DBPort                      int            `json:"dbPort" env:"DB_PORT" long:"db-port" short:"" default:"5432" description:"DB port"`
	DBUser                      string         `json:"dbUser" env:"DB_USER" long:"db-user" short:"" default:"iconrepo" description:"DB user"`
	DBPassword                  string         `json:"dbPassword" env:"DB_PASSWORD" long:"db-password" short:"" default:"iconrepo" description:"DB password"`
	DBName                      string         `json:"dbName" env:"DB_NAME" long:"db-name" short:"" default:"iconrepo" description:"Name of the database"`
	DBSchemaName                string         `json:"dbSchemaName" env:"DB_SCHEMA_NAME" long:"db-schema-name" short:"" default:"icon_repo" description:"Name of the database schemma"`
	EnableBackdoors             bool           `json:"enableBackdoors" env:"ENABLE_BACKDOORS" long:"enable-backdoors" short:"" description:"Enable backdoors"`
	PackageRootDir              string         `json:"packageRootDir" env:"PACKAGE_ROOT_DIR" long:"package-root-dir" short:"" default:"" description:"Package root dir"`
	LogLevel                    string         `json:"logLevel" env:"IGOREPO_LOG_LEVEL" long:"log-level" short:"l" default:"info"`
}

var DefaultIconRepoHome = filepath.Join(os.Getenv("HOME"), ".ui-toolbox/icon-repo")
var DefaultIconDataLocationGit = filepath.Join(DefaultIconRepoHome, "git-repo")
var DefaultConfigFilePath = filepath.Join(DefaultIconRepoHome, "config.json")

// GetConfigFilePath gets the path of the configuration file
func GetConfigFilePath() string {
	var result string
	if result = os.Getenv("ICON_REPO_CONFIG_FILE"); result != "" {
	} else {
		result = DefaultConfigFilePath
	}
	log.Infof("Configuration file: %s", result)
	return result
}

// ReadConfiguration reads the configuration file and merges it with the command line arguments
func ReadConfiguration(filePath string, clArgs []string) (Options, error) {
	optsInFile, err := ReadConfigurationFromFile(filePath)
	if err != nil {
		return Options{}, err
	}
	return parseFlagsMergeSettings(clArgs, optsInFile), nil
}

// ReadConfigurationFromFile reads configuration from a file (JSON for now)
func ReadConfigurationFromFile(filePath string) (Options, error) {
	var optsInFile = Options{}

	_, fileStatError := os.Stat(filePath)
	if fileStatError != nil {
		return optsInFile, fmt.Errorf("failed to locate configuration file %v: %w", filePath, fileStatError)
	}

	fileContent, fileReadError := os.ReadFile(filePath)
	if fileReadError != nil {
		return optsInFile, fmt.Errorf("failed to read configuration file %v: %w", filePath, fileReadError)
	}

	unmarshalError := json.Unmarshal(fileContent, &optsInFile)
	if unmarshalError != nil {
		return optsInFile, fmt.Errorf("failed to parse configuration file %v: %w", filePath, unmarshalError)
	}

	return optsInFile, nil
}

func GetDefaultConfiguration() Options {
	options := parseCommandLineArgs([]string{})
	return options
}

func ParseCommandLineArgs(clArgs []string) Options {
	options := parseCommandLineArgs(clArgs)
	return options
}

func parseCommandLineArgs(clArgs []string) Options {
	logger := log.WithField("prefix", "parseCommandLineArgs")

	var opts = Options{}
	parser := flags.NewParser(&opts, flags.Default)
	rest, parseError := parser.ParseArgs(clArgs)
	logger.Info("Command line arguments not parsed: ", rest)
	if parseError != nil {
		logger.Fatal(parseError)
	}

	opts.IconDataLocationGit = DefaultIconDataLocationGit

	return opts
}

func parseFlagsMergeSettings(clArgs []string, optsInFile Options) Options {
	opts := parseCommandLineArgs(clArgs)

	if err := mergo.Merge(&optsInFile, opts, mergo.WithOverride); err != nil {
		panic(err)
	}

	log.Infof("Options: %v", optsInFile)

	return optsInFile
}
