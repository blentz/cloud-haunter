package types

import (
	log "github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

type IFilterConfig interface {
	GetFilterValues(fType FilterEntityType, cloud CloudType, property FilterConfigProperty) []string
}

type FilterEntityType string

const (
	ExcludeAccess   = FilterEntityType("excludeAccess")
	IncludeAccess   = FilterEntityType("includeAccess")
	ExcludeInstance = FilterEntityType("excludeInstance")
	IncludeInstance = FilterEntityType("includeInstance")
)

type FilterConfigProperty string

const (
	Name  = FilterConfigProperty("name")
	Owner = FilterConfigProperty("owner")
	Label = FilterConfigProperty("label")
)

// FilterConfig structure that stores the information provided by the exclude/include flag
type FilterConfig struct {
	ExcludeAccess   *FilterAccessConfig   `yaml:"excludeAccess"`
	IncludeAccess   *FilterAccessConfig   `yaml:"includeAccess"`
	ExcludeInstance *FilterInstanceConfig `yaml:"excludeInstance"`
	IncludeInstance *FilterInstanceConfig `yaml:"includeInstance"`
	ExcludeImage    *FilterImageConfig    `yaml:"excludeImage"`
	IncludeImage    *FilterImageConfig    `yaml:"includeImage"`
	ExcludeDatabase *FilterDatabaseConfig `yaml:"excludeDatabase"`
	IncludeDatabase *FilterDatabaseConfig `yaml:"includeDatabase"`
	ExcludeDisk     *FilterDiskConfig     `yaml:"excludeDisk"`
	IncludeDisk     *FilterDiskConfig     `yaml:"includeDisk"`
	ExcludeStack    *FilterStackConfig    `yaml:"excludeStack"`
	IncludeStack    *FilterStackConfig    `yaml:"includeStack"`
}

// FilterAccessConfig filter properties for access items
type FilterAccessConfig struct {
	Aws struct {
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"aws"`
	Azure struct {
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"azure"`
	Gcp struct {
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"gcp"`
}

// FilterInstanceConfig filter properties for instances
type FilterInstanceConfig struct {
	Aws struct {
		Labels []string `yaml:"labels"`
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"aws"`
	Azure struct {
		Labels []string `yaml:"labels"`
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"azure"`
	Gcp struct {
		Labels []string `yaml:"labels"`
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"gcp"`
}

func (filterConfig FilterConfig) GetFilterValues(fType FilterEntityType, cloud CloudType, property FilterConfigProperty) []string {
	log.Debugf("fType: %s, cloud: %s, property :%s", fType, cloud, property)
	typeProperty := strings.ToUpper(string(string(fType)[0])) + string(fType)[1:]
	cloudProperty := string(string(cloud)[0]) + strings.ToLower(string(cloud)[1:])
	propertyProperty := strings.ToUpper(string(string(property)[0])) + string(property)[1:] + "s"
	log.Debugf("FilterEntityType: %s, CloudProperty: %s, FilterConfigProperty: %s", typeProperty, cloudProperty, propertyProperty)

	if typeField := reflect.ValueOf(filterConfig).FieldByName(typeProperty); typeField.IsValid() && !typeField.IsNil() {
		if cloudField := reflect.Indirect(typeField).FieldByName(cloudProperty); cloudField.IsValid() {
			if propertyField := reflect.Indirect(cloudField).FieldByName(propertyProperty); propertyField.IsValid() {
				return propertyField.Interface().([]string)
			}
		}
	}
	return nil
}

// FilterImageConfig filter properties for image items
type FilterImageConfig struct {
	Aws struct {
		Labels []string `yaml:"labels"`
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"aws"`
	Azure struct {
		Labels []string `yaml:"labels"`
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"azure"`
	Gcp struct {
		Labels []string `yaml:"labels"`
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"gcp"`
}

// FilterDatabaseConfig filter properties for image items
type FilterDatabaseConfig struct {
	Aws struct {
		Labels []string `yaml:"labels"`
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"aws"`
	Azure struct {
		Labels []string `yaml:"labels"`
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"azure"`
	Gcp struct {
		Labels []string `yaml:"labels"`
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"gcp"`
}

// FilterDiskConfig filter properties for image items
type FilterDiskConfig struct {
	Aws struct {
		Labels []string `yaml:"labels"`
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"aws"`
	Azure struct {
		Labels []string `yaml:"labels"`
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"azure"`
	Gcp struct {
		Labels []string `yaml:"labels"`
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"gcp"`
}

// FilterStackConfig filter properties for image items
type FilterStackConfig struct {
	Aws struct {
		Labels []string `yaml:"labels"`
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"aws"`
	Azure struct {
		Labels []string `yaml:"labels"`
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"azure"`
	Gcp struct {
		Labels []string `yaml:"labels"`
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"gcp"`
}

// FilterClusterConfig filter properties for instances
type FilterClusterConfig struct {
	Aws struct {
		Labels []string `yaml:"labels"`
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"aws"`
	Azure struct {
		Labels []string `yaml:"labels"`
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"azure"`
	Gcp struct {
		Labels []string `yaml:"labels"`
		Names  []string `yaml:"names"`
		Owners []string `yaml:"owners"`
	} `yaml:"gcp"`
}
