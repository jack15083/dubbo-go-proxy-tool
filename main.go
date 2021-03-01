package main

import (
	"errors"
	"flag"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

// ServiceConfig is the configuration of the service provider
type ServiceConfig struct {
	InterfaceName string `required:"true"  yaml:"interface"  json:"interface,omitempty" property:"interface"`
}

type ProviderConfig struct {
	Services map[string]ServiceConfig `yaml:"services" json:"services,omitempty" property:"services"`
}

type methodConfig struct {
	ServiceName  string              `json:"service_name"`
	InterfaceUrl string              `json:"interface_url"`
	Params       []map[string]string `json:"params"`
	Group        string              `json:"group"`
	UsedAppName  string              `json:"used_app_name"`
	Desc         string              `json:"desc"`
}

var (
	providerConf ProviderConfig
	methods      []methodConfig
	group        string
)

func main() {
	fmt.Println("start success")
	exPath, _ := os.Getwd()
	//exPath := filepath.Dir(dir)
	var (
		appPath string
	)
	flag.StringVar(&appPath, "p", exPath, "App path")
	flag.StringVar(&group, "g", "", "App group env")
	flag.Parse()
	if appPath == "" {
		panic("Config path is empty")
	}

	confPath := appPath + "/profiles/pro/provider.yml"
	_, err := UnmarshalYMLConfig(confPath, &providerConf)
	if err != nil {
		log.Fatalf("setting.Setup, fail to parse '%s': %v", confPath, err)
	}

	ReadFile(appPath + "/app/service")
	jsonByte, _ := jsoniter.MarshalIndent(methods, "", "  ")
	err = os.MkdirAll(appPath+"/app/docs", 755)
	if err != nil {
		panic(err)
	}

	filePath := appPath + "/app/docs/config.go"

	_ = os.Remove(filePath)
	docs := fmt.Sprintf("package docs\n\nvar methodDocs = `%s`\n", string(jsonByte))
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	_, err = file.WriteString(docs)
	if err != nil {
		panic(err)
	}
	file.Close()
	fmt.Println("complete!...")
}

func ReadFile(path string) {
	fs, _ := ioutil.ReadDir(path)
	for _, f := range fs {
		filePath := path + "/" + f.Name()
		if f.IsDir() {
			ReadFile(filePath)
			continue
		}
		processFile(filePath)
	}
}

func processFile(filePath string) {
	var (
		providerName  string
		interfaceName string
		appName       string
	)

	b, _ := ioutil.ReadFile(filePath)
	reg, _ := regexp.Compile(`func \(.*?\*(.*?Provider)\)`)
	content := string(b)
	matches := reg.FindStringSubmatch(content)
	if len(matches) < 2 {
		return
	}
	providerName = strings.TrimSpace(matches[1])

	if c, ok := providerConf.Services[providerName]; ok {
		interfaceName = c.InterfaceName
	}

	if interfaceName == "" {
		return
	}

	reg, _ = regexp.Compile(`la.kaike\.(.*?)\.`)
	matches = reg.FindStringSubmatch(interfaceName)
	if len(matches) < 2 {
		return
	}
	appName = matches[1]

	//func (h *HelloProvider) SayHello(ctx context.Context, name string) (string, error) {
	reg, _ = regexp.Compile(`(?is)\s+(//(\s+|\s?)@desc\s+(.*?)\s+//(\s+|\s?)@used\s+(.*?)\s+)func \(.*?` + providerName + `\) (.*?)\((.*?)\) \(.*?error\)\s?\{`)
	matchesAll := reg.FindAllStringSubmatch(content, 100)
	for _, row := range matchesAll {
		if len(row) < 8 {
			continue
		}

		desc := strings.TrimSpace(row[3])
		usedAppName := strings.TrimSpace(row[5])
		methodName := strings.TrimSpace(row[6])
		paramsString := row[7]
		params := strings.Split(paramsString, ",")
		if usedAppName == "" {
			continue
		}

		var parTypeList []map[string]string
		for _, row := range params {
			par := strings.Split(strings.TrimSpace(row), " ")
			if len(par) != 2 {
				break
			}

			parKey := strings.TrimSpace(par[0])
			parType := strings.TrimSpace(par[1])
			regSlice, _ := regexp.Compile(`^\[\]`)
			if regSlice.MatchString(parType) {
				parType = "slice"
			}

			regMap, _ := regexp.Compile(`^map`)
			if regMap.MatchString(parType) {
				parType = "map"
			}

			switch parType {
			case "context.Context":
				continue
			case "slice":
				parType = "java.util.List"
			case "map":
				parType = "java.util.Map"
			default:
				regDTO, _ := regexp.Compile(`DTO$`)
				if regDTO.MatchString(parType) {
					regParReplace, _ := regexp.Compile(`\d\.`)
					parType = regParReplace.ReplaceAllString(parType, ".")
					parType = "la.kaike." + appName + "." + parType
				}
			}
			parTypeList = append(parTypeList, map[string]string{parKey: parType})
		}

		c := methodConfig{
			Desc:         desc,
			InterfaceUrl: interfaceName,
			Params:       parTypeList,
			ServiceName:  appName + "." + providerName + "." + methodName,
			UsedAppName:  usedAppName,
			Group:        group,
		}

		methods = append(methods, c)
	}
}

// unmarshalYMLConfig Load yml config byte from file , then unmarshal to object
func UnmarshalYMLConfig(confProFile string, out interface{}) ([]byte, error) {
	confFileStream, err := LoadYMLConfig(confProFile)
	if err != nil {
		return confFileStream, err
	}
	return confFileStream, yaml.Unmarshal(confFileStream, out)
}

// loadYMLConfig Load yml config byte from file
func LoadYMLConfig(confProFile string) ([]byte, error) {
	if len(confProFile) == 0 {
		return nil, errors.New("No config file ")
	}

	if path.Ext(confProFile) != ".yml" {
		return nil, errors.New("application configure file name{%v} suffix must be .yml, " + confProFile)
	}

	return ioutil.ReadFile(confProFile)
}
