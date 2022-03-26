/*
 * Copyright (c) 2022, Nadun De Silva. All Rights Reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *   http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func NewFromFile(configFile string) (*Conf, error) {
	f, err := os.Open(filepath.Clean(configFile))
	if err != nil {
		return nil, err
	}

	var confType string
	fileExtension := filepath.Ext(configFile)
	if len(fileExtension) > 1 {
		confType = fileExtension[1:]
	} else {
		confType = "yaml"
	}

	return New(f, confType)
}

func New(reader io.Reader, configType string) (*Conf, error) {
	expandedReader, err := replaceEnv(reader)
	if err != nil {
		return nil, err
	}

	viperConf := viper.New()
	viperConf.SetConfigType(configType)
	err = viperConf.ReadConfig(expandedReader)
	if err != nil {
		return nil, err
	}
	viperConf.WatchConfig()

	conf := &Conf{}
	err = viperConf.Unmarshal(conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func replaceEnv(reader io.Reader) (io.Reader, error) {
	rawCfgBuf := &strings.Builder{}
	_, err := io.Copy(rawCfgBuf, reader)
	if err != nil {
		return nil, err
	}

	var missingEnvs []string
	replacedConf := os.Expand(rawCfgBuf.String(), func(s string) string {
		v, ok := os.LookupEnv(s)
		if !ok {
			missingEnvs = append(missingEnvs, s)
			return ""
		}
		return v
	})
	if len(missingEnvs) > 0 {
		return nil, fmt.Errorf("missing env(s): %s", strings.Join(missingEnvs, ","))
	}

	return bytes.NewBuffer([]byte(replacedConf)), nil
}
