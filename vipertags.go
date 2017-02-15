package vipertags

import (
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"os"

	"github.com/Sirupsen/logrus"
	"github.com/fatih/structs"
	"github.com/k0kubun/pp"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

func setField(field *structs.Field, val interface{}) {
	switch field.Value().(type) {
	case bool:
		field.Set(cast.ToBool(val))
	case string:
		field.Set(cast.ToString(val))
	case int64, int32, int16, int8, int:
		field.Set(cast.ToInt(val))
	case float64, float32:
		field.Set(cast.ToFloat64(val))
	case time.Time:
		field.Set(cast.ToTime(val))
	case time.Duration:
		field.Set(cast.ToDuration(val))
	case []string:
		field.Set(cast.ToStringSlice(val))
	default:
		field.Set(val)
	}
}

func buildConfiguration(st0 interface{}, prefix0 string) interface{} {
	st := structs.New(st0)
	for _, field := range st.Fields() {
		configTagValue := field.Tag("config")

		if configTagValue == "-" {
			continue
		}

		defaultTagValue := field.Tag("default")
		envTagValue := field.Tag("env")

		prefix := prefix0
		if configTagValue != "" {
			prefix = prefix + configTagValue
		}

		if field.Kind() == reflect.Struct {
			buildConfiguration(field.Value(), prefix)
			continue
		}
		if field.Kind() == reflect.Map {
			pp.Println("map == ", field)
			continue
		}
		if field.Kind() == reflect.Array || field.Kind() == reflect.Slice {
			logrus.Fatal("Vipertags parsing of array and slice not currently working...")
			t := reflect.Indirect(reflect.ValueOf(field.Value()))

			slice := reflect.MakeSlice(t.Type(), t.Len(), t.Len())
			slices := reflect.New(slice.Type())
			slices.Elem().Set(slice)

			for ii := 0; ii < t.Len(); ii++ {
				e := t.Index(ii).Interface()
				elem := buildConfiguration(e, prefix)
				slices.Elem().Index(ii).Set(reflect.ValueOf(elem))
			}
			pp.Println(slices.Interface())
			field.Set(slices.Interface())
			continue
		}
		if defaultTagValue != "" && configTagValue != "" {
			viper.SetDefault(configTagValue, defaultTagValue)
		}
		if defaultTagValue != "" && configTagValue == "" {
			setField(field, defaultTagValue)
		}

		if envTagValue != "" && configTagValue != "" {
			viper.BindEnv(configTagValue, envTagValue)
		}
		if envTagValue != "" && configTagValue == "" {
			if e := os.Getenv(envTagValue); e != "" {
				setField(field, e)
			}
		}
		if configTagValue != "" {
			setField(field, viper.Get(configTagValue))
		}
	}
	return st0
}

func Fill(class interface{}) {
	err := viper.ReadInConfig()
	if err != nil {
		logrus.WithError(err).
			Fatal("Cannot find configuration file.")
	}
	buildConfiguration(class, "")
}

func Setup(fileType string, prefix string) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("conf")
	viper.AddConfigPath("config")
	viper.SetConfigType(fileType)
	viper.AutomaticEnv()
	viper.SetEnvPrefix(prefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func FromFile(filename string, prefix string) {
	Setup(strings.Replace(filepath.Ext(filename), ".", "", 1), prefix)
	viper.SetConfigFile(filename)
	err := viper.ReadInConfig()
	if err != nil {
		logrus.WithError(err).
			Fatal("Cannot find configuration file.")
	}
}
